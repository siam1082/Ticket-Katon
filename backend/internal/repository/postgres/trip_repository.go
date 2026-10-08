package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
)

type PostgresTripRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTripRepository(pool *pgxpool.Pool) *PostgresTripRepository {
	return &PostgresTripRepository{pool: pool}
}

func (r *PostgresTripRepository) SearchTrips(ctx context.Context, params repository.TripSearchParams) ([]domain.TripSearchResult, error) {
	query := `
		SELECT 
			t.id,
			o.name AS operator_name,
			v.registration_no,
			v.type::text AS transport_type,
			r.origin_city AS origin,
			r.destination_city AS destination,
			t.departure_time,
			t.arrival_time,
			t.base_fare,
			COUNT(tsi.id) AS total_seats,
			COUNT(tsi.id) FILTER (
				WHERE tsi.status::text = 'AVAILABLE' 
				   OR (tsi.status::text = 'HELD' AND tsi.held_until IS NOT NULL AND tsi.held_until < NOW())
			) AS available_seats
		FROM trips t
		JOIN routes r ON r.id = t.route_id
		JOIN vehicles v ON v.id = t.vehicle_id
		JOIN operators o ON o.id = v.operator_id
		LEFT JOIN trip_seat_inventory tsi ON tsi.trip_id = t.id
		WHERE LOWER(r.origin_city) = LOWER($1)
		  AND LOWER(r.destination_city) = LOWER($2)
		  AND DATE(t.departure_time) = $3::date
		  AND ($4 = '' OR LOWER(v.type::text) = LOWER($4))
		GROUP BY t.id, o.name, v.registration_no, v.type, r.origin_city, r.destination_city
		ORDER BY t.departure_time ASC
	`

	rows, err := r.pool.Query(ctx, query, params.Origin, params.Destination, params.TravelDate, params.TransportType)
	if err != nil {
		return nil, fmt.Errorf("search trips query: %w", err)
	}
	defer rows.Close()

	results := make([]domain.TripSearchResult, 0)
	for rows.Next() {
		var res domain.TripSearchResult
		err := rows.Scan(
			&res.ID,
			&res.OperatorName,
			&res.RegistrationNo,
			&res.TransportType,
			&res.Origin,
			&res.Destination,
			&res.DepartureTime,
			&res.ArrivalTime,
			&res.BaseFare,
			&res.TotalSeats,
			&res.AvailableSeats,
		)
		if err != nil {
			return nil, fmt.Errorf("scan trip row: %w", err)
		}
		results = append(results, res)
	}

	return results, nil
}

func (r *PostgresTripRepository) GetTripSeatLayout(ctx context.Context, tripID string) (*domain.TripSeatLayoutResponse, error) {
	queryTrip := `
		SELECT 
			t.id,
			o.name,
			v.registration_no,
			v.type::text,
			v.seat_layout_config
		FROM trips t
		JOIN vehicles v ON v.id = t.vehicle_id
		JOIN operators o ON o.id = v.operator_id
		WHERE t.id = $1
	`

	var resp domain.TripSeatLayoutResponse
	var rawLayout []byte

	err := r.pool.QueryRow(ctx, queryTrip, tripID).Scan(
		&resp.TripID,
		&resp.OperatorName,
		&resp.RegistrationNo,
		&resp.TransportType,
		&rawLayout,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get trip layout info: %w", err)
	}

	if len(rawLayout) > 0 {
		_ = json.Unmarshal(rawLayout, &resp.SeatLayoutConfig)
	}

	querySeats := `
		SELECT 
			seat_number, 
			deck_or_class, 
			fare, 
			CASE 
				WHEN status::text = 'HELD' AND held_until IS NOT NULL AND held_until < NOW() THEN 'AVAILABLE'
				ELSE status::text 
			END AS effective_status
		FROM trip_seat_inventory
		WHERE trip_id = $1
		ORDER BY seat_number ASC
	`
	rows, err := r.pool.Query(ctx, querySeats, tripID)
	if err != nil {
		return nil, fmt.Errorf("get seats query: %w", err)
	}
	defer rows.Close()

	resp.Seats = make([]domain.SeatMetadata, 0)
	for rows.Next() {
		var s domain.SeatMetadata
		var statusStr string
		if err := rows.Scan(&s.SeatNumber, &s.DeckOrClass, &s.Fare, &statusStr); err != nil {
			return nil, fmt.Errorf("scan seat row: %w", err)
		}
		s.Status = domain.SeatStatus(statusStr)
		resp.Seats = append(resp.Seats, s)
	}

	return &resp, nil
}