package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type ApplicationRepository interface {
	Create(application *models.ProjectApplication) error
	FindByID(id string) (*models.ProjectApplication, error) // Perbaikan nama
	UpdateStatus(tx *gorm.DB, applicationID uuid.UUID, status string) error
	FindAllByProjectID(projectID string) ([]models.ProjectApplication, error)
	FindAllByWorkerID(workerID uuid.UUID) ([]models.ProjectApplication, error)
}

type applicationRepository struct {
	db *gorm.DB
}

func NewApplicationRepository(db *gorm.DB) ApplicationRepository {
	return &applicationRepository{db: db}
}

func (r *applicationRepository) Create(application *models.ProjectApplication) error {
	return r.db.Create(application).Error
}

// [PERBAIKAN] Mengubah nama menjadi FindByID dan menambahkan Preload("Project").
func (r *applicationRepository) FindByID(id string) (*models.ProjectApplication, error) {
	var application models.ProjectApplication
	err := r.db.
		Preload("Project.Farmer.User").
		Preload("Worker.User").
		Where("id = ?", id).
		First(&application).Error
	if err != nil {
		return nil, err
	}
	return &application, nil
}
func (r *applicationRepository) UpdateStatus(tx *gorm.DB, applicationID uuid.UUID, status string) error {
	return tx.Model(&models.ProjectApplication{}).Where("id = ?", applicationID).Update("status", status).Error
}

func (r *applicationRepository) FindAllByProjectID(projectID string) ([]models.ProjectApplication, error) {
	var applications []models.ProjectApplication
	err := r.db.Preload("Worker.User").Where("project_id = ?", projectID).Find(&applications).Error
	return applications, err
}

func (r *applicationRepository) FindAllByWorkerID(workerID uuid.UUID) ([]models.ProjectApplication, error) {
	var applications []models.ProjectApplication
	// Lakukan Preload untuk mendapatkan data Proyek dan Petani terkait
	err := r.db.
		Preload("Project.Farmer.User").
		Where("worker_id = ?", workerID).
		Order("created_at DESC").
		Find(&applications).Error
	return applications, err
}
