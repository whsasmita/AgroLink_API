package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	ID              uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())"`
	UserID          uuid.UUID `gorm:"type:char(36);not null;index"`
	InvoiceNumber   string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	TotalAmount     float64   `gorm:"type:decimal(12,2);not null;default:0.00"`
	Status          string    `gorm:"type:enum('pending','paid','shipped','completed','cancelled');not null;default:'pending'"`
	ShippingAddress *string   `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time

	User    User             `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Items   []OrderItem      `gorm:"foreignKey:OrderID"`
	Payment *PaymentMidtrans `gorm:"foreignKey:OrderID"`
}
