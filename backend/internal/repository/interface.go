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
	ErrAlreadyExists        = errors.New("already exists")
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)

	// Email verification
	MarkEmailVerified(ctx context.Context, userID string) error
	// CreateEmailVerification replaces any older code of the user with this one.
	CreateEmailVerification(ctx context.Context, userID, codeHash string, expiresAt time.Time) error
	GetLatestEmailVerification(ctx context.Context, userID string) (*EmailVerification, error)
	IncrementVerificationAttempts(ctx context.Context, verificationID string) error
	DeleteEmailVerifications(ctx context.Context, userID string) error
}

// EmailVerification is a pending email code (only a hash of the code is stored).
type EmailVerification struct {
	ID        string
	UserID    string
	CodeHash  string
	ExpiresAt time.Time
	CreatedAt time.Time
	Attempts  int
}

// CityRepository serves the city autocomplete.
type CityRepository interface {
	// ListCities returns every distinct city used by a route (as stored).
	ListCities(ctx context.Context) ([]string, error)
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

	// GetCancellationContext reads what the cancellation preview needs (only for the booking's owner).
	GetCancellationContext(ctx context.Context, pnr, userID string) (*CancellationContext, error)
	// CancelConfirmedBooking cancels a paid booking and frees its seats in one transaction.
	// The decider runs while the booking is locked and returns the refund, or an error to refuse.
	CancelConfirmedBooking(ctx context.Context, pnr, userID string, decide RefundDecider) (*CancelOutcome, error)
}

// RefundDecider decides the refund for a cancellation, or refuses it with an error.
type RefundDecider func(totalAmount float64, departure time.Time) (percent int, amount float64, err error)

type CancellationContext struct {
	PNRCode     string
	Status      domain.BookingStatus
	TotalAmount float64
	Departure   time.Time
}

type CancelOutcome struct {
	BookingStatus domain.BookingStatus
	PaymentStatus domain.PaymentStatus
	RefundPercent int
	RefundAmount  float64
}

// AdminUserRef is the little the admin screens need to know about an existing account.
type AdminUserRef struct {
	ID         string
	Role       domain.UserRole
	OperatorID *string
}

// AdminRepository serves the demo admin screens.
type AdminRepository interface {
	// GetOperatorIDForUser returns the operator a user belongs to (nil if none).
	GetOperatorIDForUser(ctx context.Context, userID string) (*string, error)
	GetVehicle(ctx context.Context, vehicleID string) (*domain.AdminVehicle, error)
	ListVehicles(ctx context.Context, operatorID *string) ([]domain.AdminVehicle, error)
	ListRoutes(ctx context.Context) ([]domain.AdminRoute, error)
	ListUpcomingTrips(ctx context.Context, operatorID *string, limit int) ([]domain.AdminTrip, error)
	ListOperators(ctx context.Context, operatorID *string) ([]domain.AdminOperator, error)
	GetOperator(ctx context.Context, operatorID string) (*domain.AdminOperator, error)
	// CreateRoute returns ErrAlreadyExists if the same origin and destination (any letter case) exists.
	CreateRoute(ctx context.Context, origin, destination string) (*domain.AdminRoute, error)
	// CreateVehicle returns the new vehicle id, or ErrAlreadyExists if the registration number is taken.
	CreateVehicle(ctx context.Context, operatorID, registrationNo string, vtype domain.TransportType, layoutJSON string) (string, error)
	// CreateOperator returns ErrAlreadyExists if an operator with the same name exists.
	CreateOperator(ctx context.Context, name string, otype domain.TransportType, contactPhone string) (*domain.AdminOperator, error)
	ListOperatorAdmins(ctx context.Context) ([]domain.AdminOperatorAdmin, error)
	// FindUserForAdmin returns domain.ErrNotFound when no account has this email.
	FindUserForAdmin(ctx context.Context, email string) (*AdminUserRef, error)
	// PromoteToOperatorAdmin turns a CUSTOMER into an operator admin; ErrAlreadyExists if they are not a CUSTOMER.
	PromoteToOperatorAdmin(ctx context.Context, userID, operatorID string) error
	// CreateOperatorAdminUser creates a verified OPERATOR_ADMIN account; ErrAlreadyExists on duplicate email or phone.
	CreateOperatorAdminUser(ctx context.Context, operatorID, fullName, email, phone, passwordHash string) (string, error)
	// DemoteOperatorAdmin turns an operator admin back into a CUSTOMER.
	DemoteOperatorAdmin(ctx context.Context, userID string) error
	// CreateTripWithSeats inserts a trip and all of its seat inventory in one transaction.
	CreateTripWithSeats(ctx context.Context, vehicleID, routeID string, departure, arrival time.Time, baseFare float64, seats []domain.SeatSpec) (string, error)
}
