package mailer

import (
	"context"
	"strings"
	"testing"
)

func TestNewPicksMailer(t *testing.T) {
	if New("", "", "", "", "").Real() {
		t.Error("no SMTP host should give the development logger")
	}
	m := New("smtp.example.com", "", "user", "pw", "")
	smtpMailer, ok := m.(*SMTP)
	if !ok || !m.Real() {
		t.Fatalf("SMTP host should give a real mailer, got %T", m)
	}
	if smtpMailer.Port != "587" || smtpMailer.From != "user" {
		t.Errorf("defaults not applied: %+v", smtpMailer)
	}
}

func TestSMTPRejectsHeaderInjection(t *testing.T) {
	m := &SMTP{Host: "smtp.example.com", Port: "587", From: "TicketKaton <noreply@example.com>"}
	err := m.Send(context.Background(), "a@example.com", "Hello\r\nBcc: evil@example.com", "body")
	if err == nil || !strings.Contains(err.Error(), "invalid characters") {
		t.Fatalf("expected header injection to be rejected, got %v", err)
	}
	if err := m.Send(context.Background(), "a@example.com\r\nBcc: x@example.com", "Hi", "body"); err == nil {
		t.Fatal("recipient with a line break must be rejected")
	}
}

func TestLogMailerSendsNothing(t *testing.T) {
	if err := (Log{}).Send(context.Background(), "a@example.com", "Hi", "code 123456"); err != nil {
		t.Fatal(err)
	}
}
