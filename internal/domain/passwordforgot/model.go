package passwordforgot

import "time"

type PasswordForgotStore interface {
	SavePasswordResetToken(userId int, token string, exp time.Time) error
	ChangePassword(token string, newPass string) error
}
