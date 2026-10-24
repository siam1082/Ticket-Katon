package domain

import "time"

// CancellationPreview tells the passenger what cancelling would do, before they confirm.
type CancellationPreview struct {
	PNRCode       string    `json:"pnr_code"`
	CanCancel     bool      `json:"can_cancel"`
	Reason        string    `json:"reason,omitempty"`
	TotalAmount   float64   `json:"total_amount"`
	RefundPercent int       `json:"refund_percent"`
	RefundAmount  float64   `json:"refund_amount"`
	DepartureTime time.Time `json:"departure_time"`
	Policy        []string  `json:"policy"`
}

type CancellationResult struct {
	PNRCode       string        `json:"pnr_code"`
	BookingStatus BookingStatus `json:"booking_status"`
	PaymentStatus PaymentStatus `json:"payment_status"`
	RefundPercent int           `json:"refund_percent"`
	RefundAmount  float64       `json:"refund_amount"`
	Message       string        `json:"message"`
}

// CitySuggestion is one entry of the city autocomplete.
type CitySuggestion struct {
	Name  string `json:"name"`
	Match string `json:"match"` // exact, prefix, contains, alias or similar
}
