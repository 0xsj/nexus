package email

import (
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
)

// SMTPConfig holds SMTP connection parameters.
type SMTPConfig struct {
	Host     string // SMTP server host (e.g. "localhost", "mailpit")
	Port     int    // SMTP server port (e.g. 1025 for Mailpit, 587 for production)
	From     string // Sender address (e.g. "noreply@nexus.local")
	Username string // Optional — empty for Mailpit
	Password string // Optional — empty for Mailpit
}

// SMTPSender sends emails via SMTP.
type SMTPSender struct {
	config SMTPConfig
}

// NewSMTPSender creates a new SMTPSender.
func NewSMTPSender(cfg SMTPConfig) *SMTPSender {
	return &SMTPSender{config: cfg}
}

// Send sends an email with the given recipient, subject, and HTML body.
func (s *SMTPSender) Send(to, subject, body string) error {
	addr := net.JoinHostPort(s.config.Host, strconv.Itoa(s.config.Port))

	headers := []string{
		"From: " + s.config.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}
	msg := []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + body)

	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	if err := smtp.SendMail(addr, auth, s.config.From, []string{to}, msg); err != nil {
		return fmt.Errorf("sending email to %s: %w", to, err)
	}

	return nil
}
