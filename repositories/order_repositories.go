package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

// OrderRepository mendefinisikan operasi database untuk Order dan OrderItem.
type OrderRepository interface {
	CreateWithItems(tx *gorm.DB, order *models.Order, cartItems []models.Cart) error
	UpdateStatusByPaymentID(tx *gorm.DB, paymentID uuid.UUID, status string) error
	FindByID(id uuid.UUID) (*models.Order, error)
	FindAllByUserID(userID uuid.UUID) ([]models.Order, error)
	FindOrdersByPaymentID(tx *gorm.DB, paymentID uuid.UUID) ([]models.Order, error)
	CountNewOrders(since time.Time) (int64, error)
}

type orderRepository struct{ db *gorm.DB }

// NewOrderRepository adalah constructor untuk repository.
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// CreateWithItems membuat Order baru beserta semua OrderItems-nya dalam satu transaksi.
func (r *orderRepository) CreateWithItems(tx *gorm.DB, order *models.Order, cartItems []models.Cart) error {
	// 1. Buat record Order utama
	if err := tx.Create(order).Error; err != nil {
		return err
	}

	// 2. Loop melalui item keranjang dan buat OrderItem untuk masing-masing
	for _, cartItem := range cartItems {
		orderItem := models.OrderItem{
			ID:              uuid.New(),
			OrderID:         order.ID,
			ProductID:       cartItem.ProductID,
			Quantity:        cartItem.Quantity,
			PriceAtPurchase: cartItem.Product.Price, // Simpan harga saat pembelian
		}
		if err := tx.Create(&orderItem).Error; err != nil {
			return err // Jika satu item gagal, seluruh transaksi akan dibatalkan
		}
	}
	return nil
}

// UpdateStatusByPaymentID memperbarui status semua Order yang terkait dengan satu ID pembayaran.
func (r *orderRepository) UpdateStatusByPaymentID(tx *gorm.DB, paymentID uuid.UUID, status string) error {
	// 1. Cari semua OrderID dari tabel penghubung (pivot table)
	var orderIDs []uuid.UUID
	if err := tx.Table("ecommerce_payment_orders").Where("e_commerce_payment_id = ?", paymentID).Pluck("order_id", &orderIDs).Error; err != nil {
		return err
	}

	if len(orderIDs) == 0 {
		return nil // Tidak ada order yang perlu diupdate, bukan error
	}

	// 2. Update status semua order yang ditemukan
	return tx.Model(&models.Order{}).Where("id IN ?", orderIDs).Update("status", status).Error
}

// FindByID mencari satu Order berdasarkan ID-nya, termasuk semua item di dalamnya.
func (r *orderRepository) FindByID(id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items.Product").Where("id = ?", id).First(&order).Error
	return &order, err
}

// FindAllByUserID mencari semua riwayat Order milik seorang pengguna.
func (r *orderRepository) FindAllByUserID(userID uuid.UUID) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.Preload("Items.Product").Where("user_id = ?", userID).Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) FindOrdersByPaymentID(tx *gorm.DB, paymentID uuid.UUID) ([]models.Order, error) {
	var orders []models.Order

	// 1. Cari semua OrderID dari tabel penghubung (pivot table)
	var orderIDs []uuid.UUID
	if err := tx.Table("ecommerce_payment_orders").Where("e_commerce_payment_id = ?", paymentID).Pluck("order_id", &orderIDs).Error; err != nil {
		return nil, err
	}

	if len(orderIDs) == 0 {
		return orders, nil // Kembalikan slice kosong jika tidak ada order
	}
	// 2. Ambil semua data Order tersebut dan preload item-itemnya
	// Preload "Items" sangat penting agar service webhook bisa mengakses 
	// produk yang dibeli untuk mengurangi stok.
	err := tx.Preload("Items").Where("id IN ?", orderIDs).Find(&orders).Error
	return orders, err
}

func (r *orderRepository) CountNewOrders(since time.Time) (int64, error) {
	var count int64
	// Kita hitung pesanan yang 'paid' dan dibuat dalam 30 hari terakhir
	err := r.db.Model(&models.Order{}).
		Where("status = ? AND created_at > ?", "paid", since).
		Count(&count).Error
	return count, err
}