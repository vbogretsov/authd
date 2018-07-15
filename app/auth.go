package app

import (
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/robbert229/jwt"
	"github.com/vbogretsov/go-mailcd"
	"github.com/vbogretsov/go-validation"
	"github.com/vbogretsov/go-validation/rule"
	"golang.org/x/crypto/bcrypt"

	"github.com/vbogretsov/authd/model"
)

const (
	stridLen = 32
	hashCost = bcrypt.DefaultCost
	runes    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func uuidID() string {
	return uuid.New().String()
}

func strID() string {
	return randstr(stridLen)
}

func randstr(size int) string {
	str := make([]byte, size)
	for i := 0; i < size; i++ {
		str[i] = runes[rand.Intn(len(runes))]
	}
	return string(str)
}

func checkpw(hash, clear string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(clear)) == nil
}

func hashpw(user *model.User) error {
	pw, err := bcrypt.GenerateFromPassword([]byte(user.Password), hashCost)
	if err != nil {
		return err
	}

	user.Password = string(pw)
	return nil
}

func atomic(db *gorm.DB, action func(*gorm.DB) error) error {
	txn := db.Begin()

	err := txn.Error
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			txn.Rollback()
		}
	}()

	err = action(txn)
	if err != nil {
		return err
	}

	return txn.Commit().Error
}

type userfn func(*model.User) error

type confirmationMeta struct {
	link     string
	template string
}

type confirmator struct {
	ttl    time.Duration
	meta   []confirmationMeta
	timer  Timer
	sender mailcd.Sender
}

func (cn *confirmator) generate(tx *gorm.DB, user *model.User, kind int8) error {
	now := cn.timer()
	con := model.Confirmation{
		ID:      strID(),
		UserID:  user.ID,
		Kind:    kind,
		Created: now,
		Expires: now.Add(cn.ttl),
	}

	if err := tx.Create(&con).Error; err != nil {
		return err
	}

	msg := mailcd.Request{
		TemplateLang: "en",
		TemplateName: cn.meta[kind].template,
		TemplateArgs: map[string]interface{}{
			"link": cn.meta[kind].link,
			"id":   con.ID,
		},
		To: []mailcd.Address{
			{
				Email: user.Email,
			},
		},
	}

	return cn.sender.Send(msg)
}

func (cn *confirmator) confirm(tx *gorm.DB, id string, fun userfn) error {
	con := model.Confirmation{ID: id}

	res := tx.Preload("User").First(&con)
	if res.RecordNotFound() {
		return NotFoundError{Message: ConfirmationNotFound}
	}

	if res.Error != nil {
		return res.Error
	}

	if con.Expires.Sub(cn.timer()) < 0 {
		if err := tx.Delete(&con).Error; err != nil {
			return err
		}

		if err := cn.generate(tx, con.User, con.Kind); err != nil {
			return err
		}

		return ExpiredError{Message: ConfirmationExpired}
	}

	if err := fun(con.User); err != nil {
		return err
	}

	return tx.Save(con.User).Error
}

type issuer struct {
	timer      Timer
	algorithm  jwt.Algorithm
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func (ir *issuer) grant(db *gorm.DB, user *model.User) (*model.Token, error) {
	now := ir.timer()
	expires := now.Add(ir.accessTTL)

	claims := jwt.NewClaim()
	claims.Set("id", user.ID)
	claims.Set("email", user.Email)
	claims.Set("expires", expires.Unix())

	access, err := ir.algorithm.Encode(claims)
	if err != nil {
		return nil, err
	}

	refresh := model.Refresh{
		ID:      strID(),
		UserID:  user.ID,
		Created: now,
		Expires: expires,
	}

	if err := db.Create(&refresh).Error; err != nil {
		return nil, err
	}

	user.LastLogin = now
	if err := db.Save(user).Error; err != nil {
		return nil, err
	}

	token := model.Token{
		Access:  access,
		Refresh: refresh.ID,
		Expires: expires.Unix(),
	}

	return &token, nil
}

func (ir *issuer) refresh(db *gorm.DB, refreshID string) (*model.Token, error) {
	refresh := model.Refresh{ID: refreshID}

	res := db.Preload("User").First(&refresh)
	if res.RecordNotFound() {
		return nil, UnauthorizedError{Message: RefreshExpired}
	}

	if res.Error != nil {
		return nil, res.Error
	}

	if err := db.Delete(&refresh).Error; err != nil {
		return nil, err
	}

	if refresh.Expires.Sub(time.Now()) < 0 {
		return nil, UnauthorizedError{Message: RefreshExpired}
	}

	return ir.grant(db, refresh.User)
}

type accounter struct {
	rule validation.Rule
}

func (ac *accounter) create(tx *gorm.DB, user *model.User) error {
	user.ID = uuidID()

	if err := ac.rule(user); err != nil {
		return ArgumentError{Source: err}
	}

	if err := hashpw(user); err != nil {
		return err
	}

	return tx.Create(user).Error
}

func (ac *accounter) find(tx *gorm.DB, email string) (*model.User, error) {
	return findUser(tx, email)
}

func findUser(tx *gorm.DB, email string) (*model.User, error) {
	user := model.User{Email: email}

	res := tx.Where("email = ?", email).First(&user)
	if res.RecordNotFound() {
		return nil, nil
	}

	if res.Error != nil {
		return nil, res.Error
	}

	return &user, nil
}

func emailUniq(db *gorm.DB, msg string) validation.Rule {
	return func(v interface{}) error {
		email := v.(*string)

		user, err := findUser(db, *email)
		if err != nil {
			return validation.Panic{Err: err.Error()}
		}

		if user != nil {
			return errors.New(EmailExists)
		}

		return nil
	}
}

func makeUserRule(db *gorm.DB, config Config) (validation.Rule, error) {
	return validation.Struct(&model.User{}, "json", []validation.Field{
		{
			Attr: func(v interface{}) interface{} {
				return &v.(*model.User).Email
			},
			Rules: []validation.Rule{
				rule.StrEmail(EmailInvalid),
				emailUniq(db, EmailExists),
			},
		},
		{
			Attr: func(v interface{}) interface{} {
				return &v.(*model.User).Password
			},
			Rules: []validation.Rule{
				rule.StrMinLen(config.PasswordMinLen, PasswordShort),
			},
		},
	})
}
