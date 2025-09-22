package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type DeliveryHandler struct {
	deliveryService services.DeliveryService
}

func NewDeliveryHandler(service services.DeliveryService) *DeliveryHandler {
	return &DeliveryHandler{deliveryService: service}
}

// CreateDelivery membuat permintaan pengiriman baru.
func (h *DeliveryHandler) CreateDelivery(c *gin.Context) {
	var input dto.CreateDeliveryRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	currentUser := c.MustGet("user").(*models.User)
	if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only farmers can create delivery requests", nil)
		return
	}

	delivery, err := h.deliveryService.CreateDelivery(input, currentUser.Farmer.UserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create delivery request", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Delivery request created successfully", delivery)
}

// FindDrivers mencari driver yang cocok untuk sebuah pengiriman.
func (h *DeliveryHandler) FindDrivers(c *gin.Context) {
	deliveryID := c.Param("id")
	// Ambil radius dari query param, dengan default 50km
	radius, _ := strconv.Atoi(c.DefaultQuery("radius", "50"))

	currentUser := c.MustGet("user").(*models.User)
	if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: User is not a farmer", nil)
		return
	}

	drivers, err := h.deliveryService.FindAvailableDrivers(deliveryID, currentUser.Farmer.UserID, radius)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Available drivers retrieved successfully", drivers)
}

// SelectDriver memilih driver untuk sebuah pengiriman.
func (h *DeliveryHandler) SelectDriver(c *gin.Context) {
	deliveryID := c.Param("id")
	driverID := c.Param("driverId")

	currentUser := c.MustGet("user").(*models.User)
	if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: User is not a farmer", nil)
		return
	}

	contract, err := h.deliveryService.SelectDriver(deliveryID, driverID, currentUser.Farmer.UserID.String())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Driver selected and contract offered", contract)
}

func (h *DeliveryHandler) GetMyDeliveries(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	var userID uuid.UUID
	if currentUser.Role == "farmer" && currentUser.Farmer != nil {
		userID = currentUser.Farmer.UserID
	} else if currentUser.Role == "driver" && currentUser.Driver != nil {
		userID = currentUser.Driver.UserID
	} else {
        utils.ErrorResponse(c, http.StatusForbidden, "User does not have a valid role for this action", nil)
        return
    }

	deliveries, err := h.deliveryService.GetMyDeliveries(userID, currentUser.Role)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve deliveries", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Deliveries retrieved successfully", deliveries)
}
