package auth

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
	"github.com/labstack/gommon/random"
	"github.com/robbert229/jwt"
	"github.com/vbogretsov/go-mailcd"
	"github.com/vbogretsov/go-validation"
	"github.com/vbogretsov/go-validation/rule"

	"github.com/vbogretsov/authd/model"
)

// TODO(vbogretsov): localize strings

const idsize = 32

type userfn func(*model.User) error

func attremail(v interface{}) interface{} {
	return &v.(*Credentials).Email
}

func attrpassword(v interface{}) interface{} {
	return &v.(*Credentials).Password
}

type rules struct {
	user     validation.Rule
	email    validation.Rule
	password validation.Rule
}

func newrules(db *gorm.DB, conf Config) (*rules, error) {
	er := rule.StrEmail("email-invalid")
	pr := rule.StrMinLen(conf.Password.MinLen, "password-short")

	ur, err := validation.Struct(&model.User{}, "json", []validation.Field{
		{
			Attr:  attremail,
			Rules: []validation.Rule{er, emailUniq(db, "email-uniq")},
		},
		{
			Attr:  attrpassword,
			Rules: []validation.Rule{pr},
		},
	})

	if err != nil {
		return nil, err
	}

	return &rules{email: er, password: pr, user: ur}, nil
}

// Timer represets current time provider.
type Timer func() time.Time

// Auth represents authd application controller.
type Auth struct {
	db     *gorm.DB
	cfg    Config
	rules  *rules
	now    Timer
	sender mailcd.Sender
}

// New creates new application controller.
func New(cfg Config, db *gorm.DB, now Timer, sender mailcd.Sender) (*Auth, error) {
	rules, err := newrules(db, cfg)
	if err != nil {
		return nil, err
	}

	instance := Auth{
		db:     db,
		cfg:    cfg,
		rules:  rules,
		now:    now,
		sender: sender,
	}

	return &instance, nil
}

