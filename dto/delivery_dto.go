package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateDeliveryRequest adalah DTO untuk membuat permintaan pengiriman baru.
type CreateDeliveryRequest struct {
	PickupAddress      string  `json:"pickup_address" binding:"required"`
	PickupLat          float64 `json:"pickup_lat" binding:"required"`
	PickupLng          float64 `json:"pickup_lng" binding:"required"`
	DestinationAddress string  `json:"destination_address" binding:"required"`
	ItemDescription    string  `json:"item_description" binding:"required"`
	ItemWeight         float64 `json:"item_weight" binding:"required"`
}

// DriverRecommendationResponse adalah DTO untuk menampilkan driver yang direkomendasikan.
type DriverRecommendationResponse struct {
	DriverID      uuid.UUID `json:"driver_id"`
	DriverName    string    `json:"driver_name"`
	VehicleTypes  string    `json:"vehicle_types"`
	Rating        float64   `json:"rating"`
	Distance      float64   `json:"distance_km"` // Jarak dari lokasi pickup
}

type MyDeliveryResponse struct {
	DeliveryID         uuid.UUID `json:"delivery_id"`
	ItemDescription    string    `json:"item_description"`
	DestinationAddress string    `json:"destination_address"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"created_at"`
}