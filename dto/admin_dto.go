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
	TransactionID      string    `json:"transaction_id"`   // ID dari Transaction atau seed ID (misal: "INV/20250904/001")
	TransactionDate    time.Time `json:"transaction_date"`
	Layanan            string    `json:"layanan"`          // "Pekerja", "Peternak", "Tukang", "Ekspedisi", "E-Commerce", "Chatbot Premium", "Kemitraan"
	Sektor             string    `json:"sektor"`           // "Pertanian", "Peternakan", "Tukang Bangunan", "-"
	StatusTransaksi    string    `json:"status_transaksi"` // "Sukses", "Gagal"
	Status             string    `json:"status"`           // "paid", "failed"
	Keterangan         string    `json:"keterangan"`
	KomentarUser       string    `json:"komentar_user"`
	PaymentMethod      string    `json:"payment_method"`
	NominalTransaksi   float64   `json:"nominal_transaksi"`
	PersentaseKomisi   float64   `json:"persentase_komisi"`
	KeuntunganKotor    float64   `json:"keuntungan_kotor"`
	BiayaMidtrans      float64   `json:"biaya_midtrans"`
	KeuntunganBersih   float64   `json:"keuntungan_bersih"`
	TotalDiterimaMitra float64   `json:"total_diterima_mitra"`

	// Relasi
	FarmerEmail       *string `json:"farmer_email,omitempty"`
	FarmerName        *string `json:"farmer_name,omitempty"`
	WorkerEmail       *string `json:"worker_email,omitempty"`
	WorkerName        *string `json:"worker_name,omitempty"`
	DriverEmail       *string `json:"driver_email,omitempty"`
	DriverName        *string `json:"driver_name,omitempty"`
	PenjualEmail      *string `json:"penjual_email,omitempty"`
	PembeliEmail      *string `json:"pembeli_email,omitempty"`
	BuyerEmail        *string `json:"buyer_email,omitempty"`
	BuyerName         *string `json:"buyer_name,omitempty"`
	MitraEmail        *string `json:"mitra_email,omitempty"`
	MitraName         *string `json:"mitra_name,omitempty"`
	UserEmail         *string `json:"user_email,omitempty"`
	PemberiKerjaEmail *string `json:"pemberi_kerja_email,omitempty"`
	PekerjaEmail      *string `json:"pekerja_email,omitempty"`

	// Legacy & compatibility
	AmountPaid      float64 `json:"amount_paid"`
	TransactionType string  `json:"transaction_type"` // "Jasa" (Project/Delivery) atau "Produk" (E-commerce)
	ContextInfo     string  `json:"context_info"`     // Misal: "Proyek: Panen Jagung" atau "Pesanan E-commerce"
	PayerName       string  `json:"payer_name"`       // Nama Petani (Jasa) atau Pembeli (E-commerce)
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
	Type          *string   `json:"type,omitempty"`
	Skills        []string  `json:"skills,omitempty"`
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