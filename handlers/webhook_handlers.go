package handlers

import (
	"bytes" // <-- Import baru
	"io"
	"log"
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
	// [DEBUGGING] Langkah 1: Baca dan log body mentah dari request
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("ERROR reading webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot read request body"})
		return
	}
	log.Printf("--- Raw Midtrans Webhook Body ---")
	log.Println(string(bodyBytes))
	log.Printf("---------------------------------")

	// [PENTING] Kembalikan body ke request context agar bisa dibaca lagi oleh ShouldBindJSON
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))


	var notificationPayload map[string]interface{}
	
	if err := c.ShouldBindJSON(&notificationPayload); err != nil {
		log.Printf("Invalid webhook payload (JSON binding failed): %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification format"})
		return
	}
	
	orderID, _ := notificationPayload["order_id"].(string)
	transactionStatus, _ := notificationPayload["transaction_status"].(string)
	log.Printf("Received Midtrans webhook for order %s with status %s", orderID, transactionStatus)
	
	if err := h.paymentService.HandleWebhookNotification(notificationPayload); err != nil {
		log.Printf("Webhook processing failed for order %s: %v", orderID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Notification processed successfully"})
}