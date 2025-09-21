package dto

import (
	"time"

	"github.com/google/uuid"
)

// SignContractResponse adalah DTO untuk response setelah pekerja menandatangani kontrak.
type SignContractResponse struct {
	ContractID   uuid.UUID  `json:"contract_id"`
	ProjectTitle string     `json:"project_title,omitempty"` // omitempty agar tidak muncul di kontrak delivery
	DeliveryID   *uuid.UUID `json:"delivery_id,omitempty"`   // omitempty agar tidak muncul di kontrak kerja
	Status       string     `json:"status"`
	SignedAt     time.Time  `json:"signed_at"`
	Message      string     `json:"message"`
}

// MyContractResponse adalah DTO untuk menampilkan daftar kontrak milik pengguna.
type MyContractResponse struct {
	ContractID   uuid.UUID `json:"contract_id"`
	ContractType string    `json:"contract_type"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	OfferedAt    time.Time `json:"offered_at"`
}