package service

import (
	"context"
	"errors"
	"strings"
	"time"

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
	origin := strings.TrimSpace(params.Origin)
	dest := strings.TrimSpace(params.Destination)

	if origin == "" || dest == "" {
		return nil, errors.New("origin and destination are required")
	}

	if strings.EqualFold(origin, dest) {
		return nil, errors.New("origin and destination cannot be the same")
	}

	if params.TravelDate == "" {
		return nil, errors.New("travel_date (YYYY-MM-DD) is required")
	}

	parsedDate, err := time.Parse("2006-01-02", params.TravelDate)
	if err != nil {
		return nil, errors.New("invalid date format, must be YYYY-MM-DD")
	}

	// অতীতের তারিখ আটকে দেওয়া (শুধু আজকের বা ভবিষ্যতের তারিখ গ্রাহ্য হবে)
	today := time.Now().Truncate(24 * time.Hour)
	if parsedDate.Before(today) {
		return nil, errors.New("travel_date cannot be in the past")
	}

	return s.repo.SearchTrips(ctx, params)
}

func (s *TripService) GetSeats(ctx context.Context, tripID string) (*domain.TripSeatLayoutResponse, error) {
	if strings.TrimSpace(tripID) == "" {
		return nil, errors.New("trip_id is required")
	}
	return s.repo.GetTripSeatLayout(ctx, tripID)
}