package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LocationTrack struct {
	ID         uuid.UUID `gorm:"type:char(36);primary_key"`
	DeliveryID uuid.UUID `gorm:"type:char(36);not null;index"`
	Lat        float64   `gorm:"type:decimal(10,8);not null"`
	Lng        float64   `gorm:"type:decimal(11,8);not null"`
	Timestamp  time.Time
}

func (lt *LocationTrack) BeforeCreate(tx *gorm.DB) (err error) {
	if lt.ID == uuid.Nil {
		lt.ID = uuid.New()
	}
	lt.Timestamp = time.Now()
	return
}