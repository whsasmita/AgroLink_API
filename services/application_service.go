package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

// ApplyProjectInput adalah DTO untuk request body saat melamar.
type ApplyProjectInput struct {
	Message string `json:"message"`
}

type ApplicationResponse struct {
	ID              uuid.UUID `json:"id"`
	WorkerID        uuid.UUID `json:"worker_id"`
	WorkerName      string    `json:"worker_name"`
	Message         *string   `json:"message"`
	ApplicationDate time.Time `json:"application_date"`
	Status          string    `json:"status"`
}

type ApplicationService interface {
	ApplyToProject(projectID string, workerID string, input ApplyProjectInput) (*models.ProjectApplication, error)
	AcceptApplication(applicationID string, farmerID string) (*models.Contract, error)
	FindApplicationsByProjectID(projectID string, farmerID string) ([]models.ProjectApplication, error)
	// SeeAllApplication()
}

type applicationService struct {
	appRepo      repositories.ApplicationRepository
	projectRepo  repositories.ProjectRepository
	contractRepo repositories.ContractRepository
	assignRepo   repositories.AssignmentRepository
	db           *gorm.DB // Diperlukan untuk transaksi
}

// NewApplicationService sekarang menerima semua dependensi yang dibutuhkan.
func NewApplicationService(appRepo repositories.ApplicationRepository, projectRepo repositories.ProjectRepository, contractRepo repositories.ContractRepository, assignRepo repositories.AssignmentRepository, db *gorm.DB) ApplicationService {
	return &applicationService{
		appRepo:      appRepo,
		projectRepo:  projectRepo,
		contractRepo: contractRepo,
		assignRepo:   assignRepo,
		db:           db,
	}
}

func (s *applicationService) ApplyToProject(projectID string, workerID string, input ApplyProjectInput) (*models.ProjectApplication, error) {
	// 1. Validasi apakah proyek ada dan statusnya 'open'
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if project.Status != "open" {
		return nil, errors.New("this project is no longer open for applications")
	}

	// 2. Validasi untuk mencegah lamaran ganda
	alreadyApplied, err := s.projectRepo.HasWorkerApplied(projectID, workerID)
	if err != nil {
		return nil, fmt.Errorf("error checking application status: %w", err)
	}
	if alreadyApplied {
		return nil, errors.New("you have already applied to this project")
	}

	// 3. Konversi ID
	projectUUID, _ := uuid.Parse(projectID)
	workerUUID, _ := uuid.Parse(workerID)

	// 4. Buat objek lamaran baru
	application := &models.ProjectApplication{
		ProjectID: projectUUID,
		WorkerID:  workerUUID,
		Message:   &input.Message,
		Status:    "pending",
	}

	// 5. Simpan ke database
	if err := s.appRepo.Create(application); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return application, nil
}

func (s *applicationService) AcceptApplication(applicationID string, farmerID string) (*models.Contract, error) {
	// Mulai Transaksi Database
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Ambil data aplikasi
	app, err := s.appRepo.FindByID(applicationID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("application not found")
	}

	// 2. Validasi Kepemilikan & Status
	if app.Project.FarmerID.String() != farmerID {
		tx.Rollback()
		return nil, errors.New("forbidden: you are not the owner of this project")
	}
	if app.Status != "pending" {
		tx.Rollback()
		return nil, errors.New("this application is not in pending state")
	}

	// 3. Validasi Ketersediaan Pekerja
	isWorkerBusy, err := s.projectRepo.IsWorkerOnActiveProject(app.WorkerID.String())
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if isWorkerBusy {
		tx.Rollback()
		return nil, errors.New("worker is currently busy on another project")
	}

	// 4. Buat Kontrak
	contractContent := fmt.Sprintf("Kontrak kerja untuk proyek '%s' antara Petani %s dan Pekerja %s.", app.Project.Title, farmerID, app.WorkerID)
	newContract := &models.Contract{
		ProjectID: app.ProjectID,
		FarmerID:  app.Project.FarmerID,
		WorkerID:  app.WorkerID,
		Content:   contractContent,
		Status:    "pending_signature",
		SignedByFarmer: true, 
	}
	if err := s.contractRepo.Create(tx, newContract); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}

	// 5. Buat Penugasan dengan pengecekan nil
	var agreedRate float64
	if app.Project.PaymentRate != nil {
		agreedRate = *app.Project.PaymentRate
	}

	newAssignment := &models.ProjectAssignment{
		ProjectID:  app.ProjectID,
		WorkerID:   app.WorkerID,
		ContractID: newContract.ID,
		AgreedRate: agreedRate,
		Status:     "assigned",
	}
	if err := s.assignRepo.Create(tx, newAssignment); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create assignment: %w", err)
	}

	// 6. Update Status Lamaran
	if err := s.appRepo.UpdateStatus(tx, app.ID, "accepted"); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update application status: %w", err)
	}

	// Jika semua sukses, commit transaksi
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return newContract, nil
}

func (s *applicationService) FindApplicationsByProjectID(projectID string, farmerID string) ([]models.ProjectApplication, error) {
    // 1. Validasi Kepemilikan Proyek
    project, err := s.projectRepo.FindByID(projectID)
    if err != nil {
        return nil, errors.New("project not found")
    }
    if project.FarmerID.String() != farmerID {
        return nil, errors.New("forbidden: you do not own this project")
    }

    // 2. Ambil data lamaran jika validasi berhasil
    applications, err := s.appRepo.FindAllByProjectID(projectID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch applications: %w", err)
    }

    return applications, nil
}
