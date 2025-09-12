package dto

import "github.com/google/uuid"

type CreateReviewInput struct {
	ProjectID      uuid.UUID `json:"-"` // Diambil dari URL, bukan body
	ReviewerID     uuid.UUID `json:"-"` // Diambil dari token JWT
	ReviewedWorkerID uuid.UUID `json:"-"` // Diambil dari URL

	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Comment string `json:"comment"`
}