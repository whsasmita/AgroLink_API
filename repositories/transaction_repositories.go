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
	GetAllTransactions(page, limit int) ([]models.Transaction, int64, error)
	GetRevenueStats(startDate, endDate time.Time) (total float64, trend []dto.DailyDataPoint, err error)
	GetAllTransactionsNoPaging() ([]models.Transaction, error)
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

func (r *transactionRepository) GetAllTransactions(page, limit int) ([]models.Transaction, int64, error) {
    var transactions []models.Transaction
    var total int64
    offset := (page - 1) * limit

    // Ambil data dengan Preload
    err := r.db.Preload("Invoice.Farmer.User"). // Untuk nama pembayar (Petani)
        Preload("Invoice.Project").
        Preload("Invoice.Delivery").
        Order("transaction_date DESC").
        Offset(offset).Limit(limit).
        Find(&transactions).Error
    if err != nil {
        return nil, 0, err
    }
    // Hitung total item (tanpa offset/limit)
    r.db.Model(&models.Transaction{}).Count(&total)
    
    return transactions, total, nil
}

func (r *transactionRepository) GetRevenueStats(
    start, end time.Time,
) (float64, []dto.DailyDataPoint, error) {

    var total float64
    var rows []struct {
        Date  time.Time
        Value float64
    }

    // TOTAL OMZET JASA (service)
    // Jika mau pastikan hanya transaksi utama, join ke platform_profits + source_type='utama'
    if err := r.db.
        Model(&models.Transaction{}).
        Joins("JOIN platform_profits pp ON pp.transaction_id = transactions.id").
        Where("pp.source_type = ?", "utama").
        Where("transactions.transaction_date BETWEEN ? AND ?", start, end).
        Select("COALESCE(SUM(transactions.amount_paid), 0)").
        Scan(&total).Error; err != nil {
        return 0, nil, err
    }

    // TREND HARIAN
    if err := r.db.
        Model(&models.Transaction{}).
        Joins("JOIN platform_profits pp ON pp.transaction_id = transactions.id").
        Where("pp.source_type = ?", "utama").
        Where("transactions.transaction_date BETWEEN ? AND ?", start, end).
        Select(`
            DATE(transactions.transaction_date) AS date,
            COALESCE(SUM(transactions.amount_paid), 0) AS value`,
        ).
        Group("DATE(transactions.transaction_date)").
        Order("DATE(transactions.transaction_date)").
        Scan(&rows).Error; err != nil {
        return 0, nil, err
    }

    trend := make([]dto.DailyDataPoint, 0, len(rows))
    for _, r := range rows {
        trend = append(trend, dto.DailyDataPoint{
            Date:  r.Date.Format("2006-01-02"),
            Value: r.Value,
        })
    }

    return total, trend, nil
}



func (r *transactionRepository) GetAllTransactionsNoPaging() ([]models.Transaction, error) {
	var transactions []models.Transaction
	// Preload sama seperti sebelumnya
	err := r.db.
		Preload("Invoice.Project").
		Preload("Invoice.Delivery").
		Preload("Invoice.Farmer.User").
		Order("transaction_date DESC").
		Find(&transactions).Error
	return transactions, err
}