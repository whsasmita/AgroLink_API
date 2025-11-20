package seeders

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const ecommerceSeedJSONPath = "seeders/transactions_seed_ecommerce.json"
const transactionSeedJSONPath = "seeders/transactions_seed_utama.json"

type SeedTransaksiUtamaRow struct {
	IDTransaksi      string `json:"IDTransaksi"`
	Tanggal          string `json:"Tanggal"`
	PengirimPetani   string `json:"PengirimPetani"`   // nama farmer (pengirim)
	Penerima         string `json:"Penerima"`         // pekerja / driver (kalau mau dipakai nanti)
	RolePenerima     string `json:"RolePenerima"`     // worker / driver (opsional)
	MetodePembayaran string `json:"MetodePembayaran"` // misal: "Transfer Bank", "Cash"

	TotalDiterima    *float64 `json:"TotalDiterima"`    // uang yang diterima pekerja/driver
	KeuntunganKotor  *float64 `json:"KeuntunganKotor"`  // fee kotor platform
	KeuntunganBersih *float64 `json:"KeuntunganBersih"` // pendapatan bersih platform
	BiayaMidtrans    *float64 `json:"BiayaMidtrans"`    // biaya gateway, kalau ada
	TotalBayarPetani *float64 `json:"TotalBayarPetani"` // yang dibayar petani (amount + fee + biaya lainnya)
}

