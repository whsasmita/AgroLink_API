package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type CartRepository interface {
	FindByUserID(userID uuid.UUID) ([]models.Cart, error)
	FindByUserAndProduct(userID, productID uuid.UUID) (*models.Cart, error)
	FindByUserIDWithTx(tx *gorm.DB, userID uuid.UUID) ([]models.Cart, error)
	FindByUserAndProductWithTx(tx *gorm.DB, userID, productID uuid.UUID) (*models.Cart, error) // Baru
	Create(tx *gorm.DB, cartItem *models.Cart) error
	Update(tx *gorm.DB, cartItem *models.Cart) error
	Delete(tx *gorm.DB, userID, productID uuid.UUID) error
	ClearCart(tx *gorm.DB, userID uuid.UUID) error // <-- [BARU]
}

type cartRepository struct{ db *gorm.DB }

func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) FindByUserID(userID uuid.UUID) ([]models.Cart, error) {
	var cartItems []models.Cart
	err := r.db.Preload("Product").Where("user_id = ?", userID).Find(&cartItems).Error
	return cartItems, err
}

func (r *cartRepository) FindByUserIDWithTx(tx *gorm.DB, userID uuid.UUID) ([]models.Cart, error) {
	var cartItems []models.Cart
	// Preload Product sangat penting di sini untuk validasi stok dan pengelompokan
	err := tx.Preload("Product").Where("user_id = ?", userID).Find(&cartItems).Error
	return cartItems, err
}

func (r *cartRepository) FindByUserAndProduct(userID, productID uuid.UUID) (*models.Cart, error) {
	var cartItem models.Cart
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&cartItem).Error
	return &cartItem, err
}

// [BARU] Versi transaksional
func (r *cartRepository) FindByUserAndProductWithTx(tx *gorm.DB, userID, productID uuid.UUID) (*models.Cart, error) {
	var cartItem models.Cart
	err := tx.Where("user_id = ? AND product_id = ?", userID, productID).First(&cartItem).Error
	return &cartItem, err
}

func (r *cartRepository) Create(tx *gorm.DB, cartItem *models.Cart) error {
	return tx.Create(cartItem).Error
}

func (r *cartRepository) Update(tx *gorm.DB, cartItem *models.Cart) error {
	return tx.Save(cartItem).Error
}

func (r *cartRepository) Delete(tx *gorm.DB, userID, productID uuid.UUID) error {
	return tx.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.Cart{}).Error
}

func (r *cartRepository) ClearCart(tx *gorm.DB, userID uuid.UUID) error {
	return tx.Where("user_id = ?", userID).Delete(&models.Cart{}).Error
}