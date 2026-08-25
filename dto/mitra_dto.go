package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateMitraProfileRequest DTO for creating/updating Mitra profile
type CreateMitraProfileRequest struct {
	JenisMitra         string  `json:"jenis_mitra" binding:"required,oneof=perusahaan organisasi individu"`
	NamaMitra          string  `json:"nama_mitra" binding:"required"`
	DeskripsiSingkat   *string `json:"deskripsi_singkat"`
	NomorTeleponBisnis string  `json:"nomor_telepon_bisnis" binding:"required"`
	EmailBisnis        string  `json:"email_bisnis" binding:"required,email"`
	Website            *string `json:"website"`
	AlamatLengkap      string  `json:"alamat_lengkap" binding:"required"`
	Provinsi           string  `json:"provinsi" binding:"required"`
	KotaKabupaten      string  `json:"kota_kabupaten" binding:"required"`
	NPWP               *string `json:"npwp"`
	NIB                *string `json:"nib"`
	NIKKTP             *string `json:"nik_ktp"`
	DokumenLegalitas   *string `json:"dokumen_legalitas"`
	NamaBank           string  `json:"nama_bank" binding:"required"`
	NomorRekening      string  `json:"nomor_rekening" binding:"required"`
	AtasNamaRekening  string  `json:"atas_nama_rekening" binding:"required"`
	LogoMitra          *string `json:"logo_mitra"`
}

// ReviewMitraVerificationRequest DTO for Admin to verify or reject Mitra profile
type ReviewMitraVerificationRequest struct {
	Status string  `json:"status" binding:"required,oneof=verified rejected"`
	Notes  *string `json:"notes"`
}

// CreateOfferRequest DTO for Mitra initiating an offer to a Farmer
type CreateOfferRequest struct {
	FarmerID       uuid.UUID  `json:"farmer_id" binding:"required"`
	Title          string     `json:"title" binding:"required"`
	Description    string     `json:"description" binding:"required"`
	ProposedAmount float64    `json:"proposed_amount" binding:"required,gt=0"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	Notes          *string    `json:"notes"`
}

// CreateApplicationRequest DTO for Farmer submitting an application to a Mitra
type CreateApplicationRequest struct {
	MitraID        uuid.UUID  `json:"mitra_id" binding:"required"`
	Title          string     `json:"title" binding:"required"`
	Description    string     `json:"description" binding:"required"`
	ProposedAmount float64    `json:"proposed_amount" binding:"required,gt=0"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	Notes          *string    `json:"notes"`
}

// ApproveCooperationRequest DTO for recipient party approving a cooperation proposal
type ApproveCooperationRequest struct {
	AgreedAmount *float64   `json:"agreed_amount"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Notes        *string    `json:"notes"`
}

// RejectCooperationRequest DTO for recipient party rejecting a proposal
type RejectCooperationRequest struct {
	Notes *string `json:"notes"`
}

// ReviewCooperationRequest DTO for updating proposal status to 'reviewed' with optional negotiations
type ReviewCooperationRequest struct {
	Notes *string `json:"notes"`
}

// CreateMitraReviewRequest DTO for rating & reviewing a completed cooperation
type CreateMitraReviewRequest struct {
	Rating  int     `json:"rating" binding:"required,min=1,max=5"`
	Comment *string `json:"comment"`
}
