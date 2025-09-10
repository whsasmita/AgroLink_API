package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

// [DIREFACTOR] Repository ini sekarang hanya untuk mencatat bukti pembayaran
type TransactionRepository interface {
	Create(tx *models.Transaction) error
	FindByInvoiceID(invoiceID string) (*models.Transaction, error)
}

type transactionRepository struct{ db *gorm.DB }

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(tx *models.Transaction) error {
	return r.db.Create(tx).Error
}

func (r *transactionRepository) FindByInvoiceID(invoiceID string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := r.db.Where("invoice_id = ?", invoiceID).First(&transaction).Error
	return &transaction, err
}