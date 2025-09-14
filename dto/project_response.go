package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateProjectRequest adalah DTO untuk membuat proyek baru.
type CreateProjectRequest struct {
	Title          string   `json:"title" binding:"required,max=100"`
	FarmLocationID string   `json:"farm_location_id" binding:"required,uuid"`
	Description    string   `json:"description" binding:"required"`
	ProjectType    string   `json:"project_type" binding:"required,oneof=planting maintenance harvesting irrigation pest_control"`
	RequiredSkills []string `json:"required_skills"`
	WorkersNeeded  int      `json:"workers_needed" binding:"required,min=1"`
	StartDate      string   `json:"start_date" binding:"required"` // Format: "YYYY-MM-DD"
	EndDate        string   `json:"end_date" binding:"required"`   // Format: "YYYY-MM-DD"
	PaymentRate    float64  `json:"payment_rate" binding:"required,min=0"`
	PaymentType    string   `json:"payment_type" binding:"required,oneof=per_day per_hour lump_sum"`
}

type ProjectBriefResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	ProjectType string    `json:"project_type"`
	PaymentRate *float64  `json:"payment_rate"`
	PaymentType string    `json:"payment_type"`
	StartDate   time.Time `json:"start_date"`
	// Kita bisa tambahkan info ringkas petani jika perlu
	// FarmerName string `json:"farmer_name"`
}

type FarmerInfoResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	// Tambahkan field lain jika perlu, misal: avatar, rating, dll.
}

// ProjectDetailResponse adalah DTO untuk response detail proyek.
type ProjectDetailResponse struct {
	ID             uuid.UUID          `json:"id"`
	Title          string             `json:"title"`
	Description    string             `json:"description"`
	ProjectType    string             `json:"project_type"`
	RequiredSkills []string           `json:"required_skills"`
	WorkersNeeded  int                `json:"workers_needed"`
	StartDate      time.Time          `json:"start_date"`
	EndDate        time.Time          `json:"end_date"`
	PaymentRate    *float64           `json:"payment_rate"`
	PaymentType    string             `json:"payment_type"`
	Status         string             `json:"status"`
	Farmer         FarmerInfoResponse `json:"farmer"`
}

type CreateProjectResponse struct {
	ID             uuid.UUID `json:"id"`
	FarmerID       uuid.UUID `json:"farmer_id"`
    FarmerName     string    `json:"farmer_name"`
	FarmLocationID *uuid.UUID `json:"farm_location_id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	ProjectType    string    `json:"project_type"`
	RequiredSkills []string  `json:"required_skills"`
	WorkersNeeded  int       `json:"workers_needed"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	PaymentRate    *float64  `json:"payment_rate"`
	PaymentType    string    `json:"payment_type"`
	Status         string    `json:"status"`
}

type MyProjectResponse struct {
	ProjectID    uuid.UUID  `json:"project_id"`
	ProjectTitle string     `json:"project_title"`
	ProjectStatus string    `json:"project_status"`
	// Menggunakan pointer dan omitempty agar field ini tidak muncul jika invoice belum ada.
	InvoiceID    *uuid.UUID `json:"invoice_id,omitempty"` 
}