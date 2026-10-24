package service

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"
)

// Throwaway-mailbox providers are refused: a verification code sent there proves nothing.
var disposableDomains = map[string]bool{
	"mailinator.com": true, "10minutemail.com": true, "guerrillamail.com": true,
	"tempmail.com": true, "temp-mail.org": true, "yopmail.com": true,
	"trashmail.com": true, "throwawaymail.com": true, "getnada.com": true,
	"sharklasers.com": true, "dispostable.com": true, "maildrop.cc": true,
}

// DNS lookups are variables so tests can replace them.
var (
	lookupMX = func(ctx context.Context, domain string) ([]*net.MX, error) {
		return net.DefaultResolver.LookupMX(ctx, domain)
	}
	lookupHost = func(ctx context.Context, domain string) ([]string, error) {
		return net.DefaultResolver.LookupHost(ctx, domain)
	}
)

func emailDomain(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return ""
	}
	return strings.ToLower(email[at+1:])
}

func isDisposableEmail(email string) bool {
	return disposableDomains[emailDomain(email)]
}

func isDNSNotFound(err error) bool {
	var dnsErr *net.DNSError
	return errors.As(err, &dnsErr) && dnsErr.IsNotFound
}

// emailDomainCanReceiveMail reports false only when DNS clearly says the domain cannot
// receive mail (it does not exist, or has no MX and no address record). If DNS is unreachable
// we cannot tell, so we allow the address and let the verification code decide.
func emailDomainCanReceiveMail(ctx context.Context, email string) bool {
	domain := emailDomain(email)
	if domain == "" || !strings.Contains(domain, ".") {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	mx, err := lookupMX(ctx, domain)
	if err == nil && len(mx) > 0 {
		return true
	}
	if err != nil && !isDNSNotFound(err) {
		return true // DNS trouble, not proof of a bad address
	}

	// No MX: mail can still be delivered to the domain's own address record.
	addrs, err := lookupHost(ctx, domain)
	if err == nil && len(addrs) > 0 {
		return true
	}
	if err != nil && !isDNSNotFound(err) {
		return true
	}
	return false
}
