package v1_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/robbert229/jwt"

	"github.com/labstack/gommon/random"

	"github.com/stretchr/testify/require"
	mail "github.com/vbogretsov/go-mail"
	"github.com/vbogretsov/techo"

	"github.com/vbogretsov/authd/auth"

	"github.com/vbogretsov/authd/test/api/v1/apiurl"
	"github.com/vbogretsov/authd/test/api/v1/fixture"
	"github.com/vbogretsov/authd/test/api/v1/suite"
)

var (
	dbconn = flag.String("dbconn", "", "database connection string")
	ntests = flag.Int("ntests", 10, "number of goroutines for TestConcurrentAccess")
)

func maxconn() int {
	return *ntests/2 + 1
}

func checkResponse(t *testing.T, exp fixture.Response, act techo.Response) {
	require.Equal(t, exp.Code, act.Code, string(act.Body))

	expBody := exp.Body
	actBody := exp.Type

	require.NoError(t, json.Unmarshal(act.Body, actBody))
	require.Equal(t, expBody, actBody, string(act.Body))
}

func checkInbox(t *testing.T, email string, s *suite.Suite, c auth.ConfirmConfig) {
	act, ok := s.Sender.ReadMail(email)
	require.True(t, ok, "inbox %s empty", email)

	require.Contains(t, act.TemplateArgs, "id")

	exp := mail.Request{
		TemplateLang: "en",
		TemplateName: c.Template,
		TemplateArgs: map[string]interface{}{
			"link": c.Link,
			"id":   act.TemplateArgs["id"].(string),
		},
		To: []mail.Address{{Email: email}},
	}

	require.Equal(t, exp, act)
}

func checkToken(t *testing.T, email string, data []byte, s *suite.Suite) {
	token := auth.Token{}
	require.NoError(t, json.Unmarshal(data, &token))

	expiresExp := s.Timer.Now().Add(s.Config.Token.AccessTTL).Unix()
	expiresAct := token.Expires
	require.Equal(t, expiresExp, expiresAct)

	algorithm := jwt.HmacSha512(s.Config.Token.SecretKey)

	claims, err := algorithm.Decode(token.Access)
	require.NoError(t, err)

	userID, err := claims.Get("id")
	require.NoError(t, err)
	require.Len(t, userID.(string), 32)

	userEmail, err := claims.Get("email")
	require.NoError(t, err)
	require.Equal(t, email, userEmail)
}

func TestSignUp(t *testing.T) {
	for _, fx := range fixture.SignUpSet {
		s := suite.New(t, *dbconn, maxconn())

		t.Run(fx.Name, func(t *testing.T) {
			resp := s.Client.Post(apiurl.SignUp(), nil, fx.Credentials)
			checkResponse(t, fx.Response, resp)

			if resp.Code == http.StatusOK {
				checkInbox(t, fx.Credentials.Email, s, s.Config.SignUp)
			}
		})

		s.Cleanup(t)
	}
}

func TestConfirmUser(t *testing.T) {
	for _, fx := range fixture.ConfirmUserSet {
		s := suite.New(t, *dbconn, maxconn())

		t.Run(fx.Name, func(t *testing.T) {
			s.SignUp(t, *fx.Credentials)

			conID := fx.ReadID(fx.Credentials.Email, s, s.Config.SignUp)
			resp := s.Client.Post(apiurl.ConfirmUser(conID), nil, nil)
			checkResponse(t, fx.Response, resp)

			if resp.Code == http.StatusRequestTimeout {
				checkInbox(t, fx.Credentials.Email, s, s.Config.SignUp)
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
		s := suite.New(t, *dbconn, maxconn())
		t.Run(fx.Name, func(t *testing.T) {
			if fx.CreateUser {
				s.SignUp(t, *fx.Credentials)
				if fx.ConfirmUser {
					s.ConfirmUser(t, fx.Credentials.Email)
				}
			}

			resp := s.Client.Post(apiurl.SignIn(), nil, fx.Credentials)

			if resp.Code == http.StatusOK {
				checkToken(t, fx.Credentials.Email, resp.Body, s)
			} else {
				checkResponse(t, fx.Response, resp)
			}
		})
		s.Cleanup(t)
	}
}

func TestResetPassword(t *testing.T) {
	for _, fx := range fixture.ResetPasswordSet {
		s := suite.New(t, *dbconn, maxconn())
		t.Run(fx.Name, func(t *testing.T) {
			s.SignUp(t, *fx.Credentials)
			s.ConfirmUser(t, fx.Credentials.Email)

			resp := s.Client.Post(apiurl.ResetPassword(), nil, fx.Email)
			checkResponse(t, fx.Response, resp)

			if fx.HasInbox {
				checkInbox(t, fx.Credentials.Email, s, s.Config.ResetPw)
			} else if fx.Email != nil {
				_, hasmail := s.Sender.ReadMail(fx.Email.Email)
				require.False(t, hasmail)
			}
		})
		s.Cleanup(t)
	}
}

func TestUpdatePassword(t *testing.T) {
	for _, fx := range fixture.UpdatePasswordSet {
		s := suite.New(t, *dbconn, maxconn())

		t.Run(fx.Name, func(t *testing.T) {
			s.SignUp(t, *fx.Credentials)
			s.ConfirmUser(t, fx.Credentials.Email)
			s.SignIn(t, *fx.Credentials)
			s.ResetPassword(t, fx.Credentials.Email)

			conID := fx.ReadID(fx.Credentials.Email, s, s.Config.ResetPw)
			resp := s.Client.Post(apiurl.UpdatePassword(conID), nil, fx.Password)
			checkResponse(t, fx.Response, resp)

			if resp.Code == http.StatusOK {
				resp1 := s.Client.Post(apiurl.SignIn(), nil, auth.Credentials{
					Email:    fx.Credentials.Email,
					Password: fx.Password.Password,
				})
				require.Equal(t, http.StatusOK, resp1.Code)

				resp2 := s.Client.Post(apiurl.SignIn(), nil, fx.Credentials)
				require.Equal(t, http.StatusUnauthorized, resp2.Code)
			}
		})

		s.Cleanup(t)
	}
}

func TestRefresh(t *testing.T) {
	for _, fx := range fixture.RefreshSet {
		s := suite.New(t, *dbconn, maxconn())

		t.Run(fx.Name, func(t *testing.T) {
			s.SignUp(t, *fx.Credentials)
			s.ConfirmUser(t, fx.Credentials.Email)

			token := s.SignIn(t, *fx.Credentials)
			resp := s.Client.Post(apiurl.Refresh(), nil, auth.Refresh{
				Refresh: token.Refresh,
			})

			if resp.Code == http.StatusOK {
				checkToken(t, fx.Credentials.Email, resp.Body, s)
			} else {
				checkResponse(t, fx.Response, resp)
			}
		})

		s.Cleanup(t)
	}
}

func TestConcurrentAccess(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(*ntests)

	s := suite.New(t, *dbconn, maxconn())

	for i := 0; i < *ntests; i++ {
		go func() {
			t.Run("Concurrent", func(t *testing.T) {
				defer wg.Done()
				email := fmt.Sprintf(
					"%s@mail.com",
					random.String(10, random.Alphabetic))

				creds := auth.Credentials{
					Email:    email,
					Password: random.String(10, random.Alphabetic),
				}

				s.SignUp(t, creds)
				s.ConfirmUser(t, creds.Email)
				s.SignIn(t, creds)
			})
		}()
	}

	s.Cleanup(t)

	wg.Wait()
}
