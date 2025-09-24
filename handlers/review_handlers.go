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
	deliveryService services.DeliveryService
}

func NewReviewHandler(service services.ReviewService, deliverService services.DeliveryService) *ReviewHandler {
	return &ReviewHandler{reviewService: service, deliveryService: deliverService}
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

func (h *ReviewHandler) CreateDriverReview(c *gin.Context) {
	deliveryID, _ := uuid.Parse(c.Param("deliveryId"))
	
	var inputDTO dto.CreateDriverReviewInput
	if err := c.ShouldBindJSON(&inputDTO); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	currentUser := c.MustGet("user").(*models.User)
    
    // Ambil driverID dari delivery (perlu sedikit penyesuaian di service/repo)
    delivery, _ := h.deliveryService.FindByID(deliveryID.String())

	inputDTO.DeliveryID = deliveryID
	inputDTO.ReviewedDriverID = *delivery.DriverID
	inputDTO.ReviewerID = currentUser.ID // Petani yang sedang login

	review, err := h.reviewService.CreateDriverReview(inputDTO)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Driver review submitted successfully", review)
}
