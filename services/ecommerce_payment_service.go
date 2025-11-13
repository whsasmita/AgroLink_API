package services

import (
	"crypto/sha512"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)
type ECommercePaymentService interface {
	InitiatePayment(tx *gorm.DB, userID uuid.UUID, orders []models.Order, grandTotal float64) (*dto.PaymentInitiationResponse, error)
	HandleWebhook(notificationPayload map[string]interface{}) error
}

type eCommercePaymentService struct {
	paymentRepo repositories.ECommercePaymentRepository
	orderRepo   repositories.OrderRepository
	userRepo    repositories.UserRepository
	productRepo repositories.ProductRepository
	db          *gorm.DB
}

// NewECommercePaymentService adalah constructor untuk service pembayaran e-commerce.
func NewECommercePaymentService(
	paymentRepo repositories.ECommercePaymentRepository,
	orderRepo repositories.OrderRepository,
	userRepo repositories.UserRepository,
	productRepo repositories.ProductRepository,
	db *gorm.DB,
) ECommercePaymentService {
	return &eCommercePaymentService{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		userRepo:    userRepo,
		productRepo: productRepo,
		db:          db,
	}
}

// InitiatePayment dipanggil DARI DALAM transaksi CheckoutService.
// Tugasnya adalah membuat record Payment dan mendapatkan Snap Token.
func (s *eCommercePaymentService) InitiatePayment(tx *gorm.DB, userID uuid.UUID, orders []models.Order, grandTotal float64) (*dto.PaymentInitiationResponse, error) {
	// 1. Buat record pembayaran induk
	payment := &models.ECommercePayment{
		ID:         uuid.New(),
		UserID:     userID,
		GrandTotal: grandTotal,
		Status:     "pending",
		Orders:     orders, 
	}
	if err := s.paymentRepo.Create(tx, payment); err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// 2. Ambil detail pelanggan
	user, err := s.userRepo.FindByID(userID.String())
	if err != nil {
		return nil, fmt.Errorf("customer data not found: %w", err)
	}

	// 3. Buat request ke Midtrans Snap
	snapReq := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  payment.ID.String(), // Gunakan ID payment e-commerce sebagai OrderID
			GrossAmt: int64(grandTotal),
		},
		CustomerDetail: &midtrans.CustomerDetails{
			FName: user.Name,
			Email: user.Email,
			Phone: *user.PhoneNumber,
		},
	}

	snapResponse, midtransErr := config.SnapClient.CreateTransaction(snapReq)
	if midtransErr != nil {
		return nil, fmt.Errorf("failed to create midtrans snap token: %s", midtransErr.Message)
	}

	// 4. Simpan SnapToken ke record pembayaran
	payment.SnapToken = snapResponse.Token
	if err := s.paymentRepo.Update(tx, payment); err != nil {
		return nil, err
	}

	// 5. Kembalikan DTO respons
	return &dto.PaymentInitiationResponse{
		SnapToken:   snapResponse.Token,
		RedirectURL: snapResponse.RedirectURL,
		OrderID:     payment.ID.String(),
	}, nil
}

// HandleWebhook menangani notifikasi dari Midtrans khusus e-commerce.
func (s *eCommercePaymentService) HandleWebhook(notificationPayload map[string]interface{}) error {
	paymentID, _ := notificationPayload["order_id"].(string)
	if paymentID == "" {
		return fmt.Errorf("invalid payload: missing order_id")
	}

	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		return fmt.Errorf("e-commerce payment %s not found", paymentID)
	}
	
	if payment.Status == "paid" {
		log.Printf("E-commerce webhook for payment %s already processed.", paymentID)
		return nil
	}

	// ... (Validasi signature key, sama seperti di service utama) ...
	statusCode, _ := notificationPayload["status_code"].(string)
	grossAmount, _ := notificationPayload["gross_amount"].(string)
	signatureKey, _ := notificationPayload["signature_key"].(string)
	transactionStatus, _ := notificationPayload["transaction_status"].(string)
	fraudStatus, _ := notificationPayload["fraud_status"].(string)

	serverKey := os.Getenv("MIDTRANS_SERVER_KEY")
	if serverKey == "" {
		return fmt.Errorf("MIDTRANS_SERVER_KEY is not configured")
	}
	hashSource := paymentID + statusCode + grossAmount + serverKey
	hasher := sha512.New()
	hasher.Write([]byte(hashSource))
	calculatedHash := fmt.Sprintf("%x", hasher.Sum(nil))
	
	if calculatedHash != signatureKey {
		return fmt.Errorf("invalid midtrans signature for e-commerce payment %s", paymentID)
	}
	// Fungsi internal untuk memproses pembayaran yang sukses
	finalizeSuccess := func() error {
		return s.db.Transaction(func(tx *gorm.DB) error {
			// 1. Update status ECommercePayment menjadi 'paid'
			if err := s.paymentRepo.UpdateStatus(tx, paymentID, "paid"); err != nil {
				return err
			}
			
			// 2. Update status SEMUA Order yang terhubung menjadi 'paid'
			if err := s.orderRepo.UpdateStatusByPaymentID(tx, payment.ID, "paid"); err != nil {
				log.Printf("WARN: E-commerce orders status update failed for payment %s: %v", paymentID, err)
			}

			// 3. [BARU] Kurangi Stok Produk
			orders, err := s.orderRepo.FindOrdersByPaymentID(tx, payment.ID)
			if err != nil {
				log.Printf("WARN: Gagal mengambil order untuk mengurangi stok: %v", err)
				return nil // Jangan gagalkan transaksi utama
			}

			for _, order := range orders {
				for _, item := range order.Items {
					product, err := s.productRepo.FindByIDForUpdate(tx, item.ProductID)
					if err != nil {
						log.Printf("WARN: Gagal menemukan produk %s untuk mengurangi stok: %v", item.ProductID, err)
						continue
					}
					
					// Kurangi stok total DAN stok yang direservasi
					product.Stock -= item.Quantity
					product.ReservedStock -= item.Quantity
					
					if product.Stock < 0 { product.Stock = 0 }
					if product.ReservedStock < 0 { product.ReservedStock = 0 }

					if err := s.productRepo.UpdateStock(tx, product); err != nil {
						log.Printf("WARN: Gagal mengupdate stok produk %s: %v", item.ProductID, err)
					}
				}
			}

			return nil // Commit
		})
	}

	// Logika status Midtrans
	switch transactionStatus {
	case "capture", "settlement":
		if fraudStatus == "accept" || fraudStatus == "" {
			return finalizeSuccess()
		}
		return s.paymentRepo.UpdateStatus(nil, paymentID, "failed") // Gunakan tx jika ingin transaksional
	case "expire", "cancel", "deny":
		return s.paymentRepo.UpdateStatus(nil, paymentID, "failed")
	default:
		return nil // Abaikan status lain seperti "pending"
	}
}