package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Invoice merepresentasikan tagihan total untuk satu proyek.
type Invoice struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	ProjectID   *uuid.UUID `gorm:"type:char(36)"`
	DeliveryID  *uuid.UUID `gorm:"type:char(36)"` // <-- [TAMBAHAN]
	FarmerID  uuid.UUID `gorm:"type:char(36);not null"`
	Amount    float64   `gorm:"type:decimal(12,2)"`
	PlatformFee float64 `gorm:"type:decimal(10,2)"`
	TotalAmount float64   `gorm:"type:decimal(12,2)"`
	Status    string    `gorm:"type:enum('pending','paid','failed');default:'pending'"`
	DueDate   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time

	Project  *Project  `gorm:"foreignKey:ProjectID"`
	Delivery *Delivery `gorm:"foreignKey:DeliveryID"`
}

func (i *Invoice) BeforeCreate(tx *gorm.DB) (err error) {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return
}