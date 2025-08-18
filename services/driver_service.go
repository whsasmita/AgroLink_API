// File: services/driver_service.go
package services

import (
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/repositories"
)

type DriverService interface {
	GetDrivers(sortBy, order string, limit, offset int) ([]dto.DriverResponse, int64, error)
	GetDriverProfile(id string) (dto.DriverResponse, error)
}

type driverService struct {
	repo repositories.DriverRepository
}

func NewDriverService(repo repositories.DriverRepository) DriverService {
	return &driverService{repo}
}

func (s *driverService) GetDrivers(sortBy, order string, limit, offset int) ([]dto.DriverResponse, int64, error) {
	drivers, total, err := s.repo.GetDrivers(sortBy, order, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var driverResponses []dto.DriverResponse
	for _, driver := range drivers {
		// Lakukan mapping dari models.Driver ke dtos.DriverResponse
		driverResponses = append(driverResponses, dto.DriverResponse{
			UserID:           driver.UserID,
			CompanyAddress:   driver.Address,
			PricingScheme:    driver.PricingScheme,
			VehicleTypes:     driver.VehicleTypes,
			Rating:           driver.Rating,
			TotalDeliveries:  driver.TotalDeliveries,
			CreatedAt:        driver.CreatedAt,
			Name:             driver.User.Name,
			ProfilePicture:   driver.User.ProfilePicture,
			PhoneNumber:      driver.User.PhoneNumber,
		})
	}
	return driverResponses, total, nil
}

func (s *driverService) GetDriverProfile(id string) (dto.DriverResponse, error) {
	driver, err := s.repo.GetDriverByID(id)
	if err != nil {
		return dto.DriverResponse{}, err
	}

	// Lakukan mapping dari models.Driver ke dtos.DriverResponse
	response := dto.DriverResponse{
		UserID:           driver.UserID,
		CompanyAddress:   driver.Address,
		PricingScheme:    driver.PricingScheme,
		VehicleTypes:     driver.VehicleTypes,
		Rating:           driver.Rating,
		TotalDeliveries:  driver.TotalDeliveries,
		CreatedAt:        driver.CreatedAt,
		Name:             driver.User.Name,
		ProfilePicture:   driver.User.ProfilePicture,
		PhoneNumber:      driver.User.PhoneNumber,
	}

	return response, nil
}
