package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ECommercePayment struct {
	ID         uuid.UUID `gorm:"type:char(36);primary_key"`
	UserID     uuid.UUID `gorm:"type:char(36);not null;index"`
	GrandTotal float64   `gorm:"type:decimal(12,2);not null"`
	Status     string    `gorm:"type:enum('pending','paid','failed');default:'pending'"`
	SnapToken  string    `gorm:"type:text"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	// [PERBAIKAN] Relasi Many-to-Many ke Order
	User *User `gorm:"foreignKey:UserID"`
	Orders []Order `gorm:"many2many:ecommerce_payment_orders;"`
}

func (p *ECommercePayment) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}
