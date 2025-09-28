package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

type ApplicationService interface {
	ApplyToProject(projectID string, workerID string, input dto.ApplyProjectInput) (*models.ProjectApplication, error)
	// [PERUBAHAN] Mengembalikan *models.Contract, bukan DTO
	GetMyApplications(workerID uuid.UUID) ([]dto.MyApplicationResponse, error)

	RejectApplication(applicationID string, farmerID uuid.UUID) error
	AcceptApplication(applicationID string, farmerID string) (*dto.AcceptApplicationResponse, error)
	FindApplicationsByProjectID(projectID string, farmerID string) ([]models.ProjectApplication, error)
}

type applicationService struct {
	appRepo              repositories.ApplicationRepository
	projectRepo          repositories.ProjectRepository
	contractRepo         repositories.ContractRepository
	assignRepo           repositories.AssignmentRepository
	// notificationService NotificationService
	db *gorm.DB
}

// [PERUBAHAN] Dependensi transactionRepo dihapus
func NewApplicationService(appRepo repositories.ApplicationRepository, projectRepo repositories.ProjectRepository, contractRepo repositories.ContractRepository, assignRepo repositories.AssignmentRepository,  notificationService NotificationService, db *gorm.DB) ApplicationService {
	return &applicationService{
		appRepo:              appRepo,
		projectRepo:          projectRepo,
		contractRepo:         contractRepo,
		assignRepo:           assignRepo,
		db:                   db,
	}
}

func (s *applicationService) ApplyToProject(projectID string, workerID string, input dto.ApplyProjectInput) (*models.ProjectApplication, error) {
	// ... (kode ini tidak berubah)
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if project.Status != "open" {
		return nil, errors.New("this project is no longer open for applications")
	}
	alreadyApplied, err := s.projectRepo.HasWorkerApplied(projectID, workerID)
	if err != nil {
		return nil, fmt.Errorf("error checking application status: %w", err)
	}
	if alreadyApplied {
		return nil, errors.New("you have already applied to this project")
	}
	projectUUID, _ := uuid.Parse(projectID)
	workerUUID, _ := uuid.Parse(workerID)

	application := &models.ProjectApplication{
		ProjectID: projectUUID,
		WorkerID:  workerUUID,
		Message:   &input.Message,
		Status:    "pending",
	}

	if err := s.appRepo.Create(application); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}
	return application, nil
}

func (s *applicationService) FindApplicationsByProjectID(projectID string, farmerID string) ([]models.ProjectApplication, error) {
	project, err := s.projectRepo.FindByID(projectID)
	if err != nil {
		return nil, errors.New("project not found")
	}
	if project.FarmerID.String() != farmerID {
		return nil, errors.New("forbidden: you do not own this project")
	}
	applications, err := s.appRepo.FindAllByProjectID(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch applications: %w", err)
	}
	return applications, nil
}

func (s *applicationService) RejectApplication(applicationID string, farmerID uuid.UUID) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Ambil data lamaran
	application, err := s.appRepo.FindByID(applicationID)
	if err != nil {
		tx.Rollback()
		return errors.New("application not found")
	}

	// 2. Validasi
	if application.Project.FarmerID != farmerID {
		tx.Rollback()
		return errors.New("forbidden: you are not the owner of this project")
	}
	if application.Status != "pending" {
		tx.Rollback()
		return errors.New("this application has already been processed")
	}

	// 3. Ubah status menjadi 'rejected'
	appUUID, err := uuid.Parse(applicationID)
	if err != nil {
		tx.Rollback()
		return errors.New("invalid application ID format")
	}
	if err := s.appRepo.UpdateStatus(tx, appUUID, "rejected"); err != nil {
		tx.Rollback()
		return err
	}
	
	// TODO: Kirim notifikasi ke pekerja bahwa lamarannya ditolak

	return tx.Commit().Error
}



// ... (Pastikan semua import ini ada di bagian atas file Anda)

