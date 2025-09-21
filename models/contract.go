package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Contract represents digital contracts
type Contract struct {
	ID uuid.UUID `gorm:"type:char(36);primaryKey"`

	// 'work' untuk proyek, 'delivery' untuk pengiriman
	ContractType string `gorm:"type:enum('work','delivery');not null"`

	// Kontrak bisa untuk Proyek ATAU Delivery
	ProjectID *uuid.UUID `gorm:"type:char(36);index"`

	FarmerID uuid.UUID  `gorm:"type:char(36);not null;index"` // pemberi kerja selalu Petani
	WorkerID *uuid.UUID `gorm:"type:char(36);index"`          // pihak kedua jika 'work'
	DriverID *uuid.UUID `gorm:"type:char(36);index"`          // pihak kedua jika 'delivery'

	SignedByFarmer      bool       `gorm:"default:false"`
	SignedBySecondParty bool       `gorm:"default:false"`
	SignedAt            *time.Time
	Status              string     `gorm:"type:enum('pending_signature','active','completed','terminated');default:'pending_signature'"`
	CreatedAt           time.Time
	UpdatedAt           time.Time

	// Relations
	Project  *Project `gorm:"foreignKey:ProjectID;references:ID"`

	// HAS ONE Delivery: Delivery.ContractID -> Contract.ID
	// Bila ContractType = 'delivery', baris Delivery akan merefer ke kontrak ini.
	Delivery *Delivery `gorm:"foreignKey:ContractID;references:ID"`

	Farmer Farmer  `gorm:"foreignKey:FarmerID;references:UserID"`
	Worker *Worker `gorm:"foreignKey:WorkerID;references:UserID"`
	Driver *Driver `gorm:"foreignKey:DriverID;references:UserID"`
}

func (c *Contract) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
