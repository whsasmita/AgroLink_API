package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	AIChatScopePublic  = "public"
	AIChatScopePrivate = "private"

	AIChatSubscriptionStatusPending   = "pending"
	AIChatSubscriptionStatusActive    = "active"
	AIChatSubscriptionStatusExpired   = "expired"
	AIChatSubscriptionStatusCancelled = "cancelled"
)

// AIChatTurn menyimpan satu turn chat user -> AI.
type AIChatTurn struct {
	ID          uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	Scope       string     `gorm:"type:enum('public','private');index;not null" json:"scope"`
	UserID      *uuid.UUID `gorm:"type:char(36);index" json:"user_id,omitempty"`
	IPAddress   *string    `gorm:"type:varchar(45);index" json:"ip_address,omitempty"`
	UserMessage string     `gorm:"type:longtext;not null" json:"user_message"`
	AIReply     string     `gorm:"type:longtext;not null" json:"ai_reply"`
	Model       string     `gorm:"type:varchar(100);not null" json:"model"`
	CreatedAt   time.Time  `json:"created_at"`
}

// BeforeCreate memastikan turn memiliki UUID.
func (t *AIChatTurn) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// AIChatPremiumSubscription menyimpan status langganan premium Gemini.
type AIChatPremiumSubscription struct {
	ID              uuid.UUID  `gorm:"type:char(36);primary_key" json:"id"`
	UserID          uuid.UUID  `gorm:"type:char(36);uniqueIndex;not null" json:"user_id"`
	Status          string     `gorm:"type:enum('pending','active','expired','cancelled');default:'pending'" json:"status"`
	PlanName        string     `gorm:"type:varchar(50);default:'premium'" json:"plan_name"`
	Amount          float64    `gorm:"type:decimal(12,2);not null" json:"amount"`
	Currency        string     `gorm:"type:varchar(10);default:'IDR'" json:"currency"`
	MidtransOrderID string     `gorm:"type:varchar(100);uniqueIndex;not null" json:"midtrans_order_id"`
	SnapToken       string     `gorm:"type:text" json:"snap_token"`
	RedirectURL     string     `gorm:"type:text" json:"redirect_url"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	ExpiresAt       *time.Time `gorm:"index" json:"expires_at,omitempty"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}

// BeforeCreate memastikan subscription memiliki UUID.
func (s *AIChatPremiumSubscription) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
