package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlatformProfit struct {
	ID            uuid.UUID `gorm:"type:char(36);primary_key" json:"id"`
	TransactionID uuid.UUID `gorm:"type:char(36);not null" json:"transaction_id"`
	SourceType    string    `gorm:"type:enum('utama','ecommerce');not null" json:"source_type"` // dari Transaksi Utama atau Ecommerce

	GrossProfit float64 `gorm:"type:decimal(12,2);not null" json:"gross_profit"` // Keuntungan Kotor (platform fee kotor)
	GatewayFee  float64 `gorm:"type:decimal(12,2);not null;default:0" json:"gateway_fee"` // biaya midtrans / biaya gateway
	NetProfit   float64 `gorm:"type:decimal(12,2);not null" json:"net_profit"` // Keuntungan Bersih = gross - gateway_fee

	ProfitDate time.Time `json:"profit_date"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	Transaction Transaction `gorm:"foreignKey:TransactionID" json:"transaction"`
}

func (p *PlatformProfit) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	// Kalau ProfitDate belum diisi manual, fallback ke sekarang
	if p.ProfitDate.IsZero() {
		p.ProfitDate = time.Now()
	}
	return
}
