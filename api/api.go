package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/vbogretsov/go-validation"
	"github.com/vbogretsov/go-validation/jsonerr"

	"github.com/vbogretsov/authd/auth"
)

// Error represents API error.
type Error struct {
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

// Error returns string representation of an API error.
func (e Error) Error() string {
	return e.Message
}

// ErrorHandler provides centralized error handling.
func ErrorHandler(err error, c echo.Context) {
	var s int
	var e Error

	// TODO(vbogretsov): add debug mode.

	switch err.(type) {
	case auth.ArgumentError:
		s = http.StatusBadRequest
		e.Message = "validation errors"
		e.Errors = jsonerr.Errors(err.(auth.ArgumentError).Source.(validation.Errors))
	case auth.ExpiredError:
		s = http.StatusRequestTimeout
		e.Message = err.Error()
	case auth.NotFoundError:
		s = http.StatusNotFound
		e.Message = err.Error()
	case auth.UnauthorizedError:
		s = http.StatusUnauthorized
		e.Message = err.Error()
	case *echo.HTTPError:
		s = err.(*echo.HTTPError).Code
		e.Message = fmt.Sprintf("%v", err.(*echo.HTTPError).Message)
	default:
		s = http.StatusInternalServerError
		e.Message = "Internal Server Error"
	}

	c.JSON(s, e)
}

// New creates new echo instance with all the middleware configured.
func New() *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.RemoveTrailingSlash())
	return e
}
