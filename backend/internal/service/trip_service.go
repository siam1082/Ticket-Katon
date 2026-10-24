package service

import (
	"context"
	"errors"
	"log"
	"sort"
	"strings"
	"time"

	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
)

type TripService struct {
	repo        repository.TripRepository
	seatLockSvc *SeatLockService
	cities      CityLookup // optional: makes the search tolerant of spelling variants
}

func NewTripService(repo repository.TripRepository, seatLockSvc *SeatLockService) *TripService {
	return &TripService{
		repo:        repo,
		seatLockSvc: seatLockSvc,
	}
}

// NewTripServiceWithCities is NewTripService plus spelling-tolerant city search.
func NewTripServiceWithCities(repo repository.TripRepository, seatLockSvc *SeatLockService, cities CityLookup) *TripService {
	svc := NewTripService(repo, seatLockSvc)
	svc.cities = cities
	return svc
}

func (s *TripService) Search(ctx context.Context, params repository.TripSearchParams) ([]domain.TripSearchResult, error) {
	origin := strings.TrimSpace(params.Origin)
	dest := strings.TrimSpace(params.Destination)

	if origin == "" || dest == "" {
		return nil, errors.New("origin and destination are required")
	}

	if strings.EqualFold(origin, dest) || (s.cities != nil && cityKey(origin) == cityKey(dest)) {
		return nil, errors.New("origin and destination cannot be the same")
	}

	if params.TravelDate == "" {
		return nil, errors.New("travel_date (YYYY-MM-DD) is required")
	}

	parsedDate, err := time.Parse("2006-01-02", params.TravelDate)
	if err != nil {
		return nil, errors.New("invalid date format, must be YYYY-MM-DD")
	}

	// অতীতের তারিখ আটকে দেওয়া (শুধু আজকের বা ভবিষ্যতের তারিখ গ্রাহ্য হবে)
	today := time.Now().Truncate(24 * time.Hour)
	if parsedDate.Before(today) {
		return nil, errors.New("travel_date cannot be in the past")
	}

	return s.searchResolved(ctx, params)
}

// searchResolved searches every stored spelling of the typed cities (for example both
// "Chattogram" and "Chittagong") and merges the results. If a city is unknown it falls back
// to the plain search, which behaves exactly as before.
func (s *TripService) searchResolved(ctx context.Context, params repository.TripSearchParams) ([]domain.TripSearchResult, error) {
	if s.cities == nil {
		return s.repo.SearchTrips(ctx, params)
	}

	origins, err := s.cities.Resolve(ctx, params.Origin)
	if err != nil || len(origins) == 0 {
		return s.repo.SearchTrips(ctx, params)
	}
	dests, err := s.cities.Resolve(ctx, params.Destination)
	if err != nil || len(dests) == 0 {
		return s.repo.SearchTrips(ctx, params)
	}

	seen := map[string]bool{}
	merged := []domain.TripSearchResult{}
	for _, o := range origins {
		for _, d := range dests {
			p := params
			p.Origin, p.Destination = o, d
			results, err := s.repo.SearchTrips(ctx, p)
			if err != nil {
				return nil, err
			}
			for _, r := range results {
				if !seen[r.ID] {
					seen[r.ID] = true
					merged = append(merged, r)
				}
			}
		}
	}
	sort.SliceStable(merged, func(i, j int) bool { return merged[i].DepartureTime.Before(merged[j].DepartureTime) })
	return merged, nil
}

func (s *TripService) GetSeats(ctx context.Context, tripID string) (*domain.TripSeatLayoutResponse, error) {
	if strings.TrimSpace(tripID) == "" {
		return nil, errors.New("trip_id is required")
	}

	layout, err := s.repo.GetTripSeatLayout(ctx, tripID)
	if err != nil {
		return nil, err
	}

	// Seats that are AVAILABLE in Postgres but locked in Redis are shown as HELD.
	// One MGET covers all seats. If Redis fails we keep the database status
	// (fail open) so the seat map still loads; the hold step re-checks anyway.
	if s.seatLockSvc != nil && layout != nil {
		var candidates []string
		for _, seat := range layout.Seats {
			if seat.Status == domain.SeatAvailable {
				candidates = append(candidates, seat.SeatNumber)
			}
		}

		held, err := s.seatLockSvc.HeldSeats(ctx, tripID, candidates)
		if err != nil {
			log.Printf("[SEATS] redis lock lookup failed, showing database status: %v", err)
		}

		for i := range layout.Seats {
			if held[layout.Seats[i].SeatNumber] {
				layout.Seats[i].Status = domain.SeatHeld
			}
		}
	}

	return layout, nil
}
