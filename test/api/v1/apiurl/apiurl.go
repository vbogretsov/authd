package apiurl

import (
	"fmt"
	"strings"

	apiv1 "github.com/vbogretsov/authd/api/v1"
)

func SignUp() string {
	return fmt.Sprintf("%s%s", apiv1.Conf.Group, apiv1.Conf.SignUpURL)
}

func ConfirmUser(id string) string {
	url := fmt.Sprintf("%s%s", apiv1.Conf.Group, apiv1.Conf.ConfirmUserURL)
	return strings.Replace(url, ":id", id, 1)
}

func SignIn() string {
	return fmt.Sprintf("%s%s", apiv1.Conf.Group, apiv1.Conf.SignInURL)
}

func ResetPassword() string {
	return fmt.Sprintf("%s%s", apiv1.Conf.Group, apiv1.Conf.ResetPasswordURL)
}

func UpdatePassword(id string) string {
	url := fmt.Sprintf("%s%s", apiv1.Conf.Group, apiv1.Conf.UpdatePasswordURL)
	return strings.Replace(url, ":id", id, 1)
}
