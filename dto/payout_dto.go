package dto

import (
	"time"

	"github.com/google/uuid"
)

// PayoutDetailResponse adalah DTO untuk menampilkan detail payout di dashboard admin.
type PayoutDetailResponse struct {
	PayoutID     uuid.UUID `json:"payout_id"`
	WorkerID     uuid.UUID `json:"worker_id"`
	WorkerName   string    `json:"worker_name"`
	ProjectTitle string    `json:"project_title"`
	Amount       float64   `json:"amount"`
	ReleasedAt   time.Time `json:"released_at"`
	Status       string    `json:"status"`
}