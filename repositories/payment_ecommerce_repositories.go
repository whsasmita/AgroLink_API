package repositories

import (
	"time"

	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

// ECommercePaymentRepository mendefinisikan operasi database untuk pembayaran e-commerce.
type ECommercePaymentRepository interface {
	Create(tx *gorm.DB, payment *models.ECommercePayment) error
	Update(tx *gorm.DB, payment *models.ECommercePayment) error
	FindByID(id string) (*models.ECommercePayment, error)
	UpdateStatus(tx *gorm.DB, id string, status string) error
	GetAllPayments(page, limit int) ([]models.ECommercePayment, int64, error)
	GetRevenueStats(startDate, endDate time.Time) (total float64, trend []dto.DailyDataPoint, err error)
	GetAllPaymentsNoPaging() ([]models.ECommercePayment, error)
}

type eCommercePaymentRepository struct {
	db *gorm.DB
}

// NewECommercePaymentRepository adalah constructor untuk repository.
func NewECommercePaymentRepository(db *gorm.DB) ECommercePaymentRepository {
	return &eCommercePaymentRepository{db: db}
}

// Create membuat record pembayaran baru di dalam sebuah transaksi.
func (r *eCommercePaymentRepository) Create(tx *gorm.DB, payment *models.ECommercePayment) error {
	return tx.Create(payment).Error
}

// Update menyimpan perubahan pada record pembayaran di dalam sebuah transaksi.
func (r *eCommercePaymentRepository) Update(tx *gorm.DB, payment *models.ECommercePayment) error {
	return tx.Save(payment).Error
}

// FindByID mencari record pembayaran berdasarkan ID-nya.
func (r *eCommercePaymentRepository) FindByID(id string) (*models.ECommercePayment, error) {
	var payment models.ECommercePayment
	err := r.db.Where("id = ?", id).First(&payment).Error
	return &payment, err
}

// UpdateStatus memperbarui kolom status dari sebuah record pembayaran.
func (r *eCommercePaymentRepository) UpdateStatus(tx *gorm.DB, id string, status string) error {
	return tx.Model(&models.ECommercePayment{}).Where("id = ?", id).Update("status", status).Error
}

func (r *eCommercePaymentRepository) GetAllPayments(page, limit int) ([]models.ECommercePayment, int64, error) {
    var payments []models.ECommercePayment
    var total int64
    offset := (page - 1) * limit

    // Ambil data (perlu Preload User untuk nama pembeli)
    err := r.db.Preload("User").
        Order("created_at DESC").
        Offset(offset).Limit(limit).
        Find(&payments).Error
    if err != nil {
        return nil, 0, err
    }
    
    r.db.Model(&models.ECommercePayment{}).Count(&total)

    return payments, total, nil
}

func (r *eCommercePaymentRepository) GetRevenueStats(
    start, end time.Time,
) (float64, []dto.DailyDataPoint, error) {

    var total float64
    var rows []struct {
        Date  time.Time
        Value float64
    }

    // TOTAL OMZET PRODUK (e-commerce) → dari GrandTotal
    if err := r.db.
        Model(&models.ECommercePayment{}).
        Where("created_at BETWEEN ? AND ?", start, end).
        Select("COALESCE(SUM(grand_total), 0)").
        Scan(&total).Error; err != nil {
        return 0, nil, err
    }

    // TREND HARIAN
    if err := r.db.
        Model(&models.ECommercePayment{}).
        Where("created_at BETWEEN ? AND ?", start, end).
        Select(`
            DATE(created_at) AS date,
            COALESCE(SUM(grand_total), 0) AS value`,
        ).
        Group("DATE(created_at)").
        Order("DATE(created_at)").
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


func (r *eCommercePaymentRepository) GetAllPaymentsNoPaging() ([]models.ECommercePayment, error) {
	var payments []models.ECommercePayment
	err := r.db.Preload("User").
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}