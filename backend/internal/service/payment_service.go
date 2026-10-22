package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"ticket-katon-backend/internal/domain"
	"ticket-katon-backend/internal/repository"
)

const (
	// A pending booking older than the hold time (plus a minute of slack) can never be paid.
	pendingMaxAge = LockTTL + time.Minute

	demoNotice = "Demo payment simulation. No real money is transferred."
)

func ticketSeatNumbers(tickets []domain.Ticket) []string {
	seats := make([]string, 0, len(tickets))
	for _, t := range tickets {
		seats = append(seats, t.SeatNumber)
	}
	return seats
}

// releaseHolds frees the Redis holds, but only the ones this user still owns.
func (s *bookingService) releaseHolds(ctx context.Context, tripID, userID string, seats []string) {
	if len(seats) == 0 {
		return
	}
	keys := make([]string, 0, len(seats))
	for _, seat := range seats {
		keys = append(keys, seatHoldKey(tripID, seat))
	}
	if err := s.redisClient.Eval(ctx, releaseLockLua, keys, userID).Err(); err != nil {
		log.Printf("[PAYMENT] could not release holds for trip %s: %v", tripID, err)
	}
}

// InitiatePayment creates a PENDING booking and an INITIATED payment for seats the user holds.
// The amount comes from the database, never from the client.
func (s *bookingService) InitiatePayment(
	ctx context.Context,
	userID string,
	req *domain.InitiatePaymentRequest,
) (*domain.InitiatePaymentResponse, error) {
	// Same request checks as a direct booking; the transaction id is generated below.
	if err := validateConfirmRequest(&domain.ConfirmBookingRequest{
		TripID:     req.TripID,
		Seats:      req.Seats,
		Passengers: req.Passengers,
		Payment:    domain.PaymentPayload{Gateway: req.Gateway, TransactionID: "server-generated"},
	}); err != nil {
		return nil, err
	}

	remaining, err := s.verifyHolds(ctx, req.TripID, userID, req.Seats)
	if err != nil {
		return nil, err
	}

	amount, err := s.bookingRepo.GetSeatsFare(ctx, req.TripID, req.Seats)
	if err != nil {
		if errors.Is(err, repository.ErrSeatsUnavailable) {
			return nil, domain.NewBookingError(domain.ErrSeatsTaken,
				"one or more seats are not available (already booked, unknown, or the trip has departed)")
		}
		return nil, fmt.Errorf("failed to fetch seat fares: %w", err)
	}

	bookingID := uuid.New().String()
	paymentID := uuid.New().String()

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
		})
	}

	booking := &domain.Booking{
		ID:          bookingID,
		UserID:      userID,
		TripID:      req.TripID,
		TotalAmount: amount,
		Status:      domain.BookingStatusPending,
	}

	txnID := "SIM-" + strings.ToUpper(strings.ReplaceAll(uuid.New().String(), "-", ""))
	payment := &domain.Payment{
		ID:               paymentID,
		BookingID:        bookingID,
		Gateway:          req.Gateway,
		TransactionID:    txnID,
		GatewayPaymentID: txnID,
		Amount:           amount,
		Currency:         "BDT",
		Status:           domain.PaymentStatusInitiated,
	}

	var txErr error
	for attempt := 0; attempt < maxPNRAttempts; attempt++ {
		pnr, err := generatePNR()
		if err != nil {
			return nil, fmt.Errorf("failed to generate PNR: %w", err)
		}
		booking.PNRCode = pnr

		txErr = s.bookingRepo.CreatePendingBooking(ctx, booking, tickets, payment)
		if errors.Is(txErr, repository.ErrDuplicatePNR) {
			continue
		}
		break
	}
	if txErr != nil {
		if errors.Is(txErr, repository.ErrDuplicatePNR) {
			return nil, fmt.Errorf("could not generate a unique PNR, please try again")
		}
		return nil, fmt.Errorf("failed to create pending booking: %w", txErr)
	}

	expires := int(remaining.Seconds())
	if expires <= 0 {
		expires = int(LockTTL.Seconds())
	}

	return &domain.InitiatePaymentResponse{
		PaymentID:        payment.ID,
		BookingID:        booking.ID,
		Amount:           amount,
		Currency:         payment.Currency,
		Gateway:          payment.Gateway,
		Status:           payment.Status,
		ExpiresInSeconds: expires,
		Demo:             true,
		Notice:           demoNotice,
	}, nil
}

func (s *bookingService) successResponse(
	p *domain.Payment,
	b *domain.Booking,
	tickets []domain.Ticket,
) *domain.SimulatePaymentResponse {
	return &domain.SimulatePaymentResponse{
		PaymentID:     p.ID,
		Result:        "success",
		PaymentStatus: p.Status,
		BookingStatus: b.Status,
		Message:       "Demo payment successful. Your booking is confirmed.",
		Booking: &domain.ConfirmBookingResponse{
			BookingID:      b.ID,
			PNRCode:        b.PNRCode,
			TripID:         b.TripID,
			TotalAmount:    b.TotalAmount,
			DiscountAmount: b.DiscountAmount,
			Status:         b.Status,
			Tickets:        tickets,
			PaymentStatus:  p.Status,
		},
	}
}

// SimulatePayment applies the demo result (success, fail or cancel) to a pending payment.
func (s *bookingService) SimulatePayment(
	ctx context.Context,
	userID, paymentID, result string,
) (*domain.SimulatePaymentResponse, error) {
	result = strings.ToLower(strings.TrimSpace(result))
	if result != "success" && result != "fail" && result != "cancel" {
		return nil, domain.NewBookingError(domain.ErrBookingInvalid, "result must be success, fail or cancel")
	}
	if _, err := uuid.Parse(paymentID); err != nil {
		return nil, domain.NewBookingError(domain.ErrPaymentNotFound, "payment not found")
	}

	payment, booking, tickets, err := s.bookingRepo.GetPaymentForUser(ctx, paymentID, userID)
	if err != nil {
		return nil, err
	}
	seats := ticketSeatNumbers(tickets)

	// ---- fail / cancel: release everything
	if result != "success" {
		p, b, err := s.bookingRepo.FailPendingPayment(ctx, paymentID, userID, result)
		if err != nil {
			return nil, err
		}
		s.releaseHolds(ctx, booking.TripID, userID, seats)

		msg := "Demo payment failed. Your seats have been released."
		if result == "cancel" {
			msg = "Demo payment cancelled. Your seats have been released."
		}
		return &domain.SimulatePaymentResponse{
			PaymentID:     p.ID,
			Result:        result,
			PaymentStatus: p.Status,
			BookingStatus: b.Status,
			Message:       msg,
		}, nil
	}

	// ---- success
	if payment.Status == domain.PaymentStatusSuccess && booking.Status == domain.BookingStatusConfirmed {
		return s.successResponse(payment, booking, tickets), nil // repeated click: same answer
	}
	if payment.Status != domain.PaymentStatusInitiated {
		return nil, domain.NewBookingError(domain.ErrPaymentProcessed,
			"this payment was already %s", strings.ToLower(string(payment.Status)))
	}

	if _, err := s.verifyHolds(ctx, booking.TripID, userID, seats); err != nil {
		if errors.Is(err, domain.ErrHoldInvalid) {
			// The hold lapsed, so this payment can never complete.
			if _, _, ferr := s.bookingRepo.FailPendingPayment(ctx, paymentID, userID, "expired"); ferr != nil {
				log.Printf("[PAYMENT] could not mark payment %s expired: %v", paymentID, ferr)
			}
		}
		return nil, err
	}

	p, b, tix, err := s.bookingRepo.ConfirmPendingPayment(ctx, paymentID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrSeatsUnavailable) {
			if _, _, ferr := s.bookingRepo.FailPendingPayment(ctx, paymentID, userID, "seats_unavailable"); ferr != nil {
				log.Printf("[PAYMENT] could not mark payment %s failed: %v", paymentID, ferr)
			}
			s.releaseHolds(ctx, booking.TripID, userID, seats)
			return nil, domain.NewBookingError(domain.ErrSeatsTaken, "one or more seats are no longer available")
		}
		return nil, err
	}

	s.releaseHolds(ctx, b.TripID, userID, seats)
	return s.successResponse(p, b, tix), nil
}

// GetPayment returns the current state of a payment the user owns.
func (s *bookingService) GetPayment(ctx context.Context, userID, paymentID string) (*domain.SimulatePaymentResponse, error) {
	if _, err := uuid.Parse(paymentID); err != nil {
		return nil, domain.NewBookingError(domain.ErrPaymentNotFound, "payment not found")
	}

	p, b, tickets, err := s.bookingRepo.GetPaymentForUser(ctx, paymentID, userID)
	if err != nil {
		return nil, err
	}

	if p.Status == domain.PaymentStatusSuccess && b.Status == domain.BookingStatusConfirmed {
		return s.successResponse(p, b, tickets), nil
	}

	msg := "Waiting for the demo payment result."
	if p.Status == domain.PaymentStatusFailed {
		msg = "This demo payment did not complete. The seats were released."
	}
	return &domain.SimulatePaymentResponse{
		PaymentID:     p.ID,
		PaymentStatus: p.Status,
		BookingStatus: b.Status,
		Message:       msg,
	}, nil
}

// StartPendingExpiryWorker cancels abandoned PENDING bookings once a minute until ctx ends.
func (s *bookingService) StartPendingExpiryWorker(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n, err := s.bookingRepo.ExpireStalePending(ctx, pendingMaxAge)
				if err != nil {
					log.Printf("[PAYMENT] expiry worker: %v", err)
				} else if n > 0 {
					log.Printf("[PAYMENT] expiry worker: cancelled %d abandoned payment(s)", n)
				}
			}
		}
	}()
}
