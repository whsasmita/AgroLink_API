package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Transaction merepresentasikan bukti pembayaran untuk sebuah invoice.
type Transaction struct {
	ID                        uuid.UUID `gorm:"type:char(36);primary_key"`
	InvoiceID                 uuid.UUID `gorm:"type:char(36);not null"`
	PaymentGateway            string    `gorm:"type:varchar(50);default:'midtrans'"`
	PaymentGatewayReferenceID *string   `gorm:"type:varchar(255)"`
	AmountPaid                float64   `gorm:"type:decimal(12,2)"`
	PaymentMethod             *string   `gorm:"type:varchar(50)"`
	TransactionDate           time.Time

	// [PERBAIKAN] Tambahkan field relasi ini
	Invoice Invoice `gorm:"foreignKey:InvoiceID"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	t.TransactionDate = time.Now()
	return
}