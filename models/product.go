package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())"`
	Title       string    `gorm:"type:varchar(200);not null"`
	FarmerID    uuid.UUID `gorm:"type:char(36);not null;index"`
	Description string    `gorm:"type:text"`
	Location    *string   `gorm:"type:varchar(150)"`
	Category    *string   `gorm:"type:varchar(100)"`
	Price       float64   `gorm:"type:decimal(12,2);not null;default:0.00"`
	Stock       int       `gorm:"not null;default:0"`
	ReservedStock int `gorm:"not null;default:0"`
	ImageURLs   datatypes.JSON `gorm:"column:image_urls"` 
	Rating      *float64  `gorm:"type:decimal(3,2)"`

	CreatedAt time.Time
	UpdatedAt time.Time

	Farmer     Farmer      `gorm:"foreignKey:FarmerID"`
	OrderItems []OrderItem `gorm:"foreignKey:ProductID"`
	CartItems  []Cart      `gorm:"foreignKey:ProductID"`
}
