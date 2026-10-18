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


// BOOKING, TICKET & PAYMENT DOMAIN

type BookingStatus string

const (
	BookingStatusPending BookingStatus = "PENDING"
	BookingStatusConfirmed BookingStatus = "CONFIRMED"
	BookingStatusCancelled BookingStatus = "CANCELLED"
)

type Booking struct {
	ID		string		`json:"id"`
	UserID	string      `json:"user_id"`
	TripID  string      `json:"trip_id"`
	PNRCode  string      `json:"pnr_code"`
	TotalAmount  float64      `json:"total_amount"`
	DiscountAmount  float64      `json:"discount_amount"`
	Status  BookingStatus      `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type PassengerInfo struct {
	Name string `json:"name"`
	Phone string `json:"phone"`
	Gender string `json:"gender"`
	SeatNumber string `json:"seat_number"`
}

type Ticket struct {
	ID string `json:"id"`
	BookingID string  `json:"booking_id"`
	SeatNumber string  `json:"seat_number"`
	PassengerName string  `json:"passenger_name"`
	PassengerPhone string   `json:"passenger_phone"`
	PassengerGender string   `json:"passenger_gender"`
	TicketNumber  string   `json:"ticket_number"`
	CreatedAt  time.Time `json:"created_at"`
}

type PaymentGateway string

const (
	PaymentGatewayBkash  PaymentGateway = "BKASH"
	PaymentGatewayNagad  PaymentGateway = "NAGAD"
	PaymentGatewaySSLCommerz PaymentGateway = "SSLCOMMERZ"
	PaymentGatewayMock  PaymentGateway = "MOCK"
	GatewayRocket PaymentGateway = "ROCKET"
	GatewayUpay PaymentGateway = "UPAY"
)

type PaymentStatus string

const (
	PaymentStatusInitiated PaymentStatus = "INITIATED"
	PaymentStatusSuccess  PaymentStatus = "SUCCESS"
	PaymentStatusFailed  PaymentStatus = "FAILED"
)

type Payment struct {
	ID  string  `json:"id"`
	BookingID  string `json:"booking_id"`
	Gateway  PaymentGateway `json:"gateway"`
	TransactionID  string  `json:"transaction_id"`
	GatewayPaymentID string  `json:"gateway_payment_id"`
	Amount   float64  `json:"amount"`
	Currency  string  `json:"currency"`
	Status  PaymentStatus  `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type PaymentPayload struct {
	Gateway       PaymentGateway `json:"gateway"`
	TransactionID string         `json:"transaction_id"`
}

// REQUEST & RESPONSE DTOS

type ConfirmBookingRequest struct {
	TripID  string  `json:"trip_id"`
	Seats  []string  `json:"seats"`
	Passengers []PassengerInfo `json:"passengers"`
	Payment    PaymentPayload  `json:"payment"`
}

type ConfirmBookingResponse struct {
	BookingID  string   `json:"booking_id"`
	PNRCode   string   `json:"pnr_code"`
	TripID        string        `json:"trip_id"`
	TotalAmount   float64       `json:"total_amount"`
	DiscountAmount float64      `json:"discount_amount"`
	Status        BookingStatus `json:"status"`
	Tickets       []Ticket      `json:"tickets"`
	PaymentStatus PaymentStatus `json:"payment_status"`
}