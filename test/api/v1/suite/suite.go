package suite

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/robbert229/jwt"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/middleware"
	"github.com/steakknife/devnull"
	"github.com/stretchr/testify/require"
	"github.com/vbogretsov/go-mailcd"
	mailmock "github.com/vbogretsov/go-mailcd/mock"
	mocktime "github.com/vbogretsov/go-mocktime"
	"github.com/vbogretsov/techo"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/vbogretsov/authd/api"
	apiv1 "github.com/vbogretsov/authd/api/v1"
	"github.com/vbogretsov/authd/auth"
	"github.com/vbogretsov/authd/model"

	"github.com/vbogretsov/authd/test/api/v1/apiurl"
)

const DuplicatedEmail = "duplicate@mail.com"

func utc(m *mocktime.Mock) auth.Timer {
	return func() time.Time {
		return m.Now().UTC()
	}
}

var Config = auth.Config{
	Password: auth.PasswordConfig{
		MinLen: 6,
	},
	Token: auth.TokenConfig{
		Algorithm:  jwt.HmacSha256("test-secret-key"),
		AccessTTL:  time.Duration(time.Second * 10),
		RefreshTTL: time.Duration(time.Second * 100),
	},
	SignUp: auth.ConfirmationConfig{
		TTL:      time.Duration(time.Second * 200),
		Link:     "signup-link",
		Template: "signup-template",
	},
	ResetPw: auth.ConfirmationConfig{
		TTL:      time.Duration(time.Second * 100),
		Link:     "pwreset-link",
		Template: "pwreset-template",
	},
}

type Suite struct {
	Config auth.Config
	DB     *gorm.DB
	Timer  *mocktime.Mock
	Sender *mailmock.Sender
	Client *techo.Client
}

func New(t *testing.T, dbconn string) *Suite {
	var err error

	s := new(Suite)

	s.DB, err = gorm.Open("postgres", dbconn)
	require.Nil(t, err)

	s.Timer = mocktime.New()
	s.Sender = mailmock.New()
	s.Config = Config

	ap, err := auth.New(s.Config, s.DB, utc(s.Timer), s.Sender)
	require.Nil(t, err)

	middleware.DefaultLoggerConfig.Output = devnull.Writer

	e := api.New()
	apiv1.Include(ap, e)

	s.Client = techo.New(e, json.Marshal)
	s.Client.Header.Set("Content-Type", "application/json")

	duplicate := model.User{
		ID:       "test",
		Email:    DuplicatedEmail,
		Password: "123456",
	}

	require.NoError(t, s.DB.Create(&duplicate).Error)
	return s
}

func (s *Suite) Cleanup(t *testing.T) {
	require.NoError(t, s.DB.Delete(&model.User{}).Error)
}

func (s *Suite) SignUp(t *testing.T, creds auth.Credentials) {
	resp := s.Client.Post(apiurl.SignUp(), nil, creds)
	require.Equal(t, http.StatusOK, resp.Code)
}

func (s *Suite) ConfirmUser(t *testing.T, email string) {
	mail, ok := s.Sender.ReadMail(email)
	require.True(t, ok)

	conID := mail.TemplateArgs["id"].(string)

	resp := s.Client.Post(apiurl.ConfirmUser(conID), nil, nil)
	require.Equal(t, http.StatusOK, resp.Code)
}

func (s *Suite) ResetPassword(t *testing.T, email string) {
	body := auth.Email{Email: email}
	resp := s.Client.Post(apiurl.ResetPassword(), nil, body)
	require.Equal(t, http.StatusOK, resp.Code)
}

func (s *Suite) CheckInbox(t *testing.T, email string, conf auth.ConfirmationConfig) {
	act, ok := s.Sender.ReadMail(email)
	require.True(t, ok)

	require.Contains(t, act.TemplateArgs, "id")

	exp := mailcd.Request{
		TemplateLang: "en",
		TemplateName: conf.Template,
		TemplateArgs: map[string]interface{}{
			"link": conf.Link,
			"id":   act.TemplateArgs["id"].(string),
		},
		To: []mailcd.Address{{Email: email}},
	}

	require.Equal(t, exp, act)
}

func ReadValidConfirmationID(s *Suite, email string, conf auth.ConfirmationConfig) string {
	mail, _ := s.Sender.ReadMail(email)
	id := mail.TemplateArgs["id"].(string)
	return id
}

func ReadInvalidConfirmationID(s *Suite, email string, conf auth.ConfirmationConfig) string {
	return "invalid"
}

func ReadExpiredConfirmationID(s *Suite, email string, conf auth.ConfirmationConfig) string {
	delay := conf.TTL
	s.Timer.Set(s.Timer.Now().Add(delay).Add(time.Second))

	mail, _ := s.Sender.ReadMail(email)
	id := mail.TemplateArgs["id"].(string)
	return id
}

func CheckResponse(t *testing.T, bodyType interface{}, exp, act techo.Response) {
	require.Equal(t, exp.Code, act.Code)

	actBody := bodyType
	require.NoError(t, json.Unmarshal(act.Body, &actBody))

	require.Equal(t, exp.Body, actBody)
}
