package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
)

type BookingService interface {
	ConfirmBooking(ctx context.Context, userID string, req *domain.ConfirmBookingRequest) (*domain.ConfirmBookingResponse, error)
	GetBookingByPNR(ctx context.Context, pnr string) (*domain.Booking, []domain.Ticket, error)
	GetUserBookings(ctx context.Context, userID string) ([]domain.Booking, error)
}

type bookingService struct {
	bookingRepo repository.BookingRepository
	tripRepo    repository.TripRepository
	redisClient *redis.Client
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	tripRepo repository.TripRepository,
	redisClient *redis.Client,
) BookingService {
	return &bookingService{
		bookingRepo: bookingRepo,
		tripRepo:    tripRepo,
		redisClient: redisClient,
	}
}

func (s *bookingService) ConfirmBooking(ctx context.Context, userID string, req *domain.ConfirmBookingRequest) (*domain.ConfirmBookingResponse, error) {
	return nil, nil
}

func (s *bookingService) GetBookingByPNR(ctx context.Context, pnr string) (*domain.Booking, []domain.Ticket, error) {
	return nil, nil, nil
}

func (s *bookingService) GetUserBookings(ctx context.Context, userID string) ([]domain.Booking, error) {
	return nil, nil
}