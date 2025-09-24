package services

import (
	"fmt"

	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

type ReviewService interface {
	CreateReview(input dto.CreateReviewInput) (*dto.CreateReviewResponse, error)
	CreateDriverReview(input dto.CreateDriverReviewInput) (*models.Review, error)

}

type reviewService struct {
	reviewRepo  repositories.ReviewRepository
	driverRepo repositories.DriverRepository
	workerRepo  repositories.WorkerRepository
	projectRepo repositories.ProjectRepository
	deliveryRepo repositories.DeliveryRepository
	db          *gorm.DB
}

func NewReviewService(reviewRepo repositories.ReviewRepository, workerRepo repositories.WorkerRepository, projectRepo repositories.ProjectRepository,driverRepo repositories.DriverRepository, deliveryRepo repositories.DeliveryRepository, db *gorm.DB) ReviewService {
	return &reviewService{
		reviewRepo:  reviewRepo,
		workerRepo:  workerRepo,
		projectRepo: projectRepo,
		driverRepo: driverRepo,
		deliveryRepo: deliveryRepo,
		db:          db,
	}
}

func (s *reviewService) CreateReview(input dto.CreateReviewInput) (*dto.CreateReviewResponse,  error) {
	// 1. Validasi awal (di luar transaksi untuk efisiensi)
	project, err := s.projectRepo.FindByID(input.ProjectID.String())
	if err != nil {
		return nil, fmt.Errorf("project not found")
	}
	if project.Status != "completed" {
		return nil, fmt.Errorf("reviews can only be submitted for completed projects")
	}
	if project.FarmerID != input.ReviewerID {
		return nil, fmt.Errorf("forbidden: you are not the owner of this project")
	}

	count, err := s.reviewRepo.CheckExistingReview(input.ReviewerID, input.ReviewedWorkerID, input.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("error checking for existing review: %w", err)
	}
	if count > 0 {
		return nil, fmt.Errorf("you have already reviewed this worker for this project")
	}

	newReview := &models.Review{
		ProjectID:        &input.ProjectID,
		ReviewerID:       input.ReviewerID,
		ReviewedWorkerID: &input.ReviewedWorkerID,
		Rating:           input.Rating,
		Comment:          &input.Comment,
	}

	// 2. Mulai transaksi untuk semua operasi tulis
	tx := s.db.Begin()
    if tx.Error != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
    }
    defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := s.reviewRepo.Create(tx, newReview); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	// 3. Kalkulasi ulang rating di dalam transaksi yang sama
	allReviews, err := s.reviewRepo.GetReviewsByWorkerIDWithTx(tx, input.ReviewedWorkerID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get reviews for rating calculation: %w", err)
	}

	var totalRating int
	for _, r := range allReviews {
		totalRating += r.Rating
	}
	// [PERBAIKAN] Pastikan review yang baru dibuat juga dihitung
	newReviewCount := len(allReviews)

	var newAverageRating float64
	if newReviewCount > 0 {
		newAverageRating = float64(totalRating) / float64(newReviewCount)
	}

	// 4. Update rating pekerja di dalam transaksi yang sama
	if err := s.workerRepo.UpdateRating(tx, input.ReviewedWorkerID, newAverageRating, newReviewCount); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update worker rating: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	response := &dto.CreateReviewResponse{
        ID:               newReview.ID,
		ProjectID:        *newReview.ProjectID,
        ReviewerID:       newReview.ReviewerID,
        ReviewedWorkerID: *newReview.ReviewedWorkerID,
        Rating:           newReview.Rating,
        Comment:          newReview.Comment,
        CreatedAt:        newReview.CreatedAt,
        Message:          "Ulasan Anda telah berhasil dikirim.",
    }

	return response, nil

}

func (s *reviewService) CreateDriverReview(input dto.CreateDriverReviewInput) (*models.Review, error) {
    // 1. Validasi: Pastikan pengiriman sudah selesai
    delivery, err := s.deliveryRepo.FindByID(input.DeliveryID.String()) // Perlu deliveryRepo
    if err != nil {
        return nil, fmt.Errorf("delivery not found")
    }
    if delivery.Status != "delivered" {
        return nil, fmt.Errorf("reviews can only be submitted for delivered orders")
    }
    if delivery.FarmerID != input.ReviewerID {
        return nil, fmt.Errorf("forbidden: you are not the owner of this delivery")
    }

    // 2. Validasi: Mencegah ulasan ganda
    // Anda perlu membuat fungsi CheckExistingDriverReview di reviewRepo
    
    newReview := &models.Review{
        DeliveryID:       &input.DeliveryID,
        ReviewerID:       input.ReviewerID,
        ReviewedDriverID: &input.ReviewedDriverID,
        Rating:           input.Rating,
        Comment:          &input.Comment,
    }

    tx := s.db.Begin()
	if err := s.reviewRepo.Create(tx, newReview); err != nil {
		tx.Rollback(); return nil, err
	}

    // 3. Kalkulasi ulang rating driver
    // Anda perlu membuat GetReviewsByDriverID di reviewRepo
    allReviews, _ := s.reviewRepo.GetReviewsByDriverID(input.ReviewedDriverID) 
    
    var totalRating int
    for _, r := range allReviews { totalRating += r.Rating }
    newReviewCount := len(allReviews)
    newAverageRating := float64(totalRating) / float64(newReviewCount)

    if err := s.driverRepo.UpdateRating(tx, input.ReviewedDriverID, newAverageRating, newReviewCount); err != nil {
		tx.Rollback(); return nil, err
	}
    
    if err := tx.Commit().Error; err != nil {
		return nil, err
	}

    return newReview, nil
}
