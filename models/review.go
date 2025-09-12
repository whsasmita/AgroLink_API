package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Review merepresentasikan ulasan dari Petani kepada Pekerja setelah proyek selesai.
type Review struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	ProjectID uuid.UUID `gorm:"type:char(36);not null"`
	ReviewerID  uuid.UUID `gorm:"type:char(36);not null"` // Foreign key ke User (Petani)
	ReviewedWorkerID uuid.UUID `gorm:"type:char(36);not null"` // Foreign key ke Worker

	Rating  int     `gorm:"not null;check:rating >= 1 AND rating <= 5"`
	Comment *string `gorm:"type:text"`

	CreatedAt time.Time
	UpdatedAt time.Time

	// Relasi
	Project        Project `gorm:"foreignKey:ProjectID"`
	Reviewer       User    `gorm:"foreignKey:ReviewerID"`
	ReviewedWorker Worker  `gorm:"foreignKey:ReviewedWorkerID"`
}

func (r *Review) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}