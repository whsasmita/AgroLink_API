package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	Create(tx *gorm.DB, review *models.Review) error
	GetReviewsByWorkerID(workerID uuid.UUID) ([]models.Review, error)
	// [BARU] Fungsi untuk membaca review di dalam transaksi
	GetReviewsByWorkerIDWithTx(tx *gorm.DB, workerID uuid.UUID) ([]models.Review, error)
	CheckExistingReview(reviewerID, reviewedWorkerID, projectID uuid.UUID) (int64, error)
	GetReviewsByDriverID(driverID uuid.UUID) ([]models.Review, error)
	GetReviewsByDriverIDWithTx(tx *gorm.DB, driverID uuid.UUID) ([]models.Review, error)

}

type reviewRepository struct{ db *gorm.DB }

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(tx *gorm.DB, review *models.Review) error {
	return tx.Create(review).Error
}

// Fungsi ini tetap ada jika dibutuhkan di tempat lain
func (r *reviewRepository) GetReviewsByWorkerID(workerID uuid.UUID) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.Where("reviewed_worker_id = ?", workerID).Find(&reviews).Error
	return reviews, err
}

// [FUNGSI BARU]
// Versi ini menggunakan objek transaksi (tx) untuk memastikan data yang dibaca konsisten.
func (r *reviewRepository) GetReviewsByWorkerIDWithTx(tx *gorm.DB, workerID uuid.UUID) ([]models.Review, error) {
	var reviews []models.Review
	err := tx.Where("reviewed_worker_id = ?", workerID).Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) CheckExistingReview(reviewerID, reviewedWorkerID, projectID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Review{}).
		Where("reviewer_id = ? AND reviewed_worker_id = ? AND project_id = ?", reviewerID, reviewedWorkerID, projectID).
		Count(&count).Error
	return count, err
}

func (r *reviewRepository) GetReviewsByDriverID(driverID uuid.UUID) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.Where("reviewed_driver_id = ?", driverID).Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) GetReviewsByDriverIDWithTx(tx *gorm.DB, driverID uuid.UUID) ([]models.Review, error) {
	var reviews []models.Review
	err := tx.Where("reviewed_driver_id = ?", driverID).Find(&reviews).Error
	return reviews, err
}