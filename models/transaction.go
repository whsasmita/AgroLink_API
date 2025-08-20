package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key"`
	ContractID  uuid.UUID `gorm:"type:char(36);not null"`
	FromUserID  uuid.UUID `gorm:"type:char(36);not null"` // ID User Petani
	ToUserID    uuid.UUID `gorm:"type:char(36);not null"` // ID User Pekerja
	Amount      float64   `gorm:"type:decimal(10,2);not null"`
	Description *string   `gorm:"type:text"`

	// Status untuk alur pembayaran
	Status          string `gorm:"type:enum('pending','hold','released','cancelled');default:pending"`
	TransactionDate time.Time
	ReleasedAt      *time.Time

	// Relasi
	Contract Contract
	FromUser User `gorm:"foreignKey:FromUserID"`
	ToUser   User `gorm:"foreignKey:ToUserID"`
}

func (m *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return
}