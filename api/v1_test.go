package api_test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/vbogretsov/dicontainer"
	"github.com/vbogretsov/go-fixtures"

	"github.com/vbogretsov/authd/api"
	"github.com/vbogretsov/authd/app"
)

func newCfg() *app.Cfg {
	return &app.Cfg{
		Db:   "postgresql://localhost:5432/autd",
		Port: ":8000",
	}
}

func newRouter(c *app.Cfg) http.Handler {
	r := gin.New()
	apiv1 := api.NewV1(c)
	apiv1.Configure(r)
	return r
}

type UnauthorizedCase struct {
	state *dicontainer.Container
}

func NewUnauthorizedCase() *UnauthorizedCase {
	fx := dicontainer.New()
	fx.Register(t)
	fx.Register(newCfg)
	fx.Register(fixtures.NewHttpServer)
	fx.Register(fixtures.NewHttpClient)
	return &UnauthorizedCase{state: fx}
}

func (c *UnauthorizedCase) Close() {
	c.state.Close()
}

func (c *UnauthorizedCase) CreateAccountIfEmailInvalid(t *testing.T) {
	c.state.Call(func(client *fixtures.HttpClient) {
		res, err := client.Post("/v1/requests/accounts", nil)
		if err != nil {
			t.Error(err)
		}

		if r.StatusCode != http.StatusNotImplemented {
			t.Errorf("unexpected status code: %v", r.StatusCode)
		}
	})
}

func (c *UnauthorizedCase) CreateAccountIfEmailMissing(t *testing.T) {

}

func (c *UnauthorizedCase) CreateAccountIfPasswordInvalid(t *testing.T) {

}

func (c *UnauthorizedCase) CreateAccountIfPasswordMissing(t *testing.T) {

}

func (c *UnauthorizedCase) CreateAccountSuccess(t *testing.T) {

}

func TestV1(t *testing.T) {

	defer fx.Close()

	t.Run("TestCreateAccountIfEmailInvalid", c.CreateAccountIfEmailInvalid)
}
