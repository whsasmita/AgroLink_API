package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Contract represents digital contracts
type Contract struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	ProjectID uuid.UUID `gorm:"type:char(36);not null"`
	FarmerID  uuid.UUID `gorm:"type:char(36);not null"`
	WorkerID  uuid.UUID `gorm:"type:char(36);not null"`
	Content   string    `gorm:"type:text;not null"`

	// Status tanda tangan sederhana
	SignedByFarmer bool `gorm:"default:false"`
	SignedByWorker bool `gorm:"default:false"`
	SignedAt       *time.Time

	Status    string `gorm:"type:enum('draft','pending_signature','active','completed','terminated');default:draft"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relasi
	Project            Project
	Farmer             Farmer
	Worker             Worker
	ProjectAssignments []ProjectAssignment
	Transactions       []Transaction
}

// BeforeCreate hook for Contract
func (c *Contract) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// ContractTemplate represents contract templates
