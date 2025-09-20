package dto

import (
	"time"

	"github.com/google/uuid"
)

// SignContractResponse adalah DTO untuk response setelah pekerja menandatangani kontrak.
type SignContractResponse struct {
	ContractID   uuid.UUID `json:"contract_id"`
	ProjectTitle string    `json:"project_title"`
	Status       string    `json:"status"`
	SignedByWorker bool      `json:"signed_by_worker"`
	SignedAt     time.Time `json:"signed_at"`
	Message      string    `json:"message"`
	DeliveryID   *uuid.UUID `json:"delivery_id,omitempty"`
}

type MyContractResponse struct {
	ContractID   uuid.UUID `json:"contract_id"`
	ContractType string    `json:"contract_type"` // "work" atau "delivery"
	Title        string    `json:"title"`         // Judul Proyek atau deskripsi Pengiriman
	Status       string    `json:"status"`
	OfferedAt    time.Time `json:"offered_at"`
}