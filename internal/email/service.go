package email

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
)

// EmailMessage represents an email to be sent.
type EmailMessage struct {
	To      string
	Subject string
	Body    string // HTML body
}

// Sender abstracts the email delivery mechanism.
type Sender interface {
	Send(ctx context.Context, msg EmailMessage) error
}

// LogSender logs emails to stdout instead of sending them (for dev/test).
type LogSender struct{}

// Send logs the email details without actually sending.
func (s *LogSender) Send(_ context.Context, msg EmailMessage) error {
	log.Printf("[EMAIL] To: %s | Subject: %s | Body length: %d chars", msg.To, msg.Subject, len(msg.Body))
	return nil
}

// SMTPSender sends emails via SMTP.
type SMTPSender struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Send delivers the email via SMTP.
func (s *SMTPSender) Send(_ context.Context, msg EmailMessage) error {
	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	auth := smtp.PlainAuth("", s.Username, s.Password, s.Host)

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n",
		s.From, msg.To, msg.Subject)
	body := []byte(headers + msg.Body)

	return smtp.SendMail(addr, auth, s.From, []string{msg.To}, body)
}
