package domain

import "time"

type AdminVehicle struct {
	ID               string         `json:"id"`
	OperatorID       string         `json:"operator_id"`
	OperatorName     string         `json:"operator_name"`
	RegistrationNo   string         `json:"registration_no"`
	Type             TransportType  `json:"type"`
	SeatLayoutConfig map[string]any `json:"seat_layout_config"`
}

type AdminRoute struct {
	ID              string `json:"id"`
	OriginCity      string `json:"origin_city"`
	DestinationCity string `json:"destination_city"`
}

type AdminTrip struct {
	ID             string        `json:"id"`
	VehicleID      string        `json:"vehicle_id"`
	RegistrationNo string        `json:"registration_no"`
	OperatorName   string        `json:"operator_name"`
	TransportType  TransportType `json:"transport_type"`
	Origin         string        `json:"origin"`
	Destination    string        `json:"destination"`
	DepartureTime  time.Time     `json:"departure_time"`
	ArrivalTime    time.Time     `json:"arrival_time"`
	BaseFare       float64       `json:"base_fare"`
	Status         string        `json:"status"`
	TotalSeats     int           `json:"total_seats"`
	BookedSeats    int           `json:"booked_seats"`
}

// CreateTripRequest times accept RFC3339 or "YYYY-MM-DDTHH:MM" (read as Bangladesh time).
type CreateTripRequest struct {
	VehicleID     string  `json:"vehicle_id"`
	RouteID       string  `json:"route_id"`
	DepartureTime string  `json:"departure_time"`
	ArrivalTime   string  `json:"arrival_time"`
	BaseFare      float64 `json:"base_fare"`
}

type CreateTripResponse struct {
	TripID       string `json:"trip_id"`
	SeatsCreated int    `json:"seats_created"`
}

// SeatSpec is one row to insert into trip_seat_inventory.
type SeatSpec struct {
	SeatNumber  string
	DeckOrClass string
	Fare        float64
}

type AdminOperator struct {
	ID   string        `json:"id"`
	Name string        `json:"name"`
	Type TransportType `json:"type"`
}

type CreateRouteRequest struct {
	OriginCity      string `json:"origin_city"`
	DestinationCity string `json:"destination_city"`
}

// CoachConfig describes one train coach.
type CoachConfig struct {
	CoachName  string `json:"coach_name"`
	Class      string `json:"class"`  // SNIGDHA or SHOVON_CHAIR
	Layout     string `json:"layout"` // 2x2 or 2x3 (display only)
	TotalSeats int    `json:"total_seats"`
}

// CreateVehicleRequest: the operator admin's own operator is used automatically;
// a super admin must send operator_id. The vehicle type must match the operator's type.
type CreateVehicleRequest struct {
	OperatorID     string        `json:"operator_id"`
	RegistrationNo string        `json:"registration_no"`
	Type           TransportType `json:"type"`
	BusLayout      string        `json:"bus_layout"` // BUS only: 2x2 or 2x1
	Coaches        []CoachConfig `json:"coaches"`    // TRAIN only
}

type CreateOperatorRequest struct {
	Name         string        `json:"name"`
	Type         TransportType `json:"type"`
	ContactPhone string        `json:"contact_phone"`
}

// CreateOperatorAdminRequest either promotes an existing passenger (matched by email)
// or creates a new account; full_name, phone and password are only needed for a new account.
type CreateOperatorAdminRequest struct {
	OperatorID string `json:"operator_id"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	Phone      string `json:"phone"`
	Password   string `json:"password"`
}

type AdminOperatorAdmin struct {
	UserID       string `json:"user_id"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	OperatorID   string `json:"operator_id"`
	OperatorName string `json:"operator_name"`
}

type OperatorAdminResult struct {
	Mode   string `json:"mode"` // "created" or "promoted"
	UserID string `json:"user_id"`
	Notice string `json:"notice"`
}
