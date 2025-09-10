package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/services"
)

type WebhookHandler struct {
	paymentService services.PaymentService
}

func NewWebhookHandler(service services.PaymentService) *WebhookHandler {
	return &WebhookHandler{paymentService: service}
}

func (h *WebhookHandler) HandleMidtransNotification(c *gin.Context) {
	var notificationPayload map[string]interface{}
	if err := c.ShouldBindJSON(&notificationPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification format"})
		return
	}

	if err := h.paymentService.HandleWebhookNotification(notificationPayload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Notification processed"})
}