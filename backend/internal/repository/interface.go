package repository

import (
	"context"
	"errors"

	"ticket-katon-backend/internal/domain"
)

// Sentinel errors returned by BookingRepository so the service layer
// can react without depending on database driver details.
var (
	ErrSeatsUnavailable     = errors.New("one or more seats are already booked or invalid")
	ErrDuplicatePNR         = errors.New("pnr code already exists")
	ErrDuplicateTransaction = errors.New("payment transaction id already used")
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type TripSearchParams struct {
	Origin        string
	Destination   string
	TravelDate    string // Format: YYYY-MM-DD
	TransportType string // BUS, LAUNCH, or empty
}

type TripRepository interface {
	SearchTrips(ctx context.Context, params TripSearchParams) ([]domain.TripSearchResult, error)
	GetTripSeatLayout(ctx context.Context, tripID string) (*domain.TripSeatLayoutResponse, error)
}

type BookingRepository interface {
	// CreateBookingWithSeats atomically books seats, issues tickets and records the payment.
	CreateBookingWithSeats(ctx context.Context, booking *domain.Booking, tickets []domain.Ticket, payment *domain.Payment, seatNumbers []string) error

	// GetBookingByPNR returns a booking and its tickets by PNR code.
	GetBookingByPNR(ctx context.Context, pnr string) (*domain.Booking, []domain.Ticket, error)

	// GetBookingsByUserID returns a user's booking history, newest first.
	GetBookingsByUserID(ctx context.Context, userID string) ([]domain.Booking, error)

	// GetSeatsFare returns the total fare of the given seats.
	// It returns an error if any requested seat does not exist for the trip.
	GetSeatsFare(ctx context.Context, tripID string, seatNumbers []string) (float64, error)
}