type SeedEcommerceRow struct {
	IDTransaksi          string   `json:"IDTransaksi"`
	Tanggal              string   `json:"Tanggal"`
	Pembeli              string   `json:"Pembeli"`
	Produk               string   `json:"Produk"`
	Harga                *float64 `json:"Harga"`
	Ongkir               *float64 `json:"Ongkir"`
	KeuntunganKotor      *float64 `json:"KeuntunganKotor"`
	KeuntunganBersih     *float64 `json:"KeuntunganBersih"` // 👈 kolom baru
	BiayaMidtrans        *float64 `json:"BiayaMidtrans"`    // 👈 kalau di sheet baru juga ada
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

// func parseDateYMD(s string) (time.Time, error) {
// 	s = strings.TrimSpace(s)
// 	if s == "" {
// 		return time.Time{}, nil
// 	}
// 	return time.Parse("2006-01-02", s)
// }

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
	// 1. Validasi angka penting dari Excel
	if row.TotalDiterima == nil || row.KeuntunganKotor == nil || row.TotalBayarPetani == nil {
		log.Printf("Skipping utama transaction %s because amounts are incomplete", row.IDTransaksi)
		return
	}

	// 2. Parse tanggal transaksi (pakai helper fleksibel)
	txnDate, err := parseDateFlexible(row.Tanggal)
	if err != nil || txnDate.IsZero() {
		log.Printf("Invalid date for utama transaction %s: %v", row.IDTransaksi, err)
		return
	}

	amount := *row.TotalDiterima // yang diterima pekerja/driver
	platformFee := *row.KeuntunganKotor
	totalAmount := *row.TotalBayarPetani // yang dibayar petani (sesuai Excel)

	// 3. Cari farmer dari nama pengirim (petani)
	farmerID, err := findFarmerIDByName(db, row.PengirimPetani)
	if err != nil || farmerID == uuid.Nil {
		log.Printf("Farmer not found for '%s' in utama transaction %s, skipping", row.PengirimPetani, row.IDTransaksi)
		return
	}

	// 4. Cek apakah transaksi ini sudah pernah diseed (hindari duplikasi)
	refID := strings.TrimSpace(row.IDTransaksi)
	if refID != "" {
		var existingTxn models.Transaction
		if err := db.
			Where("payment_gateway_reference_id = ?", refID).
			First(&existingTxn).Error; err == nil {
			log.Printf("Utama transaction %s already exists, skipping.", row.IDTransaksi)
			return
		}
	}

	// ====== 5. Buat Invoice ======
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

	// ====== 6. Buat Transaction ======
	paymentMethod := strings.TrimSpace(row.MetodePembayaran)
	var paymentMethodPtr *string
	if paymentMethod != "" {
		paymentMethodPtr = &paymentMethod
	}

	var refIDPtr *string
	if refID != "" {
		refIDPtr = &refID
	}

	txn := models.Transaction{
		InvoiceID:                 invoice.ID,
		PaymentGateway:            "midtrans", // atau "manual", sesuaikan jika perlu
		PaymentGatewayReferenceID: refIDPtr,
		AmountPaid:                totalAmount, // yang dibayar petani (Excel)
		PaymentMethod:             paymentMethodPtr,
		TransactionDate:           txnDate,
	}

	if err := db.Create(&txn).Error; err != nil {
		log.Printf("Failed to create utama transaction %s: %v", row.IDTransaksi, err)
		return
	}

	// ====== 7. Buat PlatformProfit (Platform Fee) ======
	gross := platformFee // Keuntungan kotor platform
	var netProfit float64
	var gatewayFee float64

	// pakai KeuntunganBersih dari Excel kalau ada
	if row.KeuntunganBersih != nil {
		netProfit = *row.KeuntunganBersih
	} else {
		// fallback: netProfit = gross - biaya gateway
		netProfit = gross
	}

	// biaya gateway (BiayaMidtrans) jika ada
	if row.BiayaMidtrans != nil {
		gatewayFee = *row.BiayaMidtrans
	} else {
		// kalau tidak ada, dan KeuntunganBersih tersedia:
		if row.KeuntunganBersih != nil {
			gatewayFee = gross - netProfit
			if gatewayFee < 0 {
				gatewayFee = 0
			}
		} else {
			gatewayFee = 0
		}
	}

	profit := models.PlatformProfit{
		TransactionID: &txn.ID,
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

func SeedEcommerceTransactions(db *gorm.DB) {
	log.Println("Seeding invoices & transactions (E-Commerce) from JSON...")
	rand.Seed(time.Now().UnixNano())

	// 1. Load farmer whitelist dari DB (24 farmer dari users_seed.json)
	farmers, err := loadSeedFarmers(db)
	if err != nil {
		log.Printf("Failed to load seed farmers for ecommerce: %v", err)
		return
	}

	// 2. Baca file JSON
	data, err := os.ReadFile(ecommerceSeedJSONPath)
	if err != nil {
		log.Printf("Failed to open seed file %s: %v", ecommerceSeedJSONPath, err)
		return
	}

	var rows []SeedEcommerceRow
	if err := json.Unmarshal(data, &rows); err != nil {
		log.Printf("Failed to parse ecommerce seed JSON: %v", err)
		return
	}

	// 3. Loop setiap baris transaksi
	for _, row := range rows {
		if err := seedSingleEcommerceTransaction(db, row, farmers); err != nil {
			log.Printf("Failed to seed ecommerce txn %s: %v", row.IDTransaksi, err)
		}
	}

	log.Println("E-Commerce transactions seeding completed.")
}

func seedSingleEcommerceTransaction(db *gorm.DB, row SeedEcommerceRow, farmers []seedFarmer) error {
	// Validasi minimal
	if row.Harga == nil || row.KeuntunganKotor == nil || row.TotalDiterimaPenjual == nil {
		return fmt.Errorf("incomplete monetary fields for txn %s", row.IDTransaksi)
	}

	// Parse tanggal
	txnDate, err := parseDateFlexible(row.Tanggal)
	if err != nil || txnDate.IsZero() {
		return fmt.Errorf("invalid date '%s' for txn %s: %w", row.Tanggal, row.IDTransaksi, err)
	}

	amount := *row.TotalDiterimaPenjual   // yang diterima penjual/petani
	platformFee := *row.KeuntunganKotor  // pendapatan kotor platform
	totalAmount := amount + platformFee  // total dibayar buyer ke platform

	// Pilih farmer random dari whitelist
	farmer := pickRandomFarmer(farmers)
	if farmer.ID == uuid.Nil {
		return fmt.Errorf("no farmer available for txn %s", row.IDTransaksi)
	}

	// (Opsional) Cek duplikasi invoice
	var existingInvoice models.Invoice
	if err := db.
		Where("farmer_id = ? AND amount = ? AND total_amount = ? AND DATE(due_date) = DATE(?)",
			farmer.ID, amount, totalAmount, txnDate).
		First(&existingInvoice).Error; err == nil {
		log.Printf("Ecommerce invoice for txn %s already exists, skipping.", row.IDTransaksi)
		return nil
	}

	// ====== 1. (Opsional) Buat Invoice ======
	invoice := models.Invoice{
		FarmerID:    farmer.ID,
		Amount:      amount,
		PlatformFee: platformFee,
		TotalAmount: totalAmount,
		Status:      "paid",
		DueDate:     txnDate,
		CreatedAt:   txnDate.AddDate(0, 0, -1),
		UpdatedAt:   txnDate,
	}

	if err := db.Create(&invoice).Error; err != nil {
		return fmt.Errorf("failed to create ecommerce invoice: %w", err)
	}

	// ====== 2. TIDAK ADA Transaction ke tabel transactions ======
	// (blok txn := models.Transaction{...} dihapus)

	// ====== 3. Buat ECommercePayment ======
	buyerID, err := findUserIDByName(db, row.Pembeli)
	if err != nil || buyerID == uuid.Nil {
		buyerID = farmer.ID
	}

	refID := strings.TrimSpace(row.IDTransaksi)

	payment := models.ECommercePayment{
		ID:         uuid.New(),
		UserID:     buyerID,
		GrandTotal: totalAmount,
		Status:     "paid",
		SnapToken:  fmt.Sprintf("SEED-%s", refID),
		CreatedAt:  txnDate,
		UpdatedAt:  txnDate,
	}

	if err := db.Create(&payment).Error; err != nil {
		return fmt.Errorf("failed to create ECommercePayment: %w", err)
	}

	// ====== 4. Buat PlatformProfit refer ke ECommercePayment ======

	gross := platformFee
	var netProfit float64
	var gatewayFee float64

	if row.KeuntunganBersih != nil {
		netProfit = *row.KeuntunganBersih
		if row.BiayaMidtrans != nil {
			gatewayFee = *row.BiayaMidtrans
		} else {
			gatewayFee = gross - netProfit
			if gatewayFee < 0 {
				gatewayFee = 0
			}
		}
	} else {
		gatewayFee = 0
		netProfit = gross
	}

	profit := models.PlatformProfit{
		SourceType:         "ecommerce",
		TransactionID:      nil,
		ECommercePaymentID: &payment.ID,
		GrossProfit:        gross,
		GatewayFee:         gatewayFee,
		NetProfit:          netProfit,
		ProfitDate:         txnDate,
	}

	if err := db.Create(&profit).Error; err != nil {
		return fmt.Errorf("failed to create platform profit for ecommerce: %w", err)
	}

	return nil
}


func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			b.WriteRune('-')
		}
	}
	out := b.String()
	out = strings.Trim(out, "-")
	out = strings.ReplaceAll(out, "--", "-")
	if out == "" {
		return "product"
	}
	return out
}

