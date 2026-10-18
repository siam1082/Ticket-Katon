package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// maxSeatsPerBooking limits how many seats one booking can contain. Adjust as needed.
	maxSeatsPerBooking = 6

	// minHoldTTL: if a seat hold has less than this left, treat it as expired,
	// so the hold can't lapse (and be taken by someone else) mid-transaction.
	minHoldTTL = 5 * time.Second

	// maxPNRAttempts: retries when a generated PNR collides with an existing one.
	maxPNRAttempts = 3
)

type BookingService interface {
	ConfirmBooking(ctx context.Context, userID string, req *domain.ConfirmBookingRequest) (*domain.ConfirmBookingResponse, error)
	GetBookingByPNR(ctx context.Context, pnr string) (*domain.Booking, []domain.Ticket, error)
	GetUserBookings(ctx context.Context, userID string) ([]domain.Booking, error)
}

type bookingService struct {
	bookingRepo repository.BookingRepository
	redisClient *redis.Client
}

func NewBookingService(
	bookingRepo repository.BookingRepository,
	redisClient *redis.Client,
) BookingService {
	return &bookingService{
		bookingRepo: bookingRepo,
		redisClient: redisClient,
	}
}

// generatePNR returns a PNR like "TK-A1B2C3D4E5" ("TK-" + 10 hex chars, 40 bits of randomness).
func generatePNR() (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "TK-" + strings.ToUpper(hex.EncodeToString(b)), nil
}

// generateTicketNumber returns a ticket number like "TKT-A1-A1B2C3D4E5".
func generateTicketNumber(seatNumber string) (string, error) {
	b := make([]byte, 5)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("TKT-%s-%s", seatNumber, strings.ToUpper(hex.EncodeToString(b))), nil
}

func seatHoldKey(tripID, seat string) string {
	return fmt.Sprintf("seat_hold:%s:%s", tripID, seat)
}

// validateConfirmRequest checks the request shape before touching Redis or the database.
func validateConfirmRequest(req *domain.ConfirmBookingRequest) error {
	if strings.TrimSpace(req.TripID) == "" {
		return domain.NewBookingError(domain.ErrBookingInvalid, "trip_id is required")
	}
	if len(req.Seats) == 0 {
		return domain.NewBookingError(domain.ErrBookingInvalid, "at least one seat must be selected")
	}
	if len(req.Seats) > maxSeatsPerBooking {
		return domain.NewBookingError(domain.ErrBookingInvalid, "a maximum of %d seats can be booked at once", maxSeatsPerBooking)
	}
	if len(req.Seats) != len(req.Passengers) {
		return domain.NewBookingError(domain.ErrBookingInvalid, "seat count (%d) and passenger count (%d) mismatch", len(req.Seats), len(req.Passengers))
	}

	// Seats must be unique
	seatSet := make(map[string]bool, len(req.Seats))
	for _, seat := range req.Seats {
		if strings.TrimSpace(seat) == "" {
			return domain.NewBookingError(domain.ErrBookingInvalid, "seat number cannot be empty")
		}
		if seatSet[seat] {
			return domain.NewBookingError(domain.ErrBookingInvalid, "duplicate seat requested: %s", seat)
		}
		seatSet[seat] = true
	}

	// Each passenger must map to exactly one requested seat
	assigned := make(map[string]bool, len(req.Passengers))
	for _, p := range req.Passengers {
		if !seatSet[p.SeatNumber] {
			return domain.NewBookingError(domain.ErrBookingInvalid, "passenger assigned to unrequested seat: %s", p.SeatNumber)
		}
		if assigned[p.SeatNumber] {
			return domain.NewBookingError(domain.ErrBookingInvalid, "more than one passenger assigned to seat %s", p.SeatNumber)
		}
		assigned[p.SeatNumber] = true

		if strings.TrimSpace(p.Name) == "" || strings.TrimSpace(p.Phone) == "" {
			return domain.NewBookingError(domain.ErrBookingInvalid, "passenger name and phone are required for seat %s", p.SeatNumber)
		}
	}

	// Payment info
	if !req.Payment.Gateway.IsValid() {
		return domain.NewBookingError(domain.ErrBookingInvalid, "unsupported payment gateway: %q", req.Payment.Gateway)
	}
	if strings.TrimSpace(req.Payment.TransactionID) == "" {
		return domain.NewBookingError(domain.ErrBookingInvalid, "payment transaction_id is required")
	}

	return nil
}

// verifyHolds makes sure every seat is held by this user and the hold won't expire imminently.
func (s *bookingService) verifyHolds(ctx context.Context, tripID, userID string, seats []string) error {
	for _, seat := range seats {
		key := seatHoldKey(tripID, seat)

		heldBy, err := s.redisClient.Get(ctx, key).Result()
		if err == redis.Nil {
			return domain.NewBookingError(domain.ErrHoldInvalid, "hold on seat %s has expired", seat)
		}
		if err != nil {
			return fmt.Errorf("failed to verify seat hold: %w", err)
		}
		if heldBy != userID {
			return domain.NewBookingError(domain.ErrHoldInvalid, "seat %s is held by another user", seat)
		}

		ttl, err := s.redisClient.TTL(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("failed to verify seat hold: %w", err)
		}
		// -2: key vanished between GET and TTL. -1: no expiry set (fine).
		if ttl == -2 || (ttl >= 0 && ttl < minHoldTTL) {
			return domain.NewBookingError(domain.ErrHoldInvalid, "hold on seat %s has expired", seat)
		}
	}
	return nil
}

