package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	CreateProject(project *models.Project) error
	FindAll(pagination dto.PaginationRequest) (*[]models.Project, int64, error)
	FindByID(id string) (*models.Project, error)
	HasWorkerApplied(projectID, workerID string) (bool, error)
	 FindAllByFarmerID(farmerID uuid.UUID) ([]models.Project, error)
	// Perbaikan: Menggunakan 'workerID' agar konsisten
	IsWorkerOnActiveProject(workerID string) (bool, error)
	UpdateStatus(projectID string, status string) error
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) CreateProject(project *models.Project) error {
	return r.db.Create(project).Error
}

func (r *projectRepository) FindAllByFarmerID(farmerID uuid.UUID) ([]models.Project, error) {
    var projects []models.Project
    // [PENTING] Tambahkan Preload("Invoice") untuk mengambil data invoice terkait
    err := r.db.Preload("Invoice").Where("farmer_id = ?", farmerID).Order("created_at DESC").Find(&projects).Error
    return projects, err
}

func (r *projectRepository) FindAll(pagination dto.PaginationRequest) (*[]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	query := r.db.Model(&models.Project{}).Where("status = ?", "open")

	query.Count(&total)

	offset := (pagination.Page - 1) * pagination.Limit

	err := query.Order(pagination.Sort).
		Limit(pagination.Limit).
		Offset(offset).
		Find(&projects).Error

	if err != nil {
		return nil, 0, err
	}

	return &projects, total, nil
}

func (r *projectRepository) FindByID(id string) (*models.Project, error) {
	var project models.Project
	err := r.db.
		Preload("Farmer.User").
		Where("id = ?", id).
		First(&project).Error
	return &project, err
}

func (r *projectRepository) HasWorkerApplied(projectID, workerID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.ProjectApplication{}).
		Where("project_id = ? AND worker_id = ?", projectID, workerID).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *projectRepository) IsWorkerOnActiveProject(workerID string) (bool, error) {
	var count int64
	activeStatuses := []string{"assigned", "started"}

	err := r.db.Model(&models.ProjectAssignment{}).
		Where("worker_id = ? AND status IN ?", workerID, activeStatuses).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *projectRepository) UpdateStatus(projectID string, status string) error {
	result := r.db.Model(&models.Project{}).Where("id = ?", projectID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}