package repositories

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type OTPRepository interface {
	CreateOTP(otp *models.EmailOTP) error
	FindValidOTP(email, code string) (*models.EmailOTP, error)
	MarkAsUsed(id uuid.UUID) error
	InvalidateExistingOTPs(email string) error
}

type otpRepository struct {
	db *gorm.DB
}

func NewOTPRepository(db *gorm.DB) OTPRepository {
	return &otpRepository{db: db}
}

func (r *otpRepository) CreateOTP(otp *models.EmailOTP) error {
	return r.db.Create(otp).Error
}

func (r *otpRepository) FindValidOTP(email, code string) (*models.EmailOTP, error) {
	var otp models.EmailOTP
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	cleanCode := strings.TrimSpace(code)
	now := time.Now()

	err := r.db.Where("LOWER(email) = ? AND otp_code = ? AND is_used = ? AND expires_at > ?", cleanEmail, cleanCode, false, now).
		Order("created_at DESC").
		First(&otp).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("kode OTP tidak valid atau sudah kedaluwarsa")
		}
		return nil, err
	}

	return &otp, nil
}

func (r *otpRepository) MarkAsUsed(id uuid.UUID) error {
	return r.db.Model(&models.EmailOTP{}).Where("id = ?", id).Update("is_used", true).Error
}

func (r *otpRepository) InvalidateExistingOTPs(email string) error {
	cleanEmail := strings.ToLower(strings.TrimSpace(email))
	return r.db.Model(&models.EmailOTP{}).Where("LOWER(email) = ? AND is_used = ?", cleanEmail, false).Update("is_used", true).Error
}
