package v1

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/vbogretsov/authd/app"
	"github.com/vbogretsov/authd/model"
)

// Config represents API V1 config.
type Config struct {
	Group             string
	SignUpURL         string
	SignInURL         string
	RefreshURL        string
	ConfirmUserURL    string
	ResetPasswordURL  string
	UpdatePasswordURL string
}

// DefaultConf represents default API V1 configuration.
var DefaultConf = Config{
	Group:             "/v1/auth",
	SignInURL:         "/signin",
	RefreshURL:        "/refresh",
	SignUpURL:         "/signup",
	ConfirmUserURL:    "/signup/confirm/:id",
	ResetPasswordURL:  "/pwreset",
	UpdatePasswordURL: "/pwreset/confirm/:id",
}

// Conf represents API V1 configuration.
var Conf = DefaultConf

// StrConfig represents strings configuration.
type StrConfig struct {
	SignUp         string
	ConfirmUser    string
	ResetPassword  string
	UpdatePassword string
}

// DefaultStrConf represents default strings configuration.
var DefaultStrConf = StrConfig{
	SignUp:         "confirmation email has been sent",
	ConfirmUser:    "user has been activated",
	ResetPassword:  "password reset email has been sent",
	UpdatePassword: "password has been updated",
}

// StrConf represents strings configuration.
var StrConf = DefaultStrConf

// Refresh represents data required for token refresh.
type Refresh struct {
	Refresh string `json:"refresh"`
}

// Password represents request data for PwReset/Confirm.
type Password struct {
	Password string `json:"password"`
}

// Email represents request data for PwReset.
type Email struct {
	Email string `json:"email"`
}

// Credentials represents credentials for SignUp/SignIn.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Message represents API message response.
type Message struct {
	Message string `json:"message"`
}

// V1 represents authd API V1.
type V1 struct {
	app *app.App
}

// New creates new API V1.
func New(app *app.App) *V1 {
	return &V1{app: app}
}

// SignUp creates a new account if email is not in use and password matches
// security criteria. The activation email will be sent to the email provided.
func (v1 *V1) SignUp(c echo.Context) error {
	user := new(model.User)
	if err := c.Bind(user); err != nil {
		return err
	}

	if err := v1.app.SignUp(user); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Message{Message: StrConf.SignUp})
}

// ConfirmUser activates user account.
func (v1 *V1) ConfirmUser(c echo.Context) error {
	conID := c.Param("id")
	if err := v1.app.ConfirmUser(conID); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Message{Message: StrConf.ConfirmUser})
}

// ResetPassword sends the password reset email to the email provided.
func (v1 *V1) ResetPassword(c echo.Context) error {
	em := Email{}
	if err := c.Bind(&em); err != nil {
		return err
	}

	if err := v1.app.ResetPassword(em.Email); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Message{Message: StrConf.ResetPassword})
}

// UpdatePassword updates user password.
func (v1 *V1) UpdatePassword(c echo.Context) error {
	conID := c.Param("id")

	newpw := Password{}
	if err := c.Bind(&newpw); err != nil {
		return err
	}

	if err := v1.app.UpdatePassword(conID, newpw.Password); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Message{Message: StrConf.UpdatePassword})
}

// SignIn authenticates the credentials provided. An access token and refresh
// token will be generated.
func (v1 *V1) SignIn(c echo.Context) error {
	creds := Credentials{}
	if err := c.Bind(&creds); err != nil {
		return err
	}

	token, err := v1.app.SignIn(creds.Email, creds.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, token)
}

// Refresh refreshes an access token. A new access token and refresh token will
// be generated.
func (v1 *V1) Refresh(c echo.Context) error {
	refresh := Refresh{}
	if err := c.Bind(&refresh); err != nil {
		return err
	}

	token, err := v1.app.Refresh(refresh.Refresh)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, token)
}

// Include confirgures authd API V1 routes.
func Include(app *app.App, e *echo.Echo) {
	v1 := New(app)
	g := e.Group(Conf.Group)
	g.POST(Conf.SignUpURL, v1.SignUp)
	g.POST(Conf.ConfirmUserURL, v1.ConfirmUser)
	g.POST(Conf.ResetPasswordURL, v1.ResetPassword)
	g.POST(Conf.UpdatePasswordURL, v1.UpdatePassword)
	g.POST(Conf.SignInURL, v1.SignIn)
	g.POST(Conf.RefreshURL, v1.Refresh)
}
