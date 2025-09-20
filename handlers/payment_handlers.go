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

func (h *PaymentHandler) InitiateInvoicePayment(c *gin.Context) {
	invoiceID := c.Param("id")
	userInterface, _ := c.Get("user")
	currentUser := userInterface.(*models.User)

	response, err := h.paymentService.InitiateInvoicePayment(invoiceID, currentUser.Farmer.UserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Payment token generated successfully", response)
}

func (h *PaymentHandler) ReleaseProjectPayment(c *gin.Context) {
	projectID := c.Param("id")
	userInterface, _ := c.Get("user")
	currentUser := userInterface.(*models.User)

	if err := h.paymentService.ReleaseProjectPayment(projectID, currentUser.Farmer.UserID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Payment released and payouts initiated", nil)
}

func (h *PaymentHandler) ReleaseDeliveryPayment(c *gin.Context) {
	deliveryID := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

    if currentUser.Farmer == nil {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden: Only farmers can release payments", nil)
		return
	}

	if err := h.paymentService.ReleaseDeliveryPayment(deliveryID, currentUser.Farmer.UserID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Delivery payment released and payout scheduled", nil)
}


// Handler untuk melihat riwayat Invoice (contoh)
func (h *PaymentHandler) GetUserInvoices(c *gin.Context) {
    // Implementasi untuk mengambil daftar invoice milik user...
}