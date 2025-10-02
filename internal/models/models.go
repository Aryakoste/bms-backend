package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Society struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string            `bson:"name" json:"name" binding:"required"`
	Code         string            `bson:"code" json:"code" binding:"required"` // Unique society access code
	Address      string            `bson:"address" json:"address"`
	City         string            `bson:"city" json:"city"`
	State        string            `bson:"state" json:"state"`
	PinCode      string            `bson:"pin_code" json:"pin_code"`
	ContactEmail string            `bson:"contact_email" json:"contact_email"`
	ContactPhone string            `bson:"contact_phone" json:"contact_phone"`
	Buildings    []Building        `bson:"buildings" json:"buildings"`
	IsActive     bool              `bson:"is_active" json:"is_active"`
	CreatedAt    time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time         `bson:"updated_at" json:"updated_at"`
}

type Building struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name          string            `bson:"name" json:"name"`
	Floors        int               `bson:"floors" json:"floors"`
	UnitsPerFloor int               `bson:"units_per_floor" json:"units_per_floor"`
	SocietyID     primitive.ObjectID `bson:"society_id" json:"society_id"`
}

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string            `bson:"name" json:"name" binding:"required"`
	Email     string            `bson:"email" json:"email" binding:"required,email"`
	Password  string            `bson:"password" json:"-"`
	Role      string            `bson:"role" json:"role" binding:"required"`
	Unit      string            `bson:"unit" json:"unit"`
	Building  string            `bson:"building" json:"building"`
	Phone     string            `bson:"phone" json:"phone"`
	Avatar    string            `bson:"avatar" json:"avatar"`
	SocietyID primitive.ObjectID `bson:"society_id" json:"society_id"`       // Link to society
	SocietyCode string           `bson:"society_code" json:"society_code"`   // Society access code
	IsActive  bool              `bson:"is_active" json:"is_active"`
	CreatedAt time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at" json:"updated_at"`
}

type Visitor struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Name            string             `bson:"name" json:"name" binding:"required"`
	Phone           string             `bson:"phone" json:"phone" binding:"required"`
	Purpose         string             `bson:"purpose" json:"purpose" binding:"required"`
	HostID          primitive.ObjectID  `bson:"host_id" json:"host_id"`
	HostName        string             `bson:"host_name" json:"host_name"`
	HostUnit        string             `bson:"host_unit" json:"host_unit"`
	ExpectedTime    time.Time          `bson:"expected_time" json:"expected_time"`
	ActualArrival   *time.Time         `bson:"actual_arrival,omitempty" json:"actual_arrival,omitempty"`
	ActualDeparture *time.Time         `bson:"actual_departure,omitempty" json:"actual_departure,omitempty"`
	QRCode          string             `bson:"qr_code" json:"qr_code"`
	Status          string             `bson:"status" json:"status"` // pending, approved, rejected, completed
	VehicleNumber   string             `bson:"vehicle_number,omitempty" json:"vehicle_number,omitempty"`
	PhotoURL        string             `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
	ApprovedBy      *primitive.ObjectID `bson:"approved_by,omitempty" json:"approved_by,omitempty"`
	SocietyID       primitive.ObjectID  `bson:"society_id" json:"society_id"`       // Link to society
	SocietyCode     string             `bson:"society_code" json:"society_code"`   // Society access code
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

type MaintenanceRecord struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UnitID      primitive.ObjectID `bson:"unit_id" json:"unit_id"`
	UnitNumber  string            `bson:"unit_number" json:"unit_number"`
	Amount      float64           `bson:"amount" json:"amount"`
	Month       string            `bson:"month" json:"month"`
	DueDate     time.Time         `bson:"due_date" json:"due_date"`
	PaidDate    *time.Time        `bson:"paid_date,omitempty" json:"paid_date,omitempty"`
	Status      string            `bson:"status" json:"status"` // pending, paid, overdue
	Description string            `bson:"description" json:"description"`
	ReceiptURL  string            `bson:"receipt_url,omitempty" json:"receipt_url,omitempty"`
	SocietyID   primitive.ObjectID `bson:"society_id" json:"society_id"`       // Link to society
	SocietyCode string            `bson:"society_code" json:"society_code"`   // Society access code
	CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
}

type Amenity struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string            `bson:"name" json:"name" binding:"required"`
	Description    string            `bson:"description" json:"description"`
	BookingFee     float64           `bson:"booking_fee" json:"booking_fee"`
	Capacity       int               `bson:"capacity" json:"capacity"`
	Facilities     []string          `bson:"facilities" json:"facilities"`
	AvailableHours string            `bson:"available_hours" json:"available_hours"`
	Images         []string          `bson:"images,omitempty" json:"images,omitempty"`
	SocietyID      primitive.ObjectID `bson:"society_id" json:"society_id"`       // Link to society
	SocietyCode    string            `bson:"society_code" json:"society_code"`   // Society access code
	IsActive       bool              `bson:"is_active" json:"is_active"`
}

type AmenityBooking struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	AmenityID   primitive.ObjectID `bson:"amenity_id" json:"amenity_id"`
	AmenityName string            `bson:"amenity_name" json:"amenity_name"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	UserName    string            `bson:"user_name" json:"user_name"`
	Date        time.Time         `bson:"date" json:"date"`
	TimeSlot    string            `bson:"time_slot" json:"time_slot"`
	Status      string            `bson:"status" json:"status"` // confirmed, cancelled, completed
	TotalAmount float64           `bson:"total_amount" json:"total_amount"`
	PaymentID   string            `bson:"payment_id,omitempty" json:"payment_id,omitempty"`
	SocietyID   primitive.ObjectID `bson:"society_id" json:"society_id"`       // Link to society
	SocietyCode string            `bson:"society_code" json:"society_code"`   // Society access code
	CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
}

type Notice struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string            `bson:"title" json:"title" binding:"required"`
	Content     string            `bson:"content" json:"content" binding:"required"`
	Type        string            `bson:"type" json:"type"` // announcement, warning, urgent
	AuthorID    primitive.ObjectID `bson:"author_id" json:"author_id"`
	AuthorName  string            `bson:"author_name" json:"author_name"`
	SocietyID   primitive.ObjectID `bson:"society_id" json:"society_id"`       // Link to society
	SocietyCode string            `bson:"society_code" json:"society_code"`   // Society access code
	IsActive    bool              `bson:"is_active" json:"is_active"`
	ExpiresAt   *time.Time        `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
}

// Request/Response models
type SocietyCodeRequest struct {
	Code string `json:"code" binding:"required"`
}

type SocietyResponse struct {
	ID      primitive.ObjectID `json:"id"`
	Name    string            `json:"name"`
	Code    string            `json:"code"`
	Address string            `json:"address"`
	City    string            `json:"city"`
	State   string            `json:"state"`
}

type LoginRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	SocietyCode string `json:"society_code" binding:"required"`
}

type RegisterRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	Role        string `json:"role" binding:"required"`
	Unit        string `json:"unit"`
	Building    string `json:"building"`
	Phone       string `json:"phone"`
	SocietyCode string `json:"society_code" binding:"required"`
}

type LoginResponse struct {
	Token   string          `json:"token"`
	User    User            `json:"user"`
	Society SocietyResponse `json:"society"`
}

type VisitorApprovalRequest struct {
	Status     string `json:"status" binding:"required"`
	ApprovedBy string `json:"approved_by,omitempty"`
}

type PaymentRequest struct {
	MaintenanceID string  `json:"maintenance_id" binding:"required"`
	Amount       float64 `json:"amount" binding:"required"`
	PaymentMethod string `json:"payment_method"`
}