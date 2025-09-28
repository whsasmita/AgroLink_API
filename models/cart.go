package models

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index:idx_cart_user_product,unique"`
	ProductID uuid.UUID `gorm:"type:char(36);not null;index:idx_cart_user_product,unique"`
	Quantity  int       `gorm:"not null;default:1"`

	CreatedAt time.Time
	UpdatedAt time.Time

	User    User    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Product Product `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
