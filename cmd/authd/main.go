package main

import (
	"github.com/gin-gonic/gin"

	"github.com/vbogretsov/authd/api"
	"github.com/vbogretsov/authd/app"
)

func main() {
	cfg := app.Cfg{
		Db:   "postgresql://localhost:5432/autd",
		Port: ":8000",
	}

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	apiv1 := api.NewV1(&cfg)
	apiv1.Configure(router)

	router.Run(cfg.Port)
}
