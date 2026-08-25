package services

import (
	"crypto/sha512"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

type PaymentService interface {
	InitiateInvoicePayment(invoiceID string, farmerID uuid.UUID) (*dto.PaymentInitiationResponse, error)
	InitiateCooperationPayment(cooperationID string, mitraID uuid.UUID) (*dto.PaymentInitiationResponse, error)
	HandleWebhookNotification(notificationPayload map[string]interface{}) error
	ReleaseProjectPayment(projectID string, farmerID uuid.UUID) error
	ReleaseDeliveryPayment(deliveryID string, farmerID uuid.UUID) error
	ReleaseCooperationFunds(cooperationID string, adminID uuid.UUID) error
}

type paymentService struct {
	invoiceRepo     repositories.InvoiceRepository
	transactionRepo repositories.TransactionRepository
	payoutRepo      repositories.PayoutRepository
	assignRepo      repositories.AssignmentRepository
	projectRepo     repositories.ProjectRepository
	userRepo        repositories.UserRepository
	deliveryRepo    repositories.DeliveryRepository
	coopRepo        repositories.MitraCooperationRepository
	contractRepo    repositories.ContractRepository
	mitraRepo       repositories.MitraProfileRepository
	db              *gorm.DB
}

func NewPaymentService(
	invoiceRepo repositories.InvoiceRepository,
	transactionRepo repositories.TransactionRepository,
	payoutRepo repositories.PayoutRepository,
	assignRepo repositories.AssignmentRepository,
	projectRepo repositories.ProjectRepository,
	userRepo repositories.UserRepository,
	deliveryRepo repositories.DeliveryRepository,
	coopRepo repositories.MitraCooperationRepository,
	contractRepo repositories.ContractRepository,
	mitraRepo repositories.MitraProfileRepository,
	db *gorm.DB,
) PaymentService {
	return &paymentService{
		invoiceRepo:     invoiceRepo,
		transactionRepo: transactionRepo,
		payoutRepo:      payoutRepo,
		assignRepo:      assignRepo,
		projectRepo:     projectRepo,
		userRepo:        userRepo,
		deliveryRepo:    deliveryRepo,
		coopRepo:        coopRepo,
		contractRepo:    contractRepo,
		mitraRepo:       mitraRepo,
		db:              db,
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

	response := &dto.PaymentInitiationResponse{
		SnapToken:   snapResponse.Token,
		OrderID:     invoice.ID.String(),
		Amount:      invoice.TotalAmount,
		RedirectURL: snapResponse.RedirectURL,
	}

	return response, nil
}

func (s *paymentService) InitiateCooperationPayment(cooperationID string, mitraID uuid.UUID) (*dto.PaymentInitiationResponse, error) {
	invoice, err := s.invoiceRepo.FindByMitraCooperationID(cooperationID)
	if err != nil {
		return nil, fmt.Errorf("invoice not found for this cooperation: %w", err)
	}

	coop, err := s.coopRepo.FindByID(cooperationID)
	if err != nil {
		return nil, fmt.Errorf("cooperation not found: %w", err)
	}

	if coop.MitraID != mitraID {
		return nil, fmt.Errorf("only designated Mitra can initiate payment for this cooperation")
	}

	if coop.Status != "waiting_payment" {
		return nil, fmt.Errorf("cooperation is not awaiting payment (current status: %s)", coop.Status)
	}

	mitraUser, err := s.userRepo.FindByID(mitraID.String())
	if err != nil {
		return nil, fmt.Errorf("mitra user data not found: %w", err)
	}

	customerDetail := &midtrans.CustomerDetails{
		FName: mitraUser.Name,
		Email: mitraUser.Email,
	}
	if mitraUser.PhoneNumber != nil {
		customerDetail.Phone = *mitraUser.PhoneNumber
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

	response := &dto.PaymentInitiationResponse{
		SnapToken:   snapResponse.Token,
		OrderID:     invoice.ID.String(),
		Amount:      invoice.TotalAmount,
		RedirectURL: snapResponse.RedirectURL,
	}

	return response, nil
}

func (s *paymentService) HandleWebhookNotification(notificationPayload map[string]interface{}) error {
	orderID, ok := notificationPayload["order_id"].(string)
	if !ok || orderID == "" {
		return fmt.Errorf("invalid payload: missing or invalid order_id")
	}

	// Abaikan notifikasi tes dari Midtrans
	if strings.HasPrefix(orderID, "payment_notif_test_") {
		log.Println("Received and acknowledged Midtrans test notification. Connectivity is OK.")
		return nil
	}

	transactionStatus, _ := notificationPayload["transaction_status"].(string)
	paymentType, _ := notificationPayload["payment_type"].(string)
	statusCode, _ := notificationPayload["status_code"].(string)
	grossAmount, _ := notificationPayload["gross_amount"].(string)
	signatureKey, _ := notificationPayload["signature_key"].(string)
	transactionIDMidtrans, _ := notificationPayload["transaction_id"].(string)
	fraudStatus, _ := notificationPayload["fraud_status"].(string)

	// Validasi signature (defense in depth; handler juga validasi)
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

	// Ambil invoice
	invoice, err := s.invoiceRepo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("invoice %s not found in internal system", orderID)
	}
	if invoice.Status == "paid" {
		log.Printf("Webhook for order %s already processed, ignoring duplicate.", orderID)
		return nil
	}

	finalizeSuccess := func() error {
		// 1) Invoice -> paid
		if err := s.invoiceRepo.UpdateStatus(invoice.ID.String(), "paid"); err != nil {
			return err
		}

		// 2) Catat transaction
		newTx := &models.Transaction{
			InvoiceID:                 invoice.ID,
			AmountPaid:                invoice.TotalAmount,
			PaymentMethod:             &paymentType,
			PaymentGatewayReferenceID: &transactionIDMidtrans,
		}
		if err := s.transactionRepo.Create(newTx); err != nil {
			_ = s.invoiceRepo.UpdateStatus(invoice.ID.String(), "pending")
			return fmt.Errorf("failed to create transaction record after payment: %w", err)
		}

		// 3) Routing: Project vs Delivery vs MitraCooperation
		if invoice.ProjectID != nil {
			if err := s.projectRepo.UpdateStatus(invoice.ProjectID.String(), "in_progress"); err != nil {
				log.Printf("WARN: project status update failed for project %s: %v", invoice.ProjectID.String(), err)
			}
			return nil
		}

		if invoice.DeliveryID != nil {
			delivery, derr := s.deliveryRepo.FindByID(invoice.DeliveryID.String())
			if derr != nil {
				log.Printf("WARN: delivery load failed for delivery %s: %v", invoice.DeliveryID.String(), derr)
				return nil
			}
			if delivery.Status == "pending_payment" || delivery.Status == "pending_signature" || delivery.Status == "pending_driver" {
				delivery.Status = "in_transit"
			} else {
				delivery.Status = "in_transit"
			}
			if err := s.deliveryRepo.Update(nil, delivery); err != nil {
				log.Printf("WARN: delivery update failed for delivery %s: %v", delivery.ID.String(), err)
			}
			return nil
		}

		if invoice.MitraCooperationID != nil {
			if err := s.finalizeCooperationEscrow(invoice.MitraCooperationID.String()); err != nil {
				log.Printf("WARN: cooperation escrow finalization failed for cooperation %s: %v", invoice.MitraCooperationID.String(), err)
			}
			return nil
		}

		log.Printf("INFO: invoice %s has no ProjectID, DeliveryID nor MitraCooperationID; skipped resource status update", invoice.ID.String())
		return nil
	}

	switch transactionStatus {
	case "capture":
		switch fraudStatus {
		case "accept":
			return finalizeSuccess()
		case "challenge":
			return s.invoiceRepo.UpdateStatus(invoice.ID.String(), "pending")
		case "deny":
			return s.invoiceRepo.UpdateStatus(invoice.ID.String(), "failed")
		default:
			return s.invoiceRepo.UpdateStatus(invoice.ID.String(), "pending")
		}

	case "settlement":
		return finalizeSuccess()

	case "pending":
		return s.invoiceRepo.UpdateStatus(invoice.ID.String(), "pending")

	case "expire", "cancel", "deny":
		return s.invoiceRepo.UpdateStatus(invoice.ID.String(), "failed")

	default:
		log.Printf("Unhandled transaction_status=%s, keeping invoice pending for order %s", transactionStatus, orderID)
		return nil
	}
}

func (s *paymentService) finalizeCooperationEscrow(cooperationID string) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	coop, err := s.coopRepo.FindByID(cooperationID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("cooperation not found: %w", err)
	}

	coop.Status = "escrowed"
	if err := s.coopRepo.Update(tx, coop); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update cooperation status to escrowed: %w", err)
	}

	now := time.Now()
	contract := &models.Contract{
		ContractType:        "mitra",
		MitraCooperationID:  &coop.ID,
		FarmerID:            coop.FarmerID,
		MitraID:             &coop.MitraID,
		SignedByFarmer:      true,
		SignedBySecondParty: true,
		SignedAt:            &now,
		Status:              "active",
	}

	if err := s.contractRepo.Create(tx, contract); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to generate mitra contract: %w", err)
	}

	coop.ContractID = &contract.ID
	coop.Status = "contract_generated"
	if err := s.coopRepo.Update(tx, coop); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to link contract to cooperation: %w", err)
	}

	return tx.Commit().Error
}

