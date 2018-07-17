package v1_test

import (
	"flag"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	apiv1 "github.com/vbogretsov/authd/api/v1"
	"github.com/vbogretsov/authd/test/api/v1/apiurl"
	"github.com/vbogretsov/authd/test/api/v1/fixture"
	"github.com/vbogretsov/authd/test/api/v1/suite"
)

var dbconn = flag.String("dbconn", "", "database connection string")

func TestSignUp(t *testing.T) {
	for _, fx := range fixture.SignUpSet {
		s := suite.New(t, *dbconn)
		t.Run(fx.Name, func(t *testing.T) {
			resp := s.Client.Post(apiurl.SignUp(), nil, fx.Creadentials)
			suite.CheckResponse(t, fx.BodyType, fx.Response, resp)

			if resp.Code == http.StatusOK {
				link := s.Config.Confirmation.SignUpLink
				template := s.Config.Confirmation.SingUpTemplate
				s.CheckInbox(t, fx.Creadentials.Email, template, link)
			}
		})
		s.Cleanup(t)
	}
}

func TestConfirmUser(t *testing.T) {
	for _, fx := range fixture.ConfirmUserSet {
		s := suite.New(t, *dbconn)
		t.Run(fx.Name, func(t *testing.T) {
			s.SignUp(t, fx.Credentials)

			conID := fx.ReadID(s, fx.Credentials.Email)
			resp := s.Client.Post(apiurl.ConfirmUser(conID), nil, nil)
			suite.CheckResponse(t, fx.BodyType, fx.Response, resp)

			if resp.Code == http.StatusRequestTimeout {
				link := s.Config.Confirmation.SignUpLink
				template := s.Config.Confirmation.SingUpTemplate
				s.CheckInbox(t, fx.Credentials.Email, template, link)
			}

			if resp.Code == http.StatusOK {
				resp = s.Client.Post(apiurl.SignIn(), nil, fx.Credentials)
				require.Equal(t, http.StatusOK, resp.Code)
			}
		})
		s.Cleanup(t)
	}
}

func TestSignIn(t *testing.T) {
	for _, fx := range fixture.SignInSet {
		s := suite.New(t, *dbconn)
		t.Run(fx.Name, func(t *testing.T) {
			if fx.CreateUser {
				s.SignUp(t, *fx.Credentials)
				if fx.ConfirmUser {
					s.ConfirmUser(t, fx.Credentials.Email)
				}
			}

			resp := s.Client.Post(apiurl.SignIn(), nil, fx.Credentials)

			if resp.Code == http.StatusOK {
				// TODO: access protected resource
			} else {
				suite.CheckResponse(t, fx.BodyType, fx.Response, resp)
			}
		})
		s.Cleanup(t)
	}
}

func TestResetPassword(t *testing.T) {
	for _, fx := range fixture.ResetPasswordSet {
		s := suite.New(t, *dbconn)
		t.Run(fx.Name, func(t *testing.T) {
			s.SignUp(t, fx.Credentials)
			s.ConfirmUser(t, fx.Credentials.Email)

			resp := s.Client.Post(apiurl.ResetPassword(), nil, fx.Email)
			suite.CheckResponse(t, fx.BodyType, fx.Response, resp)

			if fx.HasInbox {
				link := s.Config.Confirmation.ResetPwLink
				template := s.Config.Confirmation.ResetPwTemplate
				s.CheckInbox(t, fx.Credentials.Email, template, link)
			} else {
				_, hasmail := s.Sender.ReadMail(fx.Email.Email)
				require.False(t, hasmail)
			}
		})
		s.Cleanup(t)
	}
}

func TestUpdatePassword(t *testing.T) {
	for _, fx := range fixture.UpdatePasswordSet {
		s := suite.New(t, *dbconn)
		t.Run(fx.Name, func(t *testing.T) {
			s.SignUp(t, fx.Credentials)
			s.ConfirmUser(t, fx.Credentials.Email)
			s.ResetPassword(t, fx.Credentials.Email)

			conID := fx.ReadID(s, fx.Credentials.Email)
			resp := s.Client.Post(apiurl.UpdatePassword(conID), nil, fx.Password)
			suite.CheckResponse(t, fx.BodyType, fx.Response, resp)

			if resp.Code == http.StatusOK {
				cred := apiv1.Credentials{
					Email:    fx.Credentials.Email,
					Password: fx.Password.Password,
				}
				resp := s.Client.Post(apiurl.SignIn(), nil, cred)
				require.Equal(t, http.StatusOK, resp.Code)
			}
		})
		s.Cleanup(t)
	}
}
