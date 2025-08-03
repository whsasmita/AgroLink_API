package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Delivery represents delivery/expedition orders
type Delivery struct {
	ID                   uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	ProjectID            *uuid.UUID `gorm:"type:char(36)" json:"project_id"`
	FarmerID             uuid.UUID  `gorm:"type:char(36);not null" json:"farmer_id"`
	ExpeditionID         uuid.UUID  `gorm:"type:char(36);not null" json:"expedition_id"`
	TrackingCode         string     `gorm:"type:varchar(50);uniqueIndex;not null" json:"tracking_code"`

	// Product Information
	ProductType      string   `gorm:"type:varchar(100);not null" json:"product_type"`
	Weight           float64  `gorm:"type:decimal(10,2);not null" json:"weight"`
	Volume           *float64 `gorm:"type:decimal(10,2)" json:"volume"`
	PackagingType    *string  `gorm:"type:varchar(50)" json:"packaging_type"`
	SpecialHandling  *string  `gorm:"type:text" json:"special_handling"`

	// Location Information
	PickupLocationLat  float64 `gorm:"type:decimal(10,8);not null" json:"pickup_location_lat"`
	PickupLocationLng  float64 `gorm:"type:decimal(11,8);not null" json:"pickup_location_lng"`
	PickupAddress      string  `gorm:"type:text;not null" json:"pickup_address"`
	DropLocationLat    float64 `gorm:"type:decimal(10,8);not null" json:"drop_location_lat"`
	DropLocationLng    float64 `gorm:"type:decimal(11,8);not null" json:"drop_location_lng"`
	DropAddress        string  `gorm:"type:text;not null" json:"drop_address"`

	// Time Information
	ScheduledPickup   time.Time  `gorm:"not null" json:"scheduled_pickup"`
	ActualPickup      *time.Time `json:"actual_pickup"`
	EstimatedDelivery *time.Time `json:"estimated_delivery"`
	ActualDelivery    *time.Time `json:"actual_delivery"`

	// Price & Status
	Price          float64 `gorm:"type:decimal(10,2);not null" json:"price"`
	DeliveryStatus string  `gorm:"type:enum('scheduled','pickup_pending','picked_up','in_transit','out_for_delivery','delivered','failed','cancelled');default:scheduled" json:"delivery_status"`

	// Additional Data
	RouteData               *string `gorm:"type:json" json:"route_data"`
	DeliveryInstructions    *string `gorm:"type:text" json:"delivery_instructions"`
	RecipientName           *string `gorm:"type:varchar(100)" json:"recipient_name"`
	RecipientPhone          *string `gorm:"type:varchar(20)" json:"recipient_phone"`
	ProofOfDelivery         *string `gorm:"type:text;comment:URL foto bukti" json:"proof_of_delivery"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Project         *Project        `gorm:"foreignKey:ProjectID"`
	Farmer          Farmer          `gorm:"foreignKey:FarmerID;constraint:OnDelete:CASCADE"`
	Expedition      Driver      `gorm:"foreignKey:ExpeditionID;constraint:OnDelete:CASCADE"`
	LocationTracks  []LocationTrack `gorm:"foreignKey:DeliveryID;constraint:OnDelete:CASCADE"`
	Contracts       []Contract      `gorm:"foreignKey:DeliveryID"`
	Transactions    []Transaction   `gorm:"foreignKey:DeliveryID"`
	Schedules       []Schedule      `gorm:"foreignKey:DeliveryID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for Delivery
func (d *Delivery) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// LocationTrack represents real-time location tracking
type LocationTrack struct {
	ID         uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	UserID     uuid.UUID `gorm:"type:char(36);not null" json:"user_id"`
	DeliveryID uuid.UUID `gorm:"type:char(36);not null" json:"delivery_id"`
	Latitude   float64   `gorm:"type:decimal(10,8);not null" json:"latitude"`
	Longitude  float64   `gorm:"type:decimal(11,8);not null" json:"longitude"`
	Timestamp  time.Time `json:"timestamp"`
	Status     *string   `gorm:"type:varchar(50)" json:"status"`
	Speed      *float64  `gorm:"type:decimal(5,2);comment:Kecepatan dalam km/h" json:"speed"`
	Bearing    *float64  `gorm:"type:decimal(5,2);comment:Arah dalam derajat" json:"bearing"`
	Accuracy   *float64  `gorm:"type:decimal(8,2);comment:Akurasi GPS dalam meter" json:"accuracy"`

	// Relationships
	User     User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Delivery Delivery `gorm:"foreignKey:DeliveryID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for LocationTrack
func (lt *LocationTrack) BeforeCreate(tx *gorm.DB) error {
	if lt.ID == uuid.Nil {
		lt.ID = uuid.New()
	}
	return nil
}