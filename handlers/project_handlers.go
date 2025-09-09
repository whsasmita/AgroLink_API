package handlers

import (
	"encoding/json"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type ProjectHandler struct {
	projectService services.ProjectService
}

func NewProjectHandler(service services.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: service}
}

func (h *ProjectHandler) CreateProject(c *gin.Context) {
	// 1. Ambil user dari konteks
	userInterface, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized: User not found in context", nil)
		return
	}
	currentUser, ok := userInterface.(*models.User)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error: Failed to process user data", nil)
		return
	}

	// 2. [PENYESUAIAN] Validasi peran berdasarkan relasi pointer.
	// Jika currentUser.Farmer adalah nil, berarti user ini bukan seorang petani.
	if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only farmers can create projects", nil)
		return
	}
	
	// 3. [PENYESUAIAN] Ambil ID Farmer dari primary key di model Farmer yang berelasi.
	farmerID := currentUser.Farmer.UserID

	// 4. Bind input JSON
	var request dto.CreateProjectRequest // Asumsi DTO ada di package `dto`
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	// 5. Panggil service dengan ID Farmer yang sudah divalidasi
	project, err := h.projectService.CreateProject(request, farmerID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create project", err)
		return
	}

	var skills []string
	if project.RequiredSkills != "" {
		_ = json.Unmarshal([]byte(project.RequiredSkills), &skills)
	}

    response := dto.CreateProjectResponse{
        ID:             project.ID,
        FarmerID:       project.FarmerID,
        FarmerName:     project.Farmer.User.Name,
        FarmLocationID: project.FarmLocationID,
        Title:          project.Title,
        Description:    project.Description,
        ProjectType:    project.ProjectType,
        RequiredSkills: skills,
        WorkersNeeded:  project.WorkersNeeded,
        StartDate:      project.StartDate,
        EndDate:        project.EndDate,
        PaymentRate:    project.PaymentRate,
        PaymentType:    project.PaymentType,
        Status:         project.Status,
    }

	utils.SuccessResponse(c, http.StatusCreated, "Project created successfully", response)
}



func (h *ProjectHandler) FindAllProjects(c *gin.Context) {
	var paginationRequest dto.PaginationRequest
	if err := c.ShouldBindQuery(&paginationRequest); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	projects, total, err := h.projectService.FindAll(paginationRequest)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get projects", err)
		return
	}

	var projectDTOs []dto.ProjectBriefResponse
	for _, p := range *projects {
		projectDTOs = append(projectDTOs, dto.ProjectBriefResponse{
			ID:          p.ID,
			Title:       p.Title,
			ProjectType: p.ProjectType,
			PaymentRate: p.PaymentRate,
			PaymentType: p.PaymentType,
			StartDate:   p.StartDate,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(paginationRequest.Limit)))
	response := dto.PaginationResponse{
		Data:       projectDTOs,
		Total:      total,
		Page:       paginationRequest.Page,
		Limit:      paginationRequest.Limit,
		TotalPages: totalPages,
	}

	utils.SuccessResponse(c, http.StatusOK, "Projects retrieved successfully", response)
}

func (h *ProjectHandler) GetProjectByID(c *gin.Context) {
	projectID := c.Param("id")

	project, err := h.projectService.FindByID(projectID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Project not found", err)
		return
	}

	var skills []string
	if err := json.Unmarshal([]byte(project.RequiredSkills), &skills); err != nil {
		skills = []string{}
	}

	response := dto.ProjectDetailResponse{
		ID:             project.ID,
		Title:          project.Title,
		Description:    project.Description,
		ProjectType:    project.ProjectType,
		RequiredSkills: skills,
		WorkersNeeded:  project.WorkersNeeded,
		StartDate:      project.StartDate,
		EndDate:        project.EndDate,
		PaymentRate:    project.PaymentRate,
		PaymentType:    project.PaymentType,
		Status:         project.Status,
		Farmer: dto.FarmerInfoResponse{
			ID:   project.Farmer.UserID,
			Name: project.Farmer.User.Name,
		},
	}

	utils.SuccessResponse(c, http.StatusOK, "Project retrieved successfully", response)
}