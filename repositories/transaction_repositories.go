package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type TransactionRepository interface {
    // ...
	FindByID(id string) (*models.Transaction, error)
	// Fungsi untuk menyimpan ID dari Midtrans setelah request berhasil dibuat
	SetPaymentGatewayReference(id string, refID string) error 
    UpdateStatusByOrderID(orderID string, status string, paymentMethod string) error
}

// Implementasi FindByID...

type transactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) SetPaymentGatewayReference(id string, refID string) error {
    return r.db.Model(&models.Transaction{}).Where("id = ?", id).Update("payment_gateway_reference_id", refID).Error
}

func (r *transactionRepository) FindByID(id string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.Where("id = ?", id).First(&transaction).Error
	return &transaction, err
}

func (r *transactionRepository) UpdateStatusByOrderID(orderID string, status string, paymentMethod string) error {
    return r.db.Model(&models.Transaction{}).Where("id = ?", orderID).Updates(map[string]interface{}{
        "status":         status,
        "payment_method": paymentMethod,
    }).Error
}