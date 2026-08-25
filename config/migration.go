package config

import (
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/seeders"
	"gorm.io/gorm"
)

// List semua model untuk migrasi.
var migrationModels = []interface{}{
	// Base user models first
	// 1. Model dasar tanpa banyak dependensi
	&models.User{},
	&models.EmailOTP{},
	&models.AIChatPremiumSubscription{},
	&models.AIChatTurn{},
	// &models.SystemSetting{},

	// 2. Model profil yang bergantung pada User
	&models.Farmer{},
	&models.Worker{},
	&models.Driver{},
	&models.MitraProfile{},

	// 3. Model utama yang bergantung pada profil
	&models.Project{},
	&models.Contract{},
	&models.Delivery{},
	&models.FarmLocation{},
	&models.MitraCooperation{},

	// 4. Model transaksi & perjanjian yang bergantung pada Project/Delivery
	&models.Invoice{},
	&models.Transaction{},
	&models.Payout{},

	// 5. Model-model pendukung yang memiliki banyak relasi
	&models.ProjectApplication{},
	&models.ProjectAssignment{},

	&models.Review{},
	&models.MitraReview{},
	&models.WorkerAvailability{},
	&models.LocationTrack{},
	&models.WebhookLog{},

	// 6. Model tambahan dari ERD e-commerce
	&models.Product{},
	&models.UserVerification{}, // Pastikan ini diaktifkan jika Anda menggunakannya
	&models.Cart{},
	&models.Order{},
	&models.OrderItem{},
	&models.ECommercePayment{},
	&models.PlatformProfit{},
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
	migrationDB := db.Session(&gorm.Session{NewDB: true})
	migrationDB.Config.DisableForeignKeyConstraintWhenMigrating = true
	for _, model := range migrationModels {
		if err := migrationDB.AutoMigrate(model); err != nil {
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
	seeders.SeedUsers(db)
	seeders.SeedNewTransactions(db)
	seeders.SeedProducts(db)
	log.Println("✅ Default data seeded successfully")
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
