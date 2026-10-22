package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
)

type adminRepository struct {
	db *pgxpool.Pool
}

func NewAdminRepository(db *pgxpool.Pool) repository.AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetOperatorIDForUser(ctx context.Context, userID string) (*string, error) {
	var operatorID *string
	err := r.db.QueryRow(ctx, `SELECT operator_id::text FROM users WHERE id = $1`, userID).Scan(&operatorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to load operator for user: %w", err)
	}
	return operatorID, nil
}

const vehicleSelect = `
	SELECT v.id, v.operator_id, o.name, v.registration_no, v.type::text, v.seat_layout_config
	FROM vehicles v
	JOIN operators o ON o.id = v.operator_id`

func scanVehicle(row pgx.Row) (*domain.AdminVehicle, error) {
	var v domain.AdminVehicle
	var vtype string
	var layout []byte
	if err := row.Scan(&v.ID, &v.OperatorID, &v.OperatorName, &v.RegistrationNo, &vtype, &layout); err != nil {
		return nil, err
	}
	v.Type = domain.TransportType(vtype)
	if err := json.Unmarshal(layout, &v.SeatLayoutConfig); err != nil {
		v.SeatLayoutConfig = map[string]any{}
	}
	return &v, nil
}

func (r *adminRepository) GetVehicle(ctx context.Context, vehicleID string) (*domain.AdminVehicle, error) {
	v, err := scanVehicle(r.db.QueryRow(ctx, vehicleSelect+` WHERE v.id = $1`, vehicleID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrReferenceNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "22P02" { // malformed uuid
			return nil, repository.ErrReferenceNotFound
		}
		return nil, fmt.Errorf("failed to load vehicle: %w", err)
	}
	return v, nil
}

func (r *adminRepository) ListVehicles(ctx context.Context, operatorID *string) ([]domain.AdminVehicle, error) {
	rows, err := r.db.Query(ctx,
		vehicleSelect+` WHERE ($1::uuid IS NULL OR v.operator_id = $1::uuid) ORDER BY o.name, v.registration_no`,
		operatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vehicles: %w", err)
	}
	defer rows.Close()

	vehicles := []domain.AdminVehicle{}
	for rows.Next() {
		v, err := scanVehicle(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vehicle: %w", err)
		}
		vehicles = append(vehicles, *v)
	}
	return vehicles, rows.Err()
}

func (r *adminRepository) ListRoutes(ctx context.Context) ([]domain.AdminRoute, error) {
	rows, err := r.db.Query(ctx, `SELECT id, origin_city, destination_city FROM routes ORDER BY origin_city, destination_city`)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %w", err)
	}
	defer rows.Close()

	routes := []domain.AdminRoute{}
	for rows.Next() {
		var rt domain.AdminRoute
		if err := rows.Scan(&rt.ID, &rt.OriginCity, &rt.DestinationCity); err != nil {
			return nil, fmt.Errorf("failed to scan route: %w", err)
		}
		routes = append(routes, rt)
	}
	return routes, rows.Err()
}

func (r *adminRepository) ListUpcomingTrips(ctx context.Context, operatorID *string, limit int) ([]domain.AdminTrip, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.vehicle_id, v.registration_no, o.name, v.type::text,
		       rt.origin_city, rt.destination_city, t.departure_time, t.arrival_time, t.base_fare,
		       t.status::text,
		       COUNT(s.id), COUNT(s.id) FILTER (WHERE s.status = 'BOOKED')
		FROM trips t
		JOIN vehicles v ON v.id = t.vehicle_id
		JOIN operators o ON o.id = v.operator_id
		JOIN routes rt ON rt.id = t.route_id
		LEFT JOIN trip_seat_inventory s ON s.trip_id = t.id
		WHERE t.departure_time > NOW()
		  AND ($1::uuid IS NULL OR v.operator_id = $1::uuid)
		GROUP BY t.id, v.registration_no, o.name, v.type, rt.origin_city, rt.destination_city
		ORDER BY t.departure_time
		LIMIT $2`, operatorID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list trips: %w", err)
	}
	defer rows.Close()

	trips := []domain.AdminTrip{}
	for rows.Next() {
		var t domain.AdminTrip
		var ttype string
		if err := rows.Scan(&t.ID, &t.VehicleID, &t.RegistrationNo, &t.OperatorName, &ttype,
			&t.Origin, &t.Destination, &t.DepartureTime, &t.ArrivalTime, &t.BaseFare,
			&t.Status, &t.TotalSeats, &t.BookedSeats); err != nil {
			return nil, fmt.Errorf("failed to scan trip: %w", err)
		}
		t.TransportType = domain.TransportType(ttype)
		trips = append(trips, t)
	}
	return trips, rows.Err()
}

func (r *adminRepository) CreateTripWithSeats(
	ctx context.Context,
	vehicleID, routeID string,
	departure, arrival time.Time,
	baseFare float64,
	seats []domain.SeatSpec,
) (string, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// The same vehicle cannot run two overlapping trips.
	var overlap bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM trips
			WHERE vehicle_id = $1
			  AND status <> 'CANCELLED'
			  AND departure_time < $3
			  AND arrival_time > $2
		)`, vehicleID, departure, arrival).Scan(&overlap)
	if err != nil {
		return "", fmt.Errorf("failed to check trip overlap: %w", err)
	}
	if overlap {
		return "", repository.ErrTripOverlap
	}

	var tripID string
	err = tx.QueryRow(ctx, `
		INSERT INTO trips (vehicle_id, route_id, departure_time, arrival_time, base_fare, status)
		VALUES ($1, $2, $3, $4, $5, 'SCHEDULED')
		RETURNING id`, vehicleID, routeID, departure, arrival, baseFare).Scan(&tripID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign key violation
			return "", repository.ErrReferenceNotFound
		}
		return "", fmt.Errorf("failed to insert trip: %w", err)
	}

	numbers := make([]string, len(seats))
	classes := make([]string, len(seats))
	fares := make([]float64, len(seats))
	for i, s := range seats {
		numbers[i], classes[i], fares[i] = s.SeatNumber, s.DeckOrClass, s.Fare
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO trip_seat_inventory (trip_id, seat_number, deck_or_class, fare)
		SELECT $1::uuid, u.seat_number, u.deck_or_class, round(u.fare::numeric, 2)
		FROM unnest($2::text[], $3::text[], $4::float8[]) AS u(seat_number, deck_or_class, fare)`,
		tripID, numbers, classes, fares)
	if err != nil {
		return "", fmt.Errorf("failed to create seat inventory: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}
	return tripID, nil
}
