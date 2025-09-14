package dto

import (
	"time"

	"github.com/google/uuid"
)

// DTO untuk menampilkan daftar pelamar kepada petani.
type ApplicationResponse struct {
	ID              uuid.UUID `json:"id"`
	WorkerID        uuid.UUID `json:"worker_id"`
	WorkerName      string    `json:"worker_name"`
	Message         *string   `json:"message"`
	ApplicationDate time.Time `json:"application_date"`
	Status          string    `json:"status"`
}

type AcceptApplicationResponse struct {
	ContractID   uuid.UUID `json:"contract_id"`
	ProjectTitle string    `json:"project_title"`
	WorkerName   string    `json:"worker_name"`
	Message      string    `json:"message"`
}

// DTO untuk body request saat pekerja melamar.
type ApplyProjectInput struct {
	Message string `json:"message"`
}

// DTO untuk response setelah pekerja berhasil melamar.
type ApplicationSubmissionResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	WorkerID  uuid.UUID `json:"worker_id"`
	Message   *string   `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

