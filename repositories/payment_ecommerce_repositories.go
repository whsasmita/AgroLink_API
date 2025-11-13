package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

// ECommercePaymentRepository mendefinisikan operasi database untuk pembayaran e-commerce.
type ECommercePaymentRepository interface {
	Create(tx *gorm.DB, payment *models.ECommercePayment) error
	Update(tx *gorm.DB, payment *models.ECommercePayment) error
	FindByID(id string) (*models.ECommercePayment, error)
	UpdateStatus(tx *gorm.DB, id string, status string) error
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