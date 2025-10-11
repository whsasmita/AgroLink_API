package services

import (
	"encoding/json" // Tambahkan import ini
	"errors"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/datatypes"
)

type ProductService interface {
	CreateProduct(input dto.CreateProductInput, farmerID uuid.UUID) (*dto.ProductResponse, error)
	GetAllProducts() ([]dto.ProductResponse, error)
	GetProductByID(productID uuid.UUID) (*dto.ProductResponse, error)
	UpdateProduct(productID uuid.UUID, input dto.UpdateProductInput, farmerID uuid.UUID) (*dto.ProductResponse, error)
	DeleteProduct(productID uuid.UUID, farmerID uuid.UUID) error
}

type productService struct {
	productRepo repositories.ProductRepository
}

func NewProductService(repo repositories.ProductRepository) ProductService {
	return &productService{productRepo: repo}
}

// Fungsi helper untuk transformasi dari Model ke DTO
func toProductResponse(product models.Product) dto.ProductResponse {
	var imageURLs []string
	if product.ImageURLs != nil {
		// [PERBAIKAN] Unmarshal dari datatypes.JSON ke slice string
		json.Unmarshal(product.ImageURLs, &imageURLs)
	}

	return dto.ProductResponse{
		ID:          product.ID,
		FarmerID:    product.FarmerID,
		FarmerName:  product.Farmer.User.Name,
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		Category:    product.Category,
		Location:    product.Location,
		ImageURLs:   imageURLs,
	}
}

func (s *productService) CreateProduct(input dto.CreateProductInput, farmerID uuid.UUID) (*dto.ProductResponse, error) {
	// [PERBAIKAN] Marshal dari slice string ke datatypes.JSON
	imageURLsJSON, err := json.Marshal(input.ImageURLs)
	if err != nil {
		return nil, err
	}

	product := models.Product{
		ID:          uuid.New(),
		FarmerID:    farmerID,
		Title:       input.Title,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		Category:    &input.Category,
		Location:    &input.Location,
		ImageURLs:   datatypes.JSON(imageURLsJSON),
	}

	if err := s.productRepo.Create(&product); err != nil {
		return nil, err
	}
	createdProduct, _ := s.productRepo.FindByID(product.ID)
	response := toProductResponse(*createdProduct)
	return &response, nil
}

func (s *productService) GetAllProducts() ([]dto.ProductResponse, error) {
	products, err := s.productRepo.FindAll()
	if err != nil {
		return nil, err
	}
	var responses []dto.ProductResponse
	for _, p := range products {
		responses = append(responses, toProductResponse(p))
	}
	return responses, nil
}

func (s *productService) GetProductByID(productID uuid.UUID) (*dto.ProductResponse, error) {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, errors.New("product not found")
	}
	response := toProductResponse(*product)
	return &response, nil
}

func (s *productService) UpdateProduct(productID uuid.UUID, input dto.UpdateProductInput, farmerID uuid.UUID) (*dto.ProductResponse, error) {
	// 1. Ambil Produk dari Database
	// Langkah ini penting untuk mendapatkan data produk yang akan diubah.
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, errors.New("product not found")
	}

	// 2. Validasi Kepemilikan (Authorization)
	// Pastikan hanya petani pemilik yang bisa mengedit produknya.
	if product.FarmerID != farmerID {
		return nil, errors.New("forbidden: you are not the owner of this product")
	}

	// 3. Ubah Data di Model
	// Terapkan perubahan dari input DTO ke model yang sudah ada.
	if input.Title != "" {
		product.Title = input.Title
	}
	if input.Description != "" {
		product.Description = input.Description
	}
	if input.Price > 0 {
		product.Price = input.Price
	}
	if input.Stock >= 0 {
		product.Stock = input.Stock
	}
	if input.Category != "" {
		product.Category = &input.Category
	}
	if input.Location != "" {
		product.Location = &input.Location
	}
	if input.ImageURLs != nil {
		imageURLsJSON, err := json.Marshal(input.ImageURLs)
		if err != nil {
			return nil, err
		}
		product.ImageURLs = imageURLsJSON
	}

	// 4. Simpan Perubahan ke Database
	// Panggil repository untuk menyimpan model yang sudah diperbarui.
	if err := s.productRepo.Update(product); err != nil {
		return nil, err
	} 
	// 5. Kembalikan data yang sudah diperbarui
	response := toProductResponse(*product)
	return &response, nil
}


func (s *productService) DeleteProduct(productID uuid.UUID, farmerID uuid.UUID) error {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return errors.New("product not found")
	}
	if product.FarmerID != farmerID {
		return errors.New("forbidden: you are not the owner of this product")
	}
	return s.productRepo.Delete(productID)
}