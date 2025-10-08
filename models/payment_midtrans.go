package models

import (
	"github.com/google/uuid"
	"time"
)

type PaymentMidtrans struct {
	ID             uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())"`
	OrderID        uuid.UUID `gorm:"type:char(36);not null;uniqueIndex"`
	JenisTransaksi string    `gorm:"type:varchar(50)"`
	Channel        string    `gorm:"type:varchar(50)"`
	Status         string    `gorm:"type:varchar(30);index"`
	Nilai          float64   `gorm:"type:decimal(12,2)"`
	EmailPelanggan *string   `gorm:"type:varchar(120)"`
	PaidAt         *time.Time
	RawPayload     *string `gorm:"type:longtext"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Order Order `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
