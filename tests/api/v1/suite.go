package test

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/middleware"
	"github.com/steakknife/devnull"
	"github.com/stretchr/testify/require"
	"github.com/vbogretsov/go-mailcd"
	mailmock "github.com/vbogretsov/go-mailcd/mock"
	mocktime "github.com/vbogretsov/go-mocktime"
	"github.com/vbogretsov/techo"

	_ "github.com/jinzhu/gorm/dialects/postgres" // gorm postgres import

	api "github.com/vbogretsov/authd/api"
	apiv1 "github.com/vbogretsov/authd/api/v1"
	app "github.com/vbogretsov/authd/app"
	model "github.com/vbogretsov/authd/model"
)

const duplicatedEmail = "duplicate@mail.com"

var dbconn = flag.String("dbconn", "", "database connection string")

var config = app.Config{
	PasswordMinLen: 6,
	SecretKey:      "test-secret-key",
	TokenTTL:       time.Duration(time.Second * 10),
	RefreshTTL:     time.Duration(time.Second * 100),
	Confirmation: app.ConfirmationConfig{
		TTL:             time.Duration(time.Second * 100),
		SignUpLink:      "signup-link",
		SingUpTemplate:  "signup-template",
		ResetPwLink:     "pwreset-link",
		ResetPwTemplate: "pwreset-template",
	},
}

func reverseSignUp() string {
	return fmt.Sprintf("%s%s", apiv1.Conf.Group, apiv1.Conf.SignUpURL)
}

func reverseConfirmUser(id string) string {
	url := fmt.Sprintf("%s%s", apiv1.Conf.Group, apiv1.Conf.ConfirmUserURL)
	return strings.Replace(url, ":id", id, 1)
}

func reverseSignIn() string {
	return fmt.Sprintf("%s%s", apiv1.Conf.Group, apiv1.Conf.SignInURL)
}

func reverseResetPassword() string {
	return fmt.Sprintf("%s%s", apiv1.Conf.Group, apiv1.Conf.ResetPasswordURL)
}

func reverseUpdatePassword(id string) string {
	url := fmt.Sprintf("%s%s", apiv1.Conf.Group, apiv1.Conf.UpdatePasswordURL)
	return strings.Replace(url, ":id", id, 1)
}

func utc(m *mocktime.Mock) app.Timer {
	return func() time.Time {
		return m.Now().UTC()
	}
}

type suite struct {
	config app.Config
	db     *gorm.DB
	timer  *mocktime.Mock
	sender *mailmock.Sender
	client *techo.Client
}

func setup(t *testing.T) *suite {
	var err error

	s := new(suite)

	s.db, err = gorm.Open("postgres", *dbconn)
	require.Nil(t, err)

	s.timer = mocktime.New()
	s.sender = mailmock.New()
	s.config = config

	ap, err := app.New(s.config, utc(s.timer), s.db, s.sender)
	require.Nil(t, err)

	middleware.DefaultLoggerConfig.Output = devnull.Writer

	e := api.New()
	apiv1.Include(ap, e)

	s.client = techo.New(e, json.Marshal)
	s.client.Header.Set("Content-Type", "application/json")

	duplicate := model.User{
		ID:       "test",
		Email:    duplicatedEmail,
		Password: "123456",
	}

	require.NoError(t, s.db.Create(&duplicate).Error)
	return s
}

func (s *suite) cleanup(t *testing.T) {
	require.NoError(t, s.db.Delete(&model.User{}).Error)
}

func (s *suite) signUp(t *testing.T, creds apiv1.Credentials) {
	resp := s.client.Post(reverseSignUp(), nil, creds)
	require.Equal(t, http.StatusOK, resp.Code)
}

func (s *suite) confirmUser(t *testing.T) {
	conID := s.sender.Inbox[0].TemplateArgs["id"].(string)
	s.sender.Reset()

	resp := s.client.Post(reverseConfirmUser(conID), nil, nil)
	require.Equal(t, http.StatusOK, resp.Code)
}

