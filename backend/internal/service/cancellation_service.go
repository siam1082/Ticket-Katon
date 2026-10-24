package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"ticket-katon-backend/internal/domain"
)

// Demo cancellation policy. Change these to change the rules.
const (
	cancelCutoff        = 2 * time.Hour  // closer than this to departure: no cancellation
	noRefundWindow      = 6 * time.Hour  // 2h to 6h before: cancel allowed, no refund
	partialRefundWindow = 24 * time.Hour // 6h to 24h before: 50% refund
	partialRefundPct    = 50
	fullRefundPct       = 90 // 24h or more before
)

var cancellationPolicy = []string{
	"24 hours or more before departure: 90% refund",
	"6 to 24 hours before departure: 50% refund",
	"2 to 6 hours before departure: cancellation allowed, no refund",
	"Less than 2 hours before departure, or after departure: cancellation is closed",
}

// computeRefund applies the policy. ok=false means cancellation is not allowed, with the reason.
func computeRefund(now, departure time.Time, total float64) (percent int, amount float64, reason string, ok bool) {
	until := departure.Sub(now)

	switch {
	case until <= cancelCutoff:
		return 0, 0, "Cancellation closes 2 hours before departure.", false
	case until < noRefundWindow:
		percent = 0
	case until < partialRefundWindow:
		percent = partialRefundPct
	default:
		percent = fullRefundPct
	}
	return percent, round2(total * float64(percent) / 100), "", true
}

func normalizePNR(pnr string) string {
	return strings.ToUpper(strings.TrimSpace(pnr))
}

// GetCancellationPreview shows what cancelling a ticket would do, without changing anything.
func (s *bookingService) GetCancellationPreview(ctx context.Context, userID, pnr string) (*domain.CancellationPreview, error) {
	c, err := s.bookingRepo.GetCancellationContext(ctx, normalizePNR(pnr), userID)
	if err != nil {
		return nil, err
	}

	preview := &domain.CancellationPreview{
		PNRCode:       c.PNRCode,
		TotalAmount:   c.TotalAmount,
		DepartureTime: c.Departure,
		Policy:        cancellationPolicy,
	}

	if c.Status != domain.BookingStatusConfirmed {
		preview.Reason = fmt.Sprintf("This ticket is already %s.", strings.ToLower(string(c.Status)))
		return preview, nil
	}

	percent, amount, reason, ok := computeRefund(time.Now(), c.Departure, c.TotalAmount)
	preview.CanCancel = ok
	preview.Reason = reason
	preview.RefundPercent = percent
	preview.RefundAmount = amount
	return preview, nil
}

// CancelBooking cancels a paid ticket the user owns, frees the seats and records a demo refund.
func (s *bookingService) CancelBooking(ctx context.Context, userID, pnr string) (*domain.CancellationResult, error) {
	pnr = normalizePNR(pnr)

	outcome, err := s.bookingRepo.CancelConfirmedBooking(ctx, pnr, userID,
		func(total float64, departure time.Time) (int, float64, error) {
			percent, amount, reason, ok := computeRefund(time.Now(), departure, total)
			if !ok {
				return 0, 0, domain.NewBookingError(domain.ErrCancelNotAllowed, "%s", reason)
			}
			return percent, amount, nil
		})
	if err != nil {
		return nil, err
	}

	msg := "Ticket cancelled. No refund applies at this time."
	if outcome.RefundAmount > 0 {
		msg = fmt.Sprintf("Ticket cancelled. A demo refund of %.2f BDT (%d%%) was recorded. No real money moves.",
			outcome.RefundAmount, outcome.RefundPercent)
	}

	return &domain.CancellationResult{
		PNRCode:       pnr,
		BookingStatus: outcome.BookingStatus,
		PaymentStatus: outcome.PaymentStatus,
		RefundPercent: outcome.RefundPercent,
		RefundAmount:  outcome.RefundAmount,
		Message:       msg,
	}, nil
}
