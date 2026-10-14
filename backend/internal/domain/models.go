package domain

import "time"

// -----------------------------------------------------------------------------
// USER DOMAIN
// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------
// TRANSPORT & VEHICLE DOMAIN
// -----------------------------------------------------------------------------

type TransportType string

const (
	TransportBus    TransportType = "BUS"
	TransportLaunch TransportType = "LAUNCH"
	TransportTrain  TransportType = "TRAIN"
)

// Specific coach, seat, and cabin comfort classes
type SeatClass string

const (
	// Bus Classes
	SeatClassBusACChair    SeatClass = "BUS_AC_CHAIR"     // Bus AC Chair (নতুন যোগ করা হলো)
	SeatClassBusNonACChair SeatClass = "BUS_NON_AC_CHAIR" // Bus Non-AC Chair
	SeatClassBusSleeper    SeatClass = "BUS_SLEEPER"      // Bus Sleeper Coach

	// Launch Classes
	SeatClassLaunchDeck        SeatClass = "LAUNCH_DECK"
	SeatClassLaunchSingleCabin SeatClass = "LAUNCH_CABIN_SINGLE"
	SeatClassLaunchDoubleCabin SeatClass = "LAUNCH_CABIN_DOUBLE"
	SeatClassLaunchVIP         SeatClass = "LAUNCH_VIP_CABIN"

	// Train Classes (Bangladesh Railway Standard)
	SeatClassTrainShovonChair SeatClass = "TRAIN_SHOVON_CHAIR" // Train Shovon Chair
	SeatClassTrainSnigdha     SeatClass = "TRAIN_SNIGDHA"      // Train Snigdha (AC Chair)
	SeatClassTrainCabin       SeatClass = "TRAIN_CABIN_BERTH"  // Train Cabin / AC Berth
)

// -----------------------------------------------------------------------------
// SEAT & INVENTORY DOMAIN
// -----------------------------------------------------------------------------

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
	ClassType SeatClass  `json:"class_type"` // e.g. BUS_AC_CHAIR, TRAIN_SNIGDHA, TRAIN_CABIN_BERTH
	Status    SeatStatus `json:"status"`
	Price     float64    `json:"price"`
	HeldBy    *string    `json:"held_by,omitempty"`
	HeldUntil *time.Time `json:"held_until,omitempty"`
}

type SeatMetadata struct {
	SeatNumber  string     `json:"seat_number"`   // e.g. "A1", "KA-12", "CABIN-204"
	DeckOrCoach string     `json:"deck_or_coach"` // e.g. "Coach KA", "Floor 2", "Upper Deck"
	ClassType   SeatClass  `json:"class_type"`    // e.g. BUS_AC_CHAIR, TRAIN_SNIGDHA, TRAIN_CABIN_BERTH
	Fare        float64    `json:"fare"`
	Status      SeatStatus `json:"status"`
}

// -----------------------------------------------------------------------------
// TRIP SEARCH & LAYOUT RESPONSES
// -----------------------------------------------------------------------------

type TripSearchResult struct {
	ID             string        `json:"id"`
	OperatorName   string        `json:"operator_name"`
	RegistrationNo string        `json:"registration_no"`
	TransportType  TransportType `json:"transport_type"` // BUS | LAUNCH | TRAIN
	CategoryName   string        `json:"category_name"`  // e.g. "Hyundai AC Business Class", "Subarna Snigdha Coach"
	Origin         string        `json:"origin"`
	Destination    string        `json:"destination"`
	DepartureTime  time.Time     `json:"departure_time"`
	ArrivalTime    time.Time     `json:"arrival_time"`
	BaseFare       float64       `json:"base_fare"`
	TotalSeats     int           `json:"total_seats"`
	AvailableSeats int           `json:"available_seats"`
}

type TripSeatLayoutResponse struct {
	TripID           string         `json:"trip_id"`
	OperatorName     string         `json:"operator_name"`
	RegistrationNo   string         `json:"registration_no"`
	TransportType    TransportType  `json:"transport_type"`
	CategoryName     string         `json:"category_name"`
	SeatLayoutConfig map[string]any `json:"seat_layout_config"` // Dynamic JSONB schema from DB
	Seats            []SeatMetadata `json:"seats"`
}