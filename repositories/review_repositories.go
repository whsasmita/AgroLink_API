package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type ReviewRepository interface {
	// [PERBAIKAN] Tambahkan *gorm.DB sebagai argumen pertama
	Create(tx *gorm.DB, review *models.Review) error
	GetReviewsByWorkerID(workerID uuid.UUID) ([]models.Review, error)
	CheckExistingReview(reviewerID, reviewedWorkerID, projectID uuid.UUID) (int64, error)
}

type reviewRepository struct{ db *gorm.DB }

func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

// [PERBAIKAN] Ubah fungsi Create untuk menerima dan menggunakan objek transaksi
func (r *reviewRepository) Create(tx *gorm.DB, review *models.Review) error {
	// Gunakan 'tx' yang dioper dari service, bukan 'r.db' global.
	// Jika 'tx' nil, GORM secara default akan menggunakan koneksi DB utama.
	return tx.Create(review).Error
}

func (r *reviewRepository) GetReviewsByWorkerID(workerID uuid.UUID) ([]models.Review, error) {
	var reviews []models.Review
	err := r.db.Where("reviewed_worker_id = ?", workerID).Find(&reviews).Error
	return reviews, err
}

func (r *reviewRepository) CheckExistingReview(reviewerID, reviewedWorkerID, projectID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.Review{}).
		Where("reviewer_id = ? AND reviewed_worker_id = ? AND project_id = ?", reviewerID, reviewedWorkerID, projectID).
		Count(&count).Error
	return count, err
}