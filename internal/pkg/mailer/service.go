package mailer

import "time"

type Service struct {
	smtp     *Mailer
	renderer *Renderer
}

func NewService(m *Mailer, r *Renderer) *Service {
	return &Service{
		smtp:     m,
		renderer: r,
	}
}

type ResetPasswordEmailData struct {
	ResetURL string
	Year int
}

func (s *Service) SendResetPassword(to string, resetURL string) error {
	body, err := s.renderer.Render("reset_password.html", ResetPasswordEmailData{
		ResetURL: resetURL,
		Year: time.Now().Year(),
	})
	if err != nil {
		return err
	}

	return s.smtp.Send(Email{
		To:          []string{to},
		Subject:     "Reset Your Password",
		Body:        body,
		ContentType: "text/html",
	})
}

type VerifyEmailData struct {
	OTP              string
	ExpiresInMinutes int
	Year             int
}

func (s *Service) SendVerifyEmail(to string, otp string) error {
	body, err := s.renderer.Render("verify_email.html", VerifyEmailData{
		OTP:              otp,
		ExpiresInMinutes: 10,
		Year:             time.Now().Year(),
	})
	if err != nil {
		return err
	}

	return s.smtp.Send(Email{
		To:          []string{to},
		Subject:     "Verify your email",
		Body:        body,
		ContentType: "text/html",
	})
}
