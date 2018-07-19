package v1

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/vbogretsov/authd/auth"
)

// Message represents API message response.
type Message struct {
	Message string `json:"message"`
}

// V1 represents authd API V1.
type V1 struct {
	app *auth.Auth
}

// New creates new API V1.
func New(app *auth.Auth) *V1 {
	return &V1{app: app}
}

// SignUp creates a new account if email is not in use and password matches
// security criteria. The activation email will be sent to the email provided.
func (v1 *V1) SignUp(c echo.Context) error {
	cred := new(auth.Credentials)
	if err := c.Bind(cred); err != nil {
		return err
	}

	if err := v1.app.SignUp(cred); err != nil {
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
	email := new(auth.Email)
	if err := c.Bind(email); err != nil {
		return err
	}

	if err := v1.app.ResetPassword(email); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Message{Message: StrConf.ResetPassword})
}

// UpdatePassword updates user password.
func (v1 *V1) UpdatePassword(c echo.Context) error {
	cid := c.Param("id")

	password := new(auth.Password)
	if err := c.Bind(password); err != nil {
		return err
	}

	if err := v1.app.UpdatePassword(cid, password); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, Message{Message: StrConf.UpdatePassword})
}

// SignIn authenticates the credentials provided. An access token and refresh
// token will be generated.
func (v1 *V1) SignIn(c echo.Context) error {
	cred := new(auth.Credentials)
	if err := c.Bind(cred); err != nil {
		return err
	}

	token, err := v1.app.SignIn(cred)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, token)
}

// Refresh refreshes an access token. A new access token and refresh token will
// be generated.
func (v1 *V1) Refresh(c echo.Context) error {
	refresh := new(auth.Refresh)
	if err := c.Bind(refresh); err != nil {
		return err
	}

	token, err := v1.app.Refresh(refresh.Refresh)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, token)
}

// Include confirgures authd API V1 routes.
func Include(auth *auth.Auth, e *echo.Echo) {
	v1 := New(auth)
	g := e.Group(Conf.Group)
	g.POST(Conf.SignUpURL, v1.SignUp)
	g.POST(Conf.ConfirmUserURL, v1.ConfirmUser)
	g.POST(Conf.ResetPasswordURL, v1.ResetPassword)
	g.POST(Conf.UpdatePasswordURL, v1.UpdatePassword)
	g.POST(Conf.SignInURL, v1.SignIn)
	g.POST(Conf.RefreshURL, v1.Refresh)
}
