package auth

import (
	"time"

	"gopkg.in/hlandau/passlib.v1/hash/bcrypt"
)

// Config represens application configuration.
type Config struct {
	Token    TokenConfig
	Password PasswordConfig
	SignUp   ConfirmConfig
	ResetPw  ConfirmConfig
}

// TokenConfig contains information required to sign and refresh tokens.
type TokenConfig struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	SecretKey  string
}

// PasswordConfig contains information required to validate, hash and verify
// a password.
type PasswordConfig struct {
	MinLen   int
	HashCost int
}

// ConfirmConfig contains information required to generate a confirmation.
type ConfirmConfig struct {
	TTL      time.Duration
	Kind     int8
	Link     string
	Template string
}

// DefaultConfig represents default configuration.
var DefaultConfig = Config{
	Token: TokenConfig{
		AccessTTL:  time.Duration(time.Hour * 1),
		RefreshTTL: time.Duration(time.Hour * 24 * 5),
	},
	Password: PasswordConfig{
		MinLen:   6,
		HashCost: bcrypt.RecommendedCost,
	},
	SignUp: ConfirmConfig{
		TTL:      time.Duration(time.Hour * 24 * 5),
		Link:     "signup",
		Template: "signup.msg",
	},
	ResetPw: ConfirmConfig{
		TTL:      time.Duration(time.Hour * 12),
		Link:     "resetpw",
		Template: "resetpw.msg",
	},
}
