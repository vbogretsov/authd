package api

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/vbogretsov/go-validation"
	"github.com/vbogretsov/go-validation/json"

	"github.com/vbogretsov/authd/auth"
)

// Error represents API error.
type Error struct {
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

// ErrorHandler provides centralized error handling.
func ErrorHandler(debug bool) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		switch err.(type) {
		case auth.ArgumentError:
			c.JSON(http.StatusBadRequest, Error{
				Message: "validation errors",
				Errors: json.New(
					err.(auth.ArgumentError).Source.(validation.Errors),
					json.DefaultFormatter,
					json.DefaultJoiner),
			})
		case auth.ExpiredError:
			c.JSON(http.StatusRequestTimeout, Error{
				Message: err.Error(),
			})
		case auth.NotFoundError:
			c.JSON(http.StatusNotFound, Error{
				Message: err.Error(),
			})
		case auth.UnauthorizedError:
			c.JSON(http.StatusUnauthorized, Error{
				Message: err.Error(),
			})
		case *echo.HTTPError:
			httpError := err.(*echo.HTTPError)
			c.JSON(httpError.Code, Error{
				Message: fmt.Sprintf("%v", httpError),
			})
		default:
			res := Error{Message: "Internal Server Error"}
			if debug {
				res.Errors = []string{err.Error()}
			}
			c.JSON(http.StatusInternalServerError, res)
		}
	}
}
