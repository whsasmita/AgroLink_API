package dto

import "github.com/google/uuid"

type DirectOfferRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Location    string  `json:"location" binding:"required"`
	StartDate   string  `json:"start_date" binding:"required"` // Format "YYYY-MM-DD"
	EndDate     string  `json:"end_date" binding:"required"`
	PaymentRate float64 `json:"payment_rate" binding:"required,gt=0"`
}

type DirectOfferResponse struct {
	ContractID   uuid.UUID `json:"contract_id"`
	ProjectTitle string    `json:"project_title"`
	WorkerName   string    `json:"worker_name"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
}