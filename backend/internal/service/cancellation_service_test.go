package service

import (
	"testing"
	"time"
)

func TestComputeRefund(t *testing.T) {
	now := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	hours := func(h float64) time.Time { return now.Add(time.Duration(h * float64(time.Hour))) }

	cases := []struct {
		name    string
		depart  time.Time
		percent int
		amount  float64
		ok      bool
	}{
		{"3 days before: 90%", hours(72), 90, 675, true},
		{"exactly 24h before: 90%", hours(24), 90, 675, true},
		{"23h before: 50%", hours(23), 50, 375, true},
		{"6h before: 50%", hours(6), 50, 375, true},
		{"5h before: no refund but allowed", hours(5), 0, 0, true},
		{"just over 2h before: no refund but allowed", hours(2.1), 0, 0, true},
		{"exactly 2h before: closed", hours(2), 0, 0, false},
		{"1h before: closed", hours(1), 0, 0, false},
		{"already departed: closed", hours(-3), 0, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			percent, amount, reason, ok := computeRefund(now, tc.depart, 750)
			if ok != tc.ok || percent != tc.percent || amount != tc.amount {
				t.Fatalf("got (%d, %v, ok=%v), want (%d, %v, ok=%v)", percent, amount, ok, tc.percent, tc.amount, tc.ok)
			}
			if !ok && reason == "" {
				t.Error("a refused cancellation must explain why")
			}
		})
	}
}

func TestComputeRefundRounding(t *testing.T) {
	now := time.Now()
	// 90% of 333.33 is 299.997, which rounds to 300.00 at two decimals.
	_, amount, _, ok := computeRefund(now, now.Add(48*time.Hour), 333.33)
	if !ok || amount != 300 {
		t.Fatalf("got %v, want 300 (90%% of 333.33, rounded to two decimals)", amount)
	}
}
