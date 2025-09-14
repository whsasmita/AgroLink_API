package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

type ApplicationService interface {
	ApplyToProject(projectID string, workerID string, input dto.ApplyProjectInput) (*models.ProjectApplication, error)
	// [PERUBAHAN] Mengembalikan *models.Contract, bukan DTO
	AcceptApplication(applicationID string, farmerID string) (*dto.AcceptApplicationResponse, error)
	FindApplicationsByProjectID(projectID string, farmerID string) ([]models.ProjectApplication, error)
}

type applicationService struct {
	appRepo      repositories.ApplicationRepository
	projectRepo  repositories.ProjectRepository
	contractRepo repositories.ContractRepository
	assignRepo   repositories.AssignmentRepository
	contractTemplateRepo repositories.ContractTemplateRepository
	// notificationService NotificationService
	db           *gorm.DB
}

// [PERUBAHAN] Dependensi transactionRepo dihapus
func NewApplicationService(appRepo repositories.ApplicationRepository, projectRepo repositories.ProjectRepository, contractRepo repositories.ContractRepository, assignRepo repositories.AssignmentRepository, contractTemplateRepo repositories.ContractTemplateRepository, notificationService NotificationService, db *gorm.DB) ApplicationService {
	return &applicationService{
		appRepo:      appRepo,
		projectRepo:  projectRepo,
		contractRepo: contractRepo,
		assignRepo:   assignRepo,
		contractTemplateRepo: contractTemplateRepo,
		db:           db,
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
	// ... (kode ini tidak berubah)
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

func (s *applicationService) AcceptApplication(applicationID string, farmerID string) (*dto.AcceptApplicationResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC recovered in AcceptApplication: %v", r)
			tx.Rollback()
		}
	}()

	app, err := s.appRepo.FindByID(applicationID)
	if err != nil {
		log.Printf("ERROR finding application by ID: %v", err)
		tx.Rollback(); return nil, errors.New("application not found")
	}

	// === Validasi (Tidak berubah) ===
	if app.Project.FarmerID.String() != farmerID {
		log.Printf("ERROR: farmer ID mismatch."); tx.Rollback(); return nil, errors.New("forbidden: you are not the owner of this project")
	}
	if app.Status != "pending" {
		log.Printf("ERROR: application status is not pending."); tx.Rollback(); return nil, errors.New("this application is not in pending state")
	}
	isWorkerBusy, err := s.projectRepo.IsWorkerOnActiveProject(app.WorkerID.String())
	if err != nil {
		log.Printf("ERROR checking worker status: %v", err); tx.Rollback(); return nil, err
	}
	if isWorkerBusy {
		log.Printf("ERROR: worker is busy."); tx.Rollback(); return nil, errors.New("worker is currently busy on another project")
	}

	// === Pembuatan Konten Kontrak (Tidak berubah) ===
	defaultTemplate, err := s.contractTemplateRepo.GetDefault()
	if err != nil {
		log.Printf("ERROR getting default template: %v", err); tx.Rollback(); return nil, fmt.Errorf("failed to get default contract template: %w", err)
	}

	safeString := func(s *string) string { if s != nil { return *s }; return "[DATA TIDAK TERSEDIA]" }
	safeFloat := func(f *float64) float64 { if f != nil { return *f }; return 0.0 }

	replacements := map[string]string{
		"{{nomor_kontrak}}": uuid.New().String(),
		"{{tanggal_pembuatan}}": time.Now().Format("2 January 2006"),
		"{{nama_petani}}": app.Project.Farmer.User.Name,
		"{{alamat_petani}}": safeString(app.Project.Farmer.Address),
		"{{telepon_petani}}": safeString(app.Project.Farmer.User.PhoneNumber),
		"{{email_petani}}": app.Project.Farmer.User.Email,
		"{{nama_pekerja}}": app.Worker.User.Name,
		"{{nik_pekerja}}": safeString(app.Worker.NationalID),
		"{{alamat_pekerja}}": safeString(app.Worker.Address),
		"{{telepon_pekerja}}": safeString(app.Worker.User.PhoneNumber),
		"{{email_pekerja}}": app.Worker.User.Email,
		"{{judul_proyek}}": app.Project.Title,
		"{{deskripsi_proyek}}": app.Project.Description,
		"{{lokasi_proyek}}": app.Project.FarmLocation.Name,
		"{{tanggal_mulai}}": app.Project.StartDate.Format("2 January 2006"),
		"{{tanggal_berakhir}}": app.Project.EndDate.Format("2 January 2006"),
		"{{jumlah_upah}}": fmt.Sprintf("%.0f", safeFloat(app.Project.PaymentRate)),
		"{{tipe_pembayaran}}": app.Project.PaymentType,
		"{{nama_bank_pekerja}}": safeString(app.Worker.BankName),
		"{{nomor_rekening_pekerja}}": safeString(app.Worker.BankAccountNumber),
		"{{nama_pemilik_rekening}}": safeString(app.Worker.BankAccountHolder),
	}
	newContent := defaultTemplate.Content
	for placeholder, value := range replacements {
		newContent = strings.ReplaceAll(newContent, placeholder, value)
	}

	// === Buat Entitas Baru (Tidak berubah) ===
	newContract := &models.Contract{
		ProjectID: app.ProjectID, FarmerID: app.Project.FarmerID, WorkerID: app.WorkerID,
		Content: newContent, Status: "pending_signature", SignedByFarmer: true,
	}
	if err := s.contractRepo.Create(tx, newContract); err != nil {
		log.Printf("ERROR creating contract: %v", err); tx.Rollback(); return nil, err
	}

	newAssignment := &models.ProjectAssignment{
		ProjectID: app.ProjectID, WorkerID: app.WorkerID, ContractID: newContract.ID,
		AgreedRate: safeFloat(app.Project.PaymentRate), Status: "assigned",
	}
	if err := s.assignRepo.Create(tx, newAssignment); err != nil {
		log.Printf("ERROR creating assignment: %v", err); tx.Rollback(); return nil, err
	}

	if err := s.appRepo.UpdateStatus(tx, app.ID, "accepted"); err != nil {
		log.Printf("ERROR updating application status: %v", err); tx.Rollback(); return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("ERROR committing transaction: %v", err); return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	// // [PERBAIKAN] Aktifkan kembali kode notifikasi
	// title := fmt.Sprintf("Selamat! Tawaran untuk Proyek '%s'", app.Project.Title)
	// message := fmt.Sprintf("Petani %s telah menerima lamaran Anda. Silakan tinjau dan tandatangani kontrak yang ditawarkan.", app.Project.Farmer.User.Name)
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









