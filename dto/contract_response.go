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
}