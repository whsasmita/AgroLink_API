package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
)

type FarmService interface {
	FindFarmByID(id string) (*models.FarmLocation, error)
	CreateFarm(input CreateFarmInput, farmerID string) (*models.FarmLocation, error)
	GetAllFarms(farmerID string) ([]models.FarmLocation, error)
	UpdateFarm(farmID string, input CreateFarmInput) (*models.FarmLocation, error)
	DeleteFarm(farmID string) error
}

// CreateFarmInput represents the input for creating a new farm
type CreateFarmInput struct {
	FarmerID       string
	Name           string  `json:"name" binding:"required"`
	Latitude       float64 `json:"latitude" binding:"required"`
	Longitude      float64 `json:"longitude" binding:"required"`
	AreaSize       float64 `json:"area_size" binding:"required"`
	CropType       *string `json:"crop_type"`
	IrrigationType *string `json:"irrigation_type"`
	Description    *string `json:"description"`
}
type farmService struct {
	FarmRepo repositories.FarmRepository
}

func NewFarmService(farmRepo repositories.FarmRepository) FarmService {
	return &farmService{
		FarmRepo: farmRepo,
	}
}

func (s *farmService) FindFarmByID(id string) (*models.FarmLocation, error) {
	farm, err := s.FarmRepo.FindByID(id)
	if err != nil {
		return nil, err // Jika ada error dari DB, teruskan.
	}

	return farm, nil

}

func (s *farmService) GetAllFarms(farmerID string) ([]models.FarmLocation, error) {
	farms, err := s.FarmRepo.FindAllByFarmerID(farmerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get farms: %w", err)
	}

	return farms, nil
}

func (s *farmService) CreateFarm(input CreateFarmInput, farmerID string) (*models.FarmLocation, error) {

	// Parse farmer ID to UUID
	farmerUUID, err := uuid.Parse(farmerID)
	if err != nil {
		return nil, fmt.Errorf("invalid farmer ID format: %w", err)
	}

	// Create new farm location
	newFarm := &models.FarmLocation{
		FarmerID:       farmerUUID,
		Name:           input.Name,
		Latitude:       input.Latitude,
		Longitude:      input.Longitude,
		AreaSize:       input.AreaSize,
		CropType:       input.CropType,
		IrrigationType: input.IrrigationType,
		Description:    input.Description,
		IsActive:       true,
		CreatedAt:      time.Now(),
	}

	// Save to database
	err = s.FarmRepo.CreateFarm(newFarm)
	if err != nil {
		return nil, fmt.Errorf("failed to create farm: %w", err)
	}

	return newFarm, nil
}

func (s *farmService) UpdateFarm(farmID string, input CreateFarmInput) (*models.FarmLocation, error) {
	// 1. Validate farm ID format
	_, err := uuid.Parse(farmID)
	if err != nil {
		// Jika ID yang diberikan tidak valid, kembalikan error.
		return nil, errors.New("invalid farm ID format")
	}

	// 2. First, check if the farm exists
	existingFarm, err := s.FarmRepo.FindByID(farmID)
	if err != nil {
		return nil, fmt.Errorf("farm not found: %w", err)
	}

	// 3. Update the existing farm with new data
	existingFarm.Name = input.Name
	existingFarm.Latitude = input.Latitude
	existingFarm.Longitude = input.Longitude
	existingFarm.AreaSize = input.AreaSize
	existingFarm.CropType = input.CropType
	existingFarm.IrrigationType = input.IrrigationType
	existingFarm.Description = input.Description
	// Keep existing FarmerID, IsActive, CreatedAt, and update UpdatedAt
	
	// 4. Save to database
	err = s.FarmRepo.Update(existingFarm)
	if err != nil {
		return nil, fmt.Errorf("failed to update farm: %w", err)
	}

	return existingFarm, nil
}

func (s *farmService) DeleteFarm(farmID string) error {
	// 1. Konversi ID dari string ke tipe uuid.UUID
	parsedID, err := uuid.Parse(farmID)
	if err != nil {
		// Jika ID yang diberikan tidak valid, kembalikan error.
		return errors.New("invalid farm ID format")
	}

	// 2. Create a farm object with the parsed ID for deletion
	farmToDelete := &models.FarmLocation{
		ID: parsedID,
	}

	// 3. Call repository delete method
	err = s.FarmRepo.Delete(farmToDelete)
	if err != nil {
		return fmt.Errorf("failed to delete farm: %w", err)
	}

	return nil
}
