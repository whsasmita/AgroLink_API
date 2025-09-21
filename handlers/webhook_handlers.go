package handlers

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/services"
	"gorm.io/datatypes"
)

type WebhookHandler struct {
	paymentService services.PaymentService
	webhookLogRepo repositories.WebhookLogRepository
}

func NewWebhookHandler(service services.PaymentService, logRepo repositories.WebhookLogRepository) *WebhookHandler {
	return &WebhookHandler{paymentService: service, webhookLogRepo: logRepo}
}

func (h *WebhookHandler) HandleMidtransNotification(c *gin.Context) {
	ctx := context.Background()

	// 1) Read raw body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("ERROR reading webhook body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot read request body"})
		return
	}
	// reset body for binding downstream
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 2) Snapshot headers
	headersMap := map[string][]string(c.Request.Header)

	// 3) Parse payload (best-effort)
	var payload map[string]interface{}
	_ = json.Unmarshal(bodyBytes, &payload)

	// 4) Pre-calc signature_valid (handler-level)
	orderID := getString(payload, "order_id")
	statusCode := getString(payload, "status_code")
	grossAmount := getString(payload, "gross_amount")
	sig := getString(payload, "signature_key")
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")

	sigValid := false
	if serverKey != "" && orderID != "" && statusCode != "" && grossAmount != "" && sig != "" {
		calc := sha512.Sum512([]byte(orderID + statusCode + grossAmount + serverKey))
		calcHex := fmt.Sprintf("%x", calc[:])
		sigValid = (calcHex == sig)
	}

	// 5) Create log entry
	logEntry := &models.WebhookLog{
		ID:                uuid.New(),
		Provider:          "midtrans",
		Event:             getString(payload, "transaction_status"),
		OrderID:           orderID,
		TransactionID:     getString(payload, "transaction_id"),
		PaymentType:       getString(payload, "payment_type"),
		StatusCode:        statusCode,
		TransactionStatus: getString(payload, "transaction_status"),
		FraudStatus:       getString(payload, "fraud_status"),
		SignatureKey:      sig,
		SignatureValid:    sigValid,
		Processed:         false,
		Headers:           datatypes.JSON([]byte(mustJSON(headersMap))),
		RawBody:           datatypes.JSON(bodyBytes),
		ParsedBody:        datatypes.JSON([]byte(mustJSON(payload))),
	}

	if err := h.webhookLogRepo.Create(ctx, logEntry); err != nil {
		log.Printf("WARNING: failed to persist webhook log: %v", err)
	}

	// 6) Call service
	if err := h.paymentService.HandleWebhookNotification(payload); err != nil {
		_ = h.webhookLogRepo.AttachError(ctx, logEntry.ID.String(), err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 7) Mark processed
	_ = h.webhookLogRepo.MarkProcessed(ctx, logEntry.ID.String())
	c.JSON(http.StatusOK, gin.H{"message": "Notification processed successfully"})
}

// helpers
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