func (s *suite) resetPassword(t *testing.T, email string) {
	body := apiv1.Email{Email: email}
	resp := s.client.Post(reverseResetPassword(), nil, body)
	require.Equal(t, http.StatusOK, resp.Code)
}

func (*suite) checkResponse(t *testing.T, bodyType interface{}, exp, act techo.Response) {
	require.Equal(t, exp.Code, act.Code)

	expBody := bodyType
	require.NoError(t, json.Unmarshal(exp.Body, &expBody))

	actBody := bodyType
	require.NoError(t, json.Unmarshal(act.Body, &actBody))

	require.Equal(t, expBody, actBody)
}

func (s *suite) checkInbox(t *testing.T, email, template, link string) {
	require.Len(t, s.sender.Inbox, 1)
	actmsg := s.sender.Inbox[0]

	require.Contains(t, actmsg.TemplateArgs, "id")

	expmsg := mailcd.Request{
		TemplateLang: "en",
		TemplateName: template,
		TemplateArgs: map[string]interface{}{
			"link": link,
			"id":   actmsg.TemplateArgs["id"].(string),
		},
		To: []mailcd.Address{{Email: email}},
	}

	require.Equal(t, expmsg, actmsg)
}

func (s *suite) testSignUp(t *testing.T, fx signUpFixture) {
	resp := s.client.Post(reverseSignUp(), nil, fx.Creadentials)
	s.checkResponse(t, fx.BodyType, fx.Response, resp)

	if resp.Code == http.StatusOK {
		link := s.config.Confirmation.SignUpLink
		template := s.config.Confirmation.SingUpTemplate
		s.checkInbox(t, fx.Creadentials.Email, template, link)
	}
}

func (s *suite) testConfirmUser(t *testing.T, fx confirmUserFixture) {
	conID := fx.ReadID(s)
	resp := s.client.Post(reverseConfirmUser(conID), nil, nil)
	s.checkResponse(t, fx.BodyType, fx.Response, resp)

	if resp.Code == http.StatusRequestTimeout {
		link := s.config.Confirmation.SignUpLink
		template := s.config.Confirmation.SingUpTemplate
		s.checkInbox(t, fx.Credentials.Email, template, link)
	}

	if resp.Code == http.StatusOK {
		resp = s.client.Post(reverseSignIn(), nil, fx.Credentials)
		require.Equal(t, http.StatusOK, resp.Code)
	}
}

func (s *suite) testSignIn(t *testing.T, fx signInFixture) {
	resp := s.client.Post(reverseSignIn(), nil, fx.Credentials)

	if resp.Code == http.StatusOK {
		// TODO: access protected resource
	} else {
		s.checkResponse(t, fx.BodyType, fx.Response, resp)
	}
}

func (s *suite) testResetPassword(t *testing.T, fx resetPasswordFixture) {
	resp := s.client.Post(reverseResetPassword(), nil, fx.Email)
	s.checkResponse(t, fx.BodyType, fx.Response, resp)

	if fx.HasInbox {
		link := s.config.Confirmation.ResetPwLink
		template := s.config.Confirmation.ResetPwTemplate
		s.checkInbox(t, fx.Credentials.Email, template, link)
	} else {
		require.Len(t, s.sender.Inbox, 0)
	}
}

func (s *suite) testUpdatePassword(t *testing.T, fx updatePasswordFixture) {
	conID := fx.ReadID(s)
	resp := s.client.Post(reverseUpdatePassword(conID), nil, fx.Password)
	s.checkResponse(t, fx.BodyType, fx.Response, resp)

	if resp.Code == http.StatusOK {
		cred := apiv1.Credentials{
			Email:    fx.Credentials.Email,
			Password: fx.Password.Password,
		}
		resp := s.client.Post(reverseSignIn(), nil, cred)
		require.Equal(t, http.StatusOK, resp.Code)
	}
}
