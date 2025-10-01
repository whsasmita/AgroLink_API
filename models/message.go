package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Message struct {
	ID             string         `gorm:"type:char(36);primaryKey" json:"id"`
	ConversationID string         `gorm:"type:char(36);index:idx_conv_created,priority:1" json:"conversation_id"`
	SenderID       string         `gorm:"type:char(36);index" json:"sender_id"`
	Body           string         `gorm:"type:text;not null" json:"body"`
	CreatedAt      time.Time      `gorm:"autoCreateTime;index:idx_conv_created,priority:2" json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *Message) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID = uuid.New().String()
	}
	return
}
