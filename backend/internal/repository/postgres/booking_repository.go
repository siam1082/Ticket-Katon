package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
)

type bookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) repository.BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) CreateBookingWithSeats(
	ctx context.Context,
	booking *domain.Booking,
	tickets []domain.Ticket,
	payment *domain.Payment,
	seatNumbers []string,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// ১. সিট ইনভেন্টরি আপডেট: নির্বাচিত সিটগুলোকে BOOKED করা
	updateSeatQuery := `
		UPDATE trip_seat_inventory
		SET status = 'BOOKED', updated_at = NOW()
		WHERE trip_id = $1 AND seat_number = ANY($2) AND status != 'BOOKED'
	`
	tag, err := tx.Exec(ctx, updateSeatQuery, booking.TripID, seatNumbers)
	if err != nil {
		return fmt.Errorf("failed to update seat inventory: %w", err)
	}
	if int(tag.RowsAffected()) != len(seatNumbers) {
		return fmt.Errorf("one or more seats are already booked or invalid")
	}

	// ২. Bookings টেবিলে বুকিং ডাটা ইনসার্ট করা
	bookingQuery := `
		INSERT INTO bookings (id, user_id, trip_id, pnr_code, total_amount, discount_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err = tx.Exec(
		ctx,
		bookingQuery,
		booking.ID,
		booking.UserID,
		booking.TripID,
		booking.PNRCode,
		booking.TotalAmount,
		booking.DiscountAmount,
		booking.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to insert booking: %w", err)
	}

	// ৩. Tickets টেবিলে প্রতিটি টিকিটের তথ্য ইনসার্ট করা
	ticketQuery := `
		INSERT INTO tickets (id, booking_id, seat_number, passenger_name, passenger_phone, passenger_gender, ticket_number, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
	`
	for _, ticket := range tickets {
		_, err = tx.Exec(
			ctx,
			ticketQuery,
			ticket.ID,
			ticket.BookingID,
			ticket.SeatNumber,
			ticket.PassengerName,
			ticket.PassengerPhone,
			ticket.PassengerGender,
			ticket.TicketNumber,
		)
		if err != nil {
			return fmt.Errorf("failed to insert ticket for seat %s: %w", ticket.SeatNumber, err)
		}
	}

	// ৪. Payments টেবিলে পেমেন্ট রেকর্ড ইনসার্ট করা
	paymentQuery := `
		INSERT INTO payments (id, booking_id, gateway, transaction_id, gateway_payment_id, amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`
	_, err = tx.Exec(
		ctx,
		paymentQuery,
		payment.ID,
		payment.BookingID,
		payment.Gateway,
		payment.TransactionID,
		payment.GatewayPaymentID,
		payment.Amount,
		payment.Currency,
		payment.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment record: %w", err)
	}

	// ট্রানজেকশন সফলভাবে সম্পন্ন করা
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *bookingRepository) GetBookingByPNR(ctx context.Context, pnr string) (*domain.Booking, []domain.Ticket, error) {
	booking := &domain.Booking{}
	bookingQuery := `
		SELECT id, user_id, trip_id, pnr_code, total_amount, discount_amount, status, created_at, updated_at
		FROM bookings
		WHERE pnr_code = $1
	`
	err := r.db.QueryRow(ctx, bookingQuery, pnr).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.TripID,
		&booking.PNRCode,
		&booking.TotalAmount,
		&booking.DiscountAmount,
		&booking.Status,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("booking not found: %w", err)
	}

	ticketsQuery := `
		SELECT id, booking_id, seat_number, passenger_name, passenger_phone, passenger_gender, ticket_number, created_at
		FROM tickets
		WHERE booking_id = $1
	`
	rows, err := r.db.Query(ctx, ticketsQuery, booking.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to query tickets: %w", err)
	}
	defer rows.Close()

	var tickets []domain.Ticket
	for rows.Next() {
		var t domain.Ticket
		if err := rows.Scan(
			&t.ID,
			&t.BookingID,
			&t.SeatNumber,
			&t.PassengerName,
			&t.PassengerPhone,
			&t.PassengerGender,
			&t.TicketNumber,
			&t.CreatedAt,
		); err != nil {
			return nil, nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		tickets = append(tickets, t)
	}

	return booking, tickets, nil
}

func (r *bookingRepository) GetBookingsByUserID(ctx context.Context, userID string) ([]domain.Booking, error) {
	query := `
		SELECT id, user_id, trip_id, pnr_code, total_amount, discount_amount, status, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user bookings: %w", err)
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(
			&b.ID,
			&b.UserID,
			&b.TripID,
			&b.PNRCode,
			&b.TotalAmount,
			&b.DiscountAmount,
			&b.Status,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, b)
	}

	return bookings, nil
}