package config

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/seeders"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const seedUserJSONPath = "seeders/users_seed.json"


// List semua model untuk migrasi.
var migrationModels = []interface{}{
	// Base user models first
	// 1. Model dasar tanpa banyak dependensi
	&models.User{},
	&models.Payout{}, // Payout di sini
	// &models.SystemSetting{},

	// 2. Model profil yang bergantung pada User
	&models.Farmer{},
	&models.Worker{},
	&models.Driver{},

	// 3. Model utama yang bergantung pada profil
	&models.Project{},
	&models.Delivery{},
	&models.FarmLocation{},

	// 4. Model transaksi & perjanjian yang bergantung pada Project/Delivery
	&models.Invoice{},
	&models.Transaction{},
	&models.Contract{},

	// 5. Model-model pendukung yang memiliki banyak relasi
	&models.ProjectApplication{},
	&models.ProjectAssignment{},

	&models.Review{},
	&models.WorkerAvailability{},
	&models.LocationTrack{},
	&models.WebhookLog{},

	// 6. Model tambahan dari ERD e-commerce
	&models.Product{},
	// &models.UserVerification{}, // Pastikan ini diaktifkan jika Anda menggunakannya
	&models.Cart{},
	&models.Order{},
	&models.OrderItem{},
	&models.ECommercePayment{},
}

// =====================================================================
// FUNGSI UTAMA MIGRASI
// =====================================================================

// RunMigrationWithReset menjalankan proses drop table, auto migrate, dan seeding.
func RunMigrationWithReset(db *gorm.DB) {
	// 1. Hapus semua tabel yang ada (Reset)
	log.Println("🔥 Dropping existing tables...")
	if err := dropAllTables(db); err != nil {
		log.Fatalf("Failed to drop tables: %v", err)
	}
	log.Println("Tables dropped successfully.")

	// 2. Jalankan AutoMigrate untuk membuat skema baru
	AutoMigrate(db)

	// 3. Jalankan Seeder untuk mengisi data awal
	SeedDefaultData(db)
}

// AutoMigrate hanya membuat atau memperbarui tabel tanpa menghapus data.
func AutoMigrate(db *gorm.DB) {
	log.Println("🔄 Running database migrations...")
	for _, model := range migrationModels {
		if err := db.AutoMigrate(model); err != nil {
			log.Fatalf("Failed to migrate %T: %v", model, err)
		}
	}
	log.Println("✅ Database migrations completed successfully")
	CreateIndexes(db)
}

func dropAllTables(db *gorm.DB) error {
	log.Println("Disabling foreign key checks...")
	// [PERBAIKAN] Matikan pemeriksaan constraint
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0;").Error; err != nil {
		return err
	}

	// Reverse order untuk menghapus tabel
	for i := len(migrationModels) - 1; i >= 0; i-- {
		model := migrationModels[i]
		if err := db.Migrator().DropTable(model); err != nil {
			log.Printf("Warning: Failed to drop table for model %T: %v", model, err)
		}
	}

	log.Println("Re-enabling foreign key checks...")
	// [PERBAIKAN] Nyalakan kembali pemeriksaan constraint
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 1;").Error; err != nil {
		return err
	}
	return nil
}

// =====================================================================
// FUNGSI SEEDING DATA
// =====================================================================

// SeedDefaultData adalah fungsi utama untuk memanggil semua seeder.
func SeedDefaultData(db *gorm.DB) {
	log.Println("🌱 Seeding default data...")
	// seedSystemSettings(db)
	seedUsers(db)
	// seedContractTemplates(db)
	// seedCompletedProjectScenario(db)
	// seedInProgressDeliveryScenario(db)
	seeders.SeedTransactionsAndInvoices(db)
	seeders.SeedEcommerceTransactionsAndInvoices(db)
	log.Println("✅ Default data seeded successfully")
}

