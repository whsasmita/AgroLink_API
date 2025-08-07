package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TODO checking untuk pemilihan tipa data untuk masalah financial
// User represents the main user table
type User struct {
	ID             uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	Name           string    `gorm:"type:varchar(100);not null" json:"name"`
	Email          string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password       string    `gorm:"type:varchar(255);not null" json:"-"`
	PhoneNumber    *string   `gorm:"type:varchar(20)" json:"phone_number"`
	Role           string    `gorm:"type:enum('farmer','worker','driver','admin','cs');not null" json:"role"`
	ProfilePicture *string   `gorm:"type:text" json:"profile_picture"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	EmailVerified  bool      `gorm:"default:false" json:"email_verified"`
	PhoneVerified  bool      `gorm:"default:false" json:"phone_verified"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships
	Farmer *Farmer `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"farmer,omitempty"`
	Worker *Worker `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"worker,omitempty"`
	Driver *Driver `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"driver,omitempty"`
}

// BeforeCreate hook to generate UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Farmer represents farmer profile details
type Farmer struct {
	UserID         uuid.UUID `gorm:"type:char(36);primary_key" json:"user_id"`
	Address        *string   `gorm:"type:text" json:"address"`
	AdditionalInfo *string   `gorm:"type:text" json:"additional_info"`
	CreatedAt      time.Time `json:"created_at"`

	// Relationships
	User          User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	FarmLocations []FarmLocation `gorm:"foreignKey:FarmerID;constraint:OnDelete:CASCADE"`
	Projects      []Project      `gorm:"foreignKey:FarmerID;constraint:OnDelete:CASCADE"`
}

// Worker represents worker profile details
type Worker struct {
	UserID               uuid.UUID `gorm:"type:char(36);primary_key" json:"user_id"`
	Skills               string    `gorm:"type:json;not null" json:"skills"` // JSON array as string
	HourlyRate           *float64  `gorm:"type:decimal(10,2)" json:"hourly_rate"`
	DailyRate            *float64  `gorm:"type:decimal(10,2)" json:"daily_rate"`
	Address              *string   `gorm:"type:text" json:"address"`
	AvailabilitySchedule *string   `gorm:"type:json" json:"availability_schedule"`
	CurrentLocationLat   *float64  `gorm:"type:decimal(10,8)" json:"current_location_lat"`
	CurrentLocationLng   *float64  `gorm:"type:decimal(11,8)" json:"current_location_lng"`
	Rating               float64   `gorm:"default:0" json:"rating"`
	TotalJobsCompleted   int       `gorm:"default:0" json:"total_jobs_completed"`
	CreatedAt            time.Time `json:"created_at"`

	// Relationships
	User                User                 `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	ProjectApplications []ProjectApplication `gorm:"foreignKey:WorkerID;constraint:OnDelete:CASCADE"`
	ProjectAssignments  []ProjectAssignment  `gorm:"foreignKey:WorkerID;constraint:OnDelete:CASCADE"`
	WorkerAvailability  []WorkerAvailability `gorm:"foreignKey:WorkerID;constraint:OnDelete:CASCADE"`
}

// Expedition represents expedition company profile details
type Driver struct {
	UserID          uuid.UUID `gorm:"type:char(36);primary_key" json:"user_id"`
	Address         *string   `gorm:"type:text" json:"company_address"`
	// ServiceAreas    string    `gorm:"type:json;not null" json:"service_areas"` // JSON array as string
	PricingScheme   string    `gorm:"type:json;not null" json:"pricing_scheme"`
	VehicleTypes    string    `gorm:"type:json;not null" json:"vehicle_types"` // JSON array as string
	Rating          float64   `gorm:"default:0" json:"rating"`
	TotalDeliveries int       `gorm:"default:0" json:"total_deliveries"`
	CreatedAt       time.Time `json:"created_at"`

	// Relationships
	User       User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Deliveries []Delivery `gorm:"foreignKey:ExpeditionID;constraint:OnDelete:CASCADE"`
}

// FarmLocation represents individual farm locations
type FarmLocation struct {
	ID             uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	FarmerID       uuid.UUID `gorm:"type:char(36);not null" json:"farmer_id"`
	Name           string    `gorm:"type:varchar(100);not null" json:"name"`
	Latitude       float64   `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude      float64   `gorm:"type:decimal(11,8);not null" json:"longitude"`
	AreaSize       float64   `gorm:"type:decimal(10,2);not null;comment:Luas dalam are" json:"area_size"`
	CropType       *string   `gorm:"type:varchar(50)" json:"crop_type"`
	IrrigationType *string   `gorm:"type:varchar(50)" json:"irrigation_type"`
	Description    *string   `gorm:"type:text" json:"description"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`

	// Relationships
	Farmer   Farmer    `gorm:"foreignKey:FarmerID;constraint:OnDelete:CASCADE"`
	Projects []Project `gorm:"foreignKey:FarmLocationID"`
}

// BeforeCreate hook for FarmLocation
func (fl *FarmLocation) BeforeCreate(tx *gorm.DB) error {
	if fl.ID == uuid.Nil {
		fl.ID = uuid.New()
	}
	return nil
}

// WorkerAvailability represents worker availability schedule
type WorkerAvailability struct {
	ID                 uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	WorkerID           uuid.UUID `gorm:"type:char(36);not null" json:"worker_id"`
	AvailableDate      time.Time `gorm:"type:date;not null" json:"available_date"`
	AvailableStartTime time.Time `gorm:"type:time;not null" json:"available_start_time"`
	AvailableEndTime   time.Time `gorm:"type:time;not null" json:"available_end_time"`
	IsBooked           bool      `gorm:"default:false" json:"is_booked"`
	BookingType        *string   `gorm:"type:enum('project','maintenance','other')" json:"booking_type"`
	Notes              *string   `gorm:"type:text" json:"notes"`
	CreatedAt          time.Time `json:"created_at"`

	// Relationships
	Worker Worker `gorm:"foreignKey:WorkerID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for WorkerAvailability
func (wa *WorkerAvailability) BeforeCreate(tx *gorm.DB) error {
	if wa.ID == uuid.Nil {
		wa.ID = uuid.New()
	}
	return nil
}