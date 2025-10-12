package services

import (
	"encoding/json" // Tambahkan import ini
	"errors"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ProductService interface {
	CreateProduct(input dto.CreateProductInput, farmerID uuid.UUID) (*dto.ProductResponse, error)
	GetAllProducts() ([]dto.ProductResponse, error)
	GetMyProducts(farmerID uuid.UUID) ([]dto.ProductResponse, error)
	GetProductByID(productID uuid.UUID) (*dto.ProductResponse, error)
	UpdateProduct(productID uuid.UUID, input dto.UpdateProductInput, farmerID uuid.UUID) (*dto.ProductResponse, error)
	DeleteProduct(productID uuid.UUID, farmerID uuid.UUID) error
}

type productService struct {
	productRepo repositories.ProductRepository
	db          *gorm.DB
}

func NewProductService(repo repositories.ProductRepository, db *gorm.DB) ProductService {
	return &productService{productRepo: repo, db: db}
}

// Fungsi helper untuk transformasi dari Model ke DTO
func toPublicProductResponse(product models.Product) dto.ProductResponse {
	var imageURLs []string
	if product.ImageURLs != nil {
		json.Unmarshal(product.ImageURLs, &imageURLs)
	}
	availableStock := product.Stock - product.ReservedStock
	if availableStock < 0 {
		availableStock = 0
	}

	return dto.ProductResponse{
		ID:             product.ID,
		FarmerID:       product.FarmerID,
		FarmerName:     product.Farmer.User.Name,
		Title:          product.Title,
		Description:    product.Description,
		Price:          product.Price,
		AvailableStock: &availableStock, // Hanya tampilkan stok tersedia
		Category:       product.Category,
		Location:       product.Location,
		ImageURLs:      imageURLs,
	}
}

// [BARU] Helper untuk tampilan pemilik (petani)
func toFarmerProductResponse(product models.Product) dto.ProductResponse {
	var imageURLs []string
	if product.ImageURLs != nil {
		json.Unmarshal(product.ImageURLs, &imageURLs)
	}
	availableStock := product.Stock - product.ReservedStock
	if availableStock < 0 {
		availableStock = 0
	}

	return dto.ProductResponse{
		ID:             product.ID,
		FarmerID:       product.FarmerID,
		FarmerName:     product.Farmer.User.Name,
		Title:          product.Title,
		Description:    product.Description,
		Price:          product.Price,
		AvailableStock: &availableStock,
		Stock:          &product.Stock,          // Tampilkan stok total
		ReservedStock:  &product.ReservedStock, // Tampilkan stok direservasi
		Category:       product.Category,
		Location:       product.Location,
		ImageURLs:      imageURLs,
	}
}


func (s *productService) GetMyProducts(farmerID uuid.UUID) ([]dto.ProductResponse, error) {
	products, err := s.productRepo.FindAllByFarmerID(farmerID)
	if err != nil {
		return nil, err
	}
	// Gunakan kembali helper 'toProductResponse' yang sudah ada
	var responses []dto.ProductResponse
	for _, p := range products {
		responses = append(responses, toFarmerProductResponse(p))
	}
	return responses, nil
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
	response := toFarmerProductResponse(*createdProduct)
	return &response, nil
}

func (s *productService) GetAllProducts() ([]dto.ProductResponse, error) {
	products, err := s.productRepo.FindAll()
	if err != nil {
		return nil, err
	}
	var responses []dto.ProductResponse
	for _, p := range products {
		responses = append(responses, toPublicProductResponse(p))
	}
	return responses, nil
}

func (s *productService) GetProductByID(productID uuid.UUID) (*dto.ProductResponse, error) {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, errors.New("product not found")
	}
	response := toPublicProductResponse(*product)
	return &response, nil
}

func (s *productService) UpdateProduct(productID uuid.UUID, input dto.UpdateProductInput, farmerID uuid.UUID) (*dto.ProductResponse, error) {
	var updatedProduct *models.Product

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Ambil & Kunci baris produk
		product, err := s.productRepo.FindByIDForUpdate(tx, productID)
		if err != nil {
			return errors.New("product not found")
		}

		// 2. Validasi Kepemilikan
		if product.FarmerID != farmerID {
			return errors.New("forbidden: you are not the owner of this product")
		}

		// 3. Ubah Data
		if input.Title != "" {
			product.Title = input.Title
		}
		// ... (update field lainnya)
		if input.Stock >= 0 {
			product.Stock = input.Stock
		}

		// 4. Simpan Perubahan
		if err := s.productRepo.Update(product); err != nil { // Seharusnya Update juga menerima tx
			return err
		}

		updatedProduct = product
		return nil // Commit transaksi
	})

	if err != nil {
		return nil, err
	}

	response := toFarmerProductResponse(*updatedProduct)
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
