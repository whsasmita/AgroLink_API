// File: repositories/driver_repository.go
package repositories

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type DriverRepository interface {
	GetDrivers(sortBy, order string, limit, offset int) ([]models.Driver, int64, error)
	GetDriverByID(id string) (models.Driver, error)
	FindNearby(lat, lng float64, radius int) ([]models.Driver, error)
	UpdateRating(tx *gorm.DB, driverID uuid.UUID, newRating float64, reviewCount int) error
	// Anda bisa menambahkan method pencarian yang lebih spesifik nanti
}

type driverRepository struct {
	db *gorm.DB
}

func NewDriverRepository(db *gorm.DB) DriverRepository {
	return &driverRepository{db}
}

// GetDrivers mengambil daftar driver dengan sorting dan pagination
func (r *driverRepository) GetDrivers(sortBy, order string, limit, offset int) ([]models.Driver, int64, error) {
	var drivers []models.Driver
	var total int64

	// Preload data User terkait
	query := r.db.Model(&models.Driver{}).Preload("User")

	// Hitung total sebelum pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count drivers: %w", err)
	}

	// Validasi sorting
	allowedSortBy := map[string]bool{
		"created_at":      true,
		"rating":          true,
		"total_deliveries": true,
	}
	if !allowedSortBy[sortBy] {
		sortBy = "rating"
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}

	// Ambil data dengan pagination
	if err := query.
		Order(fmt.Sprintf("%s %s", sortBy, order)).
		Limit(limit).
		Offset(offset).
		Find(&drivers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get drivers: %w", err)
	}

	return drivers, total, nil
}

func (r *driverRepository) GetDriverByID(id string) (models.Driver, error) {
	var driver models.Driver
	
	// Cari driver berdasarkan user ID dan preload data user
	if err := r.db.Preload("User").First(&driver, "user_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return driver, fmt.Errorf("driver with ID %s not found", id)
		}
		return driver, fmt.Errorf("failed to get driver: %w", err)
	}

	return driver, nil
}

func (r *driverRepository) FindNearby(lat, lng float64, radius int) ([]models.Driver, error) {
	var drivers []models.Driver

	haversine := fmt.Sprintf(`(6371 * acos(cos(radians(%f)) * cos(radians(current_lat)) * cos(radians(current_lng) - radians(%f)) + sin(radians(%f)) * sin(radians(current_lat))))`, lat, lng, lat)

	err := r.db.
		Preload("User"). // <-- [TAMBAHAN] Muat data User untuk mendapatkan nama driver
		Select(fmt.Sprintf("*, %s AS distance", haversine)).
		Where(fmt.Sprintf("%s <= ?", haversine), radius).
		Order("distance ASC").
		Find(&drivers).Error

	return drivers, err
}

func (r *driverRepository) UpdateRating(tx *gorm.DB, driverID uuid.UUID, newRating float64, reviewCount int) error {
    return tx.Model(&models.Driver{}).Where("user_id = ?", driverID).Updates(map[string]interface{}{
        "rating":       newRating,
        "review_count": reviewCount,
    }).Error
}