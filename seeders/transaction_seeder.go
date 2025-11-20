package seeders

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
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
