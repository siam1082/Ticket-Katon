package http

import (
	"sync"
	"time"
)

type loginAttempt struct {
	fails        int
	first        time.Time
	blockedUntil time.Time
}

// loginLimiter blocks a key (email + IP) after too many failed logins.
// It is in-memory, which is enough for a single-server demo.
type loginLimiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	m      map[string]*loginAttempt
}

func newLoginLimiter(max int, window time.Duration) *loginLimiter {
	return &loginLimiter{max: max, window: window, m: make(map[string]*loginAttempt)}
}

// blocked reports whether the key is blocked, and for how much longer.
func (l *loginLimiter) blocked(key string) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.m[key]
	if !ok {
		return 0, false
	}
	if remaining := time.Until(a.blockedUntil); remaining > 0 {
		return remaining, true
	}
	return 0, false
}

func (l *loginLimiter) fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// Drop stale entries so the map cannot grow without bound.
	if len(l.m) > 1000 {
		for k, v := range l.m {
			if now.Sub(v.first) > l.window && now.After(v.blockedUntil) {
				delete(l.m, k)
			}
		}
	}

	a, ok := l.m[key]
	if !ok || (now.Sub(a.first) > l.window && now.After(a.blockedUntil)) {
		a = &loginAttempt{first: now}
		l.m[key] = a
	}

	a.fails++
	if a.fails >= l.max {
		a.blockedUntil = now.Add(l.window)
	}
}

func (l *loginLimiter) reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.m, key)
}
