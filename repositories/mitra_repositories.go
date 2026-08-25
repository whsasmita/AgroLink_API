package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type MitraProfileRepository interface {
	Create(tx *gorm.DB, profile *models.MitraProfile) error
	FindByUserID(userID uuid.UUID) (*models.MitraProfile, error)
	Update(tx *gorm.DB, profile *models.MitraProfile) error
	FindAllVerified(req dto.PaginationRequest) ([]models.MitraProfile, int64, error)
	FindPendingVerifications(req dto.PaginationRequest) ([]models.MitraProfile, int64, error)
	UpdateStatus(tx *gorm.DB, userID uuid.UUID, status string) error
}

type mitraProfileRepository struct {
	db *gorm.DB
}

func NewMitraProfileRepository(db *gorm.DB) MitraProfileRepository {
	return &mitraProfileRepository{db: db}
}

func (r *mitraProfileRepository) Create(tx *gorm.DB, profile *models.MitraProfile) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(profile).Error
}

func (r *mitraProfileRepository) FindByUserID(userID uuid.UUID) (*models.MitraProfile, error) {
	var profile models.MitraProfile
	err := r.db.Preload("User").Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *mitraProfileRepository) Update(tx *gorm.DB, profile *models.MitraProfile) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Save(profile).Error
}

func (r *mitraProfileRepository) FindAllVerified(req dto.PaginationRequest) ([]models.MitraProfile, int64, error) {
	var profiles []models.MitraProfile
	var total int64

	query := r.db.Model(&models.MitraProfile{}).Preload("User").Where("status_verifikasi = ?", "verified")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	sort := req.Sort
	if sort == "" {
		sort = "created_at desc"
	}

	err := query.Order(sort).Offset(offset).Limit(req.Limit).Find(&profiles).Error
	return profiles, total, err
}

func (r *mitraProfileRepository) FindPendingVerifications(req dto.PaginationRequest) ([]models.MitraProfile, int64, error) {
	var profiles []models.MitraProfile
	var total int64

	query := r.db.Model(&models.MitraProfile{}).Preload("User").Where("status_verifikasi = ?", "pending")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	sort := req.Sort
	if sort == "" {
		sort = "created_at asc"
	}

	err := query.Order(sort).Offset(offset).Limit(req.Limit).Find(&profiles).Error
	return profiles, total, err
}

func (r *mitraProfileRepository) UpdateStatus(tx *gorm.DB, userID uuid.UUID, status string) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&models.MitraProfile{}).Where("user_id = ?", userID).Update("status_verifikasi", status).Error
}

// ---------------------------------------------------------------------
// MitraCooperationRepository
// ---------------------------------------------------------------------

type MitraCooperationRepository interface {
	Create(tx *gorm.DB, coop *models.MitraCooperation) error
	FindByID(id string) (*models.MitraCooperation, error)
	FindByIDWithDetails(id string) (*models.MitraCooperation, error)
	Update(tx *gorm.DB, coop *models.MitraCooperation) error
	FindAllByFarmerID(farmerID uuid.UUID, req dto.PaginationRequest) ([]models.MitraCooperation, int64, error)
	FindAllByMitraID(mitraID uuid.UUID, req dto.PaginationRequest) ([]models.MitraCooperation, int64, error)
	UpdateStatus(tx *gorm.DB, id string, status string) error
}

type mitraCooperationRepository struct {
	db *gorm.DB
}

func NewMitraCooperationRepository(db *gorm.DB) MitraCooperationRepository {
	return &mitraCooperationRepository{db: db}
}

func (r *mitraCooperationRepository) Create(tx *gorm.DB, coop *models.MitraCooperation) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(coop).Error
}

func (r *mitraCooperationRepository) FindByID(id string) (*models.MitraCooperation, error) {
	var coop models.MitraCooperation
	err := r.db.Where("id = ?", id).First(&coop).Error
	if err != nil {
		return nil, err
	}
	return &coop, nil
}

func (r *mitraCooperationRepository) FindByIDWithDetails(id string) (*models.MitraCooperation, error) {
	var coop models.MitraCooperation
	err := r.db.
		Preload("Mitra.User").
		Preload("Farmer.User").
		Preload("Contract").
		Where("id = ?", id).
		First(&coop).Error
	if err != nil {
		return nil, err
	}
	return &coop, nil
}

func (r *mitraCooperationRepository) Update(tx *gorm.DB, coop *models.MitraCooperation) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Save(coop).Error
}

func (r *mitraCooperationRepository) FindAllByFarmerID(farmerID uuid.UUID, req dto.PaginationRequest) ([]models.MitraCooperation, int64, error) {
	var coops []models.MitraCooperation
	var total int64

	query := r.db.Model(&models.MitraCooperation{}).
		Preload("Mitra.User").
		Preload("Farmer.User").
		Where("farmer_id = ?", farmerID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	sort := req.Sort
	if sort == "" {
		sort = "created_at desc"
	}

	err := query.Order(sort).Offset(offset).Limit(req.Limit).Find(&coops).Error
	return coops, total, err
}

func (r *mitraCooperationRepository) FindAllByMitraID(mitraID uuid.UUID, req dto.PaginationRequest) ([]models.MitraCooperation, int64, error) {
	var coops []models.MitraCooperation
	var total int64

	query := r.db.Model(&models.MitraCooperation{}).
		Preload("Mitra.User").
		Preload("Farmer.User").
		Where("mitra_id = ?", mitraID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	sort := req.Sort
	if sort == "" {
		sort = "created_at desc"
	}

	err := query.Order(sort).Offset(offset).Limit(req.Limit).Find(&coops).Error
	return coops, total, err
}

func (r *mitraCooperationRepository) UpdateStatus(tx *gorm.DB, id string, status string) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Model(&models.MitraCooperation{}).Where("id = ?", id).Update("status", status).Error
}

// ---------------------------------------------------------------------
// MitraReviewRepository
// ---------------------------------------------------------------------

type MitraReviewRepository interface {
	Create(tx *gorm.DB, review *models.MitraReview) error
	FindByCooperationAndReviewer(cooperationID string, reviewerID uuid.UUID) (*models.MitraReview, error)
	FindAllByCooperationID(cooperationID string) ([]models.MitraReview, error)
}

type mitraReviewRepository struct {
	db *gorm.DB
}

func NewMitraReviewRepository(db *gorm.DB) MitraReviewRepository {
	return &mitraReviewRepository{db: db}
}

func (r *mitraReviewRepository) Create(tx *gorm.DB, review *models.MitraReview) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(review).Error
}

func (r *mitraReviewRepository) FindByCooperationAndReviewer(cooperationID string, reviewerID uuid.UUID) (*models.MitraReview, error) {
	var review models.MitraReview
	err := r.db.Where("cooperation_id = ? AND reviewer_id = ?", cooperationID, reviewerID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *mitraReviewRepository) FindAllByCooperationID(cooperationID string) ([]models.MitraReview, error) {
	var reviews []models.MitraReview
	err := r.db.Preload("Reviewer").Where("cooperation_id = ?", cooperationID).Find(&reviews).Error
	return reviews, err
}
