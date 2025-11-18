package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type UserVerificationRepository interface {
	Create(verification *models.UserVerification) error
	FindPending() ([]models.UserVerification, error)
	FindByID(id uuid.UUID) (*models.UserVerification, error)
	UpdateStatus(tx *gorm.DB, verification *models.UserVerification) error
	GetApprovedDocumentsForUser(userID uuid.UUID) ([]string, error)
}

type userVerificationRepository struct{ db *gorm.DB }

func NewUserVerificationRepository(db *gorm.DB) UserVerificationRepository {
	return &userVerificationRepository{db: db}
}

func (r *userVerificationRepository) Create(verification *models.UserVerification) error {
	return r.db.Create(verification).Error
}

func (r *userVerificationRepository) FindPending() ([]models.UserVerification, error) {
	var verifications []models.UserVerification
	err := r.db.Preload("User").Where("status = ?", "pending").Find(&verifications).Error
	return verifications, err
}

func (r *userVerificationRepository) FindByID(id uuid.UUID) (*models.UserVerification, error) {
	var verification models.UserVerification
	err := r.db.Preload("User").Where("id = ?", id).First(&verification).Error
	return &verification, err
}

func (r *userVerificationRepository) UpdateStatus(tx *gorm.DB, verification *models.UserVerification) error {
	return tx.Save(verification).Error
}

// Mengambil daftar TIPE dokumen yang sudah disetujui
func (r *userVerificationRepository) GetApprovedDocumentsForUser(userID uuid.UUID) ([]string, error) {
	var documentTypes []string
	err := r.db.Model(&models.UserVerification{}).
		Where("user_id = ? AND status = ?", userID, "approved").
		Pluck("DISTINCT(document_type)", &documentTypes).Error
	return documentTypes, err
}