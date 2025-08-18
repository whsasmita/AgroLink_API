// File: dtos/driver_response.go
package dto

import (
	"time"

	"github.com/google/uuid"
)

// DriverResponse merepresentasikan format respons API untuk data driver/ekspedisi
type DriverResponse struct {
	UserID          uuid.UUID `json:"user_id"`
	Name            string    `json:"name"`
	ProfilePicture  *string   `json:"profile_picture"`
	PhoneNumber     *string   `json:"phone_number"`
	CompanyAddress  *string   `json:"company_address"`
	PricingScheme   string    `json:"pricing_scheme"`
	VehicleTypes    string    `json:"vehicle_types"`
	Rating          float64   `json:"rating"`
	TotalDeliveries int       `json:"total_deliveries"`
	CreatedAt       time.Time `json:"created_at"`
}
