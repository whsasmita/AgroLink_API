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

	SignedByFarmer bool `gorm:"default:false"`
	SignedByWorker bool `gorm:"default:false"`
	SignedAt       *time.Time

	Status    string `gorm:"type:enum('pending_signature','active','completed','terminated');default:'pending_signature'"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relasi
	Project            Project
	Farmer             Farmer
	Worker             Worker
	ProjectAssignments []ProjectAssignment
    // [PERBAIKAN] Hapus relasi ke Transaction dari sini karena sudah ditangani oleh Invoice
	// Transactions       []Transaction `gorm:"foreignKey:ContractID"` 
}

func (c *Contract) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}