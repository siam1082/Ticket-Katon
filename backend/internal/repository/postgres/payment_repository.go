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

// queryer is satisfied by both *pgxpool.Pool and pgx.Tx.
type queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

const paymentWithBookingSelect = `
	SELECT p.id, p.booking_id, p.gateway::text, COALESCE(p.transaction_id, ''), COALESCE(p.gateway_payment_id, ''),
	       p.amount, COALESCE(p.currency, 'BDT'), p.status::text, COALESCE(p.created_at, NOW()), COALESCE(p.updated_at, NOW()),
	       b.id, b.user_id, b.trip_id, b.pnr_code, b.total_amount, COALESCE(b.discount_amount, 0), b.status::text,
	       COALESCE(b.created_at, NOW()), COALESCE(b.updated_at, NOW())
	FROM payments p
	JOIN bookings b ON b.id = p.booking_id
	WHERE p.id = $1 AND b.user_id = $2`

// loadPaymentBooking reads a payment and its booking, only if the user owns it.
// With lock=true the rows are locked until the surrounding transaction ends.
func loadPaymentBooking(ctx context.Context, q queryer, paymentID, userID string, lock bool) (*domain.Payment, *domain.Booking, error) {
	query := paymentWithBookingSelect
	if lock {
		query += " FOR UPDATE OF p, b"
	}

	var p domain.Payment
	var b domain.Booking
	var gateway, pStatus, bStatus string

	err := q.QueryRow(ctx, query, paymentID, userID).Scan(
		&p.ID, &p.BookingID, &gateway, &p.TransactionID, &p.GatewayPaymentID,
		&p.Amount, &p.Currency, &pStatus, &p.CreatedAt, &p.UpdatedAt,
		&b.ID, &b.UserID, &b.TripID, &b.PNRCode, &b.TotalAmount, &b.DiscountAmount, &bStatus,
		&b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.NewBookingError(domain.ErrPaymentNotFound, "payment not found")
		}
		return nil, nil, fmt.Errorf("failed to load payment: %w", err)
	}

	p.Gateway = domain.PaymentGateway(gateway)
	p.Status = domain.PaymentStatus(pStatus)
	b.Status = domain.BookingStatus(bStatus)
	return &p, &b, nil
}

func fetchTickets(ctx context.Context, q queryer, bookingID string) ([]domain.Ticket, error) {
	rows, err := q.Query(ctx, `
		SELECT id, booking_id, seat_number, passenger_name, passenger_phone,
		       COALESCE(passenger_gender, ''), ticket_number, COALESCE(created_at, NOW())
		FROM tickets
		WHERE booking_id = $1
		ORDER BY seat_number`, bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tickets: %w", err)
	}
	defer rows.Close()

	tickets := []domain.Ticket{}
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(&t.ID, &t.BookingID, &t.SeatNumber, &t.PassengerName, &t.PassengerPhone,
			&t.PassengerGender, &t.TicketNumber, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("tickets row iteration error: %w", err)
	}
	return tickets, nil
}

