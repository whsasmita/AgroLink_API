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
	"gorm.io/gorm"
)

type PaymentService interface {
	InitiateInvoicePayment(invoiceID string, farmerID uuid.UUID) (*dto.PaymentInitiationResponse, error)
	HandleWebhookNotification(notificationPayload map[string]interface{}) error
	ReleaseProjectPayment(projectID string, farmerID uuid.UUID) error
	ReleaseDeliveryPayment(deliveryID string, farmerID uuid.UUID) error
}
type paymentService struct {
	invoiceRepo     repositories.InvoiceRepository
	transactionRepo repositories.TransactionRepository
	payoutRepo      repositories.PayoutRepository
	assignRepo      repositories.AssignmentRepository
	projectRepo     repositories.ProjectRepository
	userRepo        repositories.UserRepository
	deliveryRepo repositories.DeliveryRepository
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
	db *gorm.DB,
) PaymentService {
	return &paymentService{
		invoiceRepo:     invoiceRepo,
		transactionRepo: transactionRepo,
		payoutRepo:      payoutRepo,
		assignRepo:      assignRepo,
		projectRepo:     projectRepo,
		userRepo:        userRepo,
		deliveryRepo: deliveryRepo,
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
		// (opsional) tambahkan Items, Expiry, dsb.
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

    // 2) Catat transaction (idempotensi di level DB: tambahkan unique index jika belum)
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

    // 3) Routing: Project vs Delivery
    //    NOTE: hindari panic bila pointer nil
    if invoice.ProjectID != nil {
        // Jangan kirim tx = nil kalau repo tidak siap; sediakan variant tanpa tx
        if err := s.projectRepo.UpdateStatus(invoice.ProjectID.String(), "in_progress"); err != nil {
            log.Printf("WARN: project status update failed for project %s: %v", invoice.ProjectID.String(), err)
            // TODO: enqueue retry / outbox (jangan gagalkan webhook)
        }
        return nil
    }

    if invoice.DeliveryID != nil {
        // Ambil delivery & set status ke 'in_transit'
        delivery, derr := s.deliveryRepo.FindByID(invoice.DeliveryID.String())
        if derr != nil {
            log.Printf("WARN: delivery load failed for delivery %s: %v", invoice.DeliveryID.String(), derr)
            return nil // jangan gagalkan webhook
        }
        // Pastikan transisi sesuai enum kamu
        if delivery.Status == "pending_payment" || delivery.Status == "pending_signature" || delivery.Status == "pending_driver" {
            delivery.Status = "in_transit"
        } else {
            // fallback aman: tetap set in_transit setelah paid
            delivery.Status = "in_transit"
        }
        if err := s.deliveryRepo.Update(nil, delivery); err != nil {
            // Jika Update(nil, ...) tidak diterima, ganti ke UpdateStatus seperti di bawah (bagian repo)
            log.Printf("WARN: delivery update failed for delivery %s: %v", delivery.ID.String(), err)
        }
        return nil
    }

    // 4) Invoice tanpa project & delivery (kasus tak terduga)
    log.Printf("INFO: invoice %s has no ProjectID nor DeliveryID; skipped resource status update", invoice.ID.String())
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

func (s *paymentService) ReleaseProjectPayment(projectID string, farmerID uuid.UUID) error {
	// Mulai transaksi database
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Validasi
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

	// 2. Buat Payout untuk setiap pekerja di dalam transaksi
	for _, assignment := range assignments {
		payout := models.Payout{
			TransactionID: transaction.ID,
			WorkerID:      assignment.WorkerID,
			Amount:        assignment.AgreedRate,
		}
		// Pastikan payoutRepo.Create menerima objek 'tx'
		if err := s.payoutRepo.Create(tx, &payout); err != nil {
			tx.Rollback()
			log.Printf("CRITICAL: Failed to create payout for worker %s: %v\n", assignment.WorkerID, err)
			return fmt.Errorf("failed to create payout record")
		}
	}

	// 3. Update status project menjadi 'completed' di dalam transaksi
	if err := s.projectRepo.UpdateStatus( projectID, "completed"); err != nil {
		tx.Rollback()
		return err
	}

	// 4. Jika semua berhasil, commit transaksi
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

	// 1. Validasi: Ambil invoice berdasarkan deliveryID dan cek kepemilikan & status
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

	// 2. Ambil data terkait
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

	// 3. Buat Payout untuk Driver
	// Pastikan driver sudah terpilih di data delivery
	if delivery.DriverID == nil {
		tx.Rollback()
		return fmt.Errorf("no driver assigned to this delivery")
	}
	payout := models.Payout{
		TransactionID: transaction.ID,
		WorkerID:      *delivery.DriverID, // Menggunakan WorkerID sebagai field generik
		Amount:        invoice.Amount,     // Gaji driver adalah jumlah pokok
	}
	if err := s.payoutRepo.Create(tx, &payout); err != nil {
		tx.Rollback()
		return err
	}

	// 4. Update status Delivery menjadi 'delivered'
	delivery.Status = "delivered"
	if err := s.deliveryRepo.Update(tx, delivery); err != nil {
		tx.Rollback()
		return err
	}
	
	// 5. Commit transaksi jika semua berhasil
	return tx.Commit().Error
}






