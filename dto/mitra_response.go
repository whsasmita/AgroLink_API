package dto

import (
	"time"

	"github.com/google/uuid"
)

// MitraProfileResponse DTO for returning Mitra profile details
type MitraProfileResponse struct {
	UserID                 uuid.UUID `json:"user_id"`
	JenisMitra             string    `json:"jenis_mitra"`
	NamaMitra              string    `json:"nama_mitra"`
	DeskripsiSingkat       *string   `json:"deskripsi_singkat"`
	NomorTeleponBisnis     string    `json:"nomor_telepon_bisnis"`
	EmailBisnis            string    `json:"email_bisnis"`
	Website                *string   `json:"website"`
	AlamatLengkap          string    `json:"alamat_lengkap"`
	Provinsi               string    `json:"provinsi"`
	KotaKabupaten          string    `json:"kota_kabupaten"`
	NPWP                   *string   `json:"npwp,omitempty"`
	NIB                    *string   `json:"nib,omitempty"`
	NIKKTP                 *string   `json:"nik_ktp,omitempty"`
	DokumenLegalitas       *string   `json:"dokumen_legalitas,omitempty"`
	StatusVerifikasi       string    `json:"status_verifikasi"`
	NamaBank               string    `json:"nama_bank"`
	NomorRekening          string    `json:"nomor_rekening"`
	AtasNamaRekening      string    `json:"atas_nama_rekening"`
	LogoMitra              *string   `json:"logo_mitra"`
	RatingMitra            float64   `json:"rating_mitra"`
	TotalTransaksiBerhasil int       `json:"total_transaksi_berhasil"`
	CreatedAt              time.Time `json:"created_at"`
}

// CooperationPartyBrief Response sub-DTO for party details
type CooperationPartyBrief struct {
	UserID uuid.UUID `json:"user_id"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Phone  *string   `json:"phone,omitempty"`
}

// CooperationBriefResponse DTO for list view
type CooperationBriefResponse struct {
	ID            uuid.UUID              `json:"id"`
	Title         string                 `json:"title"`
	InitiatorType string                 `json:"initiator_type"`
	Status        string                 `json:"status"`
	ProposedAmount float64               `json:"proposed_amount"`
	AgreedAmount  *float64               `json:"agreed_amount"`
	Mitra         CooperationPartyBrief `json:"mitra"`
	Farmer        CooperationPartyBrief `json:"farmer"`
	CreatedAt     time.Time              `json:"created_at"`
}

// CooperationDetailResponse DTO for single detailed view
type CooperationDetailResponse struct {
	ID                     uuid.UUID              `json:"id"`
	MitraID                uuid.UUID              `json:"mitra_id"`
	FarmerID               uuid.UUID              `json:"farmer_id"`
	InitiatorType          string                 `json:"initiator_type"`
	Title                  string                 `json:"title"`
	Description            string                 `json:"description"`
	ProposedAmount         float64                `json:"proposed_amount"`
	AgreedAmount           *float64               `json:"agreed_amount"`
	PlatformFeePercentage float64                `json:"platform_fee_percentage"`
	StartDate              *time.Time             `json:"start_date"`
	EndDate                *time.Time             `json:"end_date"`
	Status                 string                 `json:"status"`
	Notes                  *string                `json:"notes"`
	ContractID             *uuid.UUID             `json:"contract_id"`
	Mitra                  CooperationPartyBrief `json:"mitra"`
	Farmer                 CooperationPartyBrief `json:"farmer"`
	CreatedAt              time.Time              `json:"created_at"`
	UpdatedAt              time.Time              `json:"updated_at"`
}

// MitraReviewResponse DTO for returning review item
type MitraReviewResponse struct {
	ID            uuid.UUID `json:"id"`
	CooperationID uuid.UUID `json:"cooperation_id"`
	ReviewerType  string    `json:"reviewer_type"`
	ReviewerID    uuid.UUID `json:"reviewer_id"`
	ReviewerName  string    `json:"reviewer_name"`
	Rating        int       `json:"rating"`
	Comment       *string   `json:"comment"`
	CreatedAt     time.Time `json:"created_at"`
}
