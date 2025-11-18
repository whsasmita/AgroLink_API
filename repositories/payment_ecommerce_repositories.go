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

func (r *eCommercePaymentRepository) GetRevenueStats(startDate, endDate time.Time) (float64, []dto.DailyDataPoint, error) {
	// 1. Hitung Total (Hanya yang PAID)
	var total float64
	err := r.db.Model(&models.ECommercePayment{}).
		Where("status = ? AND created_at BETWEEN ? AND ?", "paid", startDate, endDate).
		Select("COALESCE(SUM(grand_total), 0)").
		Scan(&total).Error
	if err != nil {
		return 0, nil, err
	}

	// 2. Hitung Tren Harian
	var trend []dto.DailyDataPoint
	err = r.db.Model(&models.ECommercePayment{}).
		Select("DATE(created_at) as date, SUM(grand_total) as value").
		Where("status = ? AND created_at BETWEEN ? AND ?", "paid", startDate, endDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&trend).Error

	return total, trend, err
}

func (r *eCommercePaymentRepository) GetAllPaymentsNoPaging() ([]models.ECommercePayment, error) {
	var payments []models.ECommercePayment
	err := r.db.Preload("User").
		Order("created_at DESC").
		Find(&payments).Error
	return payments, err
}