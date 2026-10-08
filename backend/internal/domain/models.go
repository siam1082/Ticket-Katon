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