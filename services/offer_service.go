package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

type OfferService interface {
	// [PERUBAHAN] Ubah tipe data yang dikembalikan
	CreateDirectOffer(input dto.DirectOfferRequest, farmerID, workerID uuid.UUID) (*dto.DirectOfferResponse, error)
}

type offerService struct {
	projectRepo  repositories.ProjectRepository
	contractRepo repositories.ContractRepository
	assignRepo   repositories.AssignmentRepository
	userRepo     repositories.UserRepository // <-- Tambahkan dependensi ini
	db           *gorm.DB
}

// [PERUBAHAN] Perbarui konstruktor
func NewOfferService(projRepo repositories.ProjectRepository, contRepo repositories.ContractRepository, assRepo repositories.AssignmentRepository, userRepo repositories.UserRepository, db *gorm.DB) OfferService {
	return &offerService{
		projectRepo:  projRepo,
		contractRepo: contRepo,
		assignRepo:   assRepo,
		userRepo:     userRepo, // <-- Tambahkan ini
		db:           db,
	}
}

func (s *offerService) CreateDirectOffer(input dto.DirectOfferRequest, farmerID, workerID uuid.UUID) (*dto.DirectOfferResponse, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil { tx.Rollback() }
	}()

	// 1. Validasi (tidak berubah)
	isWorkerBusy, err := s.projectRepo.IsWorkerOnActiveProject(workerID.String())
	if err != nil {
		tx.Rollback(); return nil, err
	}
	if isWorkerBusy {
		tx.Rollback(); return nil, fmt.Errorf("worker is currently busy on another project")
	}

	// 2. Buat Proyek baru (tidak berubah)
	startDate, _ := time.Parse("2006-01-02", input.StartDate)
	endDate, _ := time.Parse("2006-01-02", input.EndDate)
	newProject := &models.Project{
		FarmerID:      farmerID,
		Title:         input.Title,
		Description:   input.Description,
		Location:      input.Location,
		WorkersNeeded: 1,
		StartDate:     startDate,
		EndDate:       endDate,
		PaymentRate:   &input.PaymentRate,
		PaymentType:   "per_day",
		Status:        "direct_offer",
	}
	if err := s.projectRepo.CreateProject(tx, newProject); err != nil {
		tx.Rollback(); return nil, err
	}

	// 3. Buat Kontrak (tidak berubah)
	newContract := &models.Contract{
		ProjectID:      newProject.ID,
		FarmerID:       farmerID,
		WorkerID:       workerID,
		Status:         "pending_signature",
		SignedByFarmer: true,
	}
	if err := s.contractRepo.Create(tx, newContract); err != nil {
		tx.Rollback(); return nil, err
	}

	// 4. Buat Penugasan (tidak berubah)
	newAssignment := &models.ProjectAssignment{
		ProjectID:  newProject.ID,
		WorkerID:   workerID,
		ContractID: newContract.ID,
		AgreedRate: input.PaymentRate,
		Status:     "assigned",
	}
	if err := s.assignRepo.Create(tx, newAssignment); err != nil {
		tx.Rollback(); return nil, err
	}
	
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	// [PERBAIKAN] Ambil nama pekerja untuk DTO
	workerUser, err := s.userRepo.FindByID(workerID.String())
	if err != nil {
		// Jika user tidak ditemukan, setidaknya kembalikan data yang ada
		workerUser = &models.User{Name: "[Nama Pekerja Tidak Ditemukan]"}
	}

	// Buat dan kembalikan DTO
	response := &dto.DirectOfferResponse{
		ContractID:   newContract.ID,
		ProjectTitle: newProject.Title,
		WorkerName:   workerUser.Name,
		Status:       newContract.Status,
		Message:      "Direct offer sent successfully to worker.",
	}

	return response, nil
}