func (s *paymentService) ReleaseProjectPayment(projectID string, farmerID uuid.UUID) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	invoice, err := s.invoiceRepo.FindByProjectID(projectID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("invoice not found for this project")
	}
	if invoice.FarmerID != farmerID {
		tx.Rollback()
		return fmt.Errorf("user not authorized to release this payment")
	}
	if invoice.Status != "paid" {
		tx.Rollback()
		return fmt.Errorf("payment for this project is not completed yet")
	}

	transaction, err := s.transactionRepo.FindByInvoiceID(invoice.ID.String())
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("paid transaction not found for this invoice")
	}

	assignments, err := s.assignRepo.FindAllByProjectID(projectID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("could not retrieve worker assignments")
	}

	for _, assignment := range assignments {
		payout := models.Payout{
			TransactionID: transaction.ID,
			PayeeID:       assignment.WorkerID,
			PayeeType:     "worker",
			Amount:        assignment.AgreedRate,
		}
		if err := s.payoutRepo.Create(tx, &payout); err != nil {
			tx.Rollback()
			log.Printf("CRITICAL: Failed to create payout for worker %s: %v\n", assignment.WorkerID, err)
			return fmt.Errorf("failed to create payout record")
		}
	}

	if err := s.projectRepo.UpdateStatus(projectID, "completed"); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *paymentService) ReleaseDeliveryPayment(deliveryID string, farmerID uuid.UUID) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	invoice, err := s.invoiceRepo.FindByDeliveryID(deliveryID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("invoice not found for this delivery")
	}
	if invoice.FarmerID != farmerID {
		tx.Rollback()
		return fmt.Errorf("user not authorized to release this payment")
	}
	if invoice.Status != "paid" {
		tx.Rollback()
		return fmt.Errorf("payment for this delivery is not yet settled")
	}

	transaction, err := s.transactionRepo.FindByInvoiceID(invoice.ID.String())
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("paid transaction not found for this invoice")
	}
	delivery, err := s.deliveryRepo.FindByID(deliveryID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("delivery data not found")
	}

	if delivery.DriverID == nil {
		tx.Rollback()
		return fmt.Errorf("no driver assigned to this delivery")
	}
	payout := models.Payout{
		TransactionID: transaction.ID,
		PayeeID:       *delivery.DriverID,
		PayeeType:     "driver",
		Amount:        invoice.Amount,
	}
	if err := s.payoutRepo.Create(tx, &payout); err != nil {
		tx.Rollback()
		return err
	}

	delivery.Status = "delivered"
	if err := s.deliveryRepo.Update(tx, delivery); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (s *paymentService) ReleaseCooperationFunds(cooperationID string, adminID uuid.UUID) error {
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	coop, err := s.coopRepo.FindByID(cooperationID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("cooperation not found")
	}

	if coop.Status != "escrowed" && coop.Status != "contract_generated" {
		tx.Rollback()
		return fmt.Errorf("cooperation is not in escrowed or contract_generated state (current: %s)", coop.Status)
	}

	invoice, err := s.invoiceRepo.FindByMitraCooperationID(cooperationID)
	if err != nil || invoice.Status != "paid" {
		tx.Rollback()
		return fmt.Errorf("paid invoice not found for this cooperation")
	}

	transaction, err := s.transactionRepo.FindByInvoiceID(invoice.ID.String())
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("paid transaction record not found")
	}

	payout := models.Payout{
		TransactionID: transaction.ID,
		PayeeID:       coop.FarmerID,
		PayeeType:     "farmer",
		Amount:        invoice.Amount,
	}

	if err := s.payoutRepo.Create(tx, &payout); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create payout for farmer: %w", err)
	}

	coop.Status = "completed"
	if err := s.coopRepo.Update(tx, coop); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update cooperation status to completed: %w", err)
	}

	mitraProfile, err := s.mitraRepo.FindByUserID(coop.MitraID)
	if err == nil && mitraProfile != nil {
		mitraProfile.TotalTransaksiBerhasil++
		_ = s.mitraRepo.Update(tx, mitraProfile)
	}

	return tx.Commit().Error
}
