package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Delivery represents delivery/expedition orders
type Delivery struct {
    ID                 uuid.UUID `gorm:"type:char(36);primary_key"`
    FarmerID           uuid.UUID `gorm:"type:char(36);not null"`
    DriverID           *uuid.UUID
    ContractID         *uuid.UUID
    
    PickupAddress      string
    PickupLat          float64
    PickupLng          float64
    DestinationAddress string
    
    ItemDescription    string
    ItemWeight         float64 // dalam kg
    
    Status             string `gorm:"type:enum('pending_driver', 'pending_signature', 'pending_payment', 'in_transit', 'delivered', 'cancelled');default:'pending_driver'"`

    // Relasi
    Contract *Contract
}


// BeforeCreate hook for Delivery
func (d *Delivery) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// LocationTrack represents real-time location tracking
