package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type PayoutRepository interface {
	Create(payout *models.Payout) error
	FindPendingPayouts() ([]models.Payout, error)
}

type payoutRepository struct{ db *gorm.DB }

func NewPayoutRepository(db *gorm.DB) PayoutRepository {
	return &payoutRepository{db: db}
}

func (r *payoutRepository) Create(payout *models.Payout) error {
	return r.db.Create(payout).Error
}
func (r *payoutRepository) FindPendingPayouts() ([]models.Payout, error) {
	var payouts []models.Payout
	err := r.db.
		Preload("Worker.User").
		Preload("Transaction.Invoice.Project"). // Preload berantai untuk mendapatkan judul proyek
		Where("status = ?", "pending_disbursement").
		Order("released_at asc"). // Tampilkan yang paling lama menunggu di atas
		Find(&payouts).Error
	return payouts, err
}
