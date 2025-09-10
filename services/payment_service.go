package services

import (
	"crypto/sha512"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/repositories"
)

// DTO untuk response ke frontend
type PaymentInitiationResponse struct {
	SnapToken string `json:"snap_token"`
}

type PaymentService interface {
	InitiatePayment(transactionID string, farmerID uuid.UUID) (*PaymentInitiationResponse, error)
	HandleWebhookNotification(notificationPayload map[string]interface{}) error
}

type paymentService struct {
	transactionRepo repositories.TransactionRepository
	userRepo        repositories.UserRepository
}

func NewPaymentService(txRepo repositories.TransactionRepository, userRepo repositories.UserRepository) PaymentService {
	return &paymentService{transactionRepo: txRepo, userRepo: userRepo}
}

func (s *paymentService) InitiatePayment(transactionID string, farmerID uuid.UUID) (*PaymentInitiationResponse, error) {
	tx, err := s.transactionRepo.FindByID(transactionID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	if tx.FromUserID != farmerID {
		return nil, fmt.Errorf("user not authorized for this transaction")
	}

	// Ambil data user dan tangani error jika tidak ditemukan
	farmerUser, err := s.userRepo.FindByID(farmerID.String())
	if err != nil {
		return nil, fmt.Errorf("farmer data not found for transaction: %w", err)
	}

	// Buat customer detail dengan aman, tangani jika nomor telepon nil
	customerDetail := &midtrans.CustomerDetails{
		FName: farmerUser.Name,
		Email: farmerUser.Email,
	}
	if farmerUser.PhoneNumber != nil {
		customerDetail.Phone = *farmerUser.PhoneNumber
	}

	// Buat request ke Midtrans Snap API
	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  tx.ID.String(),
			GrossAmt: int64(tx.TotalAmount),
		},
		CustomerDetail: customerDetail,
	}

	// Fix: Pastikan variabel err tidak ter-redeclare dan memiliki scope yang benar
	snapResponse, snapErr := config.SnapClient.CreateTransaction(snapReq)
	if snapErr != nil {
		return nil, fmt.Errorf("failed to create midtrans snap token: %w", snapErr)
	}

	return &PaymentInitiationResponse{SnapToken: snapResponse.Token}, nil
}

func (s *paymentService) HandleWebhookNotification(notificationPayload map[string]interface{}) error {
	orderID, _ := notificationPayload["order_id"].(string)
	transactionStatus, _ := notificationPayload["transaction_status"].(string)
	paymentType, _ := notificationPayload["payment_type"].(string)
	statusCode, _ := notificationPayload["status_code"].(string)
	grossAmount, _ := notificationPayload["gross_amount"].(string)
	signatureKey, _ := notificationPayload["signature_key"].(string)

	// 1. Validasi Signature Key (Keamanan Wajib)
	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	hashSource := orderID + statusCode + grossAmount + serverKey
	hasher := sha512.New()
	hasher.Write([]byte(hashSource))
	calculatedHash := fmt.Sprintf("%x", hasher.Sum(nil))

	if calculatedHash != signatureKey {
		return fmt.Errorf("invalid midtrans signature")
	}

	// 2. Update status transaksi berdasarkan notifikasi
	var newStatus string
	switch transactionStatus {
	case "settlement":
		newStatus = "hold" // Pembayaran berhasil, dana ditahan (escrow)
	case "cancel", "expire", "deny":
		newStatus = "failed" // Pembayaran gagal
	default:
		// Status lain (pending, challenge) tidak memerlukan aksi di backend,
		// jadi kita anggap notifikasi ini berhasil diproses.
		return nil
	}

	// 3. Simpan perubahan ke database
	if err := s.transactionRepo.UpdateStatusByOrderID(orderID, newStatus, paymentType); err != nil {
		return fmt.Errorf("failed to update transaction status for order %s: %w", orderID, err)
	}
	
	return nil
}