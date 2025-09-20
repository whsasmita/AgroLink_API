package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

type DeliveryService interface {
	CreateDelivery(input dto.CreateDeliveryRequest, farmerID uuid.UUID) (*models.Delivery, error)
	FindAvailableDrivers(deliveryID string, farmerID uuid.UUID, radius int) ([]dto.DriverRecommendationResponse, error)
	SelectDriver(deliveryID, driverID, farmerID string) (*models.Contract, error)
}

type deliveryService struct {
	deliveryRepo repositories.DeliveryRepository
	driverRepo   repositories.DriverRepository
	contractRepo repositories.ContractRepository
	db           *gorm.DB // Diperlukan untuk transaksi
}

// [PERUBAHAN] Tambahkan 'db' ke konstruktor
func NewDeliveryService(
	deliveryRepo repositories.DeliveryRepository,
	driverRepo repositories.DriverRepository,
	contractRepo repositories.ContractRepository,
	db *gorm.DB,
) DeliveryService {
	return &deliveryService{
		deliveryRepo: deliveryRepo,
		driverRepo:   driverRepo,
		contractRepo: contractRepo,
		db:           db,
	}
}

// CreateDelivery membuat permintaan pengiriman baru dari petani.
func (s *deliveryService) CreateDelivery(input dto.CreateDeliveryRequest, farmerID uuid.UUID) (*models.Delivery, error) {
	newDelivery := &models.Delivery{
		FarmerID:           farmerID,
		PickupAddress:      input.PickupAddress,
		PickupLat:          input.PickupLat,
		PickupLng:          input.PickupLng,
		DestinationAddress: input.DestinationAddress,
		ItemDescription:    input.ItemDescription,
		ItemWeight:         input.ItemWeight,
		Status:             "pending_driver",
	}

	if err := s.deliveryRepo.Create(newDelivery); err != nil {
		return nil, fmt.Errorf("failed to create delivery request: %w", err)
	}
	return newDelivery, nil
}

// FindAvailableDrivers adalah logika inti untuk mencari driver yang cocok.
func (s *deliveryService) FindAvailableDrivers(deliveryID string, farmerID uuid.UUID, radius int) ([]dto.DriverRecommendationResponse, error) {
	delivery, err := s.deliveryRepo.FindByID(deliveryID)
	if err != nil {
		return nil, fmt.Errorf("delivery request not found")
	}

	if delivery.FarmerID != farmerID {
		return nil, fmt.Errorf("forbidden: you do not own this delivery request")
	}

	// 1. Cari driver terdekat menggunakan Haversine
	nearbyDrivers, err := s.driverRepo.FindNearby(delivery.PickupLat, delivery.PickupLng, radius)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby drivers: %w", err)
	}

	// 2. TODO: Filter lebih lanjut berdasarkan rute yang cocok (untuk V2)
	// Untuk saat ini, kita tampilkan semua yang terdekat.

	// 3. Transformasi ke DTO
	var recommendations []dto.DriverRecommendationResponse
	for _, driver := range nearbyDrivers {
		recommendations = append(recommendations, dto.DriverRecommendationResponse{
			DriverID:     driver.UserID,
			DriverName:   driver.User.Name,
			VehicleTypes: driver.VehicleTypes,
			Rating:       driver.Rating,
			Distance:     driver.Distance,
		})
	}
	return recommendations, nil
}

// SelectDriver memilih driver dan membuatkan kontrak untuk pengiriman.
func (s *deliveryService) SelectDriver(deliveryID, driverID, farmerID string) (*models.Contract, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Ambil data delivery & validasi
	delivery, err := s.deliveryRepo.FindByID(deliveryID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("delivery not found")
	}
	if delivery.FarmerID.String() != farmerID {
		tx.Rollback()
		return nil, errors.New("forbidden: you do not own this delivery")
	}
	if delivery.Status != "pending_driver" {
		tx.Rollback()
		return nil, errors.New("this delivery is no longer waiting for a driver")
	}

	driverUUID, _ := uuid.Parse(driverID)
	farmerUUID, _ := uuid.Parse(farmerID)

	// 2. Buat Kontrak baru dengan tipe 'delivery'
	newContract := &models.Contract{
		ContractType:   "delivery",
		DeliveryID:     &delivery.ID,
		FarmerID:       farmerUUID,
		DriverID:       &driverUUID,
		Status:         "pending_signature",
		SignedByFarmer: true,
	}
	if err := s.contractRepo.Create(tx, newContract); err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}

	// 3. Update status Delivery
	delivery.DriverID = &driverUUID
	delivery.ContractID = &newContract.ID
	delivery.Status = "pending_signature"
	if err := s.deliveryRepo.Update(tx, delivery); err != nil { // Asumsi repo.Update menerima 'tx'
		tx.Rollback()
		return nil, fmt.Errorf("failed to update delivery status: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", err)
	}

	return newContract, nil
}
