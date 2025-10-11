package config

import (
	"fmt"
	"log"
	"time"

	"github.com/whsasmita/AgroLink_API/models"
	"golang.org/x/crypto/bcrypt" // Tambahkan import bcrypt
	"gorm.io/gorm"
)

// List semua model untuk migrasi.
var migrationModels = []interface{}{
	// Base user models first
	// 1. Model dasar tanpa banyak dependensi
	&models.User{},
	// &models.SystemSetting{},

	// 2. Model profil yang bergantung pada User
	&models.Farmer{},
	&models.Worker{},
	&models.Driver{},

	// 3. Model utama yang bergantung pada profil
	&models.Project{},
	&models.Delivery{},     // Bergantung pada Driver & Farmer
	&models.FarmLocation{}, // Bergantung pada Farmer

	// 4. Model transaksi & perjanjian yang bergantung pada Project/Delivery
	&models.Invoice{},
	&models.Transaction{},
	&models.Contract{},

	// 5. Model-model pendukung yang memiliki banyak relasi
	&models.ProjectApplication{},
	&models.ProjectAssignment{},
	&models.Payout{},
	&models.Review{},
	&models.WorkerAvailability{},
	&models.LocationTrack{},
	&models.WebhookLog{},

	// 6. Model tambahan dari ERD e-commerce
	&models.Product{},
	// &models.UserVerification{},
	&models.Cart{},
	&models.Order{},
	&models.OrderItem{},
	&models.PaymentMidtrans{},
}

// =====================================================================
// FUNGSI UTAMA MIGRASI
// =====================================================================

// RunMigrationWithReset menjalankan proses drop table, auto migrate, dan seeding.
// SANGAT BERBAHAYA UNTUK PRODUKSI. Gunakan hanya untuk development.
func RunMigrationWithReset() {
	// 1. Hapus semua tabel yang ada (Reset)
	log.Println("🔥 Dropping existing tables...")
	if err := dropAllTables(DB); err != nil {
		log.Fatalf("Failed to drop tables: %v", err)
	}
	log.Println("Tables dropped successfully.")

	// 2. Jalankan AutoMigrate untuk membuat skema baru
	AutoMigrate()

	// 3. Jalankan Seeder untuk mengisi data awal
	SeedDefaultData()
}

// AutoMigrate hanya membuat atau memperbarui tabel tanpa menghapus data.
func AutoMigrate() {
	log.Println("🔄 Running database migrations...")

	for _, model := range migrationModels {
		if err := DB.AutoMigrate(model); err != nil {
			log.Fatalf("Failed to migrate %T: %v", model, err)
		}
	}

	log.Println("✅ Database migrations completed successfully")
	// Panggil CreateIndexes di sini jika Anda ingin index dibuat setiap kali migrasi berjalan
	CreateIndexes()
}

// dropAllTables menghapus semua tabel dalam urutan terbalik untuk menghindari error foreign key.
func dropAllTables(db *gorm.DB) error {
	// Reverse order untuk menghapus tabel dengan foreign key terlebih dahulu
	for i := len(migrationModels) - 1; i >= 0; i-- {
		model := migrationModels[i]
		if err := db.Migrator().DropTable(model); err != nil {
			log.Printf("Warning: Failed to drop table for model %T: %v", model, err)
		}
	}
	return nil
}

// =====================================================================
// FUNGSI SEEDING DATA
// =====================================================================

// SeedDefaultData adalah fungsi utama untuk memanggil semua seeder.
func SeedDefaultData() {
	log.Println("🌱 Seeding default data...")
	seedSystemSettings()
	seedUsers() // <-- Panggil seeder pengguna baru
	// seedContractTemplates()
	seedCompletedProjectScenario()
	seedInProgressDeliveryScenario()
	log.Println("✅ Default data seeded successfully")
}

// seedUsers membuat data dummy untuk pengguna (Admin, Farmer, Worker).
func seedUsers() {
	log.Println("Creating seed users...")

	// [PERBAIKAN] Tambahkan *models.Driver ke dalam struct
	usersToSeed := []struct {
		User     models.User
		Farmer   *models.Farmer
		Worker   *models.Worker
		Driver   *models.Driver // <-- Tambahkan ini
		Password string
	}{
		// 1. Admin User
		{
			User:     models.User{Name: "Admin User", Email: "admin@agrolink.com", Role: "admin", EmailVerified: true},
			Password: "password123",
		},
		// 2. Farmer User
		{
			User:     models.User{Name: "Budi Petani", Email: "farmer1@agrolink.com", Role: "farmer", EmailVerified: true, PhoneNumber: StringPtr("081082083099")},
			Farmer:   &models.Farmer{Address: StringPtr("Desa Sukamaju No. 10")},
			Password: "password123",
		},
		// 3. Worker User
		{
			User: models.User{
				Name:          "Joko Pekerja",
				Email:         "worker1@agrolink.com",
				Role:          "worker",
				EmailVerified: true,
				PhoneNumber:   StringPtr("081234567890"),
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
		// [BARU] 4. Driver User
		{
			User: models.User{
				Name:          "Eka Supir",
				Email:         "driver1@agrolink.com",
				Role:          "driver",
				EmailVerified: true,
				PhoneNumber:   StringPtr("085678901234"),
			},
			Driver: &models.Driver{
				Address: StringPtr("Jalan Logistik No. 1, Denpasar"),
				// Simpan sebagai string JSON
				PricingScheme: `{"per_km": 5000, "base_fare": 20000}`,
				VehicleTypes:  `["pickup", "truk engkel"]`,
				// Lokasi awal driver
				CurrentLat: Float64Ptr(-8.65),
				CurrentLng: Float64Ptr(115.21),
			},
			Password: "password123",
		},
	}

	for _, data := range usersToSeed {
		var existingUser models.User
		if err := DB.Where("email = ?", data.User.Email).First(&existingUser).Error; err == nil {
			log.Printf("User with email %s already exists, skipping seed.", data.User.Email)
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(data.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", data.User.Email, err)
			continue
		}
		data.User.Password = string(hashedPassword)

		// [PERBAIKAN] Hubungkan profil Farmer, Worker, dan Driver
		if data.Farmer != nil {
			data.User.Farmer = data.Farmer
		}
		if data.Worker != nil {
			data.User.Worker = data.Worker
		}
		if data.Driver != nil {
			data.User.Driver = data.Driver
		}

		if err := DB.Create(&data.User).Error; err != nil {
			log.Printf("Failed to create user %s: %v", data.User.Email, err)
		}
	}
}

func CreateIndexes() {
	log.Println("🔄 Creating database indexes...")

	// Define indexes that need to be created manually
	indexes := []string{
		// User indexes
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at)",

		// Project indexes
		"CREATE INDEX IF NOT EXISTS idx_projects_farmer_id ON projects(farmer_id)",
		"CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status)",
		"CREATE INDEX IF NOT EXISTS idx_projects_start_date ON projects(start_date)",
		"CREATE INDEX IF NOT EXISTS idx_projects_project_type ON projects(project_type)",

		// Worker indexes
		"CREATE INDEX IF NOT EXISTS idx_workers_rating ON workers(rating)",
		"CREATE INDEX IF NOT EXISTS idx_workers_work_radius ON workers(work_radius)",

		// Delivery indexes
		"CREATE INDEX IF NOT EXISTS idx_deliveries_status ON deliveries(delivery_status)",
		"CREATE INDEX IF NOT EXISTS idx_deliveries_tracking ON deliveries(tracking_code)",
		"CREATE INDEX IF NOT EXISTS idx_deliveries_scheduled_pickup ON deliveries(scheduled_pickup)",

		// Transaction indexes
		"CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_from_user ON transactions(from_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_to_user ON transactions(to_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(transaction_date)",

		// Notification indexes
		"CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, read_status)",
		"CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at)",

		// Schedule indexes
		"CREATE INDEX IF NOT EXISTS idx_schedules_user_id ON schedules(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_schedules_start_datetime ON schedules(start_datetime)",
		"CREATE INDEX IF NOT EXISTS idx_schedules_status ON schedules(status)",

		// Review indexes
		"CREATE INDEX IF NOT EXISTS idx_reviews_reviewed_user ON reviews(reviewed_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_reviews_rating ON reviews(rating)",
		"CREATE INDEX IF NOT EXISTS idx_reviews_created_at ON reviews(created_at)",

		// Location tracking indexes
		"CREATE INDEX IF NOT EXISTS idx_location_tracks_delivery ON location_tracks(delivery_id, timestamp)",
		"CREATE INDEX IF NOT EXISTS idx_location_tracks_user ON location_tracks(user_id)",

		// Support ticket indexes
		"CREATE INDEX IF NOT EXISTS idx_support_tickets_user ON support_tickets(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_support_tickets_status ON support_tickets(status)",
		"CREATE INDEX IF NOT EXISTS idx_support_tickets_created_at ON support_tickets(created_at)",

		// Activity log indexes
		"CREATE INDEX IF NOT EXISTS idx_activity_logs_user ON activity_logs(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_activity_logs_action ON activity_logs(action)",
		"CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at)",

		// Session indexes
		"CREATE INDEX IF NOT EXISTS idx_user_sessions_user ON user_sessions(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token)",
		"CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions(expires_at)",
	}

	// Execute index creation
	for _, indexSQL := range indexes {
		if err := DB.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: Failed to create index: %s - %v", indexSQL, err)
		}
	}

	log.Println("✅ Database indexes created successfully")
}

func seedCompletedProjectScenario() {
	log.Println("Creating a completed project scenario for rating test...")

	// Gunakan transaksi agar semua data dibuat atau tidak sama sekali
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 1. Ambil data Petani dan Pekerja yang sudah ada
		var farmerUser, workerUser models.User
		if err := tx.Preload("Farmer").Where("email = ?", "farmer1@agrolink.com").First(&farmerUser).Error; err != nil {
			return fmt.Errorf("seeder failed: could not find farmer1@agrolink.com")
		}
		if err := tx.Preload("Worker").Where("email = ?", "worker1@agrolink.com").First(&workerUser).Error; err != nil {
			return fmt.Errorf("seeder failed: could not find worker1@agrolink.com")
		}

		// 2. Buat Proyek dengan status "completed" (tanpa FarmLocation)
		project := models.Project{
			FarmerID:      farmerUser.Farmer.UserID,
			Title:         "Proyek Panen Jagung (Selesai)",
			Description:   "Proyek ini sudah selesai dan siap untuk di-review oleh petani.",
			Location:      "Sawah Seeder, Bali",
			WorkersNeeded: 1,
			StartDate:     time.Now().AddDate(0, 0, -10), // 10 hari yang lalu
			EndDate:       time.Now().AddDate(0, 0, -1),  // Kemarin
			PaymentRate:   Float64Ptr(125000),
			PaymentType:   "per_day",
			Status:        "open", // Langsung set status selesai
		}
		if err := tx.Create(&project).Error; err != nil {
			return err
		}

		// 3. Buat Kontrak
		contract := models.Contract{
			ProjectID:           &project.ID,
			FarmerID:            farmerUser.Farmer.UserID,
			WorkerID:            &workerUser.Worker.UserID,
			SignedByFarmer:      true,
			SignedBySecondParty: true,
			ContractType:        "work",
			Status:              "completed",
		}
		if err := tx.Create(&contract).Error; err != nil {
			return err
		}

		// 4. Buat Penugasan (Assignment)
		assignment := models.ProjectAssignment{
			ProjectID:  project.ID,
			WorkerID:   workerUser.Worker.UserID,
			ContractID: contract.ID,
			AgreedRate: *project.PaymentRate,
			Status:     "completed",
		}
		if err := tx.Create(&assignment).Error; err != nil {
			return err
		}

		return nil // Commit transaksi
	})

	if err != nil {
		log.Fatalf("Failed to seed completed project scenario: %v", err)
	}
}

func seedInProgressDeliveryScenario() {
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
			ContractType:        "delivery",
			FarmerID:            farmerUser.Farmer.UserID,
			DriverID:            &driverUser.Driver.UserID,
			SignedByFarmer:      true,
			SignedBySecondParty: true, // Asumsikan driver langsung setuju
			Status:              "active",
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
			DueDate:     time.Now().Add(48 * time.Hour),
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

// seedSystemSettings... (fungsi Anda yang sudah ada)
func seedSystemSettings() {
	// ... implementasi Anda ...
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
