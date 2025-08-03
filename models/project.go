package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Project represents agricultural projects
type Project struct {
	ID                   uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	FarmerID             uuid.UUID  `gorm:"type:char(36);not null" json:"farmer_id"`
	FarmLocationID       *uuid.UUID `gorm:"type:char(36)" json:"farm_location_id"`
	Title                string     `gorm:"type:varchar(100);not null" json:"title"`
	Description          string     `gorm:"type:text;not null" json:"description"`
	ProjectType          string     `gorm:"type:enum('planting','maintenance','harvesting','irrigation','pest_control');not null" json:"project_type"`
	RequiredSkills       string     `gorm:"type:json;not null" json:"required_skills"` // JSON array as string
	WorkersNeeded        int        `gorm:"default:1" json:"workers_needed"`
	StartDate            time.Time  `gorm:"type:date;not null" json:"start_date"`
	EndDate              time.Time  `gorm:"type:date;not null" json:"end_date"`
	BudgetMin            *float64   `gorm:"type:decimal(10,2)" json:"budget_min"`
	BudgetMax            *float64   `gorm:"type:decimal(10,2)" json:"budget_max"`
	UrgencyLevel         string     `gorm:"type:enum('low','medium','high','urgent');default:medium" json:"urgency_level"`
	Status               string     `gorm:"type:enum('draft','open','in_progress','completed','cancelled');default:draft" json:"status"`
	CompletionPercentage int        `gorm:"default:0" json:"completion_percentage"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Relationships
	Farmer              Farmer               `gorm:"foreignKey:FarmerID;constraint:OnDelete:CASCADE"`
	FarmLocation        *FarmLocation        `gorm:"foreignKey:FarmLocationID"`
	ProjectApplications []ProjectApplication `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	ProjectAssignments  []ProjectAssignment  `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Contracts           []Contract           `gorm:"foreignKey:ProjectID"`
	Transactions        []Transaction        `gorm:"foreignKey:ProjectID"`
	Schedules           []Schedule           `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// BeforeCreate hook for Project
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ProjectApplication represents worker applications to projects
type ProjectApplication struct {
	ID                      uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	ProjectID               uuid.UUID  `gorm:"type:char(36);not null" json:"project_id"`
	WorkerID                uuid.UUID  `gorm:"type:char(36);not null" json:"worker_id"`
	ApplicationDate         time.Time  `json:"application_date"`
	ProposedRate            *float64   `gorm:"type:decimal(10,2)" json:"proposed_rate"`
	EstimatedCompletionDays *int       `json:"estimated_completion_days"`
	Status                  string     `gorm:"type:enum('pending','accepted','rejected','withdrawn');default:pending" json:"status"`
	Note                    *string    `gorm:"type:text" json:"note"`
	CoverLetter             *string    `gorm:"type:text" json:"cover_letter"`
	ReviewedAt              *time.Time `json:"reviewed_at"`
	ReviewedBy              *uuid.UUID `gorm:"type:char(36)" json:"reviewed_by"`

	// Relationships
	Project      Project `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Worker       Worker  `gorm:"foreignKey:WorkerID;constraint:OnDelete:CASCADE"`
	ReviewedUser *User   `gorm:"foreignKey:ReviewedBy"`
}

// BeforeCreate hook for ProjectApplication
func (pa *ProjectApplication) BeforeCreate(tx *gorm.DB) error {
	if pa.ID == uuid.Nil {
		pa.ID = uuid.New()
	}
	return nil
}

// ProjectAssignment represents worker assignments to projects
type ProjectAssignment struct {
	ID              uuid.UUID  `gorm:"type:char(36);primary_key;default:(UUID())" json:"id"`
	ProjectID       uuid.UUID  `gorm:"type:char(36);not null" json:"project_id"`
	WorkerID        uuid.UUID  `gorm:"type:char(36);not null" json:"worker_id"`
	ContractID      *uuid.UUID `gorm:"type:char(36)" json:"contract_id"`
	AssignedAt      time.Time  `json:"assigned_at"`
	StartDate       *time.Time `gorm:"type:date" json:"start_date"`
	EndDate         *time.Time `gorm:"type:date" json:"end_date"`
	AgreedRate      *float64   `gorm:"type:decimal(10,2)" json:"agreed_rate"`
	Status          string     `gorm:"type:enum('assigned','started','paused','completed','terminated');default:assigned" json:"status"`
	CompletionNotes *string    `gorm:"type:text" json:"completion_notes"`
	CompletedAt     *time.Time `json:"completed_at"`

	// Relationships
	Project  Project   `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
	Worker   Worker    `gorm:"foreignKey:WorkerID;constraint:OnDelete:CASCADE"`
	Contract *Contract `gorm:"foreignKey:ContractID"`
}

// BeforeCreate hook for ProjectAssignment
func (pa *ProjectAssignment) BeforeCreate(tx *gorm.DB) error {
	if pa.ID == uuid.Nil {
		pa.ID = uuid.New()
	}
	return nil
}
