package domain

import "time"

type UserRole string

const (
	RoleCustomer      UserRole = "CUSTOMER"
	RoleOperatorAdmin UserRole = "OPERATOR_ADMIN"
	RoleSuperAdmin    UserRole = "SUPER_ADMIN"
)

type User struct {
	ID           string    `json:"id"`
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SeatStatus string

const (
	SeatAvailable SeatStatus = "AVAILABLE"
	SeatHeld      SeatStatus = "HELD"
	SeatBooked    SeatStatus = "BOOKED"
)

type SeatInventory struct {
	ID        string     `json:"id"`
	TripID    string     `json:"trip_id"`
	SeatCode  string     `json:"seat_code"`
	Status    SeatStatus `json:"status"`
	Price     float64    `json:"price"`
	HeldBy    *string    `json:"held_by,omitempty"`
	HeldUntil *time.Time `json:"held_until,omitempty"`
}

type SeatMetadata struct {
	SeatNumber  string     `json:"seat_number"`
	DeckOrClass string     `json:"deck_or_class"`
	Fare        float64    `json:"fare"`
	Status      SeatStatus `json:"status"`
}

type TripSearchResult struct {
	ID             string    `json:"id"`
	OperatorName   string    `json:"operator_name"`
	RegistrationNo string    `json:"registration_no"`
	TransportType  string    `json:"transport_type"`
	Origin         string    `json:"origin"`
	Destination    string    `json:"destination"`
	DepartureTime  time.Time `json:"departure_time"`
	ArrivalTime    time.Time `json:"arrival_time"`
	BaseFare       float64   `json:"base_fare"`
	TotalSeats     int       `json:"total_seats"`
	AvailableSeats int       `json:"available_seats"`
}

type TripSeatLayoutResponse struct {
	TripID           string         `json:"trip_id"`
	OperatorName     string         `json:"operator_name"`
	RegistrationNo   string         `json:"registration_no"`
	TransportType    string         `json:"transport_type"`
	SeatLayoutConfig map[string]any `json:"seat_layout_config"`
	Seats            []SeatMetadata `json:"seats"`
}