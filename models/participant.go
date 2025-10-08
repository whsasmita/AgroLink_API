package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ConversationParticipant struct {
	ID             string     `gorm:"type:char(36);primaryKey" json:"id"`
	ConversationID string     `gorm:"type:char(36);index" json:"conversation_id"`
	UserID         string     `gorm:"type:char(36);index" json:"user_id"`
	Role           string     `gorm:"type:enum('member','admin');default:'member'" json:"role"`
	JoinedAt       time.Time  `gorm:"autoCreateTime" json:"joined_at"`
	LastReadMsgID  *string    `gorm:"type:char(36)" json:"last_read_message_id"`

	// Unique: one user per conversation
	_ struct{} `gorm:"uniqueIndex:uniq_conv_user,priority:1" json:"-"`
	// Unique key part 2:
	// ConversationID sebagai priority:2 otomatis ikut dari field di atas
}

func (p *ConversationParticipant) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return
}
