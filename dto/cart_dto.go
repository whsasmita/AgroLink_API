package dto

import "github.com/google/uuid"

type AddToCartInput struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,gt=0"`
}

type UpdateCartInput struct {
	Quantity int `json:"quantity" binding:"required,gte=0"` // gte=0 untuk mengizinkan penghapusan
}

// Untuk menampilkan satu item di dalam keranjang
type CartItemResponse struct {
	ProductID   uuid.UUID `json:"product_id"`
	Title       string    `json:"title"`
	Price       float64   `json:"price"`
	ImageURL    string    `json:"image_url"` // Hanya gambar pertama
	Quantity    int       `json:"quantity"`
	Subtotal    float64   `json:"subtotal"`
}

// Untuk menampilkan keseluruhan isi keranjang
type CartResponse struct {
	Items      []CartItemResponse `json:"items"`
	TotalPrice float64            `json:"total_price"`
}