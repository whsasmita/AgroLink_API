package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

type CheckoutService interface {
	CreateOrdersFromCart(userID uuid.UUID) (*dto.PaymentInitiationResponse, error)
}

type checkoutService struct {
	cartRepo       repositories.CartRepository    // Asumsi dari modul inti
	productRepo    repositories.ProductRepository // Asumsi dari modul inti
	orderRepo      repositories.OrderRepository
	paymentService ECommercePaymentService
	db             *gorm.DB
}

func NewCheckoutService(
	cartRepo repositories.CartRepository,
	productRepo repositories.ProductRepository,
	orderRepo repositories.OrderRepository,
	paymentService ECommercePaymentService,
	db *gorm.DB,
) CheckoutService {
	return &checkoutService{
		cartRepo:       cartRepo,
		productRepo:    productRepo,
		orderRepo:      orderRepo,
		paymentService: paymentService,
		db:             db,
	}
}

func (s *checkoutService) CreateOrdersFromCart(userID uuid.UUID) (*dto.PaymentInitiationResponse, error) {
	var snapResponse *dto.PaymentInitiationResponse

	err := s.db.Transaction(func(tx *gorm.DB) error {
		cartItems, err := s.cartRepo.FindByUserIDWithTx(tx, userID)
		if err != nil {
			return err
		}
		if len(cartItems) == 0 {
			return errors.New("cart is empty")
		}

		itemsByFarmer := make(map[uuid.UUID][]models.Cart)
		var grandTotal float64
		var createdOrders []models.Order

		// 2. Validasi Stok, Kelompokkan Item & Hitung Total
		for _, item := range cartItems {
			product, err := s.productRepo.FindByIDForUpdate(tx, item.ProductID)
			if err != nil {
				return fmt.Errorf("product %s not found", item.ProductID)
			}
			availableStock := product.Stock - product.ReservedStock
			if availableStock < item.Quantity {
				return fmt.Errorf("insufficient stock for product: %s", product.Title)
			}
			// product.ReservedStock -= item.Quantity // Hapus reservasi keranjang
			// if err := s.productRepo.UpdateStock(tx, product); err != nil {
			// 	return err
			// }
			itemsByFarmer[item.Product.FarmerID] = append(itemsByFarmer[item.Product.FarmerID], item)
			grandTotal += float64(item.Quantity) * product.Price
		}

		// 3. Buat Order Terpisah untuk Setiap Petani
		for farmerID, items := range itemsByFarmer {
			var orderTotal float64
			for _, item := range items {
				orderTotal += float64(item.Quantity) * item.Product.Price
			}

			newOrder := models.Order{
				FarmerID: farmerID,
				UserID:      userID,
				InvoiceNumber: fmt.Sprintf("ORD-%d", time.Now().UnixNano()),
				TotalAmount:   orderTotal,
			}
			if err := s.orderRepo.CreateWithItems(tx, &newOrder, items); err != nil {
				return err
			}
			createdOrders = append(createdOrders, newOrder)
		}

		// 4. Kosongkan Keranjang
		if err := s.cartRepo.ClearCart(tx, userID); err != nil {
			return err
		}
		paymentResponse, err := s.paymentService.InitiatePayment(tx, userID, createdOrders, grandTotal)
		if err != nil {
			return err
		}

		snapResponse = paymentResponse
		return nil // Commit
	})

	return snapResponse, err
}
