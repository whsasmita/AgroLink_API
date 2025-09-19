package services

import (
	"crypto/sha512"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
)

type PaymentService interface {
	InitiateInvoicePayment(invoiceID string, farmerID uuid.UUID) (*dto.PaymentInitiationResponse, error)
	HandleWebhookNotification(notificationPayload map[string]interface{}) error
	ReleaseProjectPayment(projectID string, farmerID uuid.UUID) error
}

type paymentService struct {
	invoiceRepo     repositories.InvoiceRepository
	transactionRepo repositories.TransactionRepository
	payoutRepo      repositories.PayoutRepository
	assignRepo      repositories.AssignmentRepository
	projectRepo     repositories.ProjectRepository
	userRepo        repositories.UserRepository
}

func NewPaymentService(
	invoiceRepo repositories.InvoiceRepository,
	transactionRepo repositories.TransactionRepository,
	payoutRepo repositories.PayoutRepository,
	assignRepo repositories.AssignmentRepository,
	projectRepo repositories.ProjectRepository,
	userRepo repositories.UserRepository,
) PaymentService {
	return &paymentService{
		invoiceRepo:     invoiceRepo,
		transactionRepo: transactionRepo,
		payoutRepo:      payoutRepo,
		assignRepo:      assignRepo,
		projectRepo:     projectRepo,
		userRepo:        userRepo,
	}
}

func (s *paymentService) InitiateInvoicePayment(invoiceID string, farmerID uuid.UUID) (*dto.PaymentInitiationResponse, error) {
	invoice, err := s.invoiceRepo.FindByID(invoiceID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}
	if invoice.FarmerID != farmerID {
		return nil, fmt.Errorf("user not authorized for this invoice")
	}
	if invoice.Status != "pending" {
		return nil, fmt.Errorf("invoice has already been processed")
	}

	farmerUser, err := s.userRepo.FindByID(farmerID.String())
	if err != nil {
		return nil, fmt.Errorf("farmer data not found for transaction: %w", err)
	}

	customerDetail := &midtrans.CustomerDetails{
		FName: farmerUser.Name,
		Email: farmerUser.Email,
	}
	if farmerUser.PhoneNumber != nil {
		customerDetail.Phone = *farmerUser.PhoneNumber
	}

	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  invoice.ID.String(),
			GrossAmt: int64(invoice.TotalAmount),
		},
		CustomerDetail: customerDetail,
	}

	snapResponse, midtransErr := config.SnapClient.CreateTransaction(snapReq)
	if midtransErr != nil {
		return nil, fmt.Errorf("failed to create midtrans snap token: %s (StatusCode: %d)", midtransErr.Message, midtransErr.StatusCode)
	}

	// [PERBAIKAN] Buat dan isi DTO response secara lengkap
	response := &dto.PaymentInitiationResponse{
		SnapToken:   snapResponse.Token,
		OrderID:     invoice.ID.String(),        // Ambil dari invoice
		Amount:      invoice.TotalAmount,        // Ambil dari invoice
		RedirectURL: snapResponse.RedirectURL,   // Ambil dari response Midtrans
	}

	return response, nil
}


func (s *paymentService) HandleWebhookNotification(notificationPayload map[string]interface{}) error {
	orderID, ok := notificationPayload["order_id"].(string)
	if !ok || orderID == "" {
		return fmt.Errorf("invalid payload: missing or invalid order_id")
	}

	if strings.HasPrefix(orderID, "payment_notif_test_") {
		log.Println("Received and acknowledged Midtrans test notification. Finding a real pending invoice to process...")
		pendingInvoice, err := s.invoiceRepo.FindFirstPending()
		if err != nil {
			return fmt.Errorf("test notification received, but no pending invoice found to process: %w", err)
		}
		orderID = pendingInvoice.ID.String()
		log.Printf("Swapped test order_id with real pending invoice_id: %s", orderID)
	}

	transactionStatus, _ := notificationPayload["transaction_status"].(string)
	paymentType, _ := notificationPayload["payment_type"].(string)
	statusCode, _ := notificationPayload["status_code"].(string)
	grossAmount, _ := notificationPayload["gross_amount"].(string)
	signatureKey, _ := notificationPayload["signature_key"].(string)
	transactionIDMidtrans, _ := notificationPayload["transaction_id"].(string)

	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	if serverKey == "" {
		return fmt.Errorf("MIDTRANS_SERVER_KEY is not configured")
	}
	hashSource := orderID + statusCode + grossAmount + serverKey
	hasher := sha512.New()
	hasher.Write([]byte(hashSource))
	calculatedHash := fmt.Sprintf("%x", hasher.Sum(nil))

	if calculatedHash != signatureKey {
		return fmt.Errorf("invalid midtrans signature")
	}

	invoice, err := s.invoiceRepo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("invoice %s not found in internal system", orderID)
	}
	if invoice.Status == "paid" {
		log.Printf("Webhook for order %s already processed, ignoring duplicate.", orderID)
		return nil
	}

	if transactionStatus == "settlement" {
		if err := s.invoiceRepo.UpdateStatus(invoice.ID.String(), "paid"); err != nil {
			return err
		}

		newTx := &models.Transaction{
			InvoiceID:                 invoice.ID,
			AmountPaid:                invoice.TotalAmount,
			PaymentMethod:             &paymentType,
			PaymentGatewayReferenceID: &transactionIDMidtrans,
		}
		if err := s.transactionRepo.Create(newTx); err != nil {
			s.invoiceRepo.UpdateStatus(invoice.ID.String(), "pending")
			return fmt.Errorf("failed to create transaction record after payment: %w", err)
		}
		return s.projectRepo.UpdateStatus(invoice.ProjectID.String(), "in_progress")

	} else if transactionStatus == "expire" || transactionStatus == "cancel" || transactionStatus == "deny" {
		return s.invoiceRepo.UpdateStatus(invoice.ID.String(), "failed")
	}

	return nil
}


func (s *paymentService) ReleaseProjectPayment(projectID string, farmerID uuid.UUID) error {
	invoice, err := s.invoiceRepo.FindByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("invoice not found for this project")
	}

	if invoice.FarmerID != farmerID {
		return fmt.Errorf("user not authorized to release this payment")
	}
	if invoice.Status != "paid" {
		return fmt.Errorf("payment for this project is not completed yet")
	}

	transaction, err := s.transactionRepo.FindByInvoiceID(invoice.ID.String())
	if err != nil {
		return fmt.Errorf("paid transaction not found for this invoice")
	}

	assignments, err := s.assignRepo.FindAllByProjectID(projectID)
	if err != nil {
		return fmt.Errorf("could not retrieve worker assignments")
	}

	// Buat Payout untuk setiap pekerja
	for _, assignment := range assignments {
		payout := models.Payout{
			TransactionID: transaction.ID,
			WorkerID:      assignment.WorkerID,
			Amount:        assignment.AgreedRate, // Gaji sesuai kesepakatan
		}
		if err := s.payoutRepo.Create(&payout); err != nil {
			// Di produksi, ini harusnya masuk ke sistem antrian untuk dicoba lagi
			fmt.Printf("CRITICAL: Failed to create payout for worker %s: %v\n", assignment.WorkerID, err)
		}
	}

	// Update status project menjadi 'completed'
	return s.projectRepo.UpdateStatus(projectID, "completed")
}