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
