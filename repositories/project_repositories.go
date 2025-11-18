package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type ProjectRepository interface {
	CreateProject(tx *gorm.DB, project *models.Project) error
	FindAll(pagination dto.PaginationRequest) (*[]models.Project, int64, error)
	FindByID(id string) (*models.Project, error)
	FindAllByFarmerID(farmerID uuid.UUID) ([]models.Project, error)
	HasWorkerApplied(projectID, workerID string) (bool, error)
	IsWorkerOnActiveProject(workerID string) (bool, error)
	UpdateStatus( projectID string, status string) error
	CountActiveContracts(projectID string) (int64, error)
	CountActiveProjects() (int64, error)
}

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &projectRepository{db: db}
}

// CreateProject menyimpan record proyek baru, bisa berjalan di dalam transaksi.
func (r *projectRepository) CreateProject(tx *gorm.DB, project *models.Project) error {
	db := r.db
	if tx != nil {
		db = tx
	}
	return db.Create(project).Error
}

// FindAll mengambil daftar proyek yang 'open' dengan pagination.
func (r *projectRepository) FindAll(pagination dto.PaginationRequest) (*[]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	query := r.db.Model(&models.Project{}).Where("status = ?", "open")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

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

// FindByID mencari satu project berdasarkan ID-nya, memuat relasi penting.
func (r *projectRepository) FindByID(id string) (*models.Project, error) {
	var project models.Project
	err := r.db.
		Preload("Farmer.User").
		Preload("Invoice").
		Where("id = ?", id).
		First(&project).Error
	return &project, err
}

// FindAllByFarmerID mengambil semua proyek milik seorang petani.
func (r *projectRepository) FindAllByFarmerID(farmerID uuid.UUID) ([]models.Project, error) {
	var projects []models.Project
	err := r.db.Preload("Invoice").Where("farmer_id = ?", farmerID).Order("created_at DESC").Find(&projects).Error
	return projects, err
}

// HasWorkerApplied mengecek apakah seorang pekerja sudah melamar ke proyek tertentu.
func (r *projectRepository) HasWorkerApplied(projectID, workerID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.ProjectApplication{}).
		Where("project_id = ? AND worker_id = ?", projectID, workerID).
		Count(&count).Error
	return count > 0, err
}

// IsWorkerOnActiveProject mengecek apakah seorang pekerja sedang terikat pada proyek aktif.
func (r *projectRepository) IsWorkerOnActiveProject(workerID string) (bool, error) {
	var count int64
	activeStatuses := []string{"assigned", "started"}
	err := r.db.Model(&models.ProjectAssignment{}).
		Where("worker_id = ? AND status IN ?", workerID, activeStatuses).
		Count(&count).Error
	return count > 0, err
}

// UpdateStatus memperbarui kolom status dari sebuah proyek.
func (r *projectRepository) UpdateStatus( projectID string, status string) error {
	db := r.db
	result := db.Model(&models.Project{}).Where("id = ?", projectID).Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *projectRepository) CountActiveContracts(projectID string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Contract{}).
		Where("project_id = ? AND status = ?", projectID, "active").
		Count(&count).Error
	return count, err
}

func (r *projectRepository) CountActiveProjects() (int64, error) {
	var count int64
	err := r.db.Model(&models.Project{}).
		Where("status = ?", "in_progress").
		Count(&count).Error
	return count, err
}