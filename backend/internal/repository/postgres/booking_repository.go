package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

// mapUniqueViolation converts Postgres unique-violation errors (code 23505)
// into repository sentinel errors. Returns nil if err is not one we care about.
func mapUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		name := strings.ToLower(pgErr.ConstraintName)
		switch {
		case strings.Contains(name, "pnr"):
			return repository.ErrDuplicatePNR
		case strings.Contains(name, "transaction"):
			return repository.ErrDuplicateTransaction
		}
	}
	return nil
}

func (r *bookingRepository) GetSeatsFare(ctx context.Context, tripID string, seatNumbers []string) (float64, error) {
	// Only seats that are still sellable: not booked, trip scheduled and not yet departed.
	query := `
		SELECT COALESCE(SUM(s.fare), 0), COUNT(*)
		FROM trip_seat_inventory s
		JOIN trips t ON t.id = s.trip_id
		WHERE s.trip_id = $1
		  AND s.seat_number = ANY($2)
		  AND s.status <> 'BOOKED'
		  AND t.status = 'SCHEDULED'
		  AND t.departure_time > NOW()
	`
	var totalFare float64
	var found int
	if err := r.db.QueryRow(ctx, query, tripID, seatNumbers).Scan(&totalFare, &found); err != nil {
		return 0, fmt.Errorf("failed to calculate seats fare: %w", err)
	}
	if found != len(seatNumbers) {
		return 0, repository.ErrSeatsUnavailable
	}
	return totalFare, nil
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

	// 1. Mark the selected seats as BOOKED (only if not already booked)
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
		return repository.ErrSeatsUnavailable
	}

	// 2. Insert booking
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
		string(booking.Status),
	)
	if err != nil {
		if mapped := mapUniqueViolation(err); mapped != nil {
			return mapped
		}
		return fmt.Errorf("failed to insert booking: %w", err)
	}

	// 3. Insert tickets
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

	// 4. Insert payment record
	paymentQuery := `
		INSERT INTO payments (id, booking_id, gateway, transaction_id, gateway_payment_id, amount, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`
	_, err = tx.Exec(
		ctx,
		paymentQuery,
		payment.ID,
		payment.BookingID,
		string(payment.Gateway),
		payment.TransactionID,
		payment.GatewayPaymentID,
		payment.Amount,
		payment.Currency,
		string(payment.Status),
	)
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

func (r *bookingRepository) GetBookingByPNR(ctx context.Context, pnr string) (*domain.Booking, []domain.Ticket, error) {
	booking := &domain.Booking{}
	bookingQuery := `
		SELECT id, user_id, trip_id, pnr_code, total_amount, discount_amount, status, created_at, updated_at
		FROM bookings
		WHERE pnr_code = $1
	`
	var statusStr string
	err := r.db.QueryRow(ctx, bookingQuery, pnr).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.TripID,
		&booking.PNRCode,
		&booking.TotalAmount,
		&booking.DiscountAmount,
		&statusStr,
		&booking.CreatedAt,
		&booking.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, domain.NewBookingError(domain.ErrBookingNotFound, "booking not found")
		}
		return nil, nil, fmt.Errorf("database query error: %w", err)
	}
	booking.Status = domain.BookingStatus(statusStr)

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
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("tickets row iteration error: %w", err)
	}

	return booking, tickets, nil
}

func splitSeats(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

// GetBookingsByUserID lists the user's paid bookings with their trip details.
// Bookings whose payment never succeeded (failed, cancelled, expired) are not tickets
// and are left out.
func (r *bookingRepository) GetBookingsByUserID(ctx context.Context, userID string) ([]domain.BookingSummary, error) {
	query := `
		SELECT b.id, b.user_id, b.trip_id, b.pnr_code, b.total_amount, COALESCE(b.discount_amount, 0), b.status::text,
		       COALESCE(b.created_at, NOW()), COALESCE(b.updated_at, NOW()),
		       v.type::text, o.name, v.registration_no, rt.origin_city, rt.destination_city,
		       t.departure_time, t.arrival_time,
		       COALESCE((SELECT string_agg(tk.seat_number, ',' ORDER BY tk.seat_number)
		                 FROM tickets tk WHERE tk.booking_id = b.id), ''),
		       COALESCE((SELECT p.status::text FROM payments p
		                 WHERE p.booking_id = b.id AND p.status IN ('SUCCESS', 'REFUNDED')
		                 ORDER BY p.created_at DESC NULLS LAST LIMIT 1), '')
		FROM bookings b
		JOIN trips t ON t.id = b.trip_id
		JOIN vehicles v ON v.id = t.vehicle_id
		JOIN operators o ON o.id = v.operator_id
		JOIN routes rt ON rt.id = t.route_id
		WHERE b.user_id = $1
		  AND EXISTS (SELECT 1 FROM payments p WHERE p.booking_id = b.id AND p.status IN ('SUCCESS', 'REFUNDED'))
		ORDER BY b.created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user bookings: %w", err)
	}
	defer rows.Close()

	bookings := []domain.BookingSummary{}
	for rows.Next() {
		var b domain.BookingSummary
		var status, transport, payStatus, seats string
		if err := rows.Scan(
			&b.ID, &b.UserID, &b.TripID, &b.PNRCode, &b.TotalAmount, &b.DiscountAmount, &status,
			&b.CreatedAt, &b.UpdatedAt,
			&transport, &b.OperatorName, &b.RegistrationNo, &b.Origin, &b.Destination,
			&b.DepartureTime, &b.ArrivalTime, &seats, &payStatus,
		); err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		b.Status = domain.BookingStatus(status)
		b.TransportType = domain.TransportType(transport)
		b.PaymentStatus = domain.PaymentStatus(payStatus)
		b.Seats = splitSeats(seats)
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("bookings row iteration error: %w", err)
	}

	return bookings, nil
}
