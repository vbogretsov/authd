package app

import (
	"gopkg.in/go-playground/validator.v9"
)

type Cfg struct {
	Db       string `validate:"required"`
	Mode     string `validate:"required"`
	Port     string `validate:"required"`
	LogLevel string `validate:"required"`
}

func (c *Cfg) Validate() error {
	return validator.New().Struct(c)
}
