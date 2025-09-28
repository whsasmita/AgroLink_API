package models

import (
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	ID                   uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())"`
	OrderID              uuid.UUID `gorm:"type:char(36);not null;index"`
	ProductID            uuid.UUID `gorm:"type:char(36);not null;index"`
	Quantity             int       `gorm:"not null;default:1"`
	PriceAtPurchase      float64   `gorm:"type:decimal(12,2);not null;default:0.00"`
	NameAtPurchase       string    `gorm:"type:varchar(200);not null"`
	DescriptionAtPurchase *string   `gorm:"type:text"`
	SubTotal             float64   `gorm:"type:decimal(12,2);not null;default:0.00"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Order   Order   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
}
