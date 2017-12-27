package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vbogretsov/authd/app"
)

type V1 struct {
	cfg *app.Cfg
}

func NewV1(cfg *app.Cfg) *V1 {
	return &V1{cfg: cfg}
}

func (v *V1) CreateAccount(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}

func (v *V1) ConfirmAccount(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}

func (v *V1) RequestPasswordReset(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}

func (v *V1) ResetPassword(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}

func (v *V1) Login(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}

func (v *V1) Logout(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{})
}

func (v1 *V1) Configure(r gin.IRouter) {
	g := r.Group("v1")
	g.POST("/requests/accounts", v1.CreateAccount)
	g.GET("/requests/accounts/:id", v1.ConfirmAccount)
	g.POST("/requests/credentials", v1.RequestPasswordReset)
	g.POST("/requests/credentials/:id", v1.CreateAccount)
	g.POST("/requests/tokens", v1.Login)
	g.DELETE("/requests/tokens/:id", v1.Logout)
}
