package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
)

func (r *bookingRepository) GetCancellationContext(ctx context.Context, pnr, userID string) (*repository.CancellationContext, error) {
	var c repository.CancellationContext
	var status string
	err := r.db.QueryRow(ctx, `
		SELECT b.pnr_code, b.status::text, b.total_amount, t.departure_time
		FROM bookings b
		JOIN trips t ON t.id = b.trip_id
		WHERE b.pnr_code = $1 AND b.user_id = $2`, pnr, userID).
		Scan(&c.PNRCode, &status, &c.TotalAmount, &c.Departure)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewBookingError(domain.ErrBookingNotFound, "booking not found")
		}
		return nil, fmt.Errorf("failed to load booking: %w", err)
	}
	c.Status = domain.BookingStatus(status)
	return &c, nil
}

func (r *bookingRepository) CancelConfirmedBooking(
	ctx context.Context,
	pnr, userID string,
	decide repository.RefundDecider,
) (*repository.CancelOutcome, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock the booking so two cancel clicks (or a cancel and a payment) cannot overlap.
	var bookingID, tripID, status string
	var total float64
	var departure time.Time
	err = tx.QueryRow(ctx, `
		SELECT b.id, b.trip_id, b.status::text, b.total_amount, t.departure_time
		FROM bookings b
		JOIN trips t ON t.id = b.trip_id
		WHERE b.pnr_code = $1 AND b.user_id = $2
		FOR UPDATE OF b`, pnr, userID).Scan(&bookingID, &tripID, &status, &total, &departure)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewBookingError(domain.ErrBookingNotFound, "booking not found")
		}
		return nil, fmt.Errorf("failed to load booking: %w", err)
	}
	if status != string(domain.BookingStatusConfirmed) {
		return nil, domain.NewBookingError(domain.ErrCancelNotAllowed,
			"this ticket is already %s", strings.ToLower(status))
	}

	var paymentID string
	err = tx.QueryRow(ctx, `
		SELECT id FROM payments
		WHERE booking_id = $1 AND status = 'SUCCESS'
		ORDER BY created_at DESC NULLS LAST
		LIMIT 1
		FOR UPDATE`, bookingID).Scan(&paymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewBookingError(domain.ErrCancelNotAllowed, "this ticket has no completed payment")
		}
		return nil, fmt.Errorf("failed to load payment: %w", err)
	}

	percent, amount, err := decide(total, departure)
	if err != nil {
		return nil, err
	}

	// Give the seats back.
	if _, err := tx.Exec(ctx, `
		UPDATE trip_seat_inventory
		SET status = 'AVAILABLE', held_until = NULL, updated_at = NOW()
		WHERE trip_id = $1 AND status = 'BOOKED'
		  AND seat_number IN (SELECT seat_number FROM tickets WHERE booking_id = $2)`,
		tripID, bookingID); err != nil {
		return nil, fmt.Errorf("failed to release seats: %w", err)
	}

	if _, err := tx.Exec(ctx,
		`UPDATE bookings SET status = 'CANCELLED', updated_at = NOW() WHERE id = $1`, bookingID); err != nil {
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	paymentStatus := domain.PaymentStatusSuccess
	if amount > 0 {
		paymentStatus = domain.PaymentStatusRefunded
	}
	if _, err := tx.Exec(ctx, `
		UPDATE payments
		SET status = CASE WHEN $2::numeric > 0 THEN 'REFUNDED'::payment_status ELSE status END,
		    raw_payload = COALESCE(raw_payload, '{}'::jsonb)
		                  || jsonb_build_object('cancelled_at', NOW(), 'refund_percent', $3::int, 'refund_amount', $2::numeric),
		    updated_at = NOW()
		WHERE id = $1`, paymentID, amount, percent); err != nil {
		return nil, fmt.Errorf("failed to record refund: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &repository.CancelOutcome{
		BookingStatus: domain.BookingStatusCancelled,
		PaymentStatus: paymentStatus,
		RefundPercent: percent,
		RefundAmount:  amount,
	}, nil
}
