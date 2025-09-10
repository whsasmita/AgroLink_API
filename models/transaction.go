package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	ID         uuid.UUID `gorm:"type:char(36);primary_key"`
	ContractID uuid.UUID `gorm:"type:char(36);not null"`
	FromUserID uuid.UUID `gorm:"type:char(36);not null"`
	ToUserID   uuid.UUID `gorm:"type:char(36);not null"`
	Description *string  `gorm:"type:text"`

	// Detail Keuangan
	Amount      float64 `gorm:"type:decimal(12,2);not null"` // Nilai kontrak dasar
	PlatformFee float64 `gorm:"type:decimal(10,2);not null"` // Biaya platform (misal 5%)
	TotalAmount float64 `gorm:"type:decimal(12,2);not null"` // Total yang harus dibayar petani

	// --- Integrasi Payment Gateway ---
    // Menyimpan order_id/transaction_id dari Midtrans. Sangat penting untuk rekonsiliasi via webhook.
	PaymentGatewayReferenceID *string `gorm:"type:varchar(255);index"` 
    // Menyimpan metode pembayaran yang dipilih pengguna (misal: "gopay", "bca_va").
	PaymentMethod           *string `gorm:"type:varchar(50)"`      
    // URL atau token yang digunakan frontend untuk menampilkan halaman pembayaran Midtrans.
    PaymentSnapToken          *string `gorm:"type:text"`

	// Status & Timestamps
    // Status: pending -> hold (setelah bayar) -> released (setelah proyek selesai)
	Status          string     `gorm:"type:enum('pending','hold','released','cancelled','failed');default:pending"`
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