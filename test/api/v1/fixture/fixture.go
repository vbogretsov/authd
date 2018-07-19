package fixture

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vbogretsov/go-validation"
	"github.com/vbogretsov/go-validation/jsonerr"
	"github.com/vbogretsov/techo"

	"github.com/vbogretsov/authd/api"
	apiv1 "github.com/vbogretsov/authd/api/v1"
	"github.com/vbogretsov/authd/auth"

	"github.com/vbogretsov/authd/test/api/v1/suite"
)

var (
	msgConfirmationEmailSent = apiv1.Message{
		Message: "confirmation email has been sent",
	}
	msgPwResetEmailSent = apiv1.Message{
		Message: "password reset email has been sent",
	}
	msgUserConfirmed = apiv1.Message{
		Message: "user has been activated",
	}
	msgPasswordUpdated = apiv1.Message{
		Message: "password has been updated",
	}
	errInvalidEmail = validation.StructError{
		Field: "email",
		Errors: []error{
			errors.New("invalid email address"),
		},
	}
	errDuplicatedEmail = validation.StructError{
		Field: "email",
		Errors: []error{
			errors.New("email already in use"),
		},
	}
	errPasswordShort = validation.StructError{
		Field: "password",
		Errors: []error{
			errors.New("password cannot be shorter that 6 characters"),
		},
	}
	errConfirmationNotFound = api.Error{
		Message: "confirmation not found",
	}
	errConfirmationExpired = api.Error{
		Message: "confirmation has been expired",
	}
	errUnauthorized = api.Error{
		Message: "invalid email or password",
	}
)

var defaultCredentials = auth.Credentials{
	Email:    "user@mail.com",
	Password: "123456",
}

func marshal(v interface{}) []byte {
	bts, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return bts
}

type SignUp struct {
	Name         string
	BodyType     interface{}
	Creadentials *auth.Credentials
	Response     techo.Response
}

var SignUpSet = []SignUp{
	{
		Name:     "Ok",
		BodyType: apiv1.Message{},
		Creadentials: &auth.Credentials{
			Email:    "user@mail.com",
			Password: "123456",
		},
		Response: techo.Response{
			Code: http.StatusOK,
			Body: marshal(msgConfirmationEmailSent),
		},
	},
	{
		Name:     "MissingEmail",
		BodyType: api.Error{},
		Creadentials: &auth.Credentials{
			Password: "123456",
		},
		Response: techo.Response{
			Code: http.StatusBadRequest,
			Body: marshal(api.Error{
				Message: "validation errors",
				Errors:  jsonerr.Errors([]error{errInvalidEmail}),
			}),
		},
	},
	{
		Name:         "MissingEmailAndPassword",
		BodyType:     api.Error{},
		Creadentials: nil,
		Response: techo.Response{
			Code: http.StatusBadRequest,
			Body: marshal(api.Error{
				Message: "validation errors",
				Errors: jsonerr.Errors([]error{
					errInvalidEmail,
					errPasswordShort,
				}),
			}),
		},
	},
	{
		Name:     "MissingPassword",
		BodyType: api.Error{},
		Creadentials: &auth.Credentials{
			Email: "user@mail.com",
		},
		Response: techo.Response{
			Code: http.StatusBadRequest,
			Body: marshal(api.Error{
				Message: "validation errors",
				Errors:  jsonerr.Errors([]error{errPasswordShort}),
			}),
		},
	},
	{
		Name:     "DuplicatedEmail",
		BodyType: api.Error{},
		Creadentials: &auth.Credentials{
			Email:    suite.DuplicatedEmail,
			Password: "123456",
		},
		Response: techo.Response{
			Code: http.StatusBadRequest,
			Body: marshal(api.Error{
				Message: "validation errors",
				Errors:  jsonerr.Errors([]error{errDuplicatedEmail}),
			}),
		},
	},
}

type ConfirmUser struct {
	Name        string
	Credentials *auth.Credentials
	BodyType    interface{}
	Response    techo.Response
	ReadID      func(*suite.Suite, string, auth.ConfirmationConfig) string
}

var ConfirmUserSet = []ConfirmUser{
	{
		Name:        "Ok",
		Credentials: &defaultCredentials,
		BodyType:    apiv1.Message{},
		ReadID:      suite.ReadValidConfirmationID,
		Response: techo.Response{
			Code: http.StatusOK,
			Body: marshal(msgUserConfirmed),
		},
	},
	{
		Name:        "NotFound",
		Credentials: &defaultCredentials,
		BodyType:    api.Error{},
		ReadID:      suite.ReadInvalidConfirmationID,
		Response: techo.Response{
			Code: http.StatusNotFound,
			Body: marshal(errConfirmationNotFound),
		},
	},
	{
		Name:        "Expired",
		Credentials: &defaultCredentials,
		BodyType:    api.Error{},
		ReadID:      suite.ReadExpiredConfirmationID,
		Response: techo.Response{
			Code: http.StatusRequestTimeout,
			Body: marshal(errConfirmationExpired),
		},
	},
}

type SignIn struct {
	Name        string
	Credentials *auth.Credentials
	CreateUser  bool
	ConfirmUser bool
	BodyType    interface{}
	Response    techo.Response
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
		BodyType:    api.Error{},
		Response: techo.Response{
			Code: http.StatusUnauthorized,
			Body: marshal(errUnauthorized),
		},
	},
	{
		Name:        "NotConfirmed",
		Credentials: &defaultCredentials,
		CreateUser:  false,
		ConfirmUser: false,
		BodyType:    api.Error{},
		Response: techo.Response{
			Code: http.StatusUnauthorized,
			Body: marshal(errUnauthorized),
		},
	},
	{
		Name:        "MissingEmailAndPassword",
		Credentials: nil,
		CreateUser:  false,
		ConfirmUser: false,
		BodyType:    api.Error{},
		Response: techo.Response{
			Code: http.StatusUnauthorized,
			Body: marshal(errUnauthorized),
		},
	},
}

type ResetPassword struct {
	Name        string
	Credentials *auth.Credentials
	Email       *auth.Email
	HasInbox    bool
	BodyType    interface{}
	Response    techo.Response
}

var ResetPasswordSet = []ResetPassword{
	{
		Name:        "Ok",
		Credentials: &defaultCredentials,
		Email:       &auth.Email{Email: defaultCredentials.Email},
		HasInbox:    true,
		BodyType:    apiv1.Message{},
		Response: techo.Response{
			Code: http.StatusOK,
			Body: marshal(msgPwResetEmailSent),
		},
	},
	{
		Name:        "EmailNotExists",
		Credentials: &defaultCredentials,
		Email:       &auth.Email{Email: "invalid@mail.com"},
		HasInbox:    false,
		BodyType:    apiv1.Message{},
		Response: techo.Response{
			Code: http.StatusOK,
			Body: marshal(msgPwResetEmailSent),
		},
	},
	{
		Name:        "InvalidEmail",
		Credentials: &defaultCredentials,
		Email:       &auth.Email{Email: "invalidmail.com"},
		HasInbox:    false,
		BodyType:    apiv1.Message{},
		Response: techo.Response{
			Code: http.StatusBadRequest,
			Body: marshal(api.Error{
				Message: "validation errors",
				Errors:  jsonerr.Errors([]error{errInvalidEmail}),
			}),
		},
	},
	{
		Name:        "MissingEmail",
		Credentials: &defaultCredentials,
		Email:       nil,
		HasInbox:    false,
		BodyType:    apiv1.Message{},
		Response: techo.Response{
			Code: http.StatusBadRequest,
			Body: marshal(api.Error{
				Message: "validation errors",
				Errors:  jsonerr.Errors([]error{errInvalidEmail}),
			}),
		},
	},
}

type UpdatePassword struct {
	Name        string
	Credentials *auth.Credentials
	Password    *auth.Password
	BodyType    interface{}
	Response    techo.Response
	ReadID      func(*suite.Suite, string, auth.ConfirmationConfig) string
}

var UpdatePasswordSet = []UpdatePassword{
	{
		Name:        "Ok",
		Credentials: &defaultCredentials,
		Password:    &auth.Password{Password: "654321"},
		ReadID:      suite.ReadValidConfirmationID,
		BodyType:    apiv1.Message{},
		Response: techo.Response{
			Code: http.StatusOK,
			Body: marshal(msgPasswordUpdated),
		},
	},
	{
		Name:        "NotFound",
		Credentials: &defaultCredentials,
		Password:    &auth.Password{Password: "654321"},
		ReadID:      suite.ReadInvalidConfirmationID,
		BodyType:    api.Error{},
		Response: techo.Response{
			Code: http.StatusNotFound,
			Body: marshal(errConfirmationNotFound),
		},
	},
	{
		Name:        "Expired",
		Credentials: &defaultCredentials,
		Password:    &auth.Password{Password: "654321"},
		ReadID:      suite.ReadExpiredConfirmationID,
		BodyType:    api.Error{},
		Response: techo.Response{
			Code: http.StatusRequestTimeout,
			Body: marshal(errConfirmationExpired),
		},
	},
	{
		Name:        "PasswordShort",
		Credentials: &defaultCredentials,
		Password:    &auth.Password{Password: "65432"},
		ReadID:      suite.ReadValidConfirmationID,
		BodyType:    api.Error{},
		Response: techo.Response{
			Code: http.StatusBadRequest,
			Body: marshal(api.Error{
				Message: "validation errors",
				Errors:  jsonerr.Errors([]error{errPasswordShort}),
			}),
		},
	},
}
