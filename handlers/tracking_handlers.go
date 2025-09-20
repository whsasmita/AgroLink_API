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

type TrackingHandler struct {
	trackingService services.TrackingService
}

func NewTrackingHandler(service services.TrackingService) *TrackingHandler {
	return &TrackingHandler{trackingService: service}
}

// UpdateLocation adalah handler untuk driver mengirim update lokasi.
func (h *TrackingHandler) UpdateLocation(c *gin.Context) {
	deliveryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid delivery ID", err)
		return
	}

	var input dto.UpdateLocationRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	currentUser := c.MustGet("user").(*models.User)
	if currentUser.Driver == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only drivers can update location", nil)
		return
	}

	err = h.trackingService.UpdateLocation(deliveryID, currentUser.Driver.UserID, input.Lat, input.Lng)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Location updated successfully", nil)
}

// GetLatestLocation adalah handler untuk petani mendapatkan lokasi terakhir.
func (h *TrackingHandler) GetLatestLocation(c *gin.Context) {
	deliveryID := c.Param("id")

	currentUser := c.MustGet("user").(*models.User)
	if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only farmers can track deliveries", nil)
		return
	}

	location, err := h.trackingService.GetLatestLocation(deliveryID, currentUser.Farmer.UserID)
	if err != nil {
		// Jika record tidak ditemukan, kembalikan data kosong, bukan error 500
		if err.Error() == "record not found" {
			utils.SuccessResponse(c, http.StatusOK, "No location data found yet", nil)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Latest location retrieved successfully", location)
}