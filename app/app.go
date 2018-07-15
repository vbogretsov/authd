package app

import (
	"errors"
	"time"

	"github.com/robbert229/jwt"

	"github.com/jinzhu/gorm"
	"github.com/vbogretsov/go-mailcd"

	"github.com/vbogretsov/authd/model"
)

const (
	conkindSignUp  = 0
	conkindPwReset = 1
)

// Timer represets current time provider.
type Timer func() time.Time

// Config represens application configuration.
type Config struct {
	TokenTTL       time.Duration
	RefreshTTL     time.Duration
	SecretKey      string
	PasswordMinLen int
	Confirmation   ConfirmationConfig
}

// ConfirmationConfig represents confirmation configuration seciton.
type ConfirmationConfig struct {
	TTL             time.Duration
	SignUpLink      string
	SingUpTemplate  string
	ResetPwLink     string
	ResetPwTemplate string
}

// App represents authd application controller.
type App struct {
	db          *gorm.DB
	issuer      *issuer
	confirmator *confirmator
	accounter   *accounter
}

// New creates new application controller.
func New(config Config, timer Timer, db *gorm.DB, sender mailcd.Sender) (*App, error) {
	if timer == nil {
		return nil, errors.New("timer cannot be nil")
	}

	if db == nil {
		return nil, errors.New("dbpool cannot be nil")
	}

	if sender == nil {
		return nil, errors.New("sender cannot be nil")
	}

	userRule, err := makeUserRule(db, config)
	if err != nil {
		return nil, err
	}

	cmeta := make([]confirmationMeta, 2)
	cmeta[conkindSignUp] = confirmationMeta{
		link:     config.Confirmation.SignUpLink,
		template: config.Confirmation.SingUpTemplate,
	}
	cmeta[conkindPwReset] = confirmationMeta{
		link:     config.Confirmation.ResetPwLink,
		template: config.Confirmation.ResetPwTemplate,
	}

	app := App{
		db: db,
		issuer: &issuer{
			timer:      timer,
			algorithm:  jwt.HmacSha256(config.SecretKey),
			accessTTL:  config.TokenTTL,
			refreshTTL: config.RefreshTTL,
		},
		confirmator: &confirmator{
			ttl:    config.Confirmation.TTL,
			timer:  timer,
			sender: sender,
			meta:   cmeta,
		},
		accounter: &accounter{
			rule: userRule,
		},
	}

	return &app, nil
}

// SignUp creates new user account if a user email is not in use.
// The confirmation email will be sent to user email.
func (app *App) SignUp(user *model.User) error {
	return atomic(app.db, func(tx *gorm.DB) error {
		if err := app.accounter.create(tx, user); err != nil {
			return err
		}

		if err := app.confirmator.generate(tx, user, conkindSignUp); err != nil {
			return err
		}

		return nil
	})
}

// ConfirmUser makes a user account active.
func (app *App) ConfirmUser(conID string) error {
	return atomic(app.db, func(tx *gorm.DB) error {
		return app.confirmator.confirm(tx, conID, func(user *model.User) error {
			user.Active = true
			return nil
		})
	})
}

// ResetPassword sends email with password reset link to the email provided.
// If email was sent or no user with the email provided was found returns nil.
func (app *App) ResetPassword(email string) error {
	return atomic(app.db, func(tx *gorm.DB) error {
		user, err := app.accounter.find(tx, email)
		if err != nil {
			return err
		}

		if user == nil {
			return nil
		}

		return app.confirmator.generate(tx, user, conkindPwReset)
	})
}

// UpdatePassword updates user password.
func (app *App) UpdatePassword(conID string, password string) error {
	return atomic(app.db, func(tx *gorm.DB) error {
		return app.confirmator.confirm(tx, conID, func(user *model.User) error {
			user.Password = password
			return hashpw(user)
		})
	})
}

// SignIn generates access token and refresh token if a user with the email
// provided exists, is't account is active and password matches. If the
// access token is expired, it can be refreshed using the refresh token.
func (app *App) SignIn(email, password string) (*model.Token, error) {
	user, err := app.accounter.find(app.db, email)
	if err != nil {
		return nil, err
	}

	if user == nil || !user.Active || !checkpw(user.Password, password) {
		return nil, UnauthorizedError{Message: LoginInvalid}
	}

	var token *model.Token

	err = atomic(app.db, func(tx *gorm.DB) error {
		token, err = app.issuer.grant(app.db, user)
		return err
	})

	return token, err
}

// Refresh refreshes the access token which correspons to the refresh token
// provided.
func (app *App) Refresh(refreshID string) (*model.Token, error) {
	return app.issuer.refresh(app.db, refreshID)
}
