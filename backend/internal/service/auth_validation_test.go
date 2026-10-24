package service

import (
	"context"
	"errors"
	"net"
	"strings"
	"testing"
)

func TestNormalizeBDPhone(t *testing.T) {
	good := map[string]string{
		"01712345678":      "01712345678",
		"+8801712345678":   "01712345678",
		"8801712345678":    "01712345678",
		"017 1234-5678":    "01712345678",
		" 01912345678 ":    "01912345678",
		"+880 1312 345678": "01312345678",
	}
	for in, want := range good {
		if got, ok := normalizeBDPhone(in); !ok || got != want {
			t.Errorf("normalizeBDPhone(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	for _, bad := range []string{"", "12345", "01212345678", "0171234567", "017123456789", "abcdefghijk", "+4412345678901", "02712345678"} {
		if _, ok := normalizeBDPhone(bad); ok {
			t.Errorf("normalizeBDPhone(%q) should be rejected", bad)
		}
	}
}

func TestPasswordProblem(t *testing.T) {
	for _, ok := range []string{"Demo@12345", "abcdefg1", "1234567a", "pass word 9"} {
		if msg := passwordProblem(ok); msg != "" {
			t.Errorf("%q should be accepted, got %q", ok, msg)
		}
	}
	bad := map[string]string{
		"short1":                 "at least 8",
		"onlyletters":            "letter and one number",
		"12345678":               "letter and one number",
		strings.Repeat("a1", 40): "at most 72",
	}
	for p, want := range bad {
		if msg := passwordProblem(p); !strings.Contains(msg, want) {
			t.Errorf("%q: got %q, want it to mention %q", p, msg, want)
		}
	}
}

func TestValidateRegister(t *testing.T) {
	in := RegisterInput{FullName: "  Rahim   Uddin ", Email: " Rahim@Example.COM ", Phone: "+8801712345678", Password: "Secret123"}
	if err := validateRegister(&in); err != nil {
		t.Fatalf("valid input rejected: %v", err)
	}
	if in.FullName != "Rahim Uddin" || in.Email != "rahim@example.com" || in.Phone != "01712345678" {
		t.Errorf("input not normalised: %+v", in)
	}

	bad := []RegisterInput{
		{FullName: "A", Email: "a@example.com", Phone: "01712345678", Password: "Secret123"},
		{FullName: "Rahim", Email: "not-an-email", Phone: "01712345678", Password: "Secret123"},
		{FullName: "Rahim", Email: "Rahim <r@example.com>", Phone: "01712345678", Password: "Secret123"},
		{FullName: "Rahim", Email: "r@mailinator.com", Phone: "01712345678", Password: "Secret123"},
		{FullName: "Rahim", Email: "r@example.com", Phone: "12345", Password: "Secret123"},
		{FullName: "Rahim", Email: "r@example.com", Phone: "01712345678", Password: "short"},
	}
	for i, b := range bad {
		b := b
		var ve *ValidationError
		if err := validateRegister(&b); !errors.As(err, &ve) {
			t.Errorf("case %d should fail validation, got %v", i, err)
		}
	}
}

func TestEmailDomainCheck(t *testing.T) {
	origMX, origHost := lookupMX, lookupHost
	defer func() { lookupMX, lookupHost = origMX, origHost }()

	notFound := &net.DNSError{Err: "no such host", IsNotFound: true}
	timeout := &net.DNSError{Err: "i/o timeout", IsTimeout: true}

	cases := []struct {
		name  string
		mx    func(context.Context, string) ([]*net.MX, error)
		host  func(context.Context, string) ([]string, error)
		email string
		want  bool
	}{
		{"has MX", func(context.Context, string) ([]*net.MX, error) { return []*net.MX{{Host: "mx.example.com."}}, nil },
			func(context.Context, string) ([]string, error) { return nil, notFound }, "a@example.com", true},
		{"no MX but an address record", func(context.Context, string) ([]*net.MX, error) { return nil, notFound },
			func(context.Context, string) ([]string, error) { return []string{"203.0.113.5"}, nil }, "a@example.com", true},
		{"domain does not exist", func(context.Context, string) ([]*net.MX, error) { return nil, notFound },
			func(context.Context, string) ([]string, error) { return nil, notFound }, "a@nosuchdomain.example", false},
		{"DNS unreachable: cannot tell, so allow", func(context.Context, string) ([]*net.MX, error) { return nil, timeout },
			func(context.Context, string) ([]string, error) { return nil, timeout }, "a@example.com", true},
		{"no dot in domain", nil, nil, "a@localhost", false},
		{"no domain", nil, nil, "nodomain", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mx != nil {
				lookupMX, lookupHost = tc.mx, tc.host
			}
			if got := emailDomainCanReceiveMail(context.Background(), tc.email); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestVerificationCodeHelpers(t *testing.T) {
	for i := 0; i < 50; i++ {
		code, err := newVerificationCode()
		if err != nil || !sixDigits.MatchString(code) {
			t.Fatalf("bad code %q (%v)", code, err)
		}
	}
	a := hashVerificationCode("user-1", "123456")
	if a != hashVerificationCode("user-1", "123456") {
		t.Error("hash must be stable")
	}
	if a == hashVerificationCode("user-2", "123456") || a == hashVerificationCode("user-1", "123457") {
		t.Error("hash must depend on user and code")
	}
	if strings.Contains(a, "123456") {
		t.Error("hash must not contain the code")
	}
}
