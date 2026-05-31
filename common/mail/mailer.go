package mail

import "github.com/go-gomail/gomail"

type Mailer struct {
	cfg EmailConfig
}

func NewMailer(cfg EmailConfig) *Mailer {
	return &Mailer{cfg: cfg}
}

func (m *Mailer) Send(subject, body string, to []string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.cfg.From)
	msg.SetHeader("To", to...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	d := gomail.NewDialer(
		m.cfg.SMTPHost,
		m.cfg.SMTPPort,
		m.cfg.From,
		m.cfg.Password,
	)

	return d.DialAndSend(msg)
}
