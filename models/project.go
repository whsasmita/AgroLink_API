package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Project: Mewakili sebuah "lowongan pekerjaan" dari petani.
type Project struct {
	ID             uuid.UUID  `gorm:"type:char(36);primary_key"`
	FarmerID       uuid.UUID  `gorm:"type:char(36);not null"`
	FarmLocationID *uuid.UUID `gorm:"type:char(36)"`
	Title          string     `gorm:"type:varchar(100);not null"`
	Description    string     `gorm:"type:text;not null"`
	ProjectType    string     `gorm:"type:enum('planting','maintenance','harvesting','irrigation','pest_control');not null"`
	RequiredSkills string     `gorm:"type:json"` // JSON array of strings
	WorkersNeeded  int        `gorm:"default:1"`
	StartDate      time.Time  `gorm:"type:date;not null"`
	EndDate        time.Time  `gorm:"type:date;not null"`
	PaymentRate    *float64   `gorm:"type:decimal(10,2)"`                         // Tarif pembayaran
	PaymentType    string     `gorm:"type:enum('per_day','per_hour','lump_sum')"` // Jenis pembayaran
	Status         string     `gorm:"type:enum('open','in_progress','completed','cancelled');default:open"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relasi
	Farmer              Farmer
	FarmLocation        *FarmLocation
	ProjectApplications []ProjectApplication
	ProjectAssignments  []ProjectAssignment
}

func (u *Project) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// ProjectApplication: Lamaran dari pekerja ke sebuah proyek.
type ProjectApplication struct {
	ID        uuid.UUID `gorm:"type:char(36);primary_key"`
	ProjectID uuid.UUID `gorm:"type:char(36);not null"`
	WorkerID  uuid.UUID `gorm:"type:char(36);not null"`
	Message   *string   `gorm:"type:text"` // Pesan singkat dari pelamar
	Status    string    `gorm:"type:enum('pending','accepted','rejected','withdrawn');default:pending"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relasi
	Project Project
	Worker  Worker
}

func (u *ProjectApplication) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type ProjectAssignment struct {
	ID         uuid.UUID `gorm:"type:char(36);primary_key"`
	ProjectID  uuid.UUID `gorm:"type:char(36);not null"`
	WorkerID   uuid.UUID `gorm:"type:char(36);not null"`
	ContractID uuid.UUID `gorm:"type:char(36);not null"` // Kunci penghubung ke kontrak
	AgreedRate float64   `gorm:"type:decimal(10,2)"`     // Tarif final sesuai kontrak
	Status     string    `gorm:"type:enum('assigned','started','completed','terminated');default:assigned"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	// Relasi
	Project  Project  `json:"-"`
	Worker   Worker   `json:"-"`
	Contract Contract `json:"-"`
}

func (u *ProjectAssignment) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
