package domain

import (
	"errors"
	"fmt"
)

// Booking error kinds. The HTTP handler maps these to status codes.
var (
	ErrBookingInvalid   = errors.New("invalid booking request")       // 400
	ErrHoldInvalid      = errors.New("seat hold expired or invalid")  // 409
	ErrSeatsTaken       = errors.New("seats already booked")          // 409
	ErrPaymentDuplicate = errors.New("duplicate payment transaction") // 409
	ErrBookingNotFound  = errors.New("booking not found")             // 404
	ErrPaymentNotFound  = errors.New("payment not found")             // 404
	ErrPaymentProcessed = errors.New("payment already processed")     // 409
)

// BookingError carries a client-safe message and the kind used for errors.Is checks.
type BookingError struct {
	Kind error
	Msg  string
}

func (e *BookingError) Error() string { return e.Msg }
func (e *BookingError) Unwrap() error { return e.Kind }

// NewBookingError builds a BookingError of the given kind with a formatted message.
func NewBookingError(kind error, format string, args ...any) error {
	return &BookingError{Kind: kind, Msg: fmt.Sprintf(format, args...)}
}
