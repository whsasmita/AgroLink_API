package services

import (
	"encoding/json"
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
}

type projectService struct {
	projectRepo repositories.ProjectRepository
	farmRepo    repositories.FarmRepository
}

func NewProjectService(repo repositories.ProjectRepository, farmRepo repositories.FarmRepository) ProjectService {
	return &projectService{projectRepo: repo, farmRepo:    farmRepo,}
}

func (s *projectService) CreateProject(request dto.CreateProjectRequest, farmerID uuid.UUID) (*models.Project, error) {
	_, err := s.farmRepo.FindByIDAndFarmerID(request.FarmLocationID, farmerID)
	if err != nil {
		// Jika GORM tidak menemukan record, berarti farm location tidak valid atau bukan milik farmer ini.
		return nil, fmt.Errorf("invalid farm_location_id: not found or you do not have permission")
	}

	farmLocationUUID, _ := uuid.Parse(request.FarmLocationID)
    
	startDate, err := time.Parse("2006-01-02", request.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", request.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: %w", err)
	}

	skillsJSON, err := json.Marshal(request.RequiredSkills)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal skills: %w", err)
	}

	project := &models.Project{
		FarmerID:       farmerID,
		FarmLocationID: &farmLocationUUID,
		Title:          request.Title,
		Description:    request.Description,
		ProjectType:    request.ProjectType,
		RequiredSkills: string(skillsJSON),
		WorkersNeeded:  request.WorkersNeeded,
		StartDate:      startDate,
		EndDate:        endDate,
		PaymentRate:    &request.PaymentRate,
		PaymentType:    request.PaymentType,
		Status:         "open",
	}

	if err := s.projectRepo.CreateProject(project); err != nil {
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
