package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type PaymentHandler struct {
	paymentService services.PaymentService
}

func NewPaymentHandler(service services.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: service}
}

func (h *PaymentHandler) InitiatePayment(c *gin.Context) {
	transactionID := c.Param("id")
	userInterface, _ := c.Get("user")
	currentUser := userInterface.(*models.User)

	response, err := h.paymentService.InitiatePayment(transactionID, currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Payment token generated successfully", response)
}