package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"ticket-katon-backend/internal/repository"
)

type cityRepository struct {
	db *pgxpool.Pool
}

func NewCityRepository(db *pgxpool.Pool) repository.CityRepository {
	return &cityRepository{db: db}
}

func (r *cityRepository) ListCities(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT city FROM (
			SELECT origin_city AS city FROM routes
			UNION
			SELECT destination_city FROM routes
		) c
		ORDER BY city`)
	if err != nil {
		return nil, fmt.Errorf("failed to list cities: %w", err)
	}
	defer rows.Close()

	cities := []string{}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, c)
	}
	return cities, rows.Err()
}
