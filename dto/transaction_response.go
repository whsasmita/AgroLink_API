package dto

import (
	"time"

	"github.com/google/uuid"
)

type PaymentInitiationResponse struct {
	SnapToken   string  `json:"snap_token"`
	OrderID     string  `json:"order_id"`
	Amount      float64 `json:"amount"`
	RedirectURL string  `json:"redirect_url"`
}

type CreateTransactionRequest struct {
	ContractID  string  `json:"contract_id" binding:"required"`
	ToUserID    string  `json:"to_user_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type TransactionHistoryResponse struct {
	ID              uuid.UUID `json:"id"`
	ProjectTitle    string    `json:"project_title"`
	OtherPartyName  string    `json:"other_party_name"`
	RoleInTx        string    `json:"role_in_tx"`
	TotalAmount     float64   `json:"total_amount"`
	Status          string    `json:"status"`
	TransactionDate time.Time `json:"transaction_date"`
	ReleasedAt      *time.Time `json:"released_at,omitempty"`
}