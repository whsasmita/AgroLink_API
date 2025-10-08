package models

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())"`
	Title       string    `gorm:"type:varchar(200);not null"`
	Description string    `gorm:"type:text"`
	Location    *string   `gorm:"type:varchar(150)"`
	Category    *string   `gorm:"type:varchar(100)"`
	Price       float64   `gorm:"type:decimal(12,2);not null;default:0.00"`
	Stock       int       `gorm:"not null;default:0"`
	ImageURL    *string   `gorm:"type:text"`
	Rating      *float64  `gorm:"type:decimal(3,2)"`

	CreatedAt time.Time
	UpdatedAt time.Time

	OrderItems []OrderItem `gorm:"foreignKey:ProductID"`
	CartItems  []Cart      `gorm:"foreignKey:ProductID"`
}
