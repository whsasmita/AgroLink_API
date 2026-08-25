package seeders

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/whsasmita/AgroLink_API/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const seedUserJSONPath = "seeders/users_seed.json"

type SeedUserRow struct {
	Nama   string `json:"Nama"`
	NoHP   string `json:"No HP"`
	Email  string `json:"Email"`
	Alamat string `json:"Alamat"`
	Role   string `json:"Role"`

	Password  string `json:"Password"`
	CreatedAt string `json:"CreatedAt"`
	Type      string `json:"Type,omitempty"`

	// Worker-specific
	Skills            []string `json:"Skills,omitempty"`
	DailyRate         *float64 `json:"DailyRate,omitempty"`
	NationalID        string   `json:"NationalID,omitempty"`
	BankName          string   `json:"BankName,omitempty"`
	BankAccountNumber string   `json:"BankAccountNumber,omitempty"`
	BankAccountHolder string   `json:"BankAccountHolder,omitempty"`

	// Driver-specific
	PricingScheme map[string]float64 `json:"PricingScheme,omitempty"`
	VehicleTypes  []string           `json:"VehicleTypes,omitempty"`
	CurrentLat    *float64           `json:"CurrentLat,omitempty"`
	CurrentLng    *float64           `json:"CurrentLng,omitempty"`
}

