package auth

import "testing"

func TestGenerateOTPLength(t *testing.T) {
	for i := 0; i < 100; i++ {
		otp, err := GenerateOTP()
		if err != nil {
			t.Fatal(err)
		}
		if len(otp) != 6 {
			t.Errorf("expected 6-digit code, got %q", otp)
		}
	}
}

func TestGenerateOTPNumeric(t *testing.T) {
	for i := 0; i < 100; i++ {
		otp, err := GenerateOTP()
		if err != nil {
			t.Fatal(err)
		}
		for _, c := range otp {
			if c < '0' || c > '9' {
				t.Errorf("non-digit character %q in %q", c, otp)
			}
		}
	}
}
