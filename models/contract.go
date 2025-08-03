package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Contract represents digital contracts
type Contract struct {
	ID          uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	FarmerID    *uuid.UUID `gorm:"type:char(36)" json:"farmer_id"`
	WorkerID    *uuid.UUID `gorm:"type:char(36)" json:"worker_id"`
	ExpeditionID *uuid.UUID `gorm:"type:char(36)" json:"expedition_id"`
	ProjectID   *uuid.UUID `gorm:"type:char(36)" json:"project_id"`
	DeliveryID  *uuid.UUID `gorm:"type:char(36)" json:"delivery_id"`

	ContractType         string  `gorm:"type:enum('work','delivery','maintenance');not null" json:"contract_type"`
	TemplateUsed         *string `gorm:"type:varchar(100)" json:"template_used"`
	Content              string  `gorm:"type:text;not null" json:"content"`
	TermsAndConditions   *string `gorm:"type:text" json:"terms_and_conditions"`

	// Digital Signatures
	SignedByFarmer           bool    `gorm:"default:false" json:"signed_by_farmer"`
	SignedByWorker           bool    `gorm:"default:false" json:"signed_by_worker"`
	SignedByExpedition       bool    `gorm:"default:false" json:"signed_by_expedition"`
	DigitalSignatureFarmer   *string `gorm:"type:text" json:"digital_signature_farmer"`
	DigitalSignatureWorker   *string `gorm:"type:text" json:"digital_signature_worker"`
	DigitalSignatureExpedition *string `gorm:"type:text" json:"digital_signature_expedition"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	SignedAt  *time.Time `json:"signed_at"`
	ExpiresAt *time.Time `json:"expires_at"`

	// Status
	IsActive bool   `gorm:"default:true" json:"is_active"`
	Status   string `gorm:"type:enum('draft','pending_signature','active','completed','terminated','expired');default:draft" json:"status"`

	// Relationships
	Farmer           *Farmer            `gorm:"foreignKey:FarmerID;constraint:OnDelete:CASCADE"`
	Worker           *Worker            `gorm:"foreignKey:WorkerID;constraint:OnDelete:CASCADE"`
	Expedition       *Driver        `gorm:"foreignKey:ExpeditionID;constraint:OnDelete:CASCADE"`
	Project          *Project           `gorm:"foreignKey:ProjectID"`
	Delivery         *Delivery          `gorm:"foreignKey:DeliveryID"`
	ProjectAssignments []ProjectAssignment `gorm:"foreignKey:ContractID"`
	Transactions     []Transaction      `gorm:"foreignKey:ContractID"`
}

// BeforeCreate hook for Contract
func (c *Contract) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// ContractTemplate represents contract templates
type ContractTemplate struct {
	ID              uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	Name            string    `gorm:"type:varchar(100);not null" json:"name"`
	Category        string    `gorm:"type:enum('work','delivery','maintenance');not null" json:"category"`
	TemplateContent string    `gorm:"type:text;not null" json:"template_content"`
	Variables       *string   `gorm:"type:json;comment:Daftar variabel yang bisa diubah" json:"variables"`
	IsDefault       bool      `gorm:"default:false" json:"is_default"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	CreatedBy       *uuid.UUID `gorm:"type:char(36)" json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`

	// Relationships
	Creator *User `gorm:"foreignKey:CreatedBy"`
}

// BeforeCreate hook for ContractTemplate
func (ct *ContractTemplate) BeforeCreate(tx *gorm.DB) error {
	if ct.ID == uuid.Nil {
		ct.ID = uuid.New()
	}
	return nil
}

// Transaction represents financial transactions
type Transaction struct {
	ID             uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	FromUserID     uuid.UUID  `gorm:"type:char(36);not null" json:"from_user_id"`
	ToUserID       uuid.UUID  `gorm:"type:char(36);not null" json:"to_user_id"`
	ProjectID      *uuid.UUID `gorm:"type:char(36)" json:"project_id"`
	DeliveryID     *uuid.UUID `gorm:"type:char(36)" json:"delivery_id"`
	ContractID     *uuid.UUID `gorm:"type:char(36)" json:"contract_id"`

	TransactionType  string   `gorm:"type:enum('work_payment','delivery_payment','platform_fee','refund');not null" json:"transaction_type"`
	Amount           float64  `gorm:"type:decimal(10,2);not null" json:"amount"`
	PlatformFee      float64  `gorm:"type:decimal(10,2);default:0" json:"platform_fee"`
	TaxAmount        float64  `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	TotalAmount      float64  `gorm:"type:decimal(10,2);not null" json:"total_amount"`

	Status         string  `gorm:"type:enum('pending','hold','released','cancelled','refunded');default:pending" json:"status"`
	PaymentMethod  *string `gorm:"type:varchar(50)" json:"payment_method"`

	// Hold System
	HoldUntil           *time.Time `json:"hold_until"`
	ReleaseConditions   *string    `gorm:"type:text" json:"release_conditions"`
	AutoReleaseDays     int        `gorm:"default:7" json:"auto_release_days"`

	TransactionDate time.Time  ` json:"transaction_date"`
	ReleasedAt      *time.Time `json:"released_at"`

	Description       *string `gorm:"type:text" json:"description"`
	ReferenceNumber   *string `gorm:"type:varchar(100);uniqueIndex" json:"reference_number"`

	// Relationships
	FromUser     User         `gorm:"foreignKey:FromUserID;constraint:OnDelete:CASCADE"`
	ToUser       User         `gorm:"foreignKey:ToUserID;constraint:OnDelete:CASCADE"`
	Project      *Project     `gorm:"foreignKey:ProjectID"`
	Delivery     *Delivery    `gorm:"foreignKey:DeliveryID"`
	Contract     *Contract    `gorm:"foreignKey:ContractID"`
	PaymentLogs  []PaymentLog `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	Reviews      []Review     `gorm:"foreignKey:TransactionID"`
	Disputes     []Dispute    `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for Transaction
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// PaymentMethod represents user payment methods
type PaymentMethod struct {
	ID            uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	UserID        uuid.UUID `gorm:"type:char(36);not null" json:"user_id"`
	MethodType    string    `gorm:"type:enum('bank_transfer','ewallet','credit_card','debit_card','virtual_account');not null" json:"method_type"`
	MethodName    string    `gorm:"type:varchar(100);not null" json:"method_name"`
	MethodDetails string    `gorm:"type:json;not null" json:"method_details"`
	IsDefault     bool      `gorm:"default:false" json:"is_default"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for PaymentMethod
func (pm *PaymentMethod) BeforeCreate(tx *gorm.DB) error {
	if pm.ID == uuid.Nil {
		pm.ID = uuid.New()
	}
	return nil
}

// PaymentLog represents payment gateway logs
type PaymentLog struct {
	ID                    uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	TransactionID         uuid.UUID `gorm:"type:char(36);not null" json:"transaction_id"`
	GatewayName           *string   `gorm:"type:varchar(50)" json:"gateway_name"`
	GatewayTransactionID  *string   `gorm:"type:varchar(100)" json:"gateway_transaction_id"`
	GatewayResponse       *string   `gorm:"type:json" json:"gateway_response"`
	Status                string    `gorm:"type:varchar(50);not null" json:"status"`
	Amount                *float64  `gorm:"type:decimal(10,2)" json:"amount"`
	Timestamp             time.Time ` json:"timestamp"`

	// Relationships
	Transaction Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for PaymentLog
func (pl *PaymentLog) BeforeCreate(tx *gorm.DB) error {
	if pl.ID == uuid.Nil {
		pl.ID = uuid.New()
	}
	return nil
}