func (s *bookingService) ConfirmBooking(
	ctx context.Context,
	userID string,
	req *domain.ConfirmBookingRequest,
) (*domain.ConfirmBookingResponse, error) {
	// 0. Input validation
	if err := validateConfirmRequest(req); err != nil {
		return nil, err
	}

	// 1. Redis seat hold validation
	if err := s.verifyHolds(ctx, req.TripID, userID, req.Seats); err != nil {
		return nil, err
	}

	// 2. Total fare (per-seat fare from the inventory)
	totalAmount, err := s.bookingRepo.GetSeatsFare(ctx, req.TripID, req.Seats)
	if err != nil {
		if errors.Is(err, repository.ErrSeatsUnavailable) {
			return nil, domain.NewBookingError(domain.ErrBookingInvalid, "one or more seats are invalid for this trip")
		}
		return nil, fmt.Errorf("failed to fetch seat fares: %w", err)
	}

	bookingID := uuid.New().String()

	booking := &domain.Booking{
		ID:             bookingID,
		UserID:         userID,
		TripID:         req.TripID,
		DiscountAmount: 0.0,
		TotalAmount:    totalAmount,
		Status:         domain.BookingStatusConfirmed,
	}

	// 3. Build tickets
	tickets := make([]domain.Ticket, 0, len(req.Passengers))
	for _, p := range req.Passengers {
		tktNum, err := generateTicketNumber(p.SeatNumber)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ticket number: %w", err)
		}
		tickets = append(tickets, domain.Ticket{
			ID:              uuid.New().String(),
			BookingID:       bookingID,
			SeatNumber:      p.SeatNumber,
			PassengerName:   p.Name,
			PassengerPhone:  p.Phone,
			PassengerGender: p.Gender,
			TicketNumber:    tktNum,
			CreatedAt:       time.Now(),
		})
	}

	// 4. Payment record
	payment := &domain.Payment{
		ID:               uuid.New().String(),
		BookingID:        bookingID,
		Gateway:          req.Payment.Gateway,
		TransactionID:    req.Payment.TransactionID,
		GatewayPaymentID: req.Payment.TransactionID,
		Amount:           totalAmount,
		Currency:         "BDT",
		Status:           domain.PaymentStatusSuccess,
	}

	// 5. PostgreSQL ACID transaction (retry only if the generated PNR collides)
	var txErr error
	for attempt := 0; attempt < maxPNRAttempts; attempt++ {
		pnr, err := generatePNR()
		if err != nil {
			return nil, fmt.Errorf("failed to generate PNR: %w", err)
		}
		booking.PNRCode = pnr

		txErr = s.bookingRepo.CreateBookingWithSeats(ctx, booking, tickets, payment, req.Seats)
		if errors.Is(txErr, repository.ErrDuplicatePNR) {
			continue
		}
		break
	}
	if txErr != nil {
		switch {
		case errors.Is(txErr, repository.ErrSeatsUnavailable):
			return nil, domain.NewBookingError(domain.ErrSeatsTaken, "one or more seats are already booked")
		case errors.Is(txErr, repository.ErrDuplicateTransaction):
			return nil, domain.NewBookingError(domain.ErrPaymentDuplicate, "this payment transaction has already been used")
		case errors.Is(txErr, repository.ErrDuplicatePNR):
			return nil, fmt.Errorf("could not generate a unique PNR, please try again")
		}
		return nil, fmt.Errorf("booking transaction failed: %w", txErr)
	}

	// 6. Booking succeeded: release the temporary Redis holds.
	// Best effort, they expire on their own via TTL anyway.
	keys := make([]string, 0, len(req.Seats))
	for _, seat := range req.Seats {
		keys = append(keys, seatHoldKey(req.TripID, seat))
	}
	s.redisClient.Del(ctx, keys...)

	return &domain.ConfirmBookingResponse{
		BookingID:      booking.ID,
		PNRCode:        booking.PNRCode,
		TripID:         booking.TripID,
		TotalAmount:    booking.TotalAmount,
		DiscountAmount: booking.DiscountAmount,
		Status:         booking.Status,
		Tickets:        tickets,
		PaymentStatus:  payment.Status,
	}, nil
}

func (s *bookingService) GetBookingByPNR(ctx context.Context, pnr string) (*domain.Booking, []domain.Ticket, error) {
	return s.bookingRepo.GetBookingByPNR(ctx, pnr)
}

func (s *bookingService) GetUserBookings(ctx context.Context, userID string) ([]domain.Booking, error) {
	return s.bookingRepo.GetBookingsByUserID(ctx, userID)
}
