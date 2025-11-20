package dto

import (
	"time"

	"github.com/google/uuid"
)

type MarkPayoutCompletedInput struct {
	TransferProofURL string `json:"transfer_proof_url" binding:"required,url"`
}

type ReviewVerificationInput struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Notes  string `json:"notes"` // Catatan dari admin (misal: "Foto KTP buram")
}

type TransactionDetailResponse struct {
	TransactionID   string    `json:"transaction_id"`   // ID dari Transaction atau ECommercePayment
	TransactionDate time.Time `json:"transaction_date"`
	AmountPaid      float64   `json:"amount_paid"`
	PaymentMethod   string    `json:"payment_method"`
	TransactionType string    `json:"transaction_type"` // "Jasa" (Project/Delivery) atau "Produk" (E-commerce)
	ContextInfo     string    `json:"context_info"`     // Misal: "Proyek: Panen Jagung" atau "Pesanan E-commerce"
	PayerName       string    `json:"payer_name"`       // Nama Petani (Jasa) atau Pembeli (E-commerce)
}

// PaginationResponse adalah wrapper untuk data yang dipaginasi
type AdminPaginationResponse struct {
	Data       interface{} `json:"data"`
	TotalItems int64       `json:"total_items"`
	TotalPages int         `json:"total_pages"`
	CurrentPage int        `json:"current_page"`
	Stats       *UserRoleStatsResponse `json:"stats,omitempty"`
}

type UserDetailResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	PhoneNumber   *string   `json:"phone_number"`
	Role          string    `json:"role"`
	IsActive      bool      `json:"is_active"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

type RevenueAnalyticsResponse struct {
	TotalRevenue      float64          `json:"total_revenue"`
	RevenueByService  float64          `json:"revenue_by_service"` // Dari Project/Delivery
	RevenueByProduct  float64          `json:"revenue_by_product"` // Dari E-commerce
	DailyTrend        []DailyDataPoint `json:"daily_trend"`        // Gabungan untuk grafik
}

type UserRoleStatsResponse struct {
	TotalUsers   int64 `json:"total_users"`   // semua user non-admin
	TotalGeneral int64 `json:"total_general"` // role = 'general'
	TotalFarmer  int64 `json:"total_farmer"`  // role = 'farmer'
	TotalWorker  int64 `json:"total_worker"`  // role = 'worker'
}