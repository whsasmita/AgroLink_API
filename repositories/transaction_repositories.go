package repositories

import (
	"time"

	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

// [DIREFACTOR] Repository ini sekarang hanya untuk mencatat bukti pembayaran
type TransactionRepository interface {
	Create(tx *models.Transaction) error
	FindByInvoiceID(invoiceID string) (*models.Transaction, error)
	GetTotalRevenue(since time.Time) (float64, error)
	GetDailyRevenueTrend(since time.Time) ([]dto.DailyDataPoint, error)
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

// GetTotalRevenue menghitung total pendapatan dari transaksi yang berhasil.
func (r *transactionRepository) GetTotalRevenue(since time.Time) (float64, error) {
	var totalRevenue float64
	// Asumsi 'transaction_date' diisi saat transaksi dibuat
	err := r.db.Model(&models.Transaction{}).
		Where("transaction_date > ?", since).
		Pluck("SUM(amount_paid)", &totalRevenue).Error
	return totalRevenue, err
}

// GetDailyRevenueTrend menghitung total pendapatan per hari.
func (r *transactionRepository) GetDailyRevenueTrend(since time.Time) ([]dto.DailyDataPoint, error) {
	var results []dto.DailyDataPoint
	err := r.db.Model(&models.Transaction{}).
		Select("DATE(transaction_date) as date, SUM(amount_paid) as value").
		Where("transaction_date > ?", since).
		Group("DATE(transaction_date)").
		Order("date ASC").
		Scan(&results).Error
	return results, err
}