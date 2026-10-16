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

	
}
