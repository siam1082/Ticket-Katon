package domain

import "time"

// BookingSummary is one row of "My bookings": the booking plus the trip details
// the ticket card needs (so the frontend never has to guess them).
type BookingSummary struct {
	Booking
	TransportType  TransportType `json:"transport_type"`
	OperatorName   string        `json:"operator_name"`
	RegistrationNo string        `json:"registration_no"`
	Origin         string        `json:"origin"`
	Destination    string        `json:"destination"`
	DepartureTime  time.Time     `json:"departure_time"`
	ArrivalTime    time.Time     `json:"arrival_time"`
	Seats          []string      `json:"seats"`
	PaymentStatus  PaymentStatus `json:"payment_status"`
	RefundAmount   float64       `json:"refund_amount"`
}
