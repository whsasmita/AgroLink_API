package repositories

import (
	"errors"
	"fmt"

	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type WorkerRepository interface {
	GetWorkers(search, sortBy, order string, limit, offset int, minDailyRate, maxDailyRate, minHourlyRate, maxHourlyRate float64) ([]models.Worker, int64, error)
	GetWorkerByID(id string) (models.Worker, error)
}

type workerRepository struct {
	db *gorm.DB
}

func NewWorkerRepository(db *gorm.DB) WorkerRepository {
	return &workerRepository{db}
}

func (r *workerRepository) GetWorkers(search, sortBy, order string, limit, offset int, minDailyRate, maxDailyRate, minHourlyRate, maxHourlyRate float64) ([]models.Worker, int64, error) {
	var workers []models.Worker
	var total int64

	query := r.db.Model(&models.Worker{}).Preload("User")

	if search != "" {
		like := fmt.Sprintf("%%%s%%", search)
		query = query.
			Joins("JOIN users ON users.id = workers.user_id").
			Where("users.name LIKE ? OR workers.skills LIKE ? OR workers.address LIKE ?", like, like, like)
	}

	// ✅ Tambahkan filter tarif
	if minDailyRate > 0 {
		query = query.Where("daily_rate >= ?", minDailyRate)
	}
	if maxDailyRate > 0 {
		query = query.Where("daily_rate <= ?", maxDailyRate)
	}
	if minHourlyRate > 0 {
		query = query.Where("hourly_rate >= ?", minHourlyRate)
	}
	if maxHourlyRate > 0 {
		query = query.Where("hourly_rate <= ?", maxHourlyRate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count workers: %w", err)
	}

	// ... (kode sorting dan find tetap sama)
	allowedSortBy := map[string]bool{
		"created_at": true,
		"rating":     true,
		"hourly_rate": true,
		"daily_rate": true,
	}
	if !allowedSortBy[sortBy] {
		sortBy = "created_at"
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}

	if err := query.
		Order(fmt.Sprintf("%s %s", sortBy, order)).
		Limit(limit).
		Offset(offset).
		Find(&workers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get workers: %w", err)
	}

	return workers, total, nil
}


func (r *workerRepository) GetWorkerByID(id string) (models.Worker, error) {
    var worker models.Worker
    
    // Cari worker berdasarkan user ID dan preload data user
    if err := r.db.Preload("User").First(&worker, "user_id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return worker, fmt.Errorf("worker with ID %s not found", id)
        }
        return worker, fmt.Errorf("failed to get worker: %w", err)
    }

    return worker, nil
}
