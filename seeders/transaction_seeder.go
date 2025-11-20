package seeders

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

const ecommerceSeedJSONPath = "seeders/transactions_seed_ecommerce.json"
const transactionSeedJSONPath = "seeders/transactions_seed_utama.json"

type SeedTransaksiUtamaRow struct {
	IDTransaksi      string   `json:"IDTransaksi"`
	Tanggal          string   `json:"Tanggal"`
	JenisTransaksi   string   `json:"JenisTransaksi"`
	PengirimPetani   string   `json:"PengirimPetani"`
	TotalBayarPetani *float64 `json:"TotalBayarPetani"`
	PenerimaPekerja  string   `json:"PenerimaPekerja"`
	TotalDiterima    *float64 `json:"TotalDiterima"`
	KeuntunganKotor  *float64 `json:"KeuntunganKotor"`
	MetodePembayaran string   `json:"MetodePembayaran"`
	BiayaMidtrans    *float64 `json:"BiayaMidtrans"`
	KeuntunganBersih *float64 `json:"KeuntunganBersih"`
}

type SeedEcommerceRow struct {
	IDTransaksi          string   `json:"IDTransaksi"`
	Tanggal              string   `json:"Tanggal"`
	Pembeli              string   `json:"Pembeli"`
	Produk               string   `json:"Produk"`
	Harga                *float64 `json:"Harga"`
	Ongkir               *float64 `json:"Ongkir"`
	KeuntunganKotor      *float64 `json:"KeuntunganKotor"`
	Penjual              string   `json:"Penjual"`
	TotalDiterimaPenjual *float64 `json:"TotalDiterimaPenjual"`
}

func findFarmerIDByName(db *gorm.DB, name string) (uuid.UUID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return uuid.Nil, gorm.ErrRecordNotFound
	}

	var user models.User
	if err := db.Where("name = ? AND role = ?", name, "farmer").First(&user).Error; err != nil {
		return uuid.Nil, err
	}

	return user.ID, nil
}

func parseDateYMD(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse("2006-01-02", s)
}

func SeedTransactionsAndInvoices(db *gorm.DB) {
	log.Println("Seeding invoices & transactions (Transaksi Utama) from JSON...")

	data, err := os.ReadFile(transactionSeedJSONPath)
	if err != nil {
		log.Printf("Failed to open seed file %s: %v", transactionSeedJSONPath, err)
		return
	}

	var rows []SeedTransaksiUtamaRow
	if err := json.Unmarshal(data, &rows); err != nil {
		log.Printf("Failed to parse transaksi utama seed JSON: %v", err)
		return
	}

	for _, row := range rows {
		seedSingleTransaksiUtama(db, row)
	}

	log.Println("Invoice & transaction seeding (Transaksi Utama) completed.")
}

func seedSingleTransaksiUtama(db *gorm.DB, row SeedTransaksiUtamaRow) {
	// Validasi angka penting
	if row.TotalDiterima == nil || row.KeuntunganKotor == nil || row.TotalBayarPetani == nil {
		log.Printf("Skipping utama transaction %s because amounts are incomplete", row.IDTransaksi)
		return
	}

	// Parse tanggal transaksi
	txnDate, err := parseDateYMD(row.Tanggal)
	if err != nil || txnDate.IsZero() {
		log.Printf("Invalid date for utama transaction %s: %v", row.IDTransaksi, err)
		return
	}

	amount := *row.TotalDiterima      // yang diterima pekerja/driver
	platformFee := *row.KeuntunganKotor
	totalAmount := *row.TotalBayarPetani // yang dibayar petani (amount + platformFee)

	// Cari farmer dari nama pengirim (petani)
	farmerID, err := findFarmerIDByName(db, row.PengirimPetani)
	if err != nil || farmerID == uuid.Nil {
		log.Printf("Farmer not found for '%s' in utama transaction %s, skipping", row.PengirimPetani, row.IDTransaksi)
		return
	}

	// ====== Buat Invoice ======
	invoice := models.Invoice{
		FarmerID:    farmerID,
		Amount:      amount,
		PlatformFee: platformFee,
		TotalAmount: totalAmount,
		Status:      "paid",
		DueDate:     txnDate,
		CreatedAt:   txnDate.AddDate(0, 0, -1), // 1 hari sebelum transaksi
		UpdatedAt:   txnDate,
	}

	if err := db.Create(&invoice).Error; err != nil {
		log.Printf("Failed to create invoice for utama transaction %s: %v", row.IDTransaksi, err)
		return
	}

	// ====== Buat Transaction ======
	paymentMethod := strings.TrimSpace(row.MetodePembayaran)
	var paymentMethodPtr *string
	if paymentMethod != "" {
		paymentMethodPtr = &paymentMethod
	}

	refID := strings.TrimSpace(row.IDTransaksi)
	var refIDPtr *string
	if refID != "" {
		refIDPtr = &refID
	}

	txn := models.Transaction{
		InvoiceID:                 invoice.ID,
		PaymentGateway:            "midtrans",
		PaymentGatewayReferenceID: refIDPtr,
		AmountPaid:                totalAmount,    // yang dibayar petani
		PaymentMethod:             paymentMethodPtr,
		TransactionDate:           txnDate,
	}

	if err := db.Create(&txn).Error; err != nil {
		log.Printf("Failed to create utama transaction %s: %v", row.IDTransaksi, err)
		return
	}

	// ====== Buat PlatformProfit (Platform Fee) ======
	gross := platformFee                    // Keuntungan kotor platform
	gatewayFee := 0.0
	if row.BiayaMidtrans != nil {
		gatewayFee = *row.BiayaMidtrans    // biaya midtrans dari Excel
	}
	netProfit := gross - gatewayFee        // harus ≈ Keuntungan Bersih di Excel

	profit := models.PlatformProfit{
		TransactionID: txn.ID,
		SourceType:    "utama",
		GrossProfit:   gross,
		GatewayFee:    gatewayFee,
		NetProfit:     netProfit,
		ProfitDate:    txn.TransactionDate,
	}

	if err := db.Create(&profit).Error; err != nil {
		log.Printf("Failed to create platform profit for utama txn %s: %v", row.IDTransaksi, err)
	}
}


