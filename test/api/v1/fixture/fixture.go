package fixture

import (
	"net/http"
	"time"

	"github.com/vbogretsov/go-validation"

	"github.com/vbogretsov/authd/api"
	"github.com/vbogretsov/authd/api/v1"
	"github.com/vbogretsov/authd/auth"

	"github.com/vbogretsov/authd/test/api/v1/suite"
)

var (
	defaultCredentials = auth.Credentials{
		Email:    "user@mail.com",
		Password: "123456",
	}

	messageConfirmationSent = v1.Message{
		Message: "user-confirmation-sent",
	}
	messageUserActivated = v1.Message{
		Message: "user-activated",
	}
	messagePwresetSent = v1.Message{
		Message: "password-reset-sent",
	}
	messagePasswordUpdated = v1.Message{
		Message: "password-updated",
	}

	errorEmailInvalid = validationError{
		Path:  ".email",
		Error: "email-invalid",
	}
	errorEmailUniq = validationError{
		Path:  ".email",
		Error: "email-uniq",
	}
	errorPasswordShort = validationError{
		Path:  ".password",
		Error: "password-short",
		Params: validation.Params{
			"minLen": float64(suite.Config.Password.MinLen),
		},
	}
	errorConfirmationNotFound = api.Error{
		Message: "confirmation-notfound",
	}
	errorConfirmationExpired = api.Error{
		Message: "confirmation-expired",
	}
	errorUnauthorized = api.Error{
		Message: "login-invalid",
	}
	errorRefreshNotFound = api.Error{
		Message: "refresh-notfound",
	}
	errorRefreshExpired = api.Error{
		Message: "refresh-expired",
	}
)

func readValidCID(email string, s *suite.Suite, c auth.ConfirmConfig) string {
	mail, _ := s.Sender.ReadMail(email)
	id := mail.TemplateArgs["id"].(string)
	return id
}

func readInvalidCID(email string, s *suite.Suite, c auth.ConfirmConfig) string {
	return "invalid"
}

func readExpiredCID(email string, s *suite.Suite, c auth.ConfirmConfig) string {
	s.Timer.Set(s.Timer.Now().Add(c.TTL).Add(time.Second))

	mail, _ := s.Sender.ReadMail(email)
	id := mail.TemplateArgs["id"].(string)
	return id
}

type validationError struct {
	Path   string
	Error  string
	Params map[string]interface{}
}

type Response struct {
	Code int
	Type interface{}
	Body interface{}
}

type SignUp struct {
	Name        string
	Credentials *auth.Credentials
	Response    Response
}

var SignUpSet = []SignUp{
	{
		Name: "Ok",
		Credentials: &auth.Credentials{
			Email:    "user@mail.com",
			Password: "123456",
		},
		Response: Response{
			Code: http.StatusOK,
			Type: &v1.Message{},
			Body: &messageConfirmationSent,
		},
	},
	{
		Name: "MissingEmail",
		Credentials: &auth.Credentials{
			Password: "123456",
		},
		Response: Response{
			Code: http.StatusBadRequest,
			Type: &api.Error{Errors: &[]validationError{}},
			Body: &api.Error{
				Message: "validation errors",
				Errors:  &[]validationError{errorEmailInvalid},
			},
		},
	},
	{
		Name: "MissingPassword",
		Credentials: &auth.Credentials{
			Email: "user@mail.com",
		},
		Response: Response{
			Code: http.StatusBadRequest,
			Type: &api.Error{Errors: &[]validationError{}},
			Body: &api.Error{
				Message: "validation errors",
				Errors:  &[]validationError{errorPasswordShort},
			},
		},
	},
	{
		Name:        "MissingEmailAndPassword",
		Credentials: nil,
		Response: Response{
			Code: http.StatusBadRequest,
			Type: &api.Error{Errors: &[]validationError{}},
			Body: &api.Error{
				Message: "validation errors",
				Errors: &[]validationError{
					errorEmailInvalid,
					errorPasswordShort,
				},
			},
		},
	},
	{
		Name: "DuplicatedEmail",
		Credentials: &auth.Credentials{
			Email:    suite.DuplicatedEmail,
			Password: "123456",
		},
		Response: Response{
			Code: http.StatusBadRequest,
			Type: &api.Error{Errors: &[]validationError{}},
			Body: &api.Error{
				Message: "validation errors",
				Errors:  &[]validationError{errorEmailUniq},
			},
		},
	},
}

type ConfirmUser struct {
	Name        string
	Credentials *auth.Credentials
	Response    Response
	ReadID      func(string, *suite.Suite, auth.ConfirmConfig) string
}

var ConfirmUserSet = []ConfirmUser{
	{
		Name:        "Ok",
		Credentials: &defaultCredentials,
		ReadID:      readValidCID,
		Response: Response{
			Code: http.StatusOK,
			Type: &v1.Message{},
			Body: &messageUserActivated,
		},
	},
	{
		Name:        "NotFound",
		Credentials: &defaultCredentials,
		ReadID:      readInvalidCID,
		Response: Response{
			Code: http.StatusNotFound,
			Type: &api.Error{},
			Body: &errorConfirmationNotFound,
		},
	},
	{
		Name:        "Expired",
		Credentials: &defaultCredentials,
		ReadID:      readExpiredCID,
		Response: Response{
			Code: http.StatusRequestTimeout,
			Type: &api.Error{},
			Body: &errorConfirmationExpired,
		},
	},
}

