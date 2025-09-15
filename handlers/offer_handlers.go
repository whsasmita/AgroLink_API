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

type OfferHandler struct {
	offerService services.OfferService
}

func NewOfferHandler(service services.OfferService) *OfferHandler {
	return &OfferHandler{offerService: service}
}

func (h *OfferHandler) CreateDirectOffer(c *gin.Context) {
	workerID, err := uuid.Parse(c.Param("workerId"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid worker ID format", err)
		return
	}

	currentUser := c.MustGet("user").(*models.User)
	if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only farmers can make direct offers", nil)
		return
	}
	farmerID := currentUser.Farmer.UserID

	var input dto.DirectOfferRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	// 'response' sekarang sudah bertipe *dto.DirectOfferResponse
	response, err := h.offerService.CreateDirectOffer(input, farmerID, workerID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Direct offer sent successfully", response)
}