func categoryForProduct(title string) *string {
	t := strings.ToLower(title)

	var cat string
	switch {
	case strings.Contains(t, "kopi"):
		cat = "Kopi"
	case strings.Contains(t, "salak") || strings.Contains(t, "alpukat") ||
		strings.Contains(t, "mangga") || strings.Contains(t, "jeruk") ||
		strings.Contains(t, "apel") || strings.Contains(t, "naga") ||
		strings.Contains(t, "rambutan") || strings.Contains(t, "nanas") ||
		strings.Contains(t, "stowberry") || strings.Contains(t, "stroberi"):
		cat = "Buah"
	case strings.Contains(t, "cabe") || strings.Contains(t, "cabai") ||
		strings.Contains(t, "bawang") || strings.Contains(t, "tomat") ||
		strings.Contains(t, "kubis") || strings.Contains(t, "buncis") ||
		strings.Contains(t, "wortel") || strings.Contains(t, "kentang") ||
		strings.Contains(t, "selada") || strings.Contains(t, "jagung") ||
		strings.Contains(t, "labu"):
		cat = "Sayur"
	case strings.Contains(t, "dodol"):
		cat = "Oleh-oleh"
	default:
		cat = "Lainnya"
	}

	return &cat
}

func getDefaultFarmerID(db *gorm.DB) (uuid.UUID, error) {
	var farmer models.Farmer
	if err := db.First(&farmer).Error; err != nil {
		return uuid.Nil, err
	}
	return farmer.UserID, nil
}

var seedProductNames = []string{
	"Kopi Arabica",
	"Jeruk",
	"Dodol",
	"Salak Gula Bali",
	"Kopi Bubuk Luwak",
	"Kopi Bubuk Robusta",
	"Mangga",
	"Green Been (Natural)",
	"Green Been (Fullwash)",
	"Cabe Rawit",
	"Bawang Merah",
	"Alpukat",
	"Stowberry",
	"Kelapa",
	"Apel Manalagi",
	"Jeruk Bali",
	"Bawang Putih",
	"Jagung Manis",
	"Rambutan",
	"Jambu Biji Kristal",
	"Labu Siam",
	"Nanas",
	"Arabica",
	"Buah Naga",
	"Cabe Besar",
	"Tomat",
	"Kubis",
	"Buncis",
	"Wortel",
	"Kacang Tanah",
	"Selada",
	"Kentang",
}

