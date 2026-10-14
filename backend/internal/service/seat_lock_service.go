package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrSeatsAlreadyHeld  = errors.New("one or more selected seats are currently locked or unavailable")
	ErrLockReleaseFailed = errors.New("failed to release locks or unauthorized owner")
)

// 5 minutes lock TTL (300 seconds)
const LockTTL = 5 * time.Minute

type SeatLockService struct {
	rdb *redis.Client
}

func NewSeatLockService(rdb *redis.Client) *SeatLockService {
	return &SeatLockService{rdb: rdb}
}

// formatSeatKey standardizes the Redis key pattern: lock:trip:{trip_id}:seat:{seat_number}
func formatSeatKey(tripID string, seatNumber string) string {
	return fmt.Sprintf("lock:trip:%s:seat:%s", tripID, seatNumber)
}

// HoldSeats atomically locks all requested seats for a user for 5 minutes.
func (s *SeatLockService) HoldSeats(ctx context.Context, tripID string, seatNumbers []string, userID string) error {
	if len(seatNumbers) == 0 {
		return errors.New("no seats provided to lock")
	}

	keys := make([]string, len(seatNumbers))
	for i, seat := range seatNumbers {
		keys[i] = formatSeatKey(tripID, seat)
	}

	ttlSeconds := int(LockTTL.Seconds())

	result, err := s.rdb.Eval(ctx, acquireLockLua, keys, ttlSeconds, userID).Result()
	if err != nil {
		return fmt.Errorf("redis eval error: %w", err)
	}

	success, ok := result.(int64)
	if !ok || success != 1 {
		return ErrSeatsAlreadyHeld
	}

	return nil
}

// ReleaseSeats frees the seats if the requesting user is the one who holds them.
func (s *SeatLockService) ReleaseSeats(ctx context.Context, tripID string, seatNumbers []string, userID string) error {
	if len(seatNumbers) == 0 {
		return nil
	}

	keys := make([]string, len(seatNumbers))
	for i, seat := range seatNumbers {
		keys[i] = formatSeatKey(tripID, seat)
	}

	_, err := s.rdb.Eval(ctx, releaseLockLua, keys, userID).Result()
	if err != nil {
		return fmt.Errorf("failed to release locks: %w", err)
	}

	return nil
}

// ValidateHold checks if all requested seats are currently held by the specified user.
// Used right before executing the Step 6 database transaction.
func (s *SeatLockService) ValidateHold(ctx context.Context, tripID string, seatNumbers []string, userID string) (bool, error) {
	pipe := s.rdb.Pipeline()
	cmds := make([]*redis.StringCmd, len(seatNumbers))

	for i, seat := range seatNumbers {
		cmds[i] = pipe.Get(ctx, formatSeatKey(tripID, seat))
	}

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, err
	}

	for _, cmd := range cmds {
		val, err := cmd.Result()
		if err != nil || val != userID {
			return false, nil // Missing lock or held by another user
		}
	}

	return true, nil
}