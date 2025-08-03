package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Schedule represents scheduling system
type Schedule struct {
	ID         uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	UserID     uuid.UUID  `gorm:"type:char(36);not null" json:"user_id"`
	ProjectID  *uuid.UUID `gorm:"type:char(36)" json:"project_id"`
	DeliveryID *uuid.UUID `gorm:"type:char(36)" json:"delivery_id"`

	Title        string  `gorm:"type:varchar(200);not null" json:"title"`
	Description  *string `gorm:"type:text" json:"description"`
	ScheduleType string  `gorm:"type:enum('work','delivery','meeting','maintenance','inspection');not null" json:"schedule_type"`

	StartDatetime time.Time `gorm:"not null" json:"start_datetime"`
	EndDatetime   time.Time `gorm:"not null" json:"end_datetime"`
	AllDay        bool      `gorm:"default:false" json:"all_day"`

	LocationLat  *float64 `gorm:"type:decimal(10,8)" json:"location_lat"`
	LocationLng  *float64 `gorm:"type:decimal(11,8)" json:"location_lng"`
	LocationName *string  `gorm:"type:varchar(200)" json:"location_name"`

	Status   string `gorm:"type:enum('scheduled','in_progress','completed','cancelled','postponed');default:scheduled" json:"status"`
	Priority string `gorm:"type:enum('low','medium','high','urgent');default:medium" json:"priority"`

	// Reminder Settings
	ReminderSettings *string `gorm:"type:json;comment:Reminder dalam menit sebelum acara" json:"reminder_settings"`

	// Recurrence
	IsRecurring       bool    `gorm:"default:false" json:"is_recurring"`
	RecurrencePattern *string `gorm:"type:json" json:"recurrence_pattern"`

	Notes     *string   `gorm:"type:text" json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User                  User                   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Project               *Project               `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Delivery              *Delivery              `gorm:"foreignKey:DeliveryID;constraint:OnDelete:CASCADE"`
	ScheduleNotifications []ScheduleNotification `gorm:"foreignKey:ScheduleID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for Schedule
func (s *Schedule) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	// Set default reminder settings jika nil
	if s.ReminderSettings == nil {
		defaultReminder := `{"enabled": true, "times": [1440, 60, 15]}`
		s.ReminderSettings = &defaultReminder
	}
	return nil
}

// ScheduleNotification represents schedule notifications
type ScheduleNotification struct {
	ID               uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	ScheduleID       uuid.UUID  `gorm:"type:char(36);not null" json:"schedule_id"`
	UserID           uuid.UUID  `gorm:"type:char(36);not null" json:"user_id"`
	NotificationType string     `gorm:"type:enum('reminder','update','cancellation','conflict');not null" json:"notification_type"`
	Message          string     `gorm:"type:text;not null" json:"message"`
	ScheduledFor     time.Time  `gorm:"not null" json:"scheduled_for"`
	SentAt           *time.Time `json:"sent_at"`
	Status           string     `gorm:"type:enum('pending','sent','failed');default:pending" json:"status"`

	// Relationships
	Schedule Schedule `gorm:"foreignKey:ScheduleID;constraint:OnDelete:CASCADE"`
	User     User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for ScheduleNotification
func (sn *ScheduleNotification) BeforeCreate(tx *gorm.DB) error {
	if sn.ID == uuid.Nil {
		sn.ID = uuid.New()
	}
	return nil
}

// Notification represents general notifications
type Notification struct {
	ID       uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	UserID   uuid.UUID `gorm:"type:char(36);not null" json:"user_id"`
	Title    string    `gorm:"type:varchar(200);not null" json:"title"`
	Message  string    `gorm:"type:text;not null" json:"message"`
	Type     string    `gorm:"type:enum('info','warning','success','error');default:info" json:"type"`
	Category string    `gorm:"type:enum('job','payment','delivery','system','ai_recommendation','schedule');not null" json:"category"`

	// Related data
	RelatedEntityType *string    `gorm:"type:enum('project','delivery','transaction','contract','schedule')" json:"related_entity_type"`
	RelatedEntityID   *uuid.UUID `gorm:"type:char(36)" json:"related_entity_id"`

	// Actions
	ActionURL   *string `gorm:"type:text" json:"action_url"`
	ActionLabel *string `gorm:"type:varchar(50)" json:"action_label"`

	// Status
	ReadStatus bool       `gorm:"default:false" json:"read_status"`
	ReadAt     *time.Time `json:"read_at"`
	Archived   bool       `gorm:"default:false" json:"archived"`

	// Delivery
	PushSent  bool `gorm:"default:false" json:"push_sent"`
	EmailSent bool `gorm:"default:false" json:"email_sent"`
	SMSSent   bool `gorm:"default:false" json:"sms_sent"`

	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for Notification
func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}
