type ResetPasswordEmailData struct {
	ResetURL string
}

func (h *Handler) sendResetEmail(to string, resetURL string) error {
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/reset_password.html",
	)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	err = tmpl.Execute(&body, ResetPasswordEmailData{
		ResetURL: resetURL,
	})
	if err != nil {
		return err
	}

	return h.mailer.Send(mailer.Email{
		To:      []string{to},
		Subject: "Reset Your Password",
		Body:    body.String(),
	})
}
