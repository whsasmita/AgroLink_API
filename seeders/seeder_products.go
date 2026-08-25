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
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const productsSeedJSONPath = "seeders/products_seed.json"

type ProductSeedRow struct {
	ID            int      `json:"ID"`
	Title         string   `json:"Title"`
	ItemName      string   `json:"ItemName"`
	Satuan        string   `json:"Satuan"`
	FarmerEmail   string   `json:"FarmerEmail"`
	FarmerName    string   `json:"FarmerName"`
	Category      string   `json:"Category"`
	Location      string   `json:"Location"`
	Price         float64  `json:"Price"`
	Stock         int      `json:"Stock"`
	ReservedStock int      `json:"ReservedStock"`
	ImageURLs     []string `json:"ImageURLs"`
	Rating        float64  `json:"Rating"`
	CreatedAt     string   `json:"CreatedAt"`
	UpdatedAt     string   `json:"UpdatedAt"`
}

func SeedProducts(db *gorm.DB) {
	log.Println("🌱 Seeding products from products_seed.json for 74 agriculture farmers...")

	data, err := os.ReadFile(productsSeedJSONPath)
	if err != nil {
		log.Printf("Failed to open products seed file %s: %v", productsSeedJSONPath, err)
		return
	}

	var rows []ProductSeedRow
	if err := json.Unmarshal(data, &rows); err != nil {
		log.Printf("Failed to parse products seed JSON: %v", err)
		return
	}

	// Load all farmers from database
	var users []models.User
	if err := db.Where("role = ?", "farmer").Find(&users).Error; err != nil {
		log.Printf("Failed to load farmers: %v", err)
		return
	}

	userMap := make(map[string]uuid.UUID)
	for _, u := range users {
		userMap[strings.ToLower(strings.TrimSpace(u.Email))] = u.ID
	}

	var seededCount int
	for _, row := range rows {
		fEmail := strings.ToLower(strings.TrimSpace(row.FarmerEmail))
		farmerID, exists := userMap[fEmail]
		if !exists {
			log.Printf("Warning: Farmer email %s not found in database, skipping product %s", row.FarmerEmail, row.Title)
			continue
		}

		createdAt, err := time.Parse("2006-01-02 15:04:05", row.CreatedAt)
		if err != nil {
			createdAt = time.Now()
		}
		updatedAt, err := time.Parse("2006-01-02 15:04:05", row.UpdatedAt)
		if err != nil {
			updatedAt = createdAt
		}

		imgJSON, _ := json.Marshal(row.ImageURLs)
		ratingVal := row.Rating
		catVal := row.Category
		locVal := row.Location

		var existing models.Product
		if err := db.Where("farmer_id = ? AND title = ?", farmerID, row.Title).First(&existing).Error; err == nil {
			// Product exists, update it
			existing.Price = row.Price
			existing.Stock = row.Stock
			existing.Category = &catVal
			existing.Location = &locVal
			existing.ImageURLs = datatypes.JSON(imgJSON)
			existing.Rating = &ratingVal
			existing.CreatedAt = createdAt
			existing.UpdatedAt = updatedAt
			db.Save(&existing)
			seededCount++
			continue
		}

		product := models.Product{
			ID:            uuid.New(),
			Title:         row.Title,
			FarmerID:      farmerID,
			Description:   fmt.Sprintf("%s segar kualitas premium dipanen langsung dari kebun pertanian di %s.", row.ItemName, row.Location),
			Location:      &locVal,
			Category:      &catVal,
			Price:         row.Price,
			Stock:         row.Stock,
			ReservedStock: 0,
			ImageURLs:     datatypes.JSON(imgJSON),
			Rating:        &ratingVal,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		}

		if err := db.Create(&product).Error; err != nil {
			log.Printf("Failed to create product %s: %v", row.Title, err)
		} else {
			seededCount++
		}
	}

	log.Printf("✅ Seeding produk selesai: %d produk berhasil dimasukkan untuk 74 petani agriculture.", seededCount)
}
