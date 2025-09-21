package repositories

import (
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type DriverRouteRepository interface {
	Create(route *models.DriverRoute) error
}

type driverRouteRepository struct{ db *gorm.DB }

func NewDriverRouteRepository(db *gorm.DB) DriverRouteRepository {
	return &driverRouteRepository{db: db}
}

func (r *driverRouteRepository) Create(route *models.DriverRoute) error {
	return r.db.Create(route).Error
}