package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID              uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())"`
	UserID          uuid.UUID `gorm:"type:char(36);not null;index"`
	FarmerID        uuid.UUID `gorm:"type:char(36);not null;index"`
	InvoiceNumber   string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	TotalAmount     float64   `gorm:"type:decimal(12,2);not null;default:0.00"`
	Status          string    `gorm:"type:enum('pending','paid','shipped','completed','cancelled');not null;default:'pending'"`
	ShippingAddress *string   `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time

	User     User     `gorm:"foreignKey:UserID"`
	Farmer   Farmer   `gorm:"foreignKey:FarmerID"`
	Items    []OrderItem        `gorm:"foreignKey:OrderID"`
	Payments []ECommercePayment `gorm:"many2many:ecommerce_payment_orders;"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	// Hanya generate UUID baru jika ID-nya masih kosong (uuid.Nil)
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return
}
