package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserVerification merepresentasikan satu entri dokumen yang diunggah untuk verifikasi.
type UserVerification struct {
	ID           uuid.UUID `gorm:"type:char(36);primary_key"`
	UserID       uuid.UUID `gorm:"type:char(36);not null;index"`
	
	// Jenis dokumen yang diunggah (misal: KTP, SIM, dll.)
	DocumentType string    `gorm:"type:enum('KTP','SIM','SKU','SELFIE_KTP');not null"`
	
	// Path ke file yang disimpan di storage
	FilePath     string    `gorm:"type:text;not null"`

	// Status verifikasi oleh admin
	Status       string    `gorm:"type:enum('pending','approved','rejected');not null;default:'pending'"`
	
	// Catatan dari admin jika verifikasi ditolak
	Notes        *string   `gorm:"type:text"`
	
	// ID admin yang melakukan review (bisa null jika belum direview)
	ReviewedBy   *uuid.UUID `gorm:"type:char(36);index"`

	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Relasi
	User     User  `gorm:"foreignKey:UserID"`
	Reviewer *User `gorm:"foreignKey:ReviewedBy"` // Admin yang mereview
}

// BeforeCreate hook untuk generate UUID secara otomatis.
func (uv *UserVerification) BeforeCreate(tx *gorm.DB) (err error) {
	if uv.ID == uuid.Nil {
		uv.ID = uuid.New()
	}
	return
}