package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EmailOTP merepresentasikan data kode OTP verifikasi email pendaftaran
type EmailOTP struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	Email     string    `gorm:"type:varchar(100);index;not null" json:"email"`
	OTPCode   string    `gorm:"type:varchar(6);not null" json:"otp_code"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsUsed    bool      `gorm:"default:false" json:"is_used"`
	CreatedAt time.Time `json:"created_at"`
}

func (e *EmailOTP) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
