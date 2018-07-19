package auth

import (
	"time"

	"github.com/robbert229/jwt"

	"gopkg.in/hlandau/passlib.v1/hash/bcrypt"
)

// Config represens application configuration.
type Config struct {
	Token    TokenConfig
	Password PasswordConfig
	SignUp   ConfirmationConfig
	ResetPw  ConfirmationConfig
}

// TokenConfig contains information required to sign and refresh tokens.
type TokenConfig struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Algorithm  jwt.Algorithm
}

// PasswordConfig contains information required to validate, hash and verify
// a password.
type PasswordConfig struct {
	MinLen   int
	HashCost int
}

// ConfirmationConfig contains information required to generate a confirmation.
type ConfirmationConfig struct {
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
	SignUp: ConfirmationConfig{
		TTL:      time.Duration(time.Hour * 24 * 5),
		Link:     "signup",
		Template: "signup.msg",
	},
	ResetPw: ConfirmationConfig{
		TTL:      time.Duration(time.Hour * 12),
		Link:     "resetpw",
		Template: "resetpw.msg",
	},
}
