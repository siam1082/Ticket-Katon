package domain

// InitiatePaymentRequest starts a demo payment for seats the user currently holds.
// The amount is always computed by the server; any amount from the client is ignored.
type InitiatePaymentRequest struct {
	TripID     string          `json:"trip_id"`
	Seats      []string        `json:"seats"`
	Passengers []PassengerInfo `json:"passengers"`
	Gateway    PaymentGateway  `json:"gateway"`
}

type InitiatePaymentResponse struct {
	PaymentID        string         `json:"payment_id"`
	BookingID        string         `json:"booking_id"`
	Amount           float64        `json:"amount"`
	Currency         string         `json:"currency"`
	Gateway          PaymentGateway `json:"gateway"`
	Status           PaymentStatus  `json:"status"`
	ExpiresInSeconds int            `json:"expires_in_seconds"`
	Demo             bool           `json:"demo"`
	Notice           string         `json:"notice"`
}

// SimulatePaymentRequest.Result is one of: success, fail, cancel.
type SimulatePaymentRequest struct {
	Result string `json:"result"`
}

type SimulatePaymentResponse struct {
	PaymentID     string                  `json:"payment_id"`
	Result        string                  `json:"result"`
	PaymentStatus PaymentStatus           `json:"payment_status"`
	BookingStatus BookingStatus           `json:"booking_status"`
	Message       string                  `json:"message"`
	Booking       *ConfirmBookingResponse `json:"booking,omitempty"` // only after a successful payment
}
