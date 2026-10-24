package mailer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
)

// Mailer sends plain-text email.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
	// Real reports whether messages really leave the server (false for the development logger).
	Real() bool
}

// SMTP sends through an SMTP server on port 587 (STARTTLS) or 25.
type SMTP struct {
	Host, Port, Username, Password, From string
}

func (m *SMTP) Real() bool { return true }

func (m *SMTP) Send(ctx context.Context, to, subject, body string) error {
	// Header injection guard: none of these may contain a line break.
	if strings.ContainsAny(to+subject+m.From, "\r\n") {
		return errors.New("invalid characters in email header")
	}
	toAddr, err := mail.ParseAddress(to)
	if err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}
	fromAddr, err := mail.ParseAddress(m.From)
	if err != nil {
		return fmt.Errorf("invalid SMTP_FROM: %w", err)
	}

	msg := strings.Join([]string{
		"From: " + m.From,
		"To: " + toAddr.Address,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if m.Username != "" {
		auth = smtp.PlainAuth("", m.Username, m.Password, m.Host)
	}

	done := make(chan error, 1)
	go func() {
		done <- smtp.SendMail(net.JoinHostPort(m.Host, m.Port), auth, fromAddr.Address, []string{toAddr.Address}, []byte(msg))
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Log is the development mailer: it prints the email in the server log and sends nothing.
type Log struct{}

func (Log) Real() bool { return false }

func (Log) Send(_ context.Context, to, subject, body string) error {
	log.Printf("[MAIL:DEV] (not sent) to=%s subject=%q\n%s", to, subject, body)
	return nil
}

// New returns an SMTP mailer when a host is configured, otherwise the development logger.
func New(host, port, username, password, from string) Mailer {
	if strings.TrimSpace(host) == "" {
		return Log{}
	}
	if port == "" {
		port = "587"
	}
	if from == "" {
		from = username
	}
	return &SMTP{Host: host, Port: port, Username: username, Password: password, From: from}
}
