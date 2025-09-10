package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Review represents user reviews and ratings
type Review struct {
	ID             uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	ReviewerID     uuid.UUID  `gorm:"type:char(36);not null" json:"reviewer_id"`
	ReviewedUserID uuid.UUID  `gorm:"type:char(36);not null" json:"reviewed_user_id"`
	ProjectID      *uuid.UUID `gorm:"type:char(36)" json:"project_id"`
	DeliveryID     *uuid.UUID `gorm:"type:char(36)" json:"delivery_id"`
	TransactionID  *uuid.UUID `gorm:"type:char(36)" json:"transaction_id"`

	Rating  int     `gorm:"not null;check:rating >= 1 AND rating <= 5" json:"rating"`
	Comment *string `gorm:"type:text" json:"comment"`

	// Review Categories
	ReviewType string `gorm:"type:enum('work_quality','delivery_service','communication','punctuality','overall');default:overall" json:"review_type"`

	// Detailed Ratings
	QualityRating       *int `gorm:"check:quality_rating >= 1 AND quality_rating <= 5" json:"quality_rating"`
	CommunicationRating *int `gorm:"check:communication_rating >= 1 AND communication_rating <= 5" json:"communication_rating"`
	PunctualityRating   *int `gorm:"check:punctuality_rating >= 1 AND punctuality_rating <= 5" json:"punctuality_rating"`

	// Media
	ReviewImages *string `gorm:"type:json" json:"review_images"` // JSON array as string

	// Status
	IsAnonymous bool `gorm:"default:false" json:"is_anonymous"`
	IsFeatured  bool `gorm:"default:false" json:"is_featured"`
	IsVerified  bool `gorm:"default:false" json:"is_verified"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Reviewer     User         `gorm:"foreignKey:ReviewerID;constraint:OnDelete:CASCADE"`
	ReviewedUser User         `gorm:"foreignKey:ReviewedUserID;constraint:OnDelete:CASCADE"`
	Project      *Project     `gorm:"foreignKey:ProjectID"`
	Delivery     *Delivery    `gorm:"foreignKey:DeliveryID"`
	Transaction  *Transaction `gorm:"foreignKey:TransactionID"`
}

// BeforeCreate hook for Review
func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// SupportTicket represents customer support tickets
type SupportTicket struct {
	ID         uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	UserID     uuid.UUID  `gorm:"type:char(36);not null" json:"user_id"`
	AssignedCS *uuid.UUID `gorm:"type:char(36)" json:"assigned_cs"`

	TicketNumber string `gorm:"type:varchar(20);uniqueIndex;not null" json:"ticket_number"`
	Category     string `gorm:"type:enum('technical','payment','delivery','account','dispute','other');not null" json:"category"`
	Subject      string `gorm:"type:varchar(200);not null" json:"subject"`
	Description  string `gorm:"type:text;not null" json:"description"`

	Status   string `gorm:"type:enum('open','in_progress','waiting_customer','resolved','closed');default:open" json:"status"`
	Priority string `gorm:"type:enum('low','medium','high','urgent');default:medium" json:"priority"`

	// Related entities
	RelatedProjectID     *uuid.UUID `gorm:"type:char(36)" json:"related_project_id"`
	RelatedDeliveryID    *uuid.UUID `gorm:"type:char(36)" json:"related_delivery_id"`
	RelatedTransactionID *uuid.UUID `gorm:"type:char(36)" json:"related_transaction_id"`

	// Resolution
	Resolution *string    `gorm:"type:text" json:"resolution"`
	ResolvedAt *time.Time `json:"resolved_at"`
	ClosedAt   *time.Time `json:"closed_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User               User             `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	AssignedCSUser     *User            `gorm:"foreignKey:AssignedCS"`
	RelatedProject     *Project         `gorm:"foreignKey:RelatedProjectID"`
	RelatedDelivery    *Delivery        `gorm:"foreignKey:RelatedDeliveryID"`
	RelatedTransaction *Transaction     `gorm:"foreignKey:RelatedTransactionID"`
	SupportMessages    []SupportMessage `gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for SupportTicket
func (st *SupportTicket) BeforeCreate(tx *gorm.DB) error {
	if st.ID == uuid.Nil {
		st.ID = uuid.New()
	}
	return nil
}

// SupportMessage represents messages in support tickets
type SupportMessage struct {
	ID          uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	TicketID    uuid.UUID `gorm:"type:char(36);not null" json:"ticket_id"`
	SenderID    uuid.UUID `gorm:"type:char(36);not null" json:"sender_id"`
	Message     string    `gorm:"type:text;not null" json:"message"`
	Attachments *string   `gorm:"type:json" json:"attachments"` // JSON array as string
	IsInternal  bool      `gorm:"default:false;comment:Internal CS notes" json:"is_internal"`
	CreatedAt   time.Time `json:"created_at"`

	// Relationships
	Ticket SupportTicket `gorm:"foreignKey:TicketID;constraint:OnDelete:CASCADE"`
	Sender User          `gorm:"foreignKey:SenderID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for SupportMessage
func (sm *SupportMessage) BeforeCreate(tx *gorm.DB) error {
	if sm.ID == uuid.Nil {
		sm.ID = uuid.New()
	}
	return nil
}

// Dispute represents disputes between users
type Dispute struct {
	ID            uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	TransactionID uuid.UUID  `gorm:"type:char(36);not null" json:"transaction_id"`
	RaisedBy      uuid.UUID  `gorm:"type:char(36);not null" json:"raised_by"`
	AgainstUser   uuid.UUID  `gorm:"type:char(36);not null" json:"against_user"`
	CSAssigned    *uuid.UUID `gorm:"type:char(36)" json:"cs_assigned"`

	DisputeType    string   `gorm:"type:enum('payment_issue','work_quality','delivery_problem','contract_breach','other');not null" json:"dispute_type"`
	Reason         string   `gorm:"type:text;not null" json:"reason"`
	Evidence       *string  `gorm:"type:json" json:"evidence"` // JSON array as string
	AmountDisputed *float64 `gorm:"type:decimal(10,2)" json:"amount_disputed"`

	Status   string `gorm:"type:enum('open','under_review','mediation','resolved','closed','escalated');default:open" json:"status"`
	Priority string `gorm:"type:enum('low','medium','high','urgent');default:medium" json:"priority"`

	// Resolution
	Resolution     *string  `gorm:"type:text" json:"resolution"`
	ResolutionType *string  `gorm:"type:enum('refund_to_buyer','pay_to_seller','partial_refund','no_action')" json:"resolution_type"`
	ResolvedAmount *float64 `gorm:"type:decimal(10,2)" json:"resolved_amount"`

	CreatedAt  time.Time  `json:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at"`
	ClosedAt   *time.Time `json:"closed_at"`

	// Relationships
	Transaction      Transaction `gorm:"foreignKey:TransactionID;constraint:OnDelete:CASCADE"`
	RaisedByUser     User        `gorm:"foreignKey:RaisedBy;constraint:OnDelete:CASCADE"`
	AgainstUserModel User        `gorm:"foreignKey:AgainstUser;constraint:OnDelete:CASCADE"`
	CSAssignedUser   *User       `gorm:"foreignKey:CSAssigned"`
}

// BeforeCreate hook for Dispute
func (d *Dispute) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}