// SignUp creates new user account if a user email is not in use.
// The confirmation email will be sent to user email.
func (auth *Auth) SignUp(creds *Credentials) error {
	return atomic(auth.db, func(tx *gorm.DB) error {
		if err := auth.rules.user(creds); err != nil {
			return err
		}

		user := model.User{
			ID:       random.String(idsize),
			Email:    creds.Email,
			Password: creds.Password,
			Created:  auth.now(),
			Active:   false,
		}

		if err := hashpw(&user, auth.cfg.Password.HashCost); err != nil {
			return err
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		return auth.confirmate(tx, &user, auth.cfg.SignUp)
	})
}

// ConfirmUser makes a user account active.
func (auth *Auth) ConfirmUser(cid string) error {
	return atomic(auth.db, func(tx *gorm.DB) error {
		update := func(user *model.User) error {
			user.Active = true
			return nil
		}

		return auth.confirm(tx, cid, auth.cfg.SignUp, update)
	})
}

// ResetPassword sends email with password reset link to the email provided.
// If email was sent or no user with the email provided was found returns nil.
func (auth *Auth) ResetPassword(email *Email) error {
	if err := auth.rules.email(email); err != nil {
		return err
	}

	return atomic(auth.db, func(tx *gorm.DB) error {
		user, err := findUser(tx, email.Email)
		if err != nil {
			return err
		}

		if user == nil {
			return nil
		}

		return auth.confirmate(tx, user, auth.cfg.ResetPw)
	})
}

// UpdatePassword updates user password.
func (auth *Auth) UpdatePassword(cid string, password *Password) error {
	if err := auth.rules.password(password); err != nil {
		return err
	}

	return atomic(auth.db, func(tx *gorm.DB) error {
		update := func(user *model.User) error {
			return hashpw(user, auth.cfg.Password.HashCost)
		}
		return auth.confirm(tx, cid, auth.cfg.ResetPw, update)
	})
}

// SignIn generates access token and refresh token if a user with the email
// provided exists, is't account is active and password matches. If the
// access token is expired, it can be refreshed using the refresh token.
func (auth *Auth) SignIn(cred *Credentials) (*Token, error) {
	user, err := findUser(auth.db, cred.Email)
	if err != nil {
		return nil, err
	}

	if user == nil || !user.Active || !checkpw(user.Password, cred.Password) {
		return nil, UnauthorizedError{Message: "login-invalid"}
	}

	var token *Token

	err = atomic(auth.db, func(tx *gorm.DB) error {
		token, err = auth.grant(tx, user)
		return err
	})

	return token, err
}

// Refresh refreshes the access token which correspons to the refresh token
// provided.
func (auth *Auth) Refresh(rid string) (*Token, error) {
	refresh := model.Refresh{ID: rid}

	res := auth.db.Preload("User").First(&refresh)
	if res.RecordNotFound() {
		return nil, UnauthorizedError{Message: "refresh-notfound"}
	}

	if res.Error != nil {
		return nil, res.Error
	}

	var token *Token
	var err error

	err = atomic(auth.db, func(tx *gorm.DB) error {
		if err := auth.db.Delete(&refresh).Error; err != nil {
			return err
		}

		if refresh.Expires.Sub(time.Now()) < 0 {
			return UnauthorizedError{Message: "refresh-expired"}
		}

		token, err = auth.grant(tx, refresh.User)
		return err
	})

	return token, err
}

func (auth *Auth) grant(tx *gorm.DB, user *model.User) (*Token, error) {
	now := auth.now()
	expires := now.Add(auth.cfg.Token.AccessTTL)

	claims := jwt.NewClaim()
	claims.Set("id", user.ID)
	claims.Set("email", user.Email)
	claims.Set("expires", expires.Unix())

	access, err := auth.cfg.Token.Algorithm.Encode(claims)
	if err != nil {
		return nil, err
	}

	refresh := model.Refresh{
		ID:      random.String(idsize),
		UserID:  user.ID,
		Created: now,
		Expires: expires,
	}

	if err := tx.Create(&refresh).Error; err != nil {
		return nil, err
	}

	user.LastLogin = now
	if err := tx.Save(user).Error; err != nil {
		return nil, err
	}

	token := Token{
		Access:  access,
		Refresh: refresh.ID,
		Expires: expires.Unix(),
	}

	return &token, nil
}

func (auth *Auth) confirmate(tx *gorm.DB, user *model.User, conf ConfirmationConfig) error {

	now := auth.now()
	con := model.Confirmation{
		ID:      random.String(idsize),
		UserID:  user.ID,
		Kind:    conf.Kind,
		Created: now,
		Expires: now.Add(conf.TTL),
	}

	if err := tx.Create(&con).Error; err != nil {
		return err
	}

	msg := mailcd.Request{
		TemplateLang: "en",
		TemplateName: conf.Template,
		TemplateArgs: map[string]interface{}{"link": conf.Link, "id": con.ID},
		To:           []mailcd.Address{{Email: user.Email}},
	}

	return auth.sender.Send(msg)
}

func (auth *Auth) confirm(tx *gorm.DB, id string, conf ConfirmationConfig, fun userfn) error {
	con := model.Confirmation{ID: id}

	res := tx.Preload("User").First(&con)
	if res.RecordNotFound() {
		return NotFoundError{Message: "confirmation-notfound"}
	}

	if res.Error != nil {
		return res.Error
	}

	if con.Expires.Sub(auth.now()) < 0 {
		if err := tx.Delete(&con).Error; err != nil {
			return err
		}

		if err := auth.confirmate(tx, con.User, conf); err != nil {
			return err
		}

		return ExpiredError{Message: "confirmation-expired"}
	}

	if err := fun(con.User); err != nil {
		return err
	}

	return tx.Save(con.User).Error
}

func hashpw(user *model.User, hashcost int) error {
	pw, err := bcrypt.GenerateFromPassword([]byte(user.Password), hashcost)
	if err != nil {
		return err
	}

	user.Password = string(pw)
	return nil
}

func checkpw(hash, clear string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(clear)) == nil
}

func emailUniq(db *gorm.DB, msg string) validation.Rule {
	return func(v interface{}) error {
		email := v.(*string)

		user, err := findUser(db, *email)
		if err != nil {
			return validation.Panic{Err: err.Error()}
		}

		if user != nil {
			return errors.New("email-inuse")
		}

		return nil
	}
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
