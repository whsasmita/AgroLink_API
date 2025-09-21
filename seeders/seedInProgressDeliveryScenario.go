package seeders

import (
	"fmt"
	"log"

	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

var DB *gorm.DB

func SeedInProgressDeliveryScenario() {
	log.Println("Creating an in-progress delivery scenario for tracking test...")

	// Gunakan transaksi agar semua data dibuat atau tidak sama sekali
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 1. Ambil data Petani dan Driver yang sudah ada
		var farmerUser, driverUser models.User
		if err := tx.Preload("Farmer").Where("email = ?", "farmer1@agrolink.com").First(&farmerUser).Error; err != nil {
			return fmt.Errorf("seeder failed: could not find farmer")
		}
		if err := tx.Preload("Driver").Where("email = ?", "driver1@agrolink.com").First(&driverUser).Error; err != nil {
			return fmt.Errorf("seeder failed: could not find driver")
		}

		// 2. Buat Kontrak terlebih dahulu
		contract := models.Contract{
			ContractType:   "delivery",
			FarmerID:       farmerUser.Farmer.UserID,
			DriverID:       &driverUser.Driver.UserID,
			SignedByFarmer: true,
			SignedBySecondParty: true, // Asumsikan driver langsung setuju
			Status:         "active",
		}
		if err := tx.Create(&contract).Error; err != nil {
			return err
		}

		// 3. Buat Delivery dengan status "in_transit" dan hubungkan ke kontrak
		delivery := models.Delivery{
			FarmerID:           farmerUser.Farmer.UserID,
			DriverID:           &driverUser.Driver.UserID,
			ContractID:         &contract.ID,
			PickupAddress:      "Bedugul, Bali",
			PickupLat:          -8.275,
			PickupLng:          115.163,
			DestinationAddress: "Canggu, Bali",
			ItemDescription:    "100kg Stroberi Segar",
			ItemWeight:         100.0,
			Status:             "in_transit", // Langsung set status in_transit
		}
		if err := tx.Create(&delivery).Error; err != nil {
			return err
		}

		// 4. Buat Invoice yang lunas
		invoice := models.Invoice{
			DeliveryID:  &delivery.ID,
			FarmerID:    farmerUser.Farmer.UserID,
			Amount:      200000,
			PlatformFee: 10000,
			TotalAmount: 210000,
			Status:      "paid",
		}
		if err := tx.Create(&invoice).Error; err != nil {
			return err
		}

		// 5. Buat Transaction sebagai bukti pembayaran
		transaction := models.Transaction{
			InvoiceID:     invoice.ID,
			AmountPaid:    invoice.TotalAmount,
			PaymentMethod: StringPtr("gopay"),
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		return nil // Commit transaksi
	})

	if err != nil {
		log.Fatalf("Failed to seed in-progress delivery scenario: %v", err)
	}
}

func StringPtr(s string) *string {
	return &s
}