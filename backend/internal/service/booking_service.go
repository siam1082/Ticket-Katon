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

// ১২ অক্ষরের ক্রিপ্টোগ্রাফিক ইউনিক PNR (TK- + 8 Hex chars = 11-12 chars, high entropy)
func generatePNR() (string, error) {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("TK-%s", strings.ToUpper(hex.EncodeToString(bytes))), nil
}

func generateTicketNumber(seatNumber string) (string, error) {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("TKT-%s-%s", seatNumber, strings.ToUpper(hex.EncodeToString(bytes))), nil
}

func (s *bookingService) ConfirmBooking(
	ctx context.Context,
	userID string,
	req *domain.ConfirmBookingRequest,
) (*domain.ConfirmBookingResponse, error) {
	// ০. ইনপুট ইনভ্যালিডেশন
	if len(req.Seats) == 0 {
		return nil, fmt.Errorf("at least one seat must be selected")
	}
	if len(req.Seats) != len(req.Passengers) {
		return nil, fmt.Errorf("seat count (%d) and passenger count (%d) mismatch", len(req.Seats), len(req.Passengers))
	}

	// সিট ভ্যালিডেশন: প্রতিটি প্যাসেঞ্জারের সিট সিলেক্টেড সিট লিস্টে আছে কিনা এবং ডুপ্লিকেট রোধ
	seatMap := make(map[string]bool)
	for _, seat := range req.Seats {
		if seatMap[seat] {
			return nil, fmt.Errorf("duplicate seat requested: %s", seat)
		}
		seatMap[seat] = true
	}

	for _, p := range req.Passengers {
		if !seatMap[p.SeatNumber] {
			return nil, fmt.Errorf("passenger assigned to unheld seat: %s", p.SeatNumber)
		}
	}

	// ১. Redis Seat Hold Validation (Logic Bug Fixed)
	for _, seat := range req.Seats {
		lockKey := fmt.Sprintf("seat_hold:%s:%s", req.TripID, seat)
		heldBy, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == redis.Nil {
			return nil, fmt.Errorf("hold on seat %s has expired", seat)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to verify seat hold: %w", err)
		}
		if heldBy != userID {
			return nil, fmt.Errorf("seat %s is held by another user", seat)
		}
	}

	// ২. সিট ফেয়ার হিসাব করা (লঞ্চ/বাসের সিট ভিত্তিক ডাইনামিক ফেয়ার)
	totalAmount, err := s.bookingRepo.GetSeatsFare(ctx, req.TripID, req.Seats)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch seat fares: %w", err)
	}

	bookingID := uuid.New().String()
	pnrCode, err := generatePNR()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PNR: %w", err)
	}

	booking := &domain.Booking{
		ID:             bookingID,
		UserID:         userID,
		TripID:         req.TripID,
		PNRCode:        pnrCode,
		TotalAmount:    totalAmount,
		DiscountAmount: 0.0,
		Status:         domain.BookingStatusConfirmed,
	}

	// ৩. টিকিট গঠন
	var tickets []domain.Ticket
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

	// ৪. পেমেন্ট রেকর্ড
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

	// ৫. PostgreSQL ACID ট্রানজেকশনে ডাটাবেজে এক্সিকিউট করা
	err = s.bookingRepo.CreateBookingWithSeats(ctx, booking, tickets, payment, req.Seats)
	if err != nil {
		return nil, fmt.Errorf("booking transaction failed: %w", err)
	}

	// ৬. বুকিং সফল হওয়ার পর Redis-এর টেম্পোরারি লক ক্লিয়ার করা
	for _, seat := range req.Seats {
		lockKey := fmt.Sprintf("seat_hold:%s:%s", req.TripID, seat)
		s.redisClient.Del(ctx, lockKey)
	}

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