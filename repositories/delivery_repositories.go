package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type DeliveryRepository interface {
	Create(delivery *models.Delivery) error
	FindByID(id string) (*models.Delivery, error)
	// [PERBAIKAN] Tambahkan *gorm.DB sebagai argumen
	Update(tx *gorm.DB, delivery *models.Delivery) error
}

type deliveryRepository struct{ db *gorm.DB }

func NewDeliveryRepository(db *gorm.DB) DeliveryRepository {
	return &deliveryRepository{db: db}
}

func (r *deliveryRepository) Create(delivery *models.Delivery) error {
	return r.db.Create(delivery).Error
}

func (r *deliveryRepository) FindByID(id string) (*models.Delivery, error) {
	var delivery models.Delivery
	err := r.db.Where("id = ?", id).First(&delivery).Error
	return &delivery, err
}

// [PERBAIKAN] Ubah fungsi untuk menerima dan menggunakan objek transaksi
func (r *deliveryRepository) Update(tx *gorm.DB, delivery *models.Delivery) error {
	// Gunakan 'tx' yang dioper dari service, bukan 'r.db'
	return tx.Save(delivery).Error
}