type SignIn struct {
	Name        string
	Credentials *auth.Credentials
	CreateUser  bool
	ConfirmUser bool
	Response    Response
}

var SignInSet = []SignIn{
	{
		Name:        "Ok",
		Credentials: &defaultCredentials,
		CreateUser:  true,
		ConfirmUser: true,
	},
	{
		Name:        "NotExist",
		Credentials: &defaultCredentials,
		CreateUser:  false,
		ConfirmUser: false,
		Response: Response{
			Code: http.StatusUnauthorized,
			Type: &api.Error{},
			Body: &errorUnauthorized,
		},
	},
	{
		Name:        "NotConfirmed",
		Credentials: &defaultCredentials,
		CreateUser:  false,
		ConfirmUser: false,
		Response: Response{
			Code: http.StatusUnauthorized,
			Type: &api.Error{},
			Body: &errorUnauthorized,
		},
	},
	{
		Name:        "MissingEmailAndPassword",
		Credentials: nil,
		CreateUser:  false,
		ConfirmUser: false,
		Response: Response{
			Code: http.StatusUnauthorized,
			Type: &api.Error{},
			Body: &errorUnauthorized,
		},
	},
}

type ResetPassword struct {
	Name        string
	Credentials *auth.Credentials
	Email       *auth.Email
	HasInbox    bool
	Response    Response
}

var ResetPasswordSet = []ResetPassword{
	{
		Name:        "Ok",
		Credentials: &defaultCredentials,
		Email:       &auth.Email{Email: defaultCredentials.Email},
		HasInbox:    true,
		Response: Response{
			Code: http.StatusOK,
			Type: &v1.Message{},
			Body: &messagePwresetSent,
		},
	},
	{
		Name:        "EmailNotExists",
		Credentials: &defaultCredentials,
		Email:       &auth.Email{Email: "invalid@mail.com"},
		HasInbox:    false,
		Response: Response{
			Code: http.StatusOK,
			Type: &v1.Message{},
			Body: &messagePwresetSent,
		},
	},
	{
		Name:        "InvalidEmail",
		Credentials: &defaultCredentials,
		Email:       &auth.Email{Email: "invalidmail.com"},
		HasInbox:    false,
		Response: Response{
			Code: http.StatusBadRequest,
			Type: &api.Error{Errors: &[]validationError{}},
			Body: &api.Error{
				Message: "validation errors",
				Errors:  &[]validationError{errorEmailInvalid},
			},
		},
	},
	{
		Name:        "MissingEmail",
		Credentials: &defaultCredentials,
		Email:       nil,
		HasInbox:    false,
		Response: Response{
			Code: http.StatusBadRequest,
			Type: &api.Error{Errors: &[]validationError{}},
			Body: &api.Error{
				Message: "validation errors",
				Errors:  &[]validationError{errorEmailInvalid},
			},
		},
	},
}

type UpdatePassword struct {
	Name        string
	Credentials *auth.Credentials
	Password    *auth.Password
	Response    Response
	ReadID      func(string, *suite.Suite, auth.ConfirmConfig) string
}

var UpdatePasswordSet = []UpdatePassword{
	{
		Name:        "Ok",
		Credentials: &defaultCredentials,
		Password:    &auth.Password{Password: "654321"},
		ReadID:      readValidCID,
		Response: Response{
			Code: http.StatusOK,
			Type: &v1.Message{},
			Body: &messagePasswordUpdated,
		},
	},
	{
		Name:        "NotFound",
		Credentials: &defaultCredentials,
		Password:    &auth.Password{Password: "654321"},
		ReadID:      readInvalidCID,
		Response: Response{
			Code: http.StatusNotFound,
			Type: &api.Error{},
			Body: &errorConfirmationNotFound,
		},
	},
	{
		Name:        "Expired",
		Credentials: &defaultCredentials,
		Password:    &auth.Password{Password: "654321"},
		ReadID:      readExpiredCID,
		Response: Response{
			Code: http.StatusRequestTimeout,
			Type: &api.Error{},
			Body: &errorConfirmationExpired,
		},
	},
	{
		Name:        "PasswordShort",
		Credentials: &defaultCredentials,
		Password:    &auth.Password{Password: "65432"},
		ReadID:      readValidCID,
		Response: Response{
			Code: http.StatusBadRequest,
			Type: &api.Error{Errors: &[]validationError{}},
			Body: &api.Error{
				Message: "validation errors",
				Errors:  &[]validationError{errorPasswordShort},
			},
		},
	},
}

type Refresh struct {
	Name        string
	Credentials *auth.Credentials
	Delay       time.Duration
	Response    Response
}

var RefreshSet = []Refresh{
	{
		Name:        "Ok",
		Credentials: &defaultCredentials,
	},
	{
		Name:        "Invalid",
		Credentials: &defaultCredentials,
		Response: Response{
			Code: http.StatusNotFound,
			Type: &api.Error{Errors: &[]validationError{}},
			Body: &errorRefreshNotFound,
		},
	},
	{
		Name:        "Expired",
		Credentials: &defaultCredentials,
		Delay:       suite.Config.Token.RefreshTTL,
		Response: Response{
			Code: http.StatusRequestTimeout,
			Type: &api.Error{Errors: &[]validationError{}},
			Body: &errorRefreshExpired,
		},
	},
}
