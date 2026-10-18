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

// ১২ অক্ষরের ইউনিক ক্রিপ্টোগ্রাফিক PNR জেনারেটর
func generatePNR() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return fmt.Sprintf("TK-%s", strings.ToUpper(hex.EncodeToString(bytes)))
}

// টিকিট নম্বর জেনারেটর
func generateTicketNumber(seatNumber string) string {
	bytes := make([]byte, 3)
	rand.Read(bytes)
	return fmt.Sprintf("TKT-%s-%s", seatNumber, strings.ToUpper(hex.EncodeToString(bytes)))
}

func (s *bookingService) ConfirmBooking(
	ctx context.Context,
	userID string,
	req *domain.ConfirmBookingRequest,
) (*domain.ConfirmBookingResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("booking request is required")
	}

	// ০. ইনপুট ইনভ্যালিডেশন: সিট সংখ্যা এবং প্যাসেঞ্জার সংখ্যা সমান কিনা যাচাই
	if len(req.Seats) == 0 {
		return nil, fmt.Errorf("at least one seat must be selected")
	}
	if len(req.Seats) != len(req.Passengers) {
		return nil, fmt.Errorf("number of seats (%d) does not match passenger count (%d)", len(req.Seats), len(req.Passengers))
	}
	// ১. Redis Seat Hold Validation: নিশ্চিত করা সিটগুলো এই ইউজারের দখলেই আছে
	for _, seat := range req.Seats {
		lockKey := fmt.Sprintf("seat_hold:%s:%s", req.TripID, seat)
		heldBy, err := s.redisClient.Get(ctx, lockKey).Result()
		if err == redis.Nil || heldBy != userID {
			return nil, fmt.Errorf("hold on seat %s has expired or is invalid", seat)
		} else if err != nil {
			return nil, fmt.Errorf("failed to verify seat hold: %w", err)
		}
	}

	// ২. নির্বাচিত সিটগুলোর ফেয়ার সংগ্রহ করে টোটাল ক্যালকুলেট করা
	totalAmount, err := s.bookingRepo.GetSeatsFare(ctx, req.TripID, req.Seats)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate seats fare: %w", err)
	}

	bookingID := uuid.New().String()
	pnrCode := generatePNR()

	booking := &domain.Booking{
		ID:             bookingID,
		UserID:         userID,
		TripID:         req.TripID,
		PNRCode:        pnrCode,
		TotalAmount:    totalAmount,
		DiscountAmount: 0.0,
		Status:         domain.BookingStatusConfirmed,
	}

	// ৩. প্যাসেঞ্জার টিকিট স্ট্রাকচার রেডি করা
	var tickets []domain.Ticket
	for _, p := range req.Passengers {
		tickets = append(tickets, domain.Ticket{
			ID:              uuid.New().String(),
			BookingID:       bookingID,
			SeatNumber:      p.SeatNumber,
			PassengerName:   p.Name,
			PassengerPhone:  p.Phone,
			PassengerGender: p.Gender,
			TicketNumber:    generateTicketNumber(p.SeatNumber),
			CreatedAt:       time.Now(),
		})
	}

	// ৪. পেমেন্ট রেকর্ড গঠন
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

	// ৬. বুকিং কনফার্ম হওয়ার পর Redis-এর টেম্পোরারি লক রিলিজ করা
	for _, seat := range req.Seats {
		lockKey := fmt.Sprintf("seat_hold:%s:%s", req.TripID, seat)
		s.redisClient.Del(ctx, lockKey)
	}

	return &domain.ConfirmBookingResponse{
		BookingID:     booking.ID,
		PNRCode:       booking.PNRCode,
		TripID:        booking.TripID,
		TotalAmount:   booking.TotalAmount,
		Status:        booking.Status,
		Tickets:       tickets,
		PaymentStatus: payment.Status,
	}, nil
}

func (s *bookingService) GetBookingByPNR(ctx context.Context, pnr string) (*domain.Booking, []domain.Ticket, error) {
	return s.bookingRepo.GetBookingByPNR(ctx, pnr)
}

func (s *bookingService) GetUserBookings(ctx context.Context, userID string) ([]domain.Booking, error) {
	return s.bookingRepo.GetBookingsByUserID(ctx, userID)
}
