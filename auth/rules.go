package auth

import (
	"errors"

	"github.com/jinzhu/gorm"

	"github.com/vbogretsov/go-validation"
	"github.com/vbogretsov/go-validation/rule"

	"github.com/vbogretsov/authd/model"
)

func attrUserEmail(v interface{}) interface{} {
	return &v.(*model.User).Email
}

func attrUserPassword(v interface{}) interface{} {
	return &v.(*model.User).Password
}

func attrEmail(v interface{}) interface{} {
	return &v.(*Email).Email
}

func attrPassword(v interface{}) interface{} {
	return &v.(*Password).Password
}

func emailuniq(msg string) validation.Rule {
	return func(ctx interface{}) func(interface{}) error {
		return func(v interface{}) error {
			db := ctx.(*gorm.DB)
			email := v.(*string)

			user, err := findUser(db, *email)
			if err != nil {
				return validation.Panic{Err: err}
			}

			if user != nil {
				return errors.New(msg)
			}

			return nil
		}
	}
}

type rules struct {
	user     validation.Rule
	email    func(interface{}) error
	password func(interface{}) error
}

func newRules(conf Config) *rules {
	email := rule.StrEmail("email-invalid")
	passw := rule.StrMinLen(conf.Password.MinLen, "password-short")

	ur := validation.Struct(&model.User{}, "json", []validation.Field{
		{
			Attr:  attrUserEmail,
			Rules: []validation.Rule{email, emailuniq("email-uniq")},
		},
		{
			Attr:  attrUserPassword,
			Rules: []validation.Rule{passw},
		},
	})

	em := validation.Struct(&Email{}, "json", []validation.Field{
		{
			Attr:  attrEmail,
			Rules: []validation.Rule{email},
		},
	})(nil)

	pw := validation.Struct(&Password{}, "json", []validation.Field{
		{
			Attr:  attrPassword,
			Rules: []validation.Rule{passw},
		},
	})(nil)

	return &rules{email: em, password: pw, user: ur}
}