var whitelistedFarmerEmails = []string{
	"kadek.andre.p@email.com",
	"komang.sutami@email.com",
	"niluhsugi.a@email.com",
	"ni.komang.w@email.com",
	"resminiresi631@gmail.com",
	"siti_nur2011@gmail.com",
	"tj8622@gmail.com",
	"aristya_pradana2016@gmail.com",
	"kw3917@gmail.com",
	"am7717@gmail.com",
	"miftahul.putri71@gmail.com",
	"putu.aulia78@gmail.com",
	"rd9097@gmail.com",
	"ws4773@gmail.com",
	"atmajasuditha459@gmail.com",
	"tj2073@gmail.com",
	"berliantiara415@gmail.com",
	"kw7080@gmail.com",
	"i.sudiarta45@gmail.com",
	"n.madesanti2022@gmail.com",
	"komang.ayu12@gmail.com",
	"masutianingsihdewi974@gmail.com",
	"ck9761@gmail.com",
	"kadek.krisna64@gmail.com",
}

type seedFarmer struct {
	ID    uuid.UUID
	Name  string
	Email string
}

func loadSeedFarmers(db *gorm.DB) ([]seedFarmer, error) {
	var users []models.User
	if err := db.
		Where("email IN ? AND role = ?", whitelistedFarmerEmails, "farmer").
		Find(&users).Error; err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no farmers found for whitelisted emails")
	}

	farmers := make([]seedFarmer, 0, len(users))
	for _, u := range users {
		farmers = append(farmers, seedFarmer{
			ID:    u.ID,
			Name:  u.Name,
			Email: u.Email,
		})
	}

	return farmers, nil
}

func pickRandomFarmer(farmers []seedFarmer) seedFarmer {
	if len(farmers) == 0 {
		return seedFarmer{}
	}
	idx := rand.Intn(len(farmers))
	return farmers[idx]
}

func SeedProducts(db *gorm.DB) {
	log.Println("Seeding products from ecommerce & product list (random farmers from whitelist)...")

	rand.Seed(time.Now().UnixNano())

	// 1. Load farmer whitelist dari users_seed (harus sudah diseed sebelumnya)
	farmers, err := loadSeedFarmers(db)
	if err != nil {
		log.Printf("Failed to load seed farmers: %v", err)
		return
	}

	// 2. Baca JSON transaksi ecommerce
	data, err := os.ReadFile(ecommerceSeedJSONPath)
	if err != nil {
		log.Printf("Failed to open ecommerce seed file %s: %v", ecommerceSeedJSONPath, err)
		return
	}

	var rows []SeedEcommerceRow
	if err := json.Unmarshal(data, &rows); err != nil {
		log.Printf("Failed to parse ecommerce seed JSON: %v", err)
		return
	}

	// 3. Kumpulkan statistik per Produk (abaikan Penjual, karena farmer akan diacak)
	type prodStat struct {
		Produk    string
		Count     int
		LastPrice float64
		LastDate  time.Time
	}

	stats := make(map[string]*prodStat) // key: nama produk

	for _, row := range rows {
		title := strings.TrimSpace(row.Produk)
		if title == "" || row.Harga == nil {
			continue
		}

		// parse tanggal
		txnDate, _ := time.Parse("2006-01-02", strings.TrimSpace(row.Tanggal))

		ps, ok := stats[title]
		if !ok {
			stats[title] = &prodStat{
				Produk:    title,
				Count:     1,
				LastPrice: *row.Harga,
				LastDate:  txnDate,
			}
		} else {
			ps.Count++
			ps.LastPrice = *row.Harga
			if txnDate.After(ps.LastDate) {
				ps.LastDate = txnDate
			}
		}
	}

	// 4. Seed product berdasarkan statistik transaksi (per Produk)
	for _, ps := range stats {
		// pilih farmer random dari whitelist
		farmer := pickRandomFarmer(farmers)

		// cek apakah produk ini sudah ada (per farmer + title)
		var existing models.Product
		if err := db.Where("farmer_id = ? AND title = ?", farmer.ID, ps.Produk).
			First(&existing).Error; err == nil {
			// sudah ada, skip
			continue
		}

		// tentukan rating berdasarkan jumlah pembelian
		var rating *float64
		if ps.Count > 0 {
			var r float64
			switch {
			case ps.Count >= 5:
				r = 4.9
			case ps.Count >= 3:
				r = 4.7
			case ps.Count == 2:
				r = 4.5
			default: // 1
				r = 4.2
			}
			rating = &r
		}

		// image URLs (placeholder): /public/uploads/products/<slug>.jpg
		slug := slugify(ps.Produk)
		imgArr := []string{"/public/uploads/products/" + slug + ".jpg"}
		imgJSON, _ := json.Marshal(imgArr)

		// stock default (misal 100 - jumlah terjual, minimal 10)
		baseStock := 100
		stock := baseStock - ps.Count
		if stock < 10 {
			stock = 10
		}

		product := models.Product{
			Title:         ps.Produk,
			FarmerID:      farmer.ID,
			Description:   fmt.Sprintf("%s dari %s", ps.Produk, farmer.Name),
			Location:      nil, // bisa diisi nanti
			Category:      categoryForProduct(ps.Produk),
			Price:         ps.LastPrice,
			Stock:         stock,
			ReservedStock: 0,
			ImageURLs:     datatypes.JSON(imgJSON),
			Rating:        rating,
			CreatedAt:     ps.LastDate.AddDate(0, 0, -1),
			UpdatedAt:     ps.LastDate,
		}

		if err := db.Create(&product).Error; err != nil {
			log.Printf("Failed to create product %s (farmer %s): %v", ps.Produk, farmer.Email, err)
		}
	}

	// 5. Pastikan semua product dari daftar txt juga ada:
	//    kalau belum ada di DB, buat baru dengan farmer random
	for _, title := range seedProductNames {
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}

		// cek apakah SUDAH ada product dengan title ini (apa pun farmer-nya)
		var existing models.Product
		if err := db.Where("title = ?", title).First(&existing).Error; err == nil {
			// sudah ada, skip
			continue
		}

		farmer := pickRandomFarmer(farmers)

		slug := slugify(title)
		imgArr := []string{"/public/uploads/products/" + slug + ".jpg"}
		imgJSON, _ := json.Marshal(imgArr)

		product := models.Product{
			Title:         title,
			FarmerID:      farmer.ID,
			Description:   fmt.Sprintf("%s dari %s", title, farmer.Name),
			Location:      nil,
			Category:      categoryForProduct(title),
			Price:         10000.0, // default placeholder
			Stock:         50,
			ReservedStock: 0,
			ImageURLs:     datatypes.JSON(imgJSON),
			Rating:        nil, // belum ada penjualan → belum ada rating
			CreatedAt:     time.Now().AddDate(0, 0, -7),
			UpdatedAt:     time.Now(),
		}

		if err := db.Create(&product).Error; err != nil {
			log.Printf("Failed to create product '%s' (farmer %s): %v", title, farmer.Email, err)
		}
	}

	log.Println("Product seeding completed (random farmers from whitelist).")
}

// parseDateFlexible mencoba berbagai format tanggal umum & excel serial date.
func parseDateFlexible(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}

	// 1. Coba parse Excel serial number
	if n, err := strconv.Atoi(s); err == nil {
		// Excel mulai tanggal 30 Des 1899 (bug: hitung 1900 sebagai leap year)
		excelBase := time.Date(1899, 12, 30, 0, 0, 0, 0, time.Local)
		return excelBase.AddDate(0, 0, n), nil
	}

	// 2. Coba format tanggal umum
	layouts := []string{
		"2006-01-02",
		"02/01/2006",
		"2/1/2006",
		"02-01-2006",
		"2-1-2006",
		"2006/01/02",
		"2006.01.02",
	}

	var lastErr error
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}

	return time.Time{}, fmt.Errorf("failed to parse date '%s': %v", s, lastErr)
}

// findUserIDByName mencari user berdasarkan nama (case-insensitive).
func findUserIDByName(db *gorm.DB, name string) (uuid.UUID, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return uuid.Nil, fmt.Errorf("empty name")
	}

	var user models.User

	// 1. Exact match (case-insensitive)
	if err := db.Where("LOWER(name) = LOWER(?)", name).First(&user).Error; err == nil {
		return user.ID, nil
	}

	// 2. Coba partial match (LIKE)
	like := "%" + strings.ToLower(name) + "%"
	if err := db.Where("LOWER(name) LIKE ?", like).First(&user).Error; err == nil {
		return user.ID, nil
	}

	return uuid.Nil, fmt.Errorf("user '%s' not found", name)
}
