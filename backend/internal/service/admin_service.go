package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
)

var (
	ErrAdminForbidden = errors.New("you do not have access to this resource")
	ErrAdminConflict  = errors.New("conflict")
)

// Bangladesh has no daylight saving, so a fixed offset is enough.
var bdt = time.FixedZone("BDT", 6*60*60)

type AdminService struct {
	repo repository.AdminRepository
}

func NewAdminService(repo repository.AdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

// scopeOperator returns the operator a caller is limited to (nil = no limit, i.e. SUPER_ADMIN).
func (s *AdminService) scopeOperator(ctx context.Context, userID string, role domain.UserRole) (*string, error) {
	if role == domain.RoleSuperAdmin {
		return nil, nil
	}
	if role != domain.RoleOperatorAdmin {
		return nil, ErrAdminForbidden
	}
	operatorID, err := s.repo.GetOperatorIDForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if operatorID == nil {
		return nil, ErrAdminForbidden
	}
	return operatorID, nil
}

func (s *AdminService) ListVehicles(ctx context.Context, userID string, role domain.UserRole) ([]domain.AdminVehicle, error) {
	scope, err := s.scopeOperator(ctx, userID, role)
	if err != nil {
		return nil, err
	}
	return s.repo.ListVehicles(ctx, scope)
}

func (s *AdminService) ListRoutes(ctx context.Context) ([]domain.AdminRoute, error) {
	return s.repo.ListRoutes(ctx)
}

func (s *AdminService) ListTrips(ctx context.Context, userID string, role domain.UserRole) ([]domain.AdminTrip, error) {
	scope, err := s.scopeOperator(ctx, userID, role)
	if err != nil {
		return nil, err
	}
	return s.repo.ListUpcomingTrips(ctx, scope, 100)
}

func parseAdminTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	for _, layout := range []string{"2006-01-02T15:04", "2006-01-02T15:04:05"} {
		if t, err := time.ParseInLocation(layout, value, bdt); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time %q", value)
}

// CreateTrip validates the request, builds the seat inventory from the vehicle's layout
// and stores the trip together with all of its seats in one transaction.
func (s *AdminService) CreateTrip(
	ctx context.Context,
	userID string,
	role domain.UserRole,
	req *domain.CreateTripRequest,
) (*domain.CreateTripResponse, error) {
	scope, err := s.scopeOperator(ctx, userID, role)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.VehicleID) == "" || strings.TrimSpace(req.RouteID) == "" {
		return nil, &ValidationError{"vehicle_id and route_id are required"}
	}
	departure, err := parseAdminTime(req.DepartureTime)
	if err != nil {
		return nil, &ValidationError{"departure_time must look like 2026-10-25T08:00"}
	}
	arrival, err := parseAdminTime(req.ArrivalTime)
	if err != nil {
		return nil, &ValidationError{"arrival_time must look like 2026-10-25T15:00"}
	}
	if !departure.After(time.Now()) {
		return nil, &ValidationError{"departure time must be in the future"}
	}
	if !arrival.After(departure) {
		return nil, &ValidationError{"arrival time must be after departure time"}
	}
	if arrival.Sub(departure) > 48*time.Hour {
		return nil, &ValidationError{"a trip cannot be longer than 48 hours"}
	}
	if req.BaseFare < 1 || req.BaseFare > 100000 {
		return nil, &ValidationError{"base fare must be between 1 and 100000 BDT"}
	}

	vehicle, err := s.repo.GetVehicle(ctx, req.VehicleID)
	if err != nil {
		if errors.Is(err, repository.ErrReferenceNotFound) {
			return nil, &ValidationError{"selected vehicle does not exist"}
		}
		return nil, err
	}
	if scope != nil && vehicle.OperatorID != *scope {
		return nil, ErrAdminForbidden
	}

	seats, err := buildSeatPlan(vehicle.Type, vehicle.SeatLayoutConfig, req.BaseFare)
	if err != nil {
		return nil, &ValidationError{err.Error()}
	}

	tripID, err := s.repo.CreateTripWithSeats(ctx, vehicle.ID, req.RouteID, departure, arrival, req.BaseFare, seats)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrTripOverlap):
			return nil, fmt.Errorf("%w: this vehicle already has a trip in that time window", ErrAdminConflict)
		case errors.Is(err, repository.ErrReferenceNotFound):
			return nil, &ValidationError{"selected route does not exist"}
		}
		return nil, err
	}

	return &domain.CreateTripResponse{TripID: tripID, SeatsCreated: len(seats)}, nil
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }

// buildSeatPlan creates the seat list for a trip. Seat numbers and classes follow
// backend/scripts/seed.sql, so seats created here look the same as seeded ones.
// Keep the two in sync.
func buildSeatPlan(vtype domain.TransportType, layout map[string]any, baseFare float64) ([]domain.SeatSpec, error) {
	switch vtype {
	case domain.TransportBus:
		perRow := 4
		if t, _ := layout["type"].(string); t == "2x1" {
			perRow = 3
		}
		var seats []domain.SeatSpec
		for row := 0; row < 10; row++ {
			for n := 1; n <= perRow; n++ {
				seats = append(seats, domain.SeatSpec{
					SeatNumber:  fmt.Sprintf("%c%d", 'A'+row, n),
					DeckOrClass: "AC_CHAIR",
					Fare:        round2(baseFare),
				})
			}
		}
		return seats, nil

	case domain.TransportTrain:
		coaches, _ := layout["coaches"].([]any)
		if len(coaches) == 0 {
			return nil, errors.New("this train has no coaches configured")
		}
		var seats []domain.SeatSpec
		for _, c := range coaches {
			coach, _ := c.(map[string]any)
			name, _ := coach["coach_name"].(string)
			class, _ := coach["class"].(string)
			total, _ := coach["total_seats"].(float64)
			if strings.TrimSpace(name) == "" || strings.TrimSpace(class) == "" || total < 1 || total > 200 {
				return nil, errors.New("this train has an invalid coach configuration")
			}
			mult := 1.0
			if class == "SNIGDHA" {
				mult = 1.4
			}
			for n := 1; n <= int(total); n++ {
				seats = append(seats, domain.SeatSpec{
					SeatNumber:  fmt.Sprintf("%s-%d", name, n),
					DeckOrClass: class,
					Fare:        round2(baseFare * mult),
				})
			}
		}
		return seats, nil

	case domain.TransportLaunch:
		groups := []struct {
			prefix      string
			first, last int
			class       string
			mult        float64
		}{
			{"DECK-", 1, 40, "DECK", 1.0},
			{"CABIN-", 201, 208, "CABIN_SINGLE", 2.0},
			{"CABIN-", 301, 304, "CABIN_DOUBLE", 3.5},
			{"CABIN-", 401, 402, "VIP_CABIN", 5.0},
		}
		var seats []domain.SeatSpec
		for _, g := range groups {
			for n := g.first; n <= g.last; n++ {
				seats = append(seats, domain.SeatSpec{
					SeatNumber:  fmt.Sprintf("%s%d", g.prefix, n),
					DeckOrClass: g.class,
					Fare:        round2(baseFare * g.mult),
				})
			}
		}
		return seats, nil
	}

	return nil, fmt.Errorf("trips for %s vehicles are not supported in this demo", vtype)
}
