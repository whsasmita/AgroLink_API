package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type PayoutRepository interface {
	Create(tx *gorm.DB ,payout *models.Payout) error
	FindPendingPayouts() ([]models.Payout, error)
	GetPendingPayoutStats() (count int64, totalAmount float64, err error)
	FindByID(id string) (*models.Payout, error)
	UpdateStatus(tx *gorm.DB, id string, status string) error
	Update(tx *gorm.DB, payout *models.Payout) error
}

type payoutRepository struct{ db *gorm.DB }

func NewPayoutRepository(db *gorm.DB) PayoutRepository {
	return &payoutRepository{db: db}
}

func (r *payoutRepository) Create(tx *gorm.DB, payout *models.Payout) error {
	return r.db.Create(payout).Error
}
func (r *payoutRepository) FindPendingPayouts() ([]models.Payout, error) {
	var payouts []models.Payout
	err := r.db.
		// Preload data Worker DAN Driver
		Preload("Transaction.Invoice.Project").
		Preload("Transaction.Invoice.Delivery").
		Where("status = ?", "pending_disbursement").
		Order("released_at asc").
		Find(&payouts).Error
	return payouts, err
}

func (r *payoutRepository) GetPendingPayoutStats() (int64, float64, error) {
	var stats struct {
		Count       int64
		TotalAmount float64
	}
	err := r.db.Model(&models.Payout{}).
		Select("COUNT(*) as count, SUM(amount) as total_amount").
		Where("status = ?", "pending_disbursement").
		Scan(&stats).Error
	return stats.Count, stats.TotalAmount, err
}

func (r *payoutRepository) FindByID(id string) (*models.Payout, error) {
	var payout models.Payout
	err := r.db.Where("id = ?", id).First(&payout).Error
	return &payout, err
}

// [FUNGSI BARU]
// UpdateStatus memperbarui kolom status dari sebuah payout di dalam transaksi.
func (r *payoutRepository) UpdateStatus(tx *gorm.DB, id string, status string) error {
	return tx.Model(&models.Payout{}).Where("id = ?", id).Update("status", status).Error
}

func (r *payoutRepository) Update(tx *gorm.DB, payout *models.Payout) error {
	return tx.Save(payout).Error
}