// seedUsers membuat data dummy untuk pengguna (Admin, Farmer, Worker).
func seedUsers(db *gorm.DB) {
	log.Println("Seeding users from JSON...")
	rand.Seed(time.Now().UnixNano())
	startDate := time.Date(2024, 9, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Now()

	// 1. Baca file JSON
	data, err := os.ReadFile(seedUserJSONPath)
	if err != nil {
		log.Printf("Failed to open seed file %s: %v", seedUserJSONPath, err)
		return
	}

	// 2. Parse JSON ke slice struct
	var rows []seeders.SeedUserRow
	if err := json.Unmarshal(data, &rows); err != nil {
		log.Printf("Failed to parse seed JSON: %v", err)
		return
	}

	for _, row := range rows {
		email := strings.TrimSpace(row.Email)
		if email == "" {
			log.Printf("Skipping row without email: %+v", row)
			continue
		}

		// 3. Cek apakah user sudah ada
		var existing models.User
		if err := db.Where("email = ?", email).First(&existing).Error; err == nil {
			log.Printf("User with email %s already exists, skipping seed.", email)
			continue
		}

		// 4. Tentukan password
		password := strings.TrimSpace(row.Password)
		if password == "" {
			password = "password123"
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", email, err)
			continue
		}

		role := strings.ToLower(strings.TrimSpace(row.Role))
		addressPtr := StringPtr(row.Alamat)

		// 5. Bentuk object User dasar
		user := models.User{
			Name:          strings.TrimSpace(row.Nama),
			Email:         email,
			Password:      string(hashedPassword),
			Role:          role,
			EmailVerified: true,
			PhoneNumber:   StringPtr(normalizePhone(row.NoHP)),
		}

		// 6. Mapping role ke relasi
		switch role {
		case "farmer":
			user.Farmer = &models.Farmer{
				Address: addressPtr,
			}

		case "worker":
			// Skills: []string → string JSON
			var skillsJSON string
			if len(row.Skills) > 0 {
				if b, err := json.Marshal(row.Skills); err != nil {
					log.Printf("Failed to marshal skills for %s: %v", email, err)
				} else {
					skillsJSON = string(b)
				}
			}

			worker := &models.Worker{
				Address:           addressPtr,
				Skills:            skillsJSON, // string JSON, sesuai contoh awalmu
				NationalID:        StringPtr(row.NationalID),
				BankName:          StringPtr(row.BankName),
				BankAccountNumber: StringPtr(row.BankAccountNumber),
				BankAccountHolder: StringPtr(row.BankAccountHolder),
			}

			if row.DailyRate != nil {
				worker.DailyRate = row.DailyRate // type *float64, cocok dengan model
			}

			user.Worker = worker

		case "driver":
			// PricingScheme: map → string JSON
			var pricingJSON string
			if row.PricingScheme != nil {
				if b, err := json.Marshal(row.PricingScheme); err != nil {
					log.Printf("Failed to marshal pricing scheme for %s: %v", email, err)
				} else {
					pricingJSON = string(b)
				}
			}

			// VehicleTypes: []string → string JSON
			var vehicleTypesJSON string
			if len(row.VehicleTypes) > 0 {
				if b, err := json.Marshal(row.VehicleTypes); err != nil {
					log.Printf("Failed to marshal vehicle types for %s: %v", email, err)
				} else {
					vehicleTypesJSON = string(b)
				}
			}

			driver := &models.Driver{
				Address:           addressPtr,
				PricingScheme:     pricingJSON,      // string JSON, sesuai contoh awalmu
				VehicleTypes:      vehicleTypesJSON, // string JSON
				BankName:          StringPtr(row.BankName),
				BankAccountNumber: StringPtr(row.BankAccountNumber),
				BankAccountHolder: StringPtr(row.BankAccountHolder),
			}

			// CurrentLat/CurrentLng: *float64 → *float64
			if row.CurrentLat != nil {
				driver.CurrentLat = row.CurrentLat
			}
			if row.CurrentLng != nil {
				driver.CurrentLng = row.CurrentLng
			}

			user.Driver = driver

		case "admin":
			// Admin tidak punya relasi khusus, cukup user saja.

		default:
			log.Printf("Unknown role '%s' for email %s, skipping.", role, email)
			continue
		}
		user.CreatedAt = randomBetween(startDate, endDate)

		// 7. Simpan user (beserta relasinya)
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user %s: %v", email, err)
		}
	}

	log.Println("User seeding from JSON completed.")
}

func CreateIndexes(db *gorm.DB) {
	log.Println("🔄 Creating database indexes...")
	indexes := []string{
		// ... (kode SQL index Anda)
	}
	for _, indexSQL := range indexes {
		if err := db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %s - %v", indexSQL, err)
		}
	}
	log.Println("✅ Database indexes created successfully")
}

