package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
)

type TrackingService interface {
	UpdateLocation(deliveryID, driverID uuid.UUID, lat, lng float64) error
	GetLatestLocation(deliveryID string, farmerID uuid.UUID) (*models.LocationTrack, error)
}

type trackingService struct {
	trackRepo    repositories.LocationTrackRepository
	deliveryRepo repositories.DeliveryRepository
}

func NewTrackingService(trackRepo repositories.LocationTrackRepository, deliveryRepo repositories.DeliveryRepository) TrackingService {
	return &trackingService{
		trackRepo:    trackRepo,
		deliveryRepo: deliveryRepo,
	}
}

// UpdateLocation membuat catatan lokasi baru untuk sebuah pengiriman.
func (s *trackingService) UpdateLocation(deliveryID, driverID uuid.UUID, lat, lng float64) error {
	// 1. Validasi: Pastikan pengiriman ada dan driver-nya benar.
	delivery, err := s.deliveryRepo.FindByID(deliveryID.String())
	if err != nil {
		return fmt.Errorf("delivery not found")
	}
	if delivery.DriverID == nil || *delivery.DriverID != driverID {
		return fmt.Errorf("forbidden: you are not assigned to this delivery")
	}
	if delivery.Status != "in_transit" {
		return fmt.Errorf("location can only be updated for deliveries that are in transit")
	}

	// 2. Buat catatan lokasi baru
	newTrack := &models.LocationTrack{
		DeliveryID: deliveryID,
		Lat:        lat,
		Lng:        lng,
	}
	return s.trackRepo.Create(newTrack)
}

// GetLatestLocation mengambil data lokasi terakhir dari sebuah pengiriman.
func (s *trackingService) GetLatestLocation(deliveryID string, farmerID uuid.UUID) (*models.LocationTrack, error) {
	// 1. Validasi: Pastikan petani adalah pemilik pengiriman.
	delivery, err := s.deliveryRepo.FindByID(deliveryID)
	if err != nil {
		return nil, fmt.Errorf("delivery not found")
	}
	if delivery.FarmerID != farmerID {
		return nil, fmt.Errorf("forbidden: you do not own this delivery")
	}

	// 2. Ambil data lokasi terakhir
	return s.trackRepo.FindLatestByDeliveryID(deliveryID)
}