func loadWorkerJobsFromTransactions() map[string]int {
	workerJobs := make(map[string]int)
	data, err := os.ReadFile("seeders/new_transaction.json")
	if err != nil {
		return workerJobs
	}
	var rows []struct {
		WorkerEmail string `json:"WorkerEmail"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		return workerJobs
	}
	for _, r := range rows {
		if r.WorkerEmail != "" {
			wEmail := strings.ToLower(strings.TrimSpace(r.WorkerEmail))
			workerJobs[wEmail]++
		}
	}
	return workerJobs
}

func calculateWorkerRating(email string, jobCount int) float64 {
	if jobCount == 0 {
		return 0.0
	}
	var sum int
	for _, c := range email {
		sum += int(c)
	}
	// Rating options between 3.5 and 4.5
	offset := (sum % 11) // 0 to 10
	return 3.5 + float64(offset)*0.1
}

func SeedUsers(db *gorm.DB) {
	log.Println("🌱 Seeding users strictly from users_seed.json...")

	data, err := os.ReadFile(seedUserJSONPath)
	if err != nil {
		log.Printf("Failed to open seed file %s: %v", seedUserJSONPath, err)
		return
	}

	var rows []SeedUserRow
	if err := json.Unmarshal(data, &rows); err != nil {
		log.Printf("Failed to parse seed JSON: %v", err)
		return
	}

	workerJobs := loadWorkerJobsFromTransactions()

	// Cache bcrypt hash for known passwords to speed up insertion while maintaining exact passwords
	pwdCache := make(map[string]string)

	for _, row := range rows {
		// 1. Cek duplikasi email sebelum insert
		var existing models.User
		if err := db.Where("email = ?", row.Email).First(&existing).Error; err == nil {
			if existing.Role == "worker" {
				jobCount := workerJobs[strings.ToLower(strings.TrimSpace(row.Email))]
				rating := calculateWorkerRating(row.Email, jobCount)
				db.Model(&models.Worker{}).Where("user_id = ?", existing.ID).Updates(map[string]interface{}{
					"total_jobs_completed": jobCount,
					"rating":               rating,
					"review_count":        jobCount,
				})
			}
			continue
		}

		// 2. Hash password sesuai isi field Password di JSON
		hashedPassword, ok := pwdCache[row.Password]
		if !ok {
			hashBytes, err := bcrypt.GenerateFromPassword([]byte(row.Password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Failed to hash password for %s: %v", row.Email, err)
				continue
			}
			hashedPassword = string(hashBytes)
			pwdCache[row.Password] = hashedPassword
		}

		// 3. Parse CreatedAt persis apa adanya
		createdAt, err := time.Parse("2006-01-02 15:04:05", row.CreatedAt)
		if err != nil {
			if t, errDate := time.Parse("2006-01-02", row.CreatedAt); errDate == nil {
				createdAt = t
			} else {
				log.Printf("Failed to parse CreatedAt '%s' for %s: %v", row.CreatedAt, row.Email, err)
				continue
			}
		}

		var phonePtr *string
		if row.NoHP != "" {
			val := row.NoHP
			phonePtr = &val
		}

		var addressPtr *string
		if row.Alamat != "" {
			val := row.Alamat
			addressPtr = &val
		}

		// 4. Buat objek User dasar
		user := models.User{
			Name:          row.Nama,
			Email:         row.Email,
			Password:      hashedPassword,
			PhoneNumber:   phonePtr,
			Role:          row.Role,
			IsActive:      true,
			EmailVerified: true,
			CreatedAt:     createdAt,
			UpdatedAt:     createdAt,
		}

		// 5. Mapping relasi profil per role tanpa memodifikasi nilai
		switch row.Role {
		case "farmer":
			user.Farmer = &models.Farmer{
				Type:      row.Type,
				Address:   addressPtr,
				CreatedAt: createdAt,
			}

		case "worker":
			var skillsJSON string
			if len(row.Skills) > 0 {
				b, err := json.Marshal(row.Skills)
				if err != nil {
					log.Printf("Failed to marshal skills for %s: %v", row.Email, err)
				} else {
					skillsJSON = string(b)
				}
			}

			jobCount := workerJobs[strings.ToLower(strings.TrimSpace(row.Email))]
			rating := calculateWorkerRating(row.Email, jobCount)
			reviewCount := jobCount

			worker := &models.Worker{
				Address:            addressPtr,
				Skills:             skillsJSON,
				DailyRate:          row.DailyRate,
				TotalJobsCompleted: jobCount,
				Rating:             rating,
				ReviewCount:        reviewCount,
				CreatedAt:          createdAt,
			}
			if row.NationalID != "" {
				val := row.NationalID
				worker.NationalID = &val
			}
			if row.BankName != "" {
				val := row.BankName
				worker.BankName = &val
			}
			if row.BankAccountNumber != "" {
				val := row.BankAccountNumber
				worker.BankAccountNumber = &val
			}
			if row.BankAccountHolder != "" {
				val := row.BankAccountHolder
				worker.BankAccountHolder = &val
			}

			user.Worker = worker

		case "driver":
			var pricingJSON string
			if row.PricingScheme != nil {
				b, err := json.Marshal(row.PricingScheme)
				if err != nil {
					log.Printf("Failed to marshal pricing scheme for %s: %v", row.Email, err)
				} else {
					pricingJSON = string(b)
				}
			}

			var vehicleTypesJSON string
			if len(row.VehicleTypes) > 0 {
				b, err := json.Marshal(row.VehicleTypes)
				if err != nil {
					log.Printf("Failed to marshal vehicle types for %s: %v", row.Email, err)
				} else {
					vehicleTypesJSON = string(b)
				}
			}

			driver := &models.Driver{
				Address:       addressPtr,
				PricingScheme: pricingJSON,
				VehicleTypes:  vehicleTypesJSON,
				CurrentLat:    row.CurrentLat,
				CurrentLng:    row.CurrentLng,
				CreatedAt:     createdAt,
			}
			if row.BankName != "" {
				val := row.BankName
				driver.BankName = &val
			}
			if row.BankAccountNumber != "" {
				val := row.BankAccountNumber
				driver.BankAccountNumber = &val
			}
			if row.BankAccountHolder != "" {
				val := row.BankAccountHolder
				driver.BankAccountHolder = &val
			}

			user.Driver = driver

		case "mitra":
			user.Mitra = &models.MitraProfile{
				JenisMitra:         "individu",
				NamaMitra:          row.Nama,
				NomorTeleponBisnis: row.NoHP,
				EmailBisnis:        row.Email,
				AlamatLengkap:      row.Alamat,
				Provinsi:           "Bali",
				KotaKabupaten:      "Denpasar",
				StatusVerifikasi:   "verified",
				NamaBank:           "BCA",
				NomorRekening:      "0000000000",
				AtasNamaRekening:   row.Nama,
				CreatedAt:          createdAt,
			}

		case "admin":
			// Admin terpisah di tabel users tanpa relasi profil tambahan

		case "general":
			// General user di tabel users tanpa relasi profil tambahan

		default:
			log.Printf("Unknown role '%s' for email %s, skipping.", row.Role, row.Email)
			continue
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to insert user %s: %v", row.Email, err)
		}
	}

	// 6. Ringkasan hasil seeding per Role
	type RoleCount struct {
		Role  string
		Count int64
	}
	var roleCounts []RoleCount
	db.Model(&models.User{}).Select("role, count(*) as count").Group("role").Order("count DESC").Scan(&roleCounts)

	var totalCount int64
	db.Model(&models.User{}).Count(&totalCount)

	log.Println("==================================================")
	log.Printf("✅ Seeding users selesai. Total user di DB: %d", totalCount)
	for _, rc := range roleCounts {
		log.Printf("   - Role '%s': %d", rc.Role, rc.Count)
	}
	log.Println("==================================================")
}