func (s *applicationService) AcceptApplication(applicationID string, farmerID string) (*dto.AcceptApplicationResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil { return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error) }
	defer func() { if r := recover(); r != nil { tx.Rollback() } }()

	app, err := s.appRepo.FindByID(applicationID)
	if err != nil { tx.Rollback(); return nil, errors.New("application not found") }

	// Validasi
	if app.Project.FarmerID.String() != farmerID { tx.Rollback(); return nil, errors.New("forbidden: you are not the owner of this project") }
	if app.Status != "pending" { tx.Rollback(); return nil, errors.New("this application is not in pending state") }
	isWorkerBusy, err := s.projectRepo.IsWorkerOnActiveProject(app.WorkerID.String())
	if err != nil { tx.Rollback(); return nil, err }
	if isWorkerBusy { tx.Rollback(); return nil, errors.New("worker is currently busy on another project") }

	// Buat Kontrak (tanpa konten)
	newContract := &models.Contract{
		ContractType:   "work", // <-- Eksplisit tentukan tipe kontrak
		ProjectID:      &app.ProjectID, // <-- Gunakan pointer (&)
		FarmerID:       app.Project.FarmerID,
		WorkerID:       &app.WorkerID, // <-- Gunakan pointer (&)
		Status:         "pending_signature",
		SignedByFarmer: true,
	}
	if err := s.contractRepo.Create(tx, newContract); err != nil { tx.Rollback(); return nil, err }

	// Buat Penugasan
	var agreedRate float64
	if app.Project.PaymentRate != nil { agreedRate = *app.Project.PaymentRate }
	newAssignment := &models.ProjectAssignment{
		ProjectID:  app.ProjectID,
		WorkerID:   app.WorkerID,
		ContractID: newContract.ID,
		AgreedRate: agreedRate,
		Status:     "assigned",
	}
	if err := s.assignRepo.Create(tx, newAssignment); err != nil { tx.Rollback(); return nil, err }

	// Update Status Lamaran
	if err := s.appRepo.UpdateStatus(tx, app.ID, "accepted"); err != nil { tx.Rollback(); return nil, err }

	if err := tx.Commit().Error; err != nil { return nil, fmt.Errorf("transaction commit failed: %w", err) }

	// Kirim notifikasi
	// title := fmt.Sprintf("Selamat! Tawaran untuk Proyek '%s'", app.Project.Title)
	// message := fmt.Sprintf("Petani %s telah menerima lamaran Anda.", app.Project.Farmer.User.Name)
	// link := fmt.Sprintf("/contracts/%s", newContract.ID)
	// s.notificationService.CreateNotification(app.Worker.UserID, title, message, link, "job_offer")
		// [PERBAIKAN] Buat DTO di akhir setelah semua proses berhasil
	response := &dto.AcceptApplicationResponse{
		ContractID:   newContract.ID,
		ProjectTitle: app.Project.Title,
		WorkerName:   app.Worker.User.Name,
		Message:      "Application accepted. A contract has been created.",
	}
	return response, nil
}

func (s *applicationService) GetMyApplications(workerID uuid.UUID) ([]dto.MyApplicationResponse, error) {
	applications, err := s.appRepo.FindAllByWorkerID(workerID)
	if err != nil {
		return nil, err
	}

	var responseDTOs []dto.MyApplicationResponse
	for _, app := range applications {
		dto := dto.MyApplicationResponse{
			ApplicationID: app.ID,
			Status:        app.Status,
			AppliedAt:     app.CreatedAt,
		}
		// Pastikan data hasil preload tidak nil sebelum diakses
		if app.Project.ID != uuid.Nil {
			dto.ProjectTitle = app.Project.Title
			if app.Project.Farmer.UserID != uuid.Nil && app.Project.Farmer.User.Name != "" {
				dto.FarmerName = app.Project.Farmer.User.Name
			}
		}
		responseDTOs = append(responseDTOs, dto)
	}

	return responseDTOs, nil
}
