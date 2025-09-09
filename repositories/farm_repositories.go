package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type FarmRepository interface {
	FindByID(id string) (*models.FarmLocation, error)
	FindAllByFarmerID(farmerID string) ([]models.FarmLocation, error)
	FindByIDAndFarmerID(id string, farmerID uuid.UUID) (*models.FarmLocation, error)
	CreateFarm(farm *models.FarmLocation) error
	Update(farm *models.FarmLocation) error
	Delete(farmID *models.FarmLocation) error
}

type farmRepository struct {
	db *gorm.DB
}

func NewFarmRepository(db *gorm.DB) FarmRepository {
	return &farmRepository{db}
}

func (r *farmRepository) FindByID(id string) (*models.FarmLocation, error) {
	var farm models.FarmLocation
	err := r.db.
		Where("id = ?", id).First(&farm).Error
	return &farm, err
}

func (r *farmRepository) FindAllByFarmerID(farmerID string) ([]models.FarmLocation, error) {
	var farms []models.FarmLocation
	err := r.db.
		Where("farmer_id = ? AND is_active = ?", farmerID, true).
		Order("created_at DESC").
		Find(&farms).Error
	return farms, err
}

func (r *farmRepository) CreateFarm(farm *models.FarmLocation) error {
	return r.db.Create(farm).Error
}

func (r *farmRepository) Update(farm *models.FarmLocation) error {
	return r.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(farm).Error
}

func (r *farmRepository) Delete(farm *models.FarmLocation) error {
	return r.db.Delete(&models.FarmLocation{}, "id = ?", farm.ID).Error
}

func (r *farmRepository) FindByIDAndFarmerID(id string, farmerID uuid.UUID) (*models.FarmLocation, error) {
    var farm models.FarmLocation
    err := r.db.Where("id = ? AND farmer_id = ?", id, farmerID).First(&farm).Error
    return &farm, err
}