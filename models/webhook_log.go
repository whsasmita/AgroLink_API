package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type WebhookLog struct {
	ID                uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Provider          string    `gorm:"type:varchar(50);index;not null" json:"provider"`  // ex: "midtrans"
	Event             string    `gorm:"type:varchar(50);index" json:"event"`              // ex: transaction_status
	OrderID           string    `gorm:"type:varchar(100);index" json:"order_id"`           // invoice UUID
	TransactionID     string    `gorm:"type:varchar(64);index" json:"transaction_id"`     // midtrans transaction_id
	PaymentType       string    `gorm:"type:varchar(50)" json:"payment_type"`             // credit_card / qris / bank_transfer
	StatusCode        string    `gorm:"type:varchar(10)" json:"status_code"`              // "200", etc
	TransactionStatus string    `gorm:"type:varchar(50);index" json:"transaction_status"` // capture/settlement/pending/...
	FraudStatus       string    `gorm:"type:varchar(50)" json:"fraud_status"`             // accept/challenge/deny
	SignatureKey      string    `gorm:"type:varchar(255)" json:"signature_key"`           // from payload
	SignatureValid    bool      `gorm:"not null" json:"signature_valid"`                  // calc result
	Processed         bool      `gorm:"not null;index" json:"processed"`                  // whether service handled it
	ErrorMessage      *string   `gorm:"type:text" json:"error_message"`                   // if any

	Headers    datatypes.JSON `gorm:"type:json" json:"headers"`     // request headers snapshot
	RawBody    datatypes.JSON `gorm:"type:json" json:"raw_body"`    // whole payload
	ParsedBody datatypes.JSON `gorm:"type:json" json:"parsed_body"` // the map after bind

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (w *WebhookLog) BeforeCreate(tx *gorm.DB) (err error) {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
