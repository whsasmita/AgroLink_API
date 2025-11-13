package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type ECommerceWebhookHandler struct {
	paymentService services.ECommercePaymentService
}

func NewECommerceWebhookHandler(s services.ECommercePaymentService) *ECommerceWebhookHandler {
	return &ECommerceWebhookHandler{paymentService: s}
}

func (h *ECommerceWebhookHandler) HandleMidtransNotification(c *gin.Context) {
	var notificationPayload map[string]interface{}
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	
	if err := json.Unmarshal(bodyBytes, &notificationPayload); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload", err)
		return
	}

	if err := h.paymentService.HandleWebhook(notificationPayload); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Notification processed successfully", nil)
}