func SeedEcommerceTransactionsAndInvoices(db *gorm.DB) {
	log.Println("Seeding invoices & transactions (E-Commerce) from JSON...")

	data, err := os.ReadFile(ecommerceSeedJSONPath)
	if err != nil {
		log.Printf("Failed to open seed file %s: %v", ecommerceSeedJSONPath, err)
		return
	}

	var rows []SeedEcommerceRow
	if err := json.Unmarshal(data, &rows); err != nil {
		log.Printf("Failed to parse ecommerce transaction seed JSON: %v", err)
		return
	}

	for _, row := range rows {
		seedSingleEcommerceTransaction(db, row)
	}

	log.Println("Invoice & transaction seeding (E-Commerce) completed.")
}

func seedSingleEcommerceTransaction(db *gorm.DB, row SeedEcommerceRow) {
	// Validasi angka penting
	if row.TotalDiterimaPenjual == nil || row.KeuntunganKotor == nil {
		log.Printf("Skipping ecommerce transaction %s because amounts are incomplete", row.IDTransaksi)
		return
	}

	// Parse tanggal transaksi
	txnDate, err := parseDateYMD(row.Tanggal)
	if err != nil || txnDate.IsZero() {
		log.Printf("Invalid date for ecommerce transaction %s: %v", row.IDTransaksi, err)
		return
	}

	amount := *row.TotalDiterimaPenjual // yang diterima penjual/petani
	platformFee := *row.KeuntunganKotor
	totalAmount := amount + platformFee // yang dibayar pembeli

	// Cari farmer/penjual dari nama di kolom Penjual
	farmerID, err := findFarmerIDByName(db, row.Penjual)
	if err != nil || farmerID == uuid.Nil {
		log.Printf("Farmer (seller) not found for '%s' in ecommerce transaction %s, skipping", row.Penjual, row.IDTransaksi)
		return
	}

	// ====== Buat Invoice ======
	invoice := models.Invoice{
		FarmerID:    farmerID,
		Amount:      amount,
		PlatformFee: platformFee,
		TotalAmount: totalAmount,
		Status:      "paid",
		DueDate:     txnDate,
		CreatedAt:   txnDate.AddDate(0, 0, -1),
		UpdatedAt:   txnDate,
	}

	if err := db.Create(&invoice).Error; err != nil {
		log.Printf("Failed to create ecommerce invoice for transaction %s: %v", row.IDTransaksi, err)
		return
	}

	// ====== Buat Transaction ======
	// Metode pembayaran untuk ecommerce – bisa kamu ganti kalau punya detail spesifik
	method := "ecommerce"
	paymentMethodPtr := &method

	refID := strings.TrimSpace(row.IDTransaksi)
	var refIDPtr *string
	if refID != "" {
		refIDPtr = &refID
	}

	txn := models.Transaction{
		InvoiceID:                 invoice.ID,
		PaymentGateway:            "midtrans",
		PaymentGatewayReferenceID: refIDPtr,
		AmountPaid:                totalAmount,
		PaymentMethod:             paymentMethodPtr,
		TransactionDate:           txnDate,
	}

	if err := db.Create(&txn).Error; err != nil {
		log.Printf("Failed to create ecommerce transaction %s: %v", row.IDTransaksi, err)
		return
	}

	// ====== Buat PlatformProfit (Platform Fee) ======
	gross := platformFee // Keuntungan kotor platform
	gatewayFee := 0.0    // kalau belum ada data fee khusus untuk ecommerce, set 0
	netProfit := gross - gatewayFee

	profit := models.PlatformProfit{
		TransactionID: txn.ID,
		SourceType:    "ecommerce",
		GrossProfit:   gross,
		GatewayFee:    gatewayFee,
		NetProfit:     netProfit,
		ProfitDate:    txn.TransactionDate,
	}

	if err := db.Create(&profit).Error; err != nil {
		log.Printf("Failed to create platform profit for ecommerce txn %s: %v", row.IDTransaksi, err)
	}
}

