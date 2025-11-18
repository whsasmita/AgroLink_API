package dto

import (
	"time"

	"github.com/google/uuid"
)

// PayoutDetailResponse adalah DTO untuk menampilkan detail payout di dashboard admin.
type PayoutDetailResponse struct {
	PayoutID          uuid.UUID `json:"payout_id"`
	PayeeID           uuid.UUID `json:"payee_id"`           // ID Worker atau Driver
	PayeeName         string    `json:"payee_name"`         // Nama Worker atau Driver
	PayeeType         string    `json:"payee_type"`         // "worker" atau "driver"
	ContextTitle      string    `json:"context_title"`      // Judul Proyek atau Deskripsi Pengiriman
	Amount            float64   `json:"amount"`
	ReleasedAt        time.Time `json:"released_at"`
	BankName          string    `json:"bank_name"`
	BankAccountNumber string    `json:"bank_account_number"`
	BankAccountHolder string    `json:"bank_account_holder"`
}