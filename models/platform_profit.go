// models/platform_profit.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlatformProfit struct {
	ID                 uuid.UUID  `gorm:"type:char(36);primary_key"`
	SourceType         string     `gorm:"type:enum('utama','ecommerce');not null"`
	TransactionID      *uuid.UUID `gorm:"type:char(36)"` // dipakai utk transaksi utama
	ECommercePaymentID *uuid.UUID `gorm:"type:char(36)"` // dipakai utk ecommerce

	GrossProfit float64   `gorm:"type:decimal(12,2);not null"`
	GatewayFee  float64   `gorm:"type:decimal(12,2);not null;default:0"`
	NetProfit   float64   `gorm:"type:decimal(12,2);not null"`
	ProfitDate  time.Time `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Relasi opsional
	Transaction      *Transaction      `gorm:"foreignKey:TransactionID"`
	ECommercePayment *ECommercePayment `gorm:"foreignKey:ECommercePaymentID"`
}

func (p *PlatformProfit) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
