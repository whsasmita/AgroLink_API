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
	DocumentType string    `gorm:"type:enum('KTP','SIM','STNK','KIR','SELFIE_KTP','SKU');not null"`
	FilePath     string    `gorm:"type:text;not null"`
	Status       string    `gorm:"type:enum('pending','approved','rejected');not null;default:'pending'"`
	Notes        *string   `gorm:"type:text"`
	ReviewedBy   *uuid.UUID `gorm:"type:char(36);index"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	User     User  `gorm:"foreignKey:UserID"`
	Reviewer *User `gorm:"foreignKey:ReviewedBy"`
}

// BeforeCreate hook untuk generate UUID secara otomatis.
func (uv *UserVerification) BeforeCreate(tx *gorm.DB) (err error) {
	if (uv.ID == uuid.Nil) {
		uv.ID = uuid.New()
	}
	return
}