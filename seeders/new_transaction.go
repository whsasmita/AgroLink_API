package seeders

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

const newTransactionSeedJSONPath = "seeders/new_transaction.json"

type NewTransactionSeedRow struct {
	IDTransaksi        string   `json:"IDTransaksi"`
	Tanggal            string   `json:"Tanggal"`
	Timestamp          string   `json:"Timestamp"`
	Layanan            string   `json:"Layanan"`
	Keterangan         string   `json:"Keterangan"`
	MetodePembayaran   string   `json:"MetodePembayaran"`
	NominalTransaksi   *float64 `json:"NominalTransaksi"`
	PersentaseKomisi   *float64 `json:"PersentaseKomisi"`
	KeuntunganKotor    *float64 `json:"KeuntunganKotor"`
	BiayaMidtrans      *float64 `json:"BiayaMidtrans"`
	KeuntunganBersih   *float64 `json:"KeuntunganBersih"`
	TotalDiterimaMitra *float64 `json:"TotalDiterimaMitra"`

	FarmerEmail string `json:"FarmerEmail"`
	WorkerEmail string `json:"WorkerEmail"`
	DriverEmail string `json:"DriverEmail"`
	MitraEmail  string `json:"MitraEmail"`
	BuyerEmail  string `json:"BuyerEmail"`
}

func SeedNewTransactions(db *gorm.DB) {
	log.Println("🌱 Seeding invoices & transactions from new_transaction.json...")

	data, err := os.ReadFile(newTransactionSeedJSONPath)
	if err != nil {
		log.Printf("Failed to open seed file %s: %v", newTransactionSeedJSONPath, err)
		return
	}

	var rows []NewTransactionSeedRow
	if err := json.Unmarshal(data, &rows); err != nil {
		log.Printf("Failed to parse new transaction seed JSON: %v", err)
		return
	}

	// Load all users to resolve email -> UUID
	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		log.Printf("Failed to load users for transaction seeding: %v", err)
		return
	}

	userMap := make(map[string]uuid.UUID)
	var firstFarmerID uuid.UUID
	for _, u := range users {
		userMap[strings.ToLower(strings.TrimSpace(u.Email))] = u.ID
		if u.Role == "farmer" && firstFarmerID == uuid.Nil {
			firstFarmerID = u.ID
		}
	}

	var totalSeeded int
	var totalGMV float64
	var totalNetProfit float64

	for _, row := range rows {
		if err := seedSingleNewTransaction(db, row, userMap, firstFarmerID); err != nil {
			log.Printf("Failed to seed transaction %s: %v", row.IDTransaksi, err)
		} else {
			totalSeeded++
			if row.NominalTransaksi != nil {
				totalGMV += *row.NominalTransaksi
			}
			if row.KeuntunganBersih != nil {
				totalNetProfit += *row.KeuntunganBersih
			}
		}
	}

	log.Printf("✅ Seeding transaksi selesai: %d record diproses. Total GMV: Rp %.0f, Total Keuntungan Bersih: Rp %.0f",
		totalSeeded, totalGMV, totalNetProfit)
}

func seedSingleNewTransaction(db *gorm.DB, row NewTransactionSeedRow, userMap map[string]uuid.UUID, fallbackFarmerID uuid.UUID) error {
	if row.NominalTransaksi == nil || row.KeuntunganKotor == nil || row.TotalDiterimaMitra == nil {
		return fmt.Errorf("incomplete monetary fields")
	}

	var txnDate time.Time
	var err error
	if strings.TrimSpace(row.Timestamp) != "" {
		txnDate, err = time.Parse("2006-01-02 15:04:05", strings.TrimSpace(row.Timestamp))
	}
	if err != nil || txnDate.IsZero() {
		txnDate, err = parseDateFlexible(row.Tanggal)
		if err != nil || txnDate.IsZero() {
			return fmt.Errorf("invalid date %q: %w", row.Tanggal, err)
		}
	}

	refID := strings.TrimSpace(row.IDTransaksi)
	if refID == "" {
		return fmt.Errorf("empty transaction reference")
	}

	// Resolve farmer or payer ID
	payerID := fallbackFarmerID
	if row.FarmerEmail != "" {
		if id, ok := userMap[strings.ToLower(strings.TrimSpace(row.FarmerEmail))]; ok {
			payerID = id
		}
	} else if row.BuyerEmail != "" {
		if id, ok := userMap[strings.ToLower(strings.TrimSpace(row.BuyerEmail))]; ok {
			payerID = id
		}
	}

	amount := *row.TotalDiterimaMitra
	grossProfit := *row.KeuntunganKotor
	totalAmount := *row.NominalTransaksi
	netProfit := grossProfit
	if row.KeuntunganBersih != nil {
		netProfit = *row.KeuntunganBersih
	}
	gatewayFee := 0.0
	if row.BiayaMidtrans != nil {
		gatewayFee = *row.BiayaMidtrans
	}

	paymentMethod := strings.TrimSpace(row.MetodePembayaran)
	refIDValue := refID

	if row.Layanan == "E-Commerce" {
		var existing models.ECommercePayment
		if err := db.Where("snap_token = ?", "SEED-"+refIDValue).First(&existing).Error; err == nil {
			return nil
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("check existing ecommerce payment: %w", err)
		}

		return db.Transaction(func(tx *gorm.DB) error {
			invoice := models.Invoice{
				FarmerID:    payerID,
				Amount:      amount,
				PlatformFee: grossProfit,
				TotalAmount: totalAmount,
				Status:      "paid",
				DueDate:     txnDate,
				CreatedAt:   txnDate.AddDate(0, 0, -1),
				UpdatedAt:   txnDate,
			}
			if err := tx.Create(&invoice).Error; err != nil {
				return fmt.Errorf("create invoice: %w", err)
			}

			payment := models.ECommercePayment{
				ID:         uuid.New(),
				UserID:     payerID,
				GrandTotal: totalAmount,
				Status:     "paid",
				SnapToken:  "SEED-" + refIDValue,
				CreatedAt:  txnDate,
				UpdatedAt:  txnDate,
			}
			if err := tx.Create(&payment).Error; err != nil {
				return fmt.Errorf("create ecommerce payment: %w", err)
			}

			profit := models.PlatformProfit{
				ECommercePaymentID: &payment.ID,
				SourceType:         "ecommerce",
				GrossProfit:        grossProfit,
				GatewayFee:         gatewayFee,
				NetProfit:          netProfit,
				ProfitDate:         txnDate,
			}
			if err := tx.Create(&profit).Error; err != nil {
				return fmt.Errorf("create platform profit: %w", err)
			}

			return nil
		})
	}

	var existing models.Transaction
	if err := db.Where("payment_gateway_reference_id = ?", refID).First(&existing).Error; err == nil {
		return nil
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("check existing transaction: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		invoice := models.Invoice{
			FarmerID:    payerID,
			Amount:      amount,
			PlatformFee: grossProfit,
			TotalAmount: totalAmount,
			Status:      "paid",
			DueDate:     txnDate,
			CreatedAt:   txnDate.AddDate(0, 0, -1),
			UpdatedAt:   txnDate,
		}
		if err := tx.Create(&invoice).Error; err != nil {
			return fmt.Errorf("create invoice: %w", err)
		}

		transaction := models.Transaction{
			InvoiceID:                 invoice.ID,
			PaymentGateway:            "midtrans",
			PaymentGatewayReferenceID: &refIDValue,
			AmountPaid:                totalAmount,
			PaymentMethod:             &paymentMethod,
			TransactionDate:           txnDate,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}

		profit := models.PlatformProfit{
			TransactionID: &transaction.ID,
			SourceType:    "utama",
			GrossProfit:   grossProfit,
			GatewayFee:    gatewayFee,
			NetProfit:     netProfit,
			ProfitDate:    txnDate,
		}
		if err := tx.Create(&profit).Error; err != nil {
			return fmt.Errorf("create platform profit: %w", err)
		}

		return nil
	})
}
