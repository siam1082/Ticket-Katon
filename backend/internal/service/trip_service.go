package service

import (
	"context"
	"errors"

	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
)

type TripService struct {
	repo repository.TripRepository
}

func NewTripService(repo repository.TripRepository) *TripService {
	return &TripService{repo: repo}
}

func (s *TripService) Search(ctx context.Context, params repository.TripSearchParams) ([]domain.TripSearchResult, error) {
	if params.Origin == "" || params.Destination == "" || params.TravelDate == "" {
		return nil, errors.New("origin, destination, and travel_date are required")
	}
	return s.repo.SearchTrips(ctx, params)
}

func (s *TripService) GetSeats(ctx context.Context, tripID string) (*domain.TripSeatLayoutResponse, error) {
	if tripID == "" {
		return nil, errors.New("trip_id is required")
	}
	return s.repo.GetTripSeatLayout(ctx, tripID)
}