package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type LocationTrackRepository interface {
	Create(track *models.LocationTrack) error
	FindLatestByDeliveryID(deliveryID string) (*models.LocationTrack, error)
}

type locationTrackRepository struct{ db *gorm.DB }

func NewLocationTrackRepository(db *gorm.DB) LocationTrackRepository {
	return &locationTrackRepository{db: db}
}

func (r *locationTrackRepository) Create(track *models.LocationTrack) error {
	return r.db.Create(track).Error
}

func (r *locationTrackRepository) FindLatestByDeliveryID(deliveryID string) (*models.LocationTrack, error) {
	var track models.LocationTrack
	err := r.db.Where("delivery_id = ?", deliveryID).Order("timestamp DESC").First(&track).Error
	return &track, err
}