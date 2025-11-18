package config

import (
	"fmt"
	"log"
	"time"

	"github.com/whsasmita/AgroLink_API/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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
	seedCompletedProjectScenario(db)
	seedInProgressDeliveryScenario(db)
	log.Println("✅ Default data seeded successfully")
}

// seedUsers membuat data dummy untuk pengguna (Admin, Farmer, Worker).
func seedUsers(db *gorm.DB) {
	log.Println("Creating seed users...")
	usersToSeed := []struct {
		User     models.User
		Farmer   *models.Farmer
		Worker   *models.Worker
		Driver   *models.Driver
		Password string
	}{
		{
			User:     models.User{Name: "Admin User", Email: "admin@agrolink.com", Role: "admin", EmailVerified: true},
			Password: "password123",
		},
		{
			User:     models.User{Name: "Budi Petani", Email: "farmer1@agrolink.com", Role: "farmer", EmailVerified: true, PhoneNumber: StringPtr("081082083099")},
			Farmer:   &models.Farmer{Address: StringPtr("Desa Sukamaju No. 10")},
			Password: "password123",
		},
		{
			User: models.User{
				Name:  "Joko Pekerja", Email: "worker1@agrolink.com", Role: "worker",
				EmailVerified: true, PhoneNumber: StringPtr("081234567890"),
			},
			Worker: &models.Worker{
				Skills:            `["menanam","menyiram","panen"]`,
				DailyRate:         Float64Ptr(120000),
				Address:           StringPtr("Jalan Tani No. 15, Desa Makmur"),
				NationalID:        StringPtr("3501234567890001"),
				BankName:          StringPtr("BCA"),
				BankAccountNumber: StringPtr("1234567890"),
				BankAccountHolder: StringPtr("JOKO PEKERJA"),
			},
			Password: "password123",
		},
		{
			User: models.User{
				Name:  "Eka Supir", Email: "driver1@agrolink.com", Role: "driver",
				EmailVerified: true, PhoneNumber: StringPtr("085678901234"),
			},
			Driver: &models.Driver{
				Address:           StringPtr("Jalan Logistik No. 1, Denpasar"),
				PricingScheme:     `{"per_km": 5000, "base_fare": 20000}`,
				VehicleTypes:      `["pickup", "truk engkel"]`,
				CurrentLat:        Float64Ptr(-8.65),
				CurrentLng:        Float64Ptr(115.21),
				BankName:          StringPtr("BRI"),
				BankAccountNumber: StringPtr("0987654321"),
				BankAccountHolder: StringPtr("EKA SUPIR"),
			},
			Password: "password123",
		},
	}

	for _, data := range usersToSeed {
		var existingUser models.User
		if err := db.Where("email = ?", data.User.Email).First(&existingUser).Error; err == nil {
			log.Printf("User with email %s already exists, skipping seed.", data.User.Email)
			continue
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", data.User.Email, err)
			continue
		}
		data.User.Password = string(hashedPassword)
		if data.Farmer != nil { data.User.Farmer = data.Farmer }
		if data.Worker != nil { data.User.Worker = data.Worker }
		if data.Driver != nil { data.User.Driver = data.Driver }
		if err := db.Create(&data.User).Error; err != nil {
			log.Printf("Failed to create user %s: %v", data.User.Email, err)
		}
	}
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
			FarmerID:    farmerUser.Farmer.UserID,
			Title:       "Proyek Panen Jagung (Selesai)",
			Description: "Proyek ini sudah selesai dan siap untuk di-review oleh petani.",
			Location:    "Sawah Seeder, Bali",
			WorkersNeeded: 1,
			StartDate:   time.Now().AddDate(0, 0, -10),
			EndDate:     time.Now().AddDate(0, 0, -1),
			PaymentRate: Float64Ptr(125000),
			PaymentType: "per_day",
			Status:      "completed",
		}
		if err := tx.Create(&project).Error; err != nil { return err }

		contract := models.Contract{
			ProjectID:       &project.ID, FarmerID: farmerUser.Farmer.UserID,
			WorkerID:        &workerUser.Worker.UserID,
			ContractType:    "work", Status: "completed",
		}
		if err := tx.Create(&contract).Error; err != nil { return err }
		
		assignment := models.ProjectAssignment{
			ProjectID: project.ID, WorkerID: workerUser.Worker.UserID,
			ContractID: contract.ID, AgreedRate: *project.PaymentRate, Status: "completed",
		}
		if err := tx.Create(&assignment).Error; err != nil { return err }

		invoice := models.Invoice{
			ProjectID: &project.ID, FarmerID: farmerUser.Farmer.UserID,
			Amount: 120000, PlatformFee: 5000, TotalAmount: 125000,
			Status: "paid", DueDate: time.Now(),
		}
		if err := tx.Create(&invoice).Error; err != nil { return err }

		transaction := models.Transaction{
			InvoiceID: invoice.ID, AmountPaid: invoice.TotalAmount,
		}
		if err := tx.Create(&transaction).Error; err != nil { return err }

		payoutWorker := models.Payout{
			TransactionID: transaction.ID,
			PayeeID:       workerUser.Worker.UserID,
			PayeeType:     "worker",
			Amount:        invoice.Amount,
			Status:        "pending_disbursement",
		}
		if err := tx.Create(&payoutWorker).Error; err != nil { return err }

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
		if err := tx.Create(&contract).Error; err != nil { return err }

		delivery := models.Delivery{
			FarmerID: farmerUser.Farmer.UserID, DriverID: &driverUser.Driver.UserID,
			ContractID: &contract.ID, ItemDescription: "100kg Stroberi Segar",
			Status: "delivered",
		}
		if err := tx.Create(&delivery).Error; err != nil { return err }

		invoice := models.Invoice{
			DeliveryID: &delivery.ID, FarmerID: farmerUser.Farmer.UserID,
			Amount: 200000, PlatformFee: 10000, TotalAmount: 210000,
			Status: "paid", DueDate: time.Now(),
		}
		if err := tx.Create(&invoice).Error; err != nil { return err }

		transaction := models.Transaction{
			InvoiceID: invoice.ID, AmountPaid: invoice.TotalAmount,
		}
		if err := tx.Create(&transaction).Error; err != nil { return err }

		payoutDriver := models.Payout{
			TransactionID: transaction.ID,
			PayeeID:       driverUser.Driver.UserID,
			PayeeType:     "driver",
			Amount:        invoice.Amount,
			Status:        "pending_disbursement",
		}
		if err := tx.Create(&payoutDriver).Error; err != nil { return err }
		
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
	return &s
}

func Float64Ptr(f float64) *float64 {
	return &f
}