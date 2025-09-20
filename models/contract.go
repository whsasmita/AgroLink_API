package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Contract represents digital contracts
type Contract struct {
	ID uuid.UUID `gorm:"type:char(36);primary_key"`

	// [BARU] Menandakan jenis kontrak: 'work' untuk proyek, 'delivery' untuk pengiriman.
	ContractType string `gorm:"type:enum('work', 'delivery');not null"`

	// [DIUBAH] Menjadi pointer (*), karena kontrak bisa untuk Proyek ATAU Pengiriman.
	ProjectID  *uuid.UUID `gorm:"type:char(36)"`
	DeliveryID *uuid.UUID `gorm:"type:char(36)"` // Akan digunakan nanti

	FarmerID uuid.UUID `gorm:"type:char(36);not null"` // Pemberi kerja selalu Petani

	// [DIUBAH] Menjadi pointer (*), karena pihak kedua bisa Pekerja ATAU Driver.
	WorkerID *uuid.UUID `gorm:"type:char(36)"`
	DriverID *uuid.UUID `gorm:"type:char(36)"` // Akan digunakan nanti

	// Field lain tidak berubah
	SignedByFarmer      bool `gorm:"default:false"`
	SignedBySecondParty bool `gorm:"default:false"` // Mungkin perlu diganti nama menjadi SignedBySecondParty
	SignedAt            *time.Time
	Status              string `gorm:"type:enum('pending_signature','active','completed','terminated');default:'pending_signature'"`
	CreatedAt           time.Time
	UpdatedAt           time.Time

	// Relasi
	Project  *Project // Diubah menjadi pointer
	Delivery *Delivery
	Farmer   Farmer
	Worker   *Worker // Diubah menjadi pointer
}

func (c *Contract) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}
