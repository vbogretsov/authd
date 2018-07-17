package fixture

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/vbogretsov/go-validation"
	"github.com/vbogretsov/go-validation/jsonerr"
	"github.com/vbogretsov/techo"

	api "github.com/vbogretsov/authd/api"
	apiv1 "github.com/vbogretsov/authd/api/v1"

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

var defaultCredentials = apiv1.Credentials{
	Email:    "user@mail.com",
	Password: "123456",
}

func marshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

type SignUp struct {
	Name         string
	BodyType     interface{}
	Creadentials *apiv1.Credentials
	Response     techo.Response
}

var SignUpSet = []SignUp{
	{
		Name:     "Ok",
		BodyType: apiv1.Message{},
		Creadentials: &apiv1.Credentials{
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
		Creadentials: &apiv1.Credentials{
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
		Creadentials: &apiv1.Credentials{
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
		Creadentials: &apiv1.Credentials{
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
	Credentials apiv1.Credentials
	BodyType    interface{}
	Response    techo.Response
	ReadID      func(*suite.Suite, string) string
}

var ConfirmUserSet = []ConfirmUser{
	{
		Name:        "Ok",
		Credentials: defaultCredentials,
		BodyType:    apiv1.Message{},
		Response: techo.Response{
			Code: http.StatusOK,
			Body: marshal(msgUserConfirmed),
		},
		ReadID: suite.ReadValidConfirmationID,
	},
	{
		Name:        "NotFound",
		Credentials: defaultCredentials,
		BodyType:    api.Error{},
		Response: techo.Response{
			Code: http.StatusNotFound,
			Body: marshal(errConfirmationNotFound),
		},
		ReadID: suite.ReadInvalidConfirmationID,
	},
	{
		Name:        "Expired",
		Credentials: defaultCredentials,
		BodyType:    api.Error{},
		Response: techo.Response{
			Code: http.StatusRequestTimeout,
			Body: marshal(errConfirmationExpired),
		},
		ReadID: suite.ReadExpiredConfirmationID,
	},
}

type SignIn struct {
	Name        string
	Credentials *apiv1.Credentials
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
	Credentials apiv1.Credentials
	Email       *apiv1.Email
	HasInbox    bool
	BodyType    interface{}
	Response    techo.Response
}

var ResetPasswordSet = []ResetPassword{
	{
		Name:        "Ok",
		Credentials: defaultCredentials,
		Email:       &apiv1.Email{Email: defaultCredentials.Email},
		HasInbox:    true,
		BodyType:    apiv1.Message{},
		Response: techo.Response{
			Code: http.StatusOK,
			Body: marshal(msgPwResetEmailSent),
		},
	},
	{
		Name:        "EmailNotExists",
		Credentials: defaultCredentials,
		Email:       &apiv1.Email{Email: "invalid@mail.com"},
		HasInbox:    false,
		BodyType:    apiv1.Message{},
		Response: techo.Response{
			Code: http.StatusOK,
			Body: marshal(msgPwResetEmailSent),
		},
	},
	{
		Name:        "InvalidEmail",
		Credentials: defaultCredentials,
		Email:       &apiv1.Email{Email: "invalidmail.com"},
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
		Credentials: defaultCredentials,
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
	Credentials apiv1.Credentials
	Password    apiv1.Password
	BodyType    interface{}
	Response    techo.Response
	ReadID      func(*suite.Suite, string) string
}

var UpdatePasswordSet = []UpdatePassword{
	{
		Name:        "Ok",
		Credentials: defaultCredentials,
		ReadID:      suite.ReadValidConfirmationID,
		Password:    apiv1.Password{Password: "654321"},
		BodyType:    apiv1.Message{},
		Response: techo.Response{
			Code: http.StatusOK,
			Body: marshal(msgPasswordUpdated),
		},
	},
	{
		Name:        "NotFound",
		Credentials: defaultCredentials,
		ReadID:      suite.ReadInvalidConfirmationID,
		Password:    apiv1.Password{Password: "654321"},
		BodyType:    api.Error{},
		Response: techo.Response{
			Code: http.StatusNotFound,
			Body: marshal(errConfirmationNotFound),
		},
	},
	{
		Name:        "Expired",
		Credentials: defaultCredentials,
		ReadID:      suite.ReadExpiredConfirmationID,
		Password:    apiv1.Password{Password: "654321"},
		BodyType:    api.Error{},
		Response: techo.Response{
			Code: http.StatusNotFound,
			Body: marshal(errConfirmationNotFound),
		},
	},
	{
		Name:        "PasswordShort",
		Credentials: defaultCredentials,
		ReadID:      suite.ReadValidConfirmationID,
		Password:    apiv1.Password{Password: "65432"},
		BodyType:    api.Error{},
		Response: techo.Response{
			Code: http.StatusNotFound,
			Body: marshal(api.Error{
				Message: "validation errors",
				Errors:  jsonerr.Errors([]error{errPasswordShort}),
			}),
		},
	},
}
