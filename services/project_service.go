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
	CreateProject(request dto.CreateProjectRequest, farmerID uuid.UUID) (*models.Project, error)
	FindAll(pagination dto.PaginationRequest) (*[]models.Project, int64, error)
	FindByID(id string) (*models.Project, error)
	FindMyProjects(farmerID uuid.UUID) ([]dto.MyProjectResponse, error)
	CheckAndFinalizeProject(projectID uuid.UUID) error
	UpdateStatus(projectID string, status string) error
}

type projectService struct {
	projectRepo  repositories.ProjectRepository
	assignRepo   repositories.AssignmentRepository
	invoiceRepo  repositories.InvoiceRepository
}

// [PERBAIKAN] Konstruktor sekarang menerima semua dependensi yang dibutuhkan
func NewProjectService(
	projectRepo repositories.ProjectRepository,
	assignRepo repositories.AssignmentRepository,
	invoiceRepo repositories.InvoiceRepository,
) ProjectService {
	return &projectService{
		projectRepo:  projectRepo,
		assignRepo:   assignRepo,
		invoiceRepo:  invoiceRepo,
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
		PaymentType:   "per_day", // Menggunakan nilai tetap
		PaymentRate:   &request.PaymentRate,
		Status:        "open",
	}

	// [PERBAIKAN] Mengirim 'nil' karena ini bukan bagian dari transaksi yang lebih besar
	if err := s.projectRepo.CreateProject(nil, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

func (s *projectService) FindAll(pagination dto.PaginationRequest) (*[]models.Project, int64, error) {
	return s.projectRepo.FindAll(pagination)
}

func (s *projectService) FindByID(id string) (*models.Project, error) {
	return s.projectRepo.FindByID(id)
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
			WorkerNeeed:   p.WorkersNeeded,
		}
		if p.Invoice.ID != uuid.Nil {
			dto.InvoiceID = &p.Invoice.ID
		}
		responseDTOs = append(responseDTOs, dto)
	}
	return responseDTOs, nil
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
			ProjectID:   &project.ID,
			FarmerID:    project.FarmerID,
			Amount:      baseAmount,
			PlatformFee: platformFee,
			TotalAmount: totalAmount,
			Status:      "pending",
			DueDate:     time.Now().Add(48 * time.Hour),
		}

		if err := s.invoiceRepo.Create(nil, invoice); err != nil { // Menggunakan 'nil'
			return fmt.Errorf("failed to create invoice: %w", err)
		}

		return s.projectRepo.UpdateStatus( project.ID.String(), "waiting_payment") // Menggunakan 'nil'
	}
	return nil
}

func (s *projectService) UpdateStatus( projectID string, status string) error {
	return s.projectRepo.UpdateStatus( projectID, status)
}