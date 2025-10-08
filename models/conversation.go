package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConversationType string
const (
	ConversationDirect ConversationType = "direct"
)

type Conversation struct {
	ID        string           `gorm:"type:char(36);primaryKey" json:"id"`
	Type      ConversationType `gorm:"type:enum('direct');default:'direct';index" json:"type"`
	CreatedBy string           `gorm:"type:char(36);index" json:"created_by"`
	CreatedAt time.Time        `gorm:"autoCreateTime" json:"created_at"`
}

func (c *Conversation) BeforeCreate(tx *gorm.DB) (err error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return
}
