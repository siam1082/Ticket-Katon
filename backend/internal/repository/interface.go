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