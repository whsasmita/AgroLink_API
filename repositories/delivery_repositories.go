package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type DeliveryRepository interface {
	Create(delivery *models.Delivery) error
	FindByID(id string) (*models.Delivery, error)
	// [PERBAIKAN] Tambahkan *gorm.DB sebagai argumen
	Update(tx *gorm.DB, delivery *models.Delivery) error
	FindByContractID(contractID string) (*models.Delivery, error)
	FindAllByUserID(userID uuid.UUID, role string) ([]models.Delivery, error)
	CountActiveDeliveries() (int64, error)

}

type deliveryRepository struct{ db *gorm.DB }

func NewDeliveryRepository(db *gorm.DB) DeliveryRepository {
	return &deliveryRepository{db: db}
}

func (r *deliveryRepository) FindByContractID(contractID string) (*models.Delivery, error) {
	var delivery models.Delivery
	err := r.db.Where("contract_id = ?", contractID).First(&delivery).Error
	return &delivery, err
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

func (r *deliveryRepository) FindAllByUserID(userID uuid.UUID, role string) ([]models.Delivery, error) {
	var deliveries []models.Delivery
	query := r.db

	if role == "farmer" {
		query = query.Where("farmer_id = ?", userID)
	} else if role == "driver" {
		query = query.Where("driver_id = ?", userID)
	} else {
		// Jika peran tidak sesuai, kembalikan daftar kosong
		return deliveries, nil
	}

	err := query.Order("created_at DESC").Find(&deliveries).Error
	return deliveries, err
}

func (r *deliveryRepository) CountActiveDeliveries() (int64, error) {
	var count int64
	err := r.db.Model(&models.Delivery{}).
		Where("status = ?", "in_transit").
		Count(&count).Error
	return count, err
}