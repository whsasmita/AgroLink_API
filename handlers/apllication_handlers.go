package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type ApplicationHandler struct {
	appService services.ApplicationService
}

func NewApplicationHandler(appService services.ApplicationService) *ApplicationHandler {
	return &ApplicationHandler{appService: appService}
}

func (h *ApplicationHandler) ApplyToProject(c *gin.Context) {
	projectID := c.Param("id")
	userInterface, _ := c.Get("user")
	currentUser := userInterface.(*models.User)

	// [PERBAIKAN] Validasi peran dan ambil ID Worker yang benar
	if currentUser.Worker == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only workers can apply for projects", nil)
		return
	}
	workerID := currentUser.Worker.UserID.String()

	var input services.ApplyProjectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	application, err := h.appService.ApplyToProject(projectID, workerID, input)
	if err != nil {
        // ... (logika error Anda sudah baik)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit application", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Application submitted successfully", application)
}

func (h *ApplicationHandler) AcceptApplication(c *gin.Context) {
	applicationID := c.Param("id")
	userInterface, _ := c.Get("user")
	currentUser := userInterface.(*models.User)

	// [PERBAIKAN] Validasi peran dan ambil ID Farmer yang benar
	if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only farmers can accept applications", nil)
		return
	}
	farmerID := currentUser.Farmer.UserID.String()

	contract, err := h.appService.AcceptApplication(applicationID, farmerID)
	if err != nil {
        // ... (logika error Anda sudah baik)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process application", err)
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Application accepted and contract offered", contract)
}

func (h *ApplicationHandler) FindApplicationsByProjectID(c *gin.Context) {
    // 1. Ambil ID Proyek dari URL
    projectID := c.Param("id")

    // 2. Ambil ID Petani dari konteks JWT
    userInterface, _ := c.Get("user")
    currentUser := userInterface.(*models.User)
    if currentUser.Farmer == nil {
        utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: User is not a farmer", nil)
        return
    }
    farmerID := currentUser.Farmer.UserID.String()

    // 3. Panggil service
    applications, err := h.appService.FindApplicationsByProjectID(projectID, farmerID)
    if err != nil {
        if err.Error() == "project not found" {
            utils.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
            return
        }
        if err.Error() == "forbidden: you do not own this project" {
            utils.ErrorResponse(c, http.StatusForbidden, err.Error(), nil)
            return
        }
        utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
        return
    }

    // 4. Transformasi ke DTO Response
    var responses []services.ApplicationResponse
    for _, app := range applications {
        response := services.ApplicationResponse{
            ID:              app.ID,
            WorkerID:        app.WorkerID,
            ApplicationDate: app.CreatedAt,
            Message:         app.Message,
            Status:          app.Status,
        }
        // Pastikan data worker tidak nil sebelum diakses
        if app.Worker.UserID != uuid.Nil {
            response.WorkerName = app.Worker.User.Name
        }
        responses = append(responses, response)
    }

    utils.SuccessResponse(c, http.StatusOK, "Applications retrieved successfully", responses)
}

