package test

import (
	"testing"
)

func TestSignUp(t *testing.T) {
	for _, fx := range signUpFixtures {
		s := setup(t)
		t.Run(fx.Name, func(t *testing.T) {
			s.testSignUp(t, fx)
		})
		s.cleanup(t)
	}
}

func TestConfirmUser(t *testing.T) {
	for _, fx := range confirmUserFixtures {
		s := setup(t)
		t.Run(fx.Name, func(t *testing.T) {
			s.signUp(t, fx.Credentials)
			s.testConfirmUser(t, fx)
		})
		s.cleanup(t)
	}
}

func TestSignIn(t *testing.T) {
	for _, fx := range signInFixtures {
		s := setup(t)
		t.Run(fx.Name, func(t *testing.T) {
			if fx.CreateUser {
				s.signUp(t, *fx.Credentials)
				if fx.ConfirmUser {
					s.confirmUser(t)
				}
			}
			s.testSignIn(t, fx)
		})
		s.cleanup(t)
	}
}

func TestResetPassword(t *testing.T) {
	for _, fx := range resetPasswordFixtures {
		s := setup(t)
		t.Run(fx.Name, func(t *testing.T) {
			s.signUp(t, fx.Credentials)
			s.confirmUser(t)
			s.testResetPassword(t, fx)
		})
		s.cleanup(t)
	}
}

func TestUpdatePassword(t *testing.T) {
	for _, fx := range updatePasswordFixtures {
		s := setup(t)
		t.Run(fx.Name, func(t *testing.T) {
			s.signUp(t, fx.Credentials)
			s.confirmUser(t)
			s.resetPassword(t, fx.Credentials.Email)

		})
		s.cleanup(t)
	}
}
