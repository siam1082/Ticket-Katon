package domain

import "time"

// -----------------------------------------------------------------------------
// USER & RBAC DOMAIN
// -----------------------------------------------------------------------------

type UserRole string

const (
	RoleCustomer      UserRole = "CUSTOMER"
	RoleOperatorAdmin UserRole = "OPERATOR_ADMIN" // কাউন্টার সেলস স্টাফ ও টিকিট অফিসার
	RoleSuperAdmin    UserRole = "SUPER_ADMIN"    // প্ল্যাটফর্ম সফটওয়্যার ইঞ্জিনিয়ার ও ডেভঅপস
)

type User struct {
	ID           string    `json:"id"`
	OperatorID   *string   `json:"operator_id,omitempty"` // নির্দিষ্ট বাস/লঞ্চ/ট্রেন কোম্পানির রেফারেন্স
	FullName     string    `json:"full_name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	PasswordHash string    `json:"-"`
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// -----------------------------------------------------------------------------
// MULTIMODAL TRANSPORT & SEAT CLASSES
// -----------------------------------------------------------------------------

type TransportType string

const (
	TransportBus    TransportType = "BUS"
	TransportLaunch TransportType = "LAUNCH"
	TransportTrain  TransportType = "TRAIN"
)

type SeatClass string

const (
	// Bus Classes
	SeatClassBusACChair    SeatClass = "BUS_AC_CHAIR"
	SeatClassBusNonACChair SeatClass = "BUS_NON_AC_CHAIR"
	SeatClassBusSleeper    SeatClass = "BUS_SLEEPER"

	// Launch Classes
	SeatClassLaunchDeck        SeatClass = "LAUNCH_DECK"
	SeatClassLaunchSingleCabin SeatClass = "LAUNCH_CABIN_SINGLE"
	SeatClassLaunchDoubleCabin SeatClass = "LAUNCH_CABIN_DOUBLE"
	SeatClassLaunchVIP         SeatClass = "LAUNCH_VIP_CABIN"

	// Train Classes (Bangladesh Railway)
	SeatClassTrainShovonChair SeatClass = "TRAIN_SHOVON_CHAIR"
	SeatClassTrainSnigdha     SeatClass = "TRAIN_SNIGDHA"
	SeatClassTrainCabin       SeatClass = "TRAIN_CABIN_BERTH"
)

// -----------------------------------------------------------------------------
// SEAT INVENTORY & STATUS
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
	ClassType SeatClass  `json:"class_type"`
	Status    SeatStatus `json:"status"`
	Price     float64    `json:"price"`
	HeldBy    *string    `json:"held_by,omitempty"`
	HeldUntil *time.Time `json:"held_until,omitempty"`
}

type SeatMetadata struct {
	SeatNumber  string     `json:"seat_number"`   // যেমন: "A1", "KA-12", "CABIN-204"
	DeckOrClass string     `json:"deck_or_class"` // রিপোজিটরির সাথে ১০০% ম্যাচ রাখতে DeckOrClass
	ClassType   SeatClass  `json:"class_type"`
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
	TransportType  TransportType `json:"transport_type"`
	CategoryName   string        `json:"category_name"`
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
	SeatLayoutConfig map[string]any `json:"seat_layout_config"`
	Seats            []SeatMetadata `json:"seats"`
}