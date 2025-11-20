package dto

import "github.com/google/uuid"

type CreateProductInput struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description" binding:"required"`
	Price       float64  `json:"price" binding:"required,gt=0"`
	Stock       int      `json:"stock" binding:"required,gte=0"`
	Category    string   `json:"category"`
	Location    string   `json:"location"`
	ImageURLs   []string `json:"image_urls"`
}

type UpdateProductInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Price       float64  `json:"price,omitempty"`
	Stock       int      `json:"stock,omitempty"`
	Category    string   `json:"category"`
	Location    string   `json:"location"`
	ImageURLs   []string `json:"image_urls"`
}

type ProductResponse struct {
	ID             uuid.UUID `json:"id"`
	FarmerID       uuid.UUID `json:"farmer_id"`
	FarmerName     string    `json:"farmer_name"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Rating         *float64  `json:"rating"`
	Price          float64   `json:"price"`
	AvailableStock *int      `json:"available_stock,omitempty"` // Untuk pembeli
	Stock          *int      `json:"stock,omitempty"`           // Stok total untuk petani
	ReservedStock  *int      `json:"reserved_stock,omitempty"`  // Stok direservasi, hanya untuk petani
	Category       *string   `json:"category"`
	Location       *string   `json:"location"`
	ImageURLs      []string  `json:"image_urls"`
}
