package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductRepository interface {
	Create(product *models.Product) error
	FindAll() ([]models.Product, error)
	FindAllByFarmerID(farmerID uuid.UUID) ([]models.Product, error) // <-- [BARU]
	FindByID(id uuid.UUID) (*models.Product, error)
	Update(product *models.Product) error
	Delete(id uuid.UUID) error
	UpdateStock(tx *gorm.DB, product *models.Product) error
	FindByIDForUpdate(tx *gorm.DB, id uuid.UUID) (*models.Product, error)
}

type productRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

func (r *productRepository) FindAllByFarmerID(farmerID uuid.UUID) ([]models.Product, error) {
	var products []models.Product
	// Lakukan Preload untuk mendapatkan data relasi yang relevan
	err := r.db.Preload("Farmer.User").Where("farmer_id = ?", farmerID).Order("created_at DESC").Find(&products).Error
	return products, err
}

func (r *productRepository) FindAll() ([]models.Product, error) {
	var products []models.Product
	err := r.db.Preload("Farmer.User").Order("created_at DESC").Find(&products).Error
	return products, err
}

func (r *productRepository) FindByID(id uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := r.db.Preload("Farmer.User").Where("id = ?", id).First(&product).Error
	return &product, err
}

func (r *productRepository) Update(product *models.Product) error {
	return r.db.Save(product).Error
}

func (r *productRepository) Delete(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&models.Product{}).Error
}

func (r *productRepository) FindByIDForUpdate(tx *gorm.DB, id uuid.UUID) (*models.Product, error) {
	var product models.Product
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&product).Error
	return &product, err
}

func (r *productRepository) UpdateStock(tx *gorm.DB, product *models.Product) error {
	return tx.Model(product).Updates(map[string]interface{}{
		"stock":          product.Stock,
		"reserved_stock": product.ReservedStock,
	}).Error
}
