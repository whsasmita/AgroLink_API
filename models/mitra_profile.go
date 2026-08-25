package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MitraProfile represents the profile of an investor, company, organization, or individual partner
type MitraProfile struct {
	UserID                 uuid.UUID      `gorm:"type:char(36);primary_key" json:"user_id"`
	JenisMitra             string         `gorm:"type:enum('perusahaan','organisasi','individu');not null" json:"jenis_mitra"`
	NamaMitra              string         `gorm:"type:varchar(150);not null" json:"nama_mitra"`
	DeskripsiSingkat       *string        `gorm:"type:text" json:"deskripsi_singkat"`
	NomorTeleponBisnis     string         `gorm:"type:varchar(20);not null" json:"nomor_telepon_bisnis"`
	EmailBisnis            string         `gorm:"type:varchar(100);not null" json:"email_bisnis"`
	Website                *string        `gorm:"type:varchar(255)" json:"website"`
	AlamatLengkap          string         `gorm:"type:text;not null" json:"alamat_lengkap"`
	Provinsi               string         `gorm:"type:varchar(100);not null" json:"provinsi"`
	KotaKabupaten          string         `gorm:"type:varchar(100);not null" json:"kota_kabupaten"`
	NPWP                   *string        `gorm:"type:varchar(30)" json:"npwp"`
	NIB                    *string        `gorm:"type:varchar(30)" json:"nib"`
	NIKKTP                 *string        `gorm:"type:varchar(20)" json:"nik_ktp"`
	DokumenLegalitas       *string        `gorm:"type:text" json:"dokumen_legalitas"`
	StatusVerifikasi       string         `gorm:"type:enum('pending','verified','rejected');default:'pending'" json:"status_verifikasi"`
	NamaBank               string         `gorm:"type:varchar(50);not null" json:"nama_bank"`
	NomorRekening          string         `gorm:"type:varchar(50);not null" json:"nomor_rekening"`
	AtasNamaRekening      string         `gorm:"type:varchar(100);not null" json:"atas_nama_rekening"`
	LogoMitra              *string        `gorm:"type:text" json:"logo_mitra"`
	RatingMitra            float64        `gorm:"type:decimal(3,2);default:0" json:"rating_mitra"`
	TotalTransaksiBerhasil int            `gorm:"default:0" json:"total_transaksi_berhasil"`
	CreatedAt              time.Time      `json:"created_at"`
	UpdatedAt              time.Time      `json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
}
