package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SystemSetting represents system configuration
type SystemSetting struct {
	Key         string     `gorm:"type:varchar(100);primary_key" json:"key"`
	Value       string     `gorm:"type:text;not null" json:"value"`
	DataType    string     `gorm:"type:enum('string','number','boolean','json');default:string" json:"data_type"`
	Description *string    `gorm:"type:text" json:"description"`
	IsPublic    bool       `gorm:"default:false" json:"is_public"`
	UpdatedBy   *uuid.UUID `gorm:"type:char(36)" json:"updated_by"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	UpdatedByUser *User `gorm:"foreignKey:UpdatedBy"`
}

// ActivityLog represents user activity logging
type ActivityLog struct {
	ID         uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	UserID     *uuid.UUID `gorm:"type:char(36)" json:"user_id"`
	Action     string     `gorm:"type:varchar(100);not null" json:"action"`
	EntityType *string    `gorm:"type:varchar(50)" json:"entity_type"`
	EntityID   *uuid.UUID `gorm:"type:char(36)" json:"entity_id"`
	Details    *string    `gorm:"type:json" json:"details"`
	IPAddress  *string    `gorm:"type:varchar(45)" json:"ip_address"` // IPv6 compatible
	UserAgent  *string    `gorm:"type:text" json:"user_agent"`
	CreatedAt  time.Time  `json:"created_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID"`
}

// BeforeCreate hook for ActivityLog
func (al *ActivityLog) BeforeCreate(tx *gorm.DB) error {
	if al.ID == uuid.Nil {
		al.ID = uuid.New()
	}
	return nil
}

// UserSession represents user login sessions
type UserSession struct {
	ID           uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	UserID       uuid.UUID  `gorm:"type:char(36);not null" json:"user_id"`
	SessionToken string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"session_token"`
	DeviceInfo   *string    `gorm:"type:json" json:"device_info"`
	IPAddress    *string    `gorm:"type:varchar(45)" json:"ip_address"` // IPv6 compatible
	LocationInfo *string    `gorm:"type:json" json:"location_info"`
	IsActive     bool       `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time  `json:"created_at"`
	ExpiresAt    time.Time  `gorm:"not null" json:"expires_at"`
	LastActivity time.Time `json:"last_activity"`


	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for UserSession
func (us *UserSession) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}

	if us.LastActivity.IsZero() {
		us.LastActivity = time.Now()
	}
	return nil
}

// AIRecommendation represents AI-based recommendations
type AIRecommendation struct {
	ID               uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	UserID           uuid.UUID  `gorm:"type:char(36);not null" json:"user_id"`
	RecommendationType string   `gorm:"type:enum('worker','expedition','schedule','price','route');not null" json:"recommendation_type"`
	ContextData      string     `gorm:"type:json;not null;comment:Data input untuk rekomendasi" json:"context_data"`
	Criteria         string     `gorm:"type:json;not null" json:"criteria"`
	RecommendedItems string     `gorm:"type:json;not null" json:"recommended_items"`
	MatchScores      string     `gorm:"type:json;not null" json:"match_scores"`
	AlgorithmVersion *string    `gorm:"type:varchar(20)" json:"algorithm_version"`
	ConfidenceScore  *float64   `json:"confidence_score"`
	CreatedAt        time.Time  `json:"created_at"`
	Accepted         bool       `gorm:"default:false" json:"accepted"`
	AcceptedAt       *time.Time `json:"accepted_at"`
	FeedbackScore    *int       `gorm:"check:feedback_score >= 1 AND feedback_score <= 5" json:"feedback_score"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for AIRecommendation
func (air *AIRecommendation) BeforeCreate(tx *gorm.DB) error {
	if air.ID == uuid.Nil {
		air.ID = uuid.New()
	}
	return nil
}

// UserPreference represents user preferences for AI
type UserPreference struct {
	UserID           uuid.UUID `gorm:"type:char(36);primary_key" json:"user_id"`
	Preferences      string    `gorm:"type:json;not null;" json:"preferences"`
	BehaviorPatterns string    `gorm:"type:json;" json:"behavior_patterns"`
	SuccessHistory   string    `gorm:"type:json;'" json:"success_history"`
	LearningData     string    `gorm:"type:json;" json:"learning_data"`
	UpdatedAt        time.Time `json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

// MLTrainingData represents machine learning training data
type MLTrainingData struct {
	ID            uuid.UUID `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	UserID        *uuid.UUID `gorm:"type:char(36)" json:"user_id"`
	FeatureVector string    `gorm:"type:json;not null" json:"feature_vector"`
	Label         string    `gorm:"type:varchar(100);not null" json:"label"`
	OutcomeScore  *float64  `json:"outcome_score"`
	ModelType     string    `gorm:"type:enum('matching','pricing','scheduling','routing');not null" json:"model_type"`
	CreatedAt     time.Time `json:"created_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID"`
}

// BeforeCreate hook for MLTrainingData
func (mltd *MLTrainingData) BeforeCreate(tx *gorm.DB) error {
	if mltd.ID == uuid.Nil {
		mltd.ID = uuid.New()
	}
	return nil
}