package emailverification

import "time"

const (
	OTPExpiry       = 10 * time.Minute
	ResendCooldown  = 60 * time.Second
	MaxOTPAttempts  = 5
)

type EmailVerificationStore interface {
	SaveOTP(userId int, email string, otpHash string, exp time.Time) error
	LastOTPSentAt(userId int) (time.Time, error)
	DeleteOTPs(userId int) error
	VerifyOTP(email string, otpHash string) (int, error)
	MarkVerified(userId int) error
}
