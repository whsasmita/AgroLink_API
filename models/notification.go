package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index"`
	Title     string    `gorm:"type:varchar(255);not null"`
	Message   string    `gorm:"type:text"`
	IsRead    bool      `gorm:"default:false"`
	Link      *string   `gorm:"type:text"`
	Type      string    `gorm:"type:varchar(50)"`
	CreatedAt time.Time

	User User `gorm:"foreignKey:UserID"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) (err error) {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return
}