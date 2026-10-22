package repository

import (
	"context"
	"errors"
	"time"

	"ticket-katon-backend/internal/domain"
)

// Sentinel errors returned by BookingRepository so the service layer
// can react without depending on database driver details.
var (
	ErrSeatsUnavailable     = errors.New("one or more seats are already booked or invalid")
	ErrDuplicatePNR         = errors.New("pnr code already exists")
	ErrDuplicateTransaction = errors.New("payment transaction id already used")
	ErrTripOverlap          = errors.New("vehicle already has a trip in that time window")
	ErrReferenceNotFound    = errors.New("vehicle or route not found")
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type TripSearchParams struct {
	Origin        string
	Destination   string
	TravelDate    string // Format: YYYY-MM-DD
	TransportType string // BUS, LAUNCH, or empty
}

type TripRepository interface {
	SearchTrips(ctx context.Context, params TripSearchParams) ([]domain.TripSearchResult, error)
	GetTripSeatLayout(ctx context.Context, tripID string) (*domain.TripSeatLayoutResponse, error)
}

type BookingRepository interface {
	// CreateBookingWithSeats atomically books seats, issues tickets and records the payment.
	CreateBookingWithSeats(ctx context.Context, booking *domain.Booking, tickets []domain.Ticket, payment *domain.Payment, seatNumbers []string) error

	// GetBookingByPNR returns a booking and its tickets by PNR code.
	GetBookingByPNR(ctx context.Context, pnr string) (*domain.Booking, []domain.Ticket, error)

	// GetBookingsByUserID returns a user's booking history, newest first.
	GetBookingsByUserID(ctx context.Context, userID string) ([]domain.BookingSummary, error)

	// GetSeatsFare returns the total fare of the given seats.
	// It returns an error if any requested seat does not exist for the trip.
	GetSeatsFare(ctx context.Context, tripID string, seatNumbers []string) (float64, error)

	// CreatePendingBooking stores a PENDING booking, its tickets and an INITIATED payment.
	// Older PENDING bookings of the same user and trip are cancelled first.
	CreatePendingBooking(ctx context.Context, booking *domain.Booking, tickets []domain.Ticket, payment *domain.Payment) error

	// GetPaymentForUser returns a payment with its booking and tickets, only if the user owns it.
	GetPaymentForUser(ctx context.Context, paymentID, userID string) (*domain.Payment, *domain.Booking, []domain.Ticket, error)

	// ConfirmPendingPayment atomically marks seats BOOKED, the booking CONFIRMED and the payment SUCCESS.
	ConfirmPendingPayment(ctx context.Context, paymentID, userID string) (*domain.Payment, *domain.Booking, []domain.Ticket, error)

	// FailPendingPayment marks the payment FAILED and the booking CANCELLED. result is stored for the demo audit trail.
	FailPendingPayment(ctx context.Context, paymentID, userID, result string) (*domain.Payment, *domain.Booking, error)

	// ExpireStalePending cancels PENDING bookings older than the given age. Returns the number of payments failed.
	ExpireStalePending(ctx context.Context, olderThan time.Duration) (int64, error)
}

// AdminRepository serves the demo admin screens.
type AdminRepository interface {
	// GetOperatorIDForUser returns the operator a user belongs to (nil if none).
	GetOperatorIDForUser(ctx context.Context, userID string) (*string, error)
	GetVehicle(ctx context.Context, vehicleID string) (*domain.AdminVehicle, error)
	ListVehicles(ctx context.Context, operatorID *string) ([]domain.AdminVehicle, error)
	ListRoutes(ctx context.Context) ([]domain.AdminRoute, error)
	ListUpcomingTrips(ctx context.Context, operatorID *string, limit int) ([]domain.AdminTrip, error)
	// CreateTripWithSeats inserts a trip and all of its seat inventory in one transaction.
	CreateTripWithSeats(ctx context.Context, vehicleID, routeID string, departure, arrival time.Time, baseFare float64, seats []domain.SeatSpec) (string, error)
}
