package model

import (
	"time"
)

// User represents a user model.
type User struct {
	ID        string    `gorm:"type:varchar(36);primary_key;" db:"id"`
	Email     string    `gorm:"type:varchar(256);unique;not null;" json:"email" db:"email"`
	Password  string    `gorm:"type:varchar(256);not null;" json:"password" db:"password"`
	Active    bool      `gorm:"default:false;" db:"active"`
	Created   time.Time `gorm:"type:timestamp;not null;" db:"created"`
	Updated   time.Time `gorm:"type:timestamp;not null;" db:"updated"`
	LastLogin time.Time `gorm:"type:timestamp;" db:"last_login"`
}

// Confirmation represents a confirmation record model.
type Confirmation struct {
	ID      string    `gorm:"type:varchar(32);primary_key;"`
	UserID  string    `gorm:"type:varchar(36);not null;"`
	User    *User     `gorm:"preload;"`
	Kind    int8      `gorm:"not null;"`
	Created time.Time `gorm:"type:timestamp;not null;"`
	Expires time.Time `gorm:"type:timestamp;not null;"`
}

// Token represents a result of a successfull login. Access token is a bearer
// token.
type Token struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
	Expires int64  `json:"expires"`
}

// Refresh represents refresh token record.
type Refresh struct {
	ID      string    `gorm:"type:varchar(32);primary_key;"`
	UserID  string    `gorm:"type:varchar(36);not null;"`
	User    *User     `gorm:"preload;"`
	Created time.Time `gorm:"type:timestamp;not null;"`
	Expires time.Time `gorm:"type:timestamp;not null;"`
}
