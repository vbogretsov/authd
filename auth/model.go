package auth

// Token represents a result of a successfull login. Access token is a JWT token.
type Token struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
	Expires int64  `json:"expires"`
}

// Credentials represents a credentials arguments for SignUp/SignIn operations.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Email represents an email argument.
type Email struct {
	Email string `json:"email"`
}

// Password represents a password argument.
type Password struct {
	Password string `json:"password"`
}

// Refresh represents data required for token refresh.
type Refresh struct {
	Refresh string `json:"refresh"`
}
