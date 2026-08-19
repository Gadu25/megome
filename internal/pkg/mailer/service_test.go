package mailer

import (
	"strings"
	"testing"
)

func TestVerifyEmailTemplateRenders(t *testing.T) {
	r, err := NewRenderer("templates/*.html")
	if err != nil {
		t.Fatal(err)
	}

	body, err := r.Render("verify_email.html", VerifyEmailData{
		OTP:              "123456",
		ExpiresInMinutes: 10,
		Year:             2026,
	})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(body, "123456") {
		t.Error("rendered body missing the OTP")
	}
	if !strings.Contains(body, "10 minutes") {
		t.Error("rendered body missing the expiry note")
	}
}
