package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MitraCooperation represents a transaction/cooperation agreement between Farmer and Mitra
type MitraCooperation struct {
	ID                    uuid.UUID  `gorm:"type:char(36);primaryKey" json:"id"`
	MitraID               uuid.UUID  `gorm:"type:char(36);not null;index" json:"mitra_id"`
	FarmerID              uuid.UUID  `gorm:"type:char(36);not null;index" json:"farmer_id"`
	InitiatorType         string     `gorm:"type:enum('mitra','farmer');not null" json:"initiator_type"`
	Title                 string     `gorm:"type:varchar(150);not null" json:"title"`
	Description           string     `gorm:"type:text;not null" json:"description"`
	ProposedAmount        float64    `gorm:"type:decimal(15,2);not null" json:"proposed_amount"`
	AgreedAmount          *float64   `gorm:"type:decimal(15,2)" json:"agreed_amount"`
	PlatformFeePercentage float64    `gorm:"type:decimal(5,2);default:11.00" json:"platform_fee_percentage"`
	StartDate             *time.Time `gorm:"type:date" json:"start_date"`
	EndDate               *time.Time `gorm:"type:date" json:"end_date"`
	Status                string     `gorm:"type:enum('submitted','reviewed','waiting_payment','escrowed','contract_generated','completed','rejected','cancelled');default:'submitted'" json:"status"`
	Notes                 *string    `gorm:"type:text" json:"notes"`
	ContractID            *uuid.UUID `gorm:"type:char(36);index" json:"contract_id"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`

	// Relationships
	Mitra    *MitraProfile `gorm:"foreignKey:MitraID;references:UserID" json:"mitra,omitempty"`
	Farmer   *Farmer       `gorm:"foreignKey:FarmerID;references:UserID" json:"farmer,omitempty"`
	Contract *Contract     `gorm:"foreignKey:ContractID;references:ID;constraint:false" json:"contract,omitempty"`
}

func (mc *MitraCooperation) BeforeCreate(tx *gorm.DB) error {
	if mc.ID == uuid.Nil {
		mc.ID = uuid.New()
	}
	if mc.PlatformFeePercentage == 0 {
		mc.PlatformFeePercentage = 11.00
	}
	return nil
}

// MitraReview represents bi-directional reviews between Farmer and Mitra after completion
type MitraReview struct {
	ID            uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	CooperationID uuid.UUID `gorm:"type:char(36);not null;index" json:"cooperation_id"`
	ReviewerType  string    `gorm:"type:enum('farmer','mitra');not null" json:"reviewer_type"`
	ReviewerID    uuid.UUID `gorm:"type:char(36);not null;index" json:"reviewer_id"`
	Rating        int       `gorm:"not null" json:"rating"`
	Comment       *string   `gorm:"type:text" json:"comment"`
	CreatedAt     time.Time `json:"created_at"`

	// Relationships
	Cooperation *MitraCooperation `gorm:"foreignKey:CooperationID;references:ID" json:"cooperation,omitempty"`
	Reviewer    *User             `gorm:"foreignKey:ReviewerID;references:ID" json:"reviewer,omitempty"`
}

func (mr *MitraReview) BeforeCreate(tx *gorm.DB) error {
	if mr.ID == uuid.Nil {
		mr.ID = uuid.New()
	}
	return nil
}
