package config

import (
	"log"

	"github.com/whsasmita/AgroLink_API/models"
)

// AutoMigrate runs database migrations
func AutoMigrate() {
	log.Println("🔄 Running database migrations...")

	// Define migration order to handle foreign key dependencies
	migrationModels := []interface{}{
		// Base user models first
		&models.User{},
		&models.Farmer{},
		&models.Worker{},
		&models.Driver{},

		// Location and farm models
		&models.FarmLocation{},
		&models.WorkerAvailability{},

		// Project related models
		&models.Project{},
		&models.ProjectApplication{},
		&models.ProjectAssignment{},

		// Delivery models
		&models.Delivery{},
		&models.LocationTrack{},

		// Contract and transaction models
		&models.ContractTemplate{},
		&models.Contract{},
		&models.Transaction{},
		&models.PaymentMethod{},
		&models.PaymentLog{},

		// Schedule models
		&models.Schedule{},
		&models.ScheduleNotification{},
		&models.Notification{},

		// Review and support models
		&models.Review{},
		&models.SupportTicket{},
		&models.SupportMessage{},
		&models.Dispute{},

		// System models
		&models.SystemSetting{},
		&models.ActivityLog{},
		&models.UserSession{},

		// AI models
		&models.AIRecommendation{},
		&models.UserPreference{},
		&models.MLTrainingData{},
	}

	// Run migrations
	for _, model := range migrationModels {
		if err := DB.AutoMigrate(model); err != nil {
			log.Fatalf("Failed to migrate %T: %v", model, err)
		}
	}

	log.Println("✅ Database migrations completed successfully")
}

// CreateIndexes creates additional indexes for better performance
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

// SeedDefaultData inserts default system data
func SeedDefaultData() {
	log.Println("🌱 Seeding default data...")

	// Seed default contract templates
	seedContractTemplates()

	// Seed system settings
	seedSystemSettings()

	log.Println("✅ Default data seeded successfully")
}

// seedContractTemplates creates default contract templates
func seedContractTemplates() {
	templates := []models.ContractTemplate{
		{
			Name:     "Standard Work Contract",
			Category: "work",
			TemplateContent: `KONTRAK KERJA PERTANIAN

PIHAK PERTAMA (Petani):
Nama: {{farmer_name}}
Alamat: {{farmer_address}}
Telepon: {{farmer_phone}}

PIHAK KEDUA (Pekerja):
Nama: {{worker_name}}
Alamat: {{worker_address}}
Telepon: {{worker_phone}}

PASAL 1 - PEKERJAAN
Jenis Pekerjaan: {{project_type}}
Deskripsi: {{project_description}}
Lokasi: {{farm_location}}

PASAL 2 - WAKTU
Tanggal Mulai: {{start_date}}
Tanggal Selesai: {{end_date}}
Jam Kerja: {{working_hours}}

PASAL 3 - UPAH
Upah: Rp {{agreed_rate}} per {{rate_type}}
Metode Pembayaran: {{payment_method}}

PASAL 4 - KEWAJIBAN
Pekerja wajib:
- Bekerja sesuai jadwal yang disepakati
- Menggunakan alat pelindung diri
- Menjaga kualitas hasil kerja

Petani wajib:
- Menyediakan alat kerja yang diperlukan
- Membayar upah sesuai kesepakatan
- Menyediakan fasilitas istirahat

Kontrak ini berlaku sejak ditandatangani oleh kedua belah pihak.`,
			IsDefault: true,
			IsActive:  true,
		},
		{
			Name:     "Standard Delivery Contract",
			Category: "delivery",
			TemplateContent: `KONTRAK PENGIRIMAN HASIL PERTANIAN

PIHAK PERTAMA (Pengirim):
Nama: {{sender_name}}
Alamat: {{sender_address}}
Telepon: {{sender_phone}}

PIHAK KEDUA (Ekspedisi):
Nama Perusahaan: {{expedition_name}}
Alamat: {{expedition_address}}
Telepon: {{expedition_phone}}

PASAL 1 - BARANG
Jenis Barang: {{product_type}}
Berat: {{weight}} kg
Volume: {{volume}}
Kemasan: {{packaging_type}}

PASAL 2 - PENGIRIMAN
Alamat Penjemputan: {{pickup_address}}
Alamat Tujuan: {{delivery_address}}
Tanggal Penjemputan: {{pickup_date}}
Estimasi Tiba: {{estimated_delivery}}

PASAL 3 - BIAYA
Biaya Pengiriman: Rp {{delivery_price}}
Asuransi: {{insurance_info}}

PASAL 4 - TANGGUNG JAWAB
Ekspedisi bertanggung jawab atas:
- Keamanan barang selama pengiriman
- Ketepatan waktu pengiriman
- Penanganan khusus sesuai instruksi

Kontrak ini berlaku sejak ditandatangani oleh kedua belah pihak.`,
			IsDefault: true,
			IsActive:  true,
		},
	}

	for _, template := range templates {
		var existingTemplate models.ContractTemplate
		if err := DB.Where("name = ?", template.Name).First(&existingTemplate).Error; err != nil {
			// Template doesn't exist, create it
			if err := DB.Create(&template).Error; err != nil {
				log.Printf("Failed to create contract template %s: %v", template.Name, err)
			}
		}
	}
}

// seedSystemSettings creates default system settings
func seedSystemSettings() {
	settings := []models.SystemSetting{
		{
			Key:         "platform_fee_percentage",
			Value:       "5.0",
			DataType:    "number",
			Description: StringPtr("Platform fee percentage for transactions"),
			IsPublic:    true,
		},
		{
			Key:         "auto_release_days",
			Value:       "7",
			DataType:    "number",
			Description: StringPtr("Auto release payment after N days"),
			IsPublic:    true,
		},
		{
			Key:         "max_file_upload_size",
			Value:       "5242880",
			DataType:    "number",
			Description: StringPtr("Maximum file upload size in bytes (5MB)"),
			IsPublic:    true,
		},
		{
			Key:         "default_work_radius",
			Value:       "50",
			DataType:    "number",
			Description: StringPtr("Default work radius for workers in KM"),
			IsPublic:    true,
		},
		{
			Key:         "maintenance_mode",
			Value:       "false",
			DataType:    "boolean",
			Description: StringPtr("Enable maintenance mode"),
			IsPublic:    true,
		},
		{
			Key:         "app_version",
			Value:       "1.0.0",
			DataType:    "string",
			Description: StringPtr("Current application version"),
			IsPublic:    true,
		},
	}

	for _, setting := range settings {
		var existingSetting models.SystemSetting
		if err := DB.Where("key = ?", setting.Key).First(&existingSetting).Error; err != nil {
			// Setting doesn't exist, create it
			if err := DB.Create(&setting).Error; err != nil {
				log.Printf("Failed to create system setting %s: %v", setting.Key, err)
			}
		}
	}
}

// StringPtr returns a pointer to string
func StringPtr(s string) *string {
	return &s
}