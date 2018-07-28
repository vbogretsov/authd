package api

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// New creates new echo instance with all the middleware configured.
func New(debug bool) *echo.Echo {
	e := echo.New()
	e.HTTPErrorHandler = ErrorHandler(debug)
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.RemoveTrailingSlash())
	return e
}