func (r *bookingRepository) CreatePendingBooking(
	ctx context.Context,
	booking *domain.Booking,
	tickets []domain.Ticket,
	payment *domain.Payment,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// A new attempt supersedes this user's older pending attempts for the same trip.
	if _, err := tx.Exec(ctx, `
		UPDATE payments
		SET status = 'FAILED', raw_payload = jsonb_build_object('simulated_result', 'superseded'), updated_at = NOW()
		WHERE status = 'INITIATED'
		  AND booking_id IN (SELECT id FROM bookings WHERE user_id = $1 AND trip_id = $2 AND status = 'PENDING')`,
		booking.UserID, booking.TripID); err != nil {
		return fmt.Errorf("failed to supersede old payments: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE bookings SET status = 'CANCELLED', updated_at = NOW()
		WHERE user_id = $1 AND trip_id = $2 AND status = 'PENDING'`,
		booking.UserID, booking.TripID); err != nil {
		return fmt.Errorf("failed to supersede old bookings: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO bookings (id, user_id, trip_id, pnr_code, total_amount, discount_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`,
		booking.ID, booking.UserID, booking.TripID, booking.PNRCode,
		booking.TotalAmount, booking.DiscountAmount, string(booking.Status))
	if err != nil {
		if mapped := mapUniqueViolation(err); mapped != nil {
			return mapped
		}
		return fmt.Errorf("failed to insert booking: %w", err)
	}

	for _, t := range tickets {
		_, err = tx.Exec(ctx, `
			INSERT INTO tickets (id, booking_id, seat_number, passenger_name, passenger_phone, passenger_gender, ticket_number, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`,
			t.ID, t.BookingID, t.SeatNumber, t.PassengerName, t.PassengerPhone, t.PassengerGender, t.TicketNumber)
		if err != nil {
			return fmt.Errorf("failed to insert ticket for seat %s: %w", t.SeatNumber, err)
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO payments (id, booking_id, gateway, transaction_id, gateway_payment_id, amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`,
		payment.ID, payment.BookingID, string(payment.Gateway), payment.TransactionID, payment.GatewayPaymentID,
		payment.Amount, payment.Currency, string(payment.Status))
	if err != nil {
		if mapped := mapUniqueViolation(err); mapped != nil {
			return mapped
		}
		return fmt.Errorf("failed to insert payment record: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *bookingRepository) GetPaymentForUser(
	ctx context.Context,
	paymentID, userID string,
) (*domain.Payment, *domain.Booking, []domain.Ticket, error) {
	p, b, err := loadPaymentBooking(ctx, r.db, paymentID, userID, false)
	if err != nil {
		return nil, nil, nil, err
	}
	tickets, err := fetchTickets(ctx, r.db, b.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	return p, b, tickets, nil
}

func (r *bookingRepository) ConfirmPendingPayment(
	ctx context.Context,
	paymentID, userID string,
) (*domain.Payment, *domain.Booking, []domain.Ticket, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	p, b, err := loadPaymentBooking(ctx, tx, paymentID, userID, true)
	if err != nil {
		return nil, nil, nil, err
	}
	tickets, err := fetchTickets(ctx, tx, b.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Repeat request after success: return the same result, change nothing.
	if p.Status == domain.PaymentStatusSuccess && b.Status == domain.BookingStatusConfirmed {
		return p, b, tickets, nil
	}
	if p.Status != domain.PaymentStatusInitiated || b.Status != domain.BookingStatusPending {
		return nil, nil, nil, domain.NewBookingError(domain.ErrPaymentProcessed,
			"this payment was already %s", strings.ToLower(string(p.Status)))
	}

	seatNumbers := make([]string, 0, len(tickets))
	for _, t := range tickets {
		seatNumbers = append(seatNumbers, t.SeatNumber)
	}

	tag, err := tx.Exec(ctx, `
		UPDATE trip_seat_inventory
		SET status = 'BOOKED', updated_at = NOW()
		WHERE trip_id = $1 AND seat_number = ANY($2) AND status <> 'BOOKED'`,
		b.TripID, seatNumbers)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to update seat inventory: %w", err)
	}
	if int(tag.RowsAffected()) != len(seatNumbers) {
		return nil, nil, nil, repository.ErrSeatsUnavailable
	}

	if _, err := tx.Exec(ctx, `UPDATE bookings SET status = 'CONFIRMED', updated_at = NOW() WHERE id = $1`, b.ID); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to confirm booking: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE payments
		SET status = 'SUCCESS', raw_payload = jsonb_build_object('simulated_result', 'success'), updated_at = NOW()
		WHERE id = $1`, p.ID); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to confirm payment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	p.Status = domain.PaymentStatusSuccess
	b.Status = domain.BookingStatusConfirmed
	return p, b, tickets, nil
}

func (r *bookingRepository) FailPendingPayment(
	ctx context.Context,
	paymentID, userID, result string,
) (*domain.Payment, *domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	p, b, err := loadPaymentBooking(ctx, tx, paymentID, userID, true)
	if err != nil {
		return nil, nil, err
	}
	if p.Status != domain.PaymentStatusInitiated || b.Status != domain.BookingStatusPending {
		return nil, nil, domain.NewBookingError(domain.ErrPaymentProcessed,
			"this payment was already %s", strings.ToLower(string(p.Status)))
	}

	if _, err := tx.Exec(ctx, `
		UPDATE payments
		SET status = 'FAILED', raw_payload = jsonb_build_object('simulated_result', $2::text), updated_at = NOW()
		WHERE id = $1`, p.ID, result); err != nil {
		return nil, nil, fmt.Errorf("failed to fail payment: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE bookings SET status = 'CANCELLED', updated_at = NOW() WHERE id = $1`, b.ID); err != nil {
		return nil, nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	p.Status = domain.PaymentStatusFailed
	b.Status = domain.BookingStatusCancelled
	return p, b, nil
}

func (r *bookingRepository) ExpireStalePending(ctx context.Context, olderThan time.Duration) (int64, error) {
	tag, err := r.db.Exec(ctx, `
		WITH stale AS (
			UPDATE bookings
			SET status = 'CANCELLED', updated_at = NOW()
			WHERE status = 'PENDING' AND created_at < NOW() - make_interval(secs => $1)
			RETURNING id
		)
		UPDATE payments
		SET status = 'FAILED', raw_payload = jsonb_build_object('simulated_result', 'expired'), updated_at = NOW()
		WHERE status = 'INITIATED' AND booking_id IN (SELECT id FROM stale)`,
		olderThan.Seconds())
	if err != nil {
		return 0, fmt.Errorf("failed to expire stale pending bookings: %w", err)
	}
	return tag.RowsAffected(), nil
}
