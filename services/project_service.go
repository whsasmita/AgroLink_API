package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
)

type ProjectService interface {
	// Perbaikan: Menyesuaikan tipe data farmerID agar sesuai dengan handler
	CreateProject(request dto.CreateProjectRequest, farmerID uuid.UUID) (*models.Project, error)

	// Perbaikan: Mengembalikan model mentah dan total, sesuai kesepakatan
	FindAll(pagination dto.PaginationRequest) (*[]models.Project, int64, error)
	FindByID(id string) (*models.Project, error)
	FindMyProjects(farmerID uuid.UUID) ([]dto.MyProjectResponse, error)
	CheckAndFinalizeProject(projectID uuid.UUID) error  // <-- Fungsi baru
	UpdateStatus(projectID string, status string) error // <-- Fungsi baru
}

type projectService struct {
	projectRepo repositories.ProjectRepository
	assignRepo  repositories.AssignmentRepository
	invoiceRepo repositories.InvoiceRepository
}

// [PERUBAHAN] Perbarui konstruktor untuk menerima dependensi baru
func NewProjectService(projectRepo repositories.ProjectRepository, assignRepo repositories.AssignmentRepository, invoiceRepo repositories.InvoiceRepository) ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		assignRepo:  assignRepo,
		invoiceRepo: invoiceRepo,
	}
}

func (s *projectService) CreateProject(request dto.CreateProjectRequest, farmerID uuid.UUID) (*models.Project, error) {

	startDate, err := time.Parse("2006-01-02", request.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", request.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: %w", err)
	}

	project := &models.Project{
		FarmerID:      farmerID,
		Title:         request.Title,
		Location:      request.Location,
		Description:   request.Description,
		WorkersNeeded: request.WorkersNeeded,
		StartDate:     startDate,
		EndDate:       endDate,
		PaymentType:   request.PaymentType,
		PaymentRate:   &request.PaymentRate,
		Status:        "open",
	}

	if err := s.projectRepo.CreateProject(nil,project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}
	createdProject, err := s.projectRepo.FindByID(project.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to fetch created project: %w", err)
	}
	return createdProject, nil

}

func (s *projectService) FindAll(pagination dto.PaginationRequest) (*[]models.Project, int64, error) {
	// Service hanya meneruskan panggilan ke repository.
	// Logika transformasi DTO sepenuhnya berada di handler.
	projects, total, err := s.projectRepo.FindAll(pagination)
	if err != nil {
		return nil, 0, err
	}
	return projects, total, nil
}

func (s *projectService) FindByID(id string) (*models.Project, error) {
	project, err := s.projectRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("project with id %s not found: %w", id, err)
	}
	return project, nil
}

func (s *projectService) CheckAndFinalizeProject(projectID uuid.UUID) error {
	project, err := s.projectRepo.FindByID(projectID.String())
	if err != nil {
		return err
	}

	if project.Status != "open" {
		return nil // Proyek sudah diproses atau tidak lagi terbuka
	}

	assignments, err := s.assignRepo.FindAllByProjectID(projectID.String())
	if err != nil {
		return err
	}

	// TODO perbarui lagi konsep invoicenya
	if len(assignments) >= project.WorkersNeeded {
		var baseAmount float64
		durationDays := project.EndDate.Sub(project.StartDate).Hours()/24 + 1

		if project.PaymentType == "per_day" && project.PaymentRate != nil {
			baseAmount = *project.PaymentRate * durationDays * float64(project.WorkersNeeded)
		} else if project.PaymentType == "lump_sum" && project.PaymentRate != nil {
			baseAmount = *project.PaymentRate * float64(project.WorkersNeeded)
		}

		if baseAmount <= 0 {
			return fmt.Errorf("calculated base amount is zero or negative for project %s", projectID)
		}

		platformFee := baseAmount * 0.05
		totalAmount := baseAmount + platformFee

		invoice := &models.Invoice{
			ProjectID:   project.ID,
			FarmerID:    project.FarmerID,
			Amount:      baseAmount,
			PlatformFee: platformFee,
			TotalAmount: totalAmount,
			Status:      "pending",
			DueDate:     time.Now().Add(48 * time.Hour),
		}

		if err := s.invoiceRepo.Create(invoice); err != nil {
			return fmt.Errorf("failed to create invoice: %w", err)
		}

		return s.UpdateStatus(project.ID.String(), "waiting_payment")
	}

	return nil
}

func (s *projectService) UpdateStatus(projectID string, status string) error {
	return s.projectRepo.UpdateStatus(projectID, status)
}

func (s *projectService) FindMyProjects(farmerID uuid.UUID) ([]dto.MyProjectResponse, error) {
	projects, err := s.projectRepo.FindAllByFarmerID(farmerID)
	if err != nil {
		return nil, err
	}

	var responseDTOs []dto.MyProjectResponse
	for _, p := range projects {
		dto := dto.MyProjectResponse{
			ProjectID:     p.ID,
			ProjectTitle:  p.Title,
			ProjectStatus: p.Status,
		}
		// Jika proyek ini sudah punya invoice, tambahkan ID-nya ke DTO
		if p.Invoice.ID != uuid.Nil {
			dto.InvoiceID = &p.Invoice.ID
		}
		responseDTOs = append(responseDTOs, dto)
	}

	return responseDTOs, nil
}
