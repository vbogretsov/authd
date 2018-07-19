package v1

// Config represents API V1 config.
type Config struct {
	Group             string
	SignUpURL         string
	SignInURL         string
	RefreshURL        string
	ConfirmUserURL    string
	ResetPasswordURL  string
	UpdatePasswordURL string
}

// DefaultConf represents default API V1 configuration.
var DefaultConf = Config{
	Group:             "/v1/auth",
	SignInURL:         "/signin",
	RefreshURL:        "/refresh",
	SignUpURL:         "/signup",
	ConfirmUserURL:    "/signup/:id",
	ResetPasswordURL:  "/pwreset",
	UpdatePasswordURL: "/pwreset/:id",
}

// Conf represents API V1 configuration.
var Conf = DefaultConf

// StrConfig represents strings configuration.
type StrConfig struct {
	SignUp         string
	ConfirmUser    string
	ResetPassword  string
	UpdatePassword string
}

// DefaultStrConf represents default strings configuration.
var DefaultStrConf = StrConfig{
	SignUp:         "ok-signup",
	ConfirmUser:    "ok-confirm-user",
	ResetPassword:  "ok-resetpw",
	UpdatePassword: "ok-update-password",
}

// StrConf represents strings configuration.
var StrConf = DefaultStrConf
