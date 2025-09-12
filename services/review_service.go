package services

import (
	"fmt"

	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

type ReviewService interface {
	CreateReview(input dto.CreateReviewInput) (*models.Review, error)
}

type reviewService struct {
	reviewRepo  repositories.ReviewRepository
	workerRepo  repositories.WorkerRepository
	projectRepo repositories.ProjectRepository
	db          *gorm.DB // Untuk transaksi
}

func NewReviewService(reviewRepo repositories.ReviewRepository, workerRepo repositories.WorkerRepository, projectRepo repositories.ProjectRepository, db *gorm.DB) ReviewService {
	return &reviewService{
		reviewRepo:  reviewRepo,
		workerRepo:  workerRepo,
		projectRepo: projectRepo,
		db:          db,
	}
}

func (s *reviewService) CreateReview(input dto.CreateReviewInput) (*models.Review, error) {
	// 1. Validasi: Pastikan proyek sudah selesai
	project, err := s.projectRepo.FindByID(input.ProjectID.String())
	if err != nil {
		return nil, fmt.Errorf("project not found")
	}
	if project.Status != "completed" {
		return nil, fmt.Errorf("reviews can only be submitted for completed projects")
	}
    // Pastikan petani yang memberi review adalah pemilik proyek
    if project.FarmerID != input.ReviewerID {
        return nil, fmt.Errorf("forbidden: you are not the owner of this project")
    }

    // 2. Validasi: Pastikan petani belum pernah mereview pekerja ini untuk proyek yang sama
    count, err := s.reviewRepo.CheckExistingReview(input.ReviewerID, input.ReviewedWorkerID, input.ProjectID)
    if err != nil || count > 0 {
        return nil, fmt.Errorf("you have already reviewed this worker for this project")
    }

	newReview := &models.Review{
		ProjectID:      input.ProjectID,
		ReviewerID:     input.ReviewerID,
		ReviewedWorkerID: input.ReviewedWorkerID,
		Rating:         input.Rating,
		Comment:        &input.Comment,
	}

	// Gunakan transaksi untuk memastikan review dan update rating terjadi bersamaan
	tx := s.db.Begin()
	if err := s.reviewRepo.Create(tx, newReview); err != nil { // Note: Perlu penyesuaian di repo untuk menerima 'tx'
		tx.Rollback()
		return nil, err
	}

	// 3. Kalkulasi ulang dan update rata-rata rating
	allReviews, err := s.reviewRepo.GetReviewsByWorkerID(input.ReviewedWorkerID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	var totalRating int
	for _, r := range allReviews {
		totalRating += r.Rating
	}
	newReviewCount := len(allReviews)
	newAverageRating := float64(totalRating) / float64(newReviewCount)

	if err := s.workerRepo.UpdateRating(input.ReviewedWorkerID, newAverageRating, newReviewCount); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return newReview, nil
}