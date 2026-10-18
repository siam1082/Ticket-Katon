package repository

import (
	"context"

	"ticket-katon-backend/internal/domain"
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
	CreateBookingWithSeats(ctx context.Context, booking *domain.Booking, tickets []domain.Ticket, payment *domain.Payment, seatNumbers []string) error
	GetBookingByPNR(ctx context.Context, pnr string) (*domain.Booking, []domain.Ticket, error)
	GetBookingsByUserID(ctx context.Context, userID string) ([]domain.Booking, error)
	GetSeatsFare(ctx context.Context, tripID string, seatNumbers []string) (float64, error)
}