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
	ecommPaymentService services.ECommercePaymentService
	webhookLogRepo repositories.WebhookLogRepository
	invoiceRepo      repositories.InvoiceRepository
	ecommPaymentRepo repositories.ECommercePaymentRepository
}

func NewWebhookHandler(service services.PaymentService, logRepo repositories.WebhookLogRepository, ecommerceService services.ECommercePaymentService, invoiceRepo repositories.InvoiceRepository, ecommercePaymentRepo repositories.ECommercePaymentRepository ) *WebhookHandler {
	return &WebhookHandler{paymentService: service, webhookLogRepo: logRepo, ecommPaymentService: ecommerceService, invoiceRepo: invoiceRepo, ecommPaymentRepo: ecommercePaymentRepo}
}

func (h *WebhookHandler) HandleMidtransNotification(c *gin.Context) {
	ctx := context.Background()
	reqID := uuid.NewString()

	// 1) Read raw body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("[WEBHOOK %s] ERROR reading body: %v", reqID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot read request body"})
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Penting untuk re-buffering

	log.Printf("[WEBHOOK %s] >>> NEW MIDTRANS NOTIFICATION", reqID)
	
	// 2) Parse payload
	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		log.Printf("[WEBHOOK %s] WARN cannot parse JSON: %v", reqID, err)
	}
	payload["_req_id"] = reqID

	// 3) Validasi signature (handler-level)
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
	
	// 4) Create log entry
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
		Headers:           datatypes.JSON([]byte(mustJSON(c.Request.Header))),
		RawBody:           datatypes.JSON(bodyBytes),
		ParsedBody:        datatypes.JSON([]byte(mustJSON(payload))),
	}
	if err := h.webhookLogRepo.Create(ctx, logEntry); err != nil {
		log.Printf("[WEBHOOK %s] WARNING persist log failed: %v", reqID, err)
	}

	// 5) [ROUTING LOGIC] Panggil service yang tepat
	var serviceErr error
	routed := false

	// Cek 1: Apakah ini Invoice (Proyek/Delivery)?
	_, errInvoice := h.invoiceRepo.FindByID(orderID)
	if errInvoice == nil {
		log.Printf("[WEBHOOK %s] Routing to Core PaymentService", reqID)
		serviceErr = h.paymentService.HandleWebhookNotification(payload)
		routed = true
	}

	// Cek 2: Jika bukan, apakah ini ECommercePayment?
	if !routed {
		_, errEcomm := h.ecommPaymentRepo.FindByID(orderID)
		if errEcomm == nil {
			log.Printf("[WEBHOOK %s] Routing to E-Commerce PaymentService", reqID)
			serviceErr = h.ecommPaymentService.HandleWebhook(payload)
			routed = true
		}
	}

	// 6) Handle hasil
	if !routed {
		log.Printf("[WEBHOOK %s] SERVICE ERROR: OrderID %s not found in any service", reqID, orderID)
		_ = h.webhookLogRepo.AttachError(ctx, logEntry.ID.String(), "OrderID not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Order ID not found"})
		return
	}

	if serviceErr != nil {
		log.Printf("[WEBHOOK %s] SERVICE ERROR: %v", reqID, serviceErr)
		_ = h.webhookLogRepo.AttachError(ctx, logEntry.ID.String(), serviceErr.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": serviceErr.Error()})
		return
	}

	// 7) Mark processed
	_ = h.webhookLogRepo.MarkProcessed(ctx, logEntry.ID.String())
	log.Printf("[WEBHOOK %s] processed OK", reqID)
	c.JSON(http.StatusOK, gin.H{"message": "Notification processed successfully"})
}


// --- Helpers ---
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