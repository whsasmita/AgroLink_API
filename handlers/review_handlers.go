package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type ReviewHandler struct {
	reviewService services.ReviewService
}

func NewReviewHandler(service services.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: service}
}

func (h *ReviewHandler) CreateReview(c *gin.Context) {
	projectID, _ := uuid.Parse(c.Param("id"))
	workerID, _ := uuid.Parse(c.Param("workerId"))

	var inputDTO dto.CreateReviewInput
	if err := c.ShouldBindJSON(&inputDTO); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	userInterface, _ := c.Get("user")
	currentUser := userInterface.(*models.User)

	inputDTO.ProjectID = projectID
	inputDTO.ReviewedWorkerID = workerID
	inputDTO.ReviewerID = currentUser.ID // Petani yang sedang login

	response, err := h.reviewService.CreateReview(inputDTO)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Review submitted successfully", response)
}