package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateReviewInput struct {
	ProjectID      uuid.UUID `json:"-"` // Diambil dari URL, bukan body
	ReviewerID     uuid.UUID `json:"-"` // Diambil dari token JWT
	ReviewedWorkerID uuid.UUID `json:"-"` // Diambil dari URL

	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}


// CreateReviewResponse adalah DTO untuk response setelah berhasil membuat ulasan.
type CreateReviewResponse struct {
	ID               uuid.UUID `json:"id"`
	ProjectID        uuid.UUID `json:"project_id"`
	ReviewerID       uuid.UUID `json:"reviewer_id"`
	ReviewedWorkerID uuid.UUID `json:"reviewed_worker_id"`
	Rating           int       `json:"rating"`
	Comment          *string   `json:"comment"`
	CreatedAt        time.Time `json:"created_at"`
	Message          string    `json:"message"`
}