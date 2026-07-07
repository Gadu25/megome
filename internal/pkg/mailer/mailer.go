package mailer

import "gopkg.in/gomail.v2"

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type Mailer struct {
	dialer *gomail.Dialer
	from   string
}

func New(cfg Config) *Mailer {
	return &Mailer{
		dialer: gomail.NewDialer(
			cfg.Host,
			cfg.Port,
			cfg.Username,
			cfg.Password,
		),
		from: cfg.From,
	}
}

type Email struct {
	To          []string
	Subject     string
	Body        string
	ContentType string
}

func (m *Mailer) Send(e Email) error {
	msg := gomail.NewMessage()

	msg.SetHeader("From", m.from)
	msg.SetHeader("To", e.To...)
	msg.SetHeader("Subject", e.Subject)

	contentType := e.ContentType
	if contentType == "" {
		contentType = "text/plain"
	}

	msg.SetBody(contentType, e.Body)

	return m.dialer.DialAndSend(msg)
}
