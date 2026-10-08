package service

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthService struct {
	pool *pgxpool.Pool
	rdb  *redis.Client
}

func NewHealthService(pool *pgxpool.Pool, rdb *redis.Client) *HealthService {
	return &HealthService{pool: pool, rdb: rdb}
}

func (s *HealthService) Check(ctx context.Context) map[string]string {
	res := map[string]string{
		"status":   "UP",
		"postgres": "OK",
		"redis":    "OK",
	}

	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := s.pool.Ping(checkCtx); err != nil {
		res["postgres"] = "DOWN: " + err.Error()
		res["status"] = "DEGRADED"
	}

	if err := s.rdb.Ping(checkCtx).Err(); err != nil {
		res["redis"] = "DOWN: " + err.Error()
		res["status"] = "DEGRADED"
	}

	return res
}