func seedCompletedProjectScenario(db *gorm.DB) {
	log.Println("Creating a completed project scenario for payout test...")
	err := db.Transaction(func(tx *gorm.DB) error {
		var farmerUser, workerUser models.User
		if err := tx.Preload("Farmer").Where("email = ?", "farmer1@agrolink.com").First(&farmerUser).Error; err != nil {
			return fmt.Errorf("seeder failed: could not find farmer1@agrolink.com")
		}
		if err := tx.Preload("Worker").Where("email = ?", "worker1@agrolink.com").First(&workerUser).Error; err != nil {
			return fmt.Errorf("seeder failed: could not find worker1@agrolink.com")
		}

		project := models.Project{
			FarmerID:      farmerUser.Farmer.UserID,
			Title:         "Proyek Panen Jagung (Selesai)",
			Description:   "Proyek ini sudah selesai dan siap untuk di-review oleh petani.",
			Location:      "Sawah Seeder, Bali",
			WorkersNeeded: 1,
			StartDate:     time.Now().AddDate(0, 0, -10),
			EndDate:       time.Now().AddDate(0, 0, -1),
			PaymentRate:   Float64Ptr(125000),
			PaymentType:   "per_day",
			Status:        "completed",
		}
		if err := tx.Create(&project).Error; err != nil {
			return err
		}

		contract := models.Contract{
			ProjectID: &project.ID, FarmerID: farmerUser.Farmer.UserID,
			WorkerID:     &workerUser.Worker.UserID,
			ContractType: "work", Status: "completed",
		}
		if err := tx.Create(&contract).Error; err != nil {
			return err
		}

		assignment := models.ProjectAssignment{
			ProjectID: project.ID, WorkerID: workerUser.Worker.UserID,
			ContractID: contract.ID, AgreedRate: *project.PaymentRate, Status: "completed",
		}
		if err := tx.Create(&assignment).Error; err != nil {
			return err
		}

		invoice := models.Invoice{
			ProjectID: &project.ID, FarmerID: farmerUser.Farmer.UserID,
			Amount: 120000, PlatformFee: 5000, TotalAmount: 125000,
			Status: "paid", DueDate: time.Now(),
		}
		if err := tx.Create(&invoice).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			InvoiceID: invoice.ID, AmountPaid: invoice.TotalAmount,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		payoutWorker := models.Payout{
			TransactionID: transaction.ID,
			PayeeID:       workerUser.Worker.UserID,
			PayeeType:     "worker",
			Amount:        invoice.Amount,
			Status:        "pending_disbursement",
		}
		if err := tx.Create(&payoutWorker).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Failed to seed completed project scenario: %v", err)
	}
}




func seedInProgressDeliveryScenario(db *gorm.DB) {
	log.Println("Creating a COMPLETED delivery scenario for payout test...")
	err := db.Transaction(func(tx *gorm.DB) error {
		var farmerUser, driverUser models.User
		if err := tx.Preload("Farmer").Where("email = ?", "farmer1@agrolink.com").First(&farmerUser).Error; err != nil {
			return fmt.Errorf("seeder failed: could not find farmer")
		}
		if err := tx.Preload("Driver").Where("email = ?", "driver1@agrolink.com").First(&driverUser).Error; err != nil {
			return fmt.Errorf("seeder failed: could not find driver")
		}

		contract := models.Contract{
			ContractType: "delivery", FarmerID: farmerUser.Farmer.UserID,
			DriverID: &driverUser.Driver.UserID, Status: "completed",
		}
		if err := tx.Create(&contract).Error; err != nil {
			return err
		}

		delivery := models.Delivery{
			FarmerID: farmerUser.Farmer.UserID, DriverID: &driverUser.Driver.UserID,
			ContractID: &contract.ID, ItemDescription: "100kg Stroberi Segar",
			Status: "delivered",
		}
		if err := tx.Create(&delivery).Error; err != nil {
			return err
		}

		invoice := models.Invoice{
			DeliveryID: &delivery.ID, FarmerID: farmerUser.Farmer.UserID,
			Amount: 200000, PlatformFee: 10000, TotalAmount: 210000,
			Status: "paid", DueDate: time.Now(),
		}
		if err := tx.Create(&invoice).Error; err != nil {
			return err
		}

		transaction := models.Transaction{
			InvoiceID: invoice.ID, AmountPaid: invoice.TotalAmount,
		}
		if err := tx.Create(&transaction).Error; err != nil {
			return err
		}

		payoutDriver := models.Payout{
			TransactionID: transaction.ID,
			PayeeID:       driverUser.Driver.UserID,
			PayeeType:     "driver",
			Amount:        invoice.Amount,
			Status:        "pending_disbursement",
		}
		if err := tx.Create(&payoutDriver).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Failed to seed in-progress delivery scenario: %v", err)
	}
}

func seedSystemSettings(db *gorm.DB) {
	// ... (implementasi Anda)
}

// =====================================================================
// HELPER FUNCTIONS
// =====================================================================

func StringPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}

func Float64Ptr(f float64) *float64 {
	return &f
}

func normalizePhone(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	// kalau sudah mulai dengan 0, biarkan
	if strings.HasPrefix(raw, "0") {
		return raw
	}
	// data dari Excel kita kebanyakan tanpa 0 di depan (contoh: 8214...)
	return "0" + raw
}

func randomBetween(start, end time.Time) time.Time {
	// buat interval dalam detik
	delta := end.Unix() - start.Unix()
	if delta <= 0 {
		return start
	}
	// angka acak dalam range delta
	sec := rand.Int63n(delta)
	return time.Unix(start.Unix()+sec, 0)
}