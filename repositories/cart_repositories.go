package repositories

import (
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/models"
	"gorm.io/gorm"
)

type CartRepository interface {
	FindByUserID(userID uuid.UUID) ([]models.Cart, error)
	FindByUserAndProduct(userID, productID uuid.UUID) (*models.Cart, error)
	Create(cartItem *models.Cart) error
	Update(cartItem *models.Cart) error
	Delete(userID, productID uuid.UUID) error
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

func (r *cartRepository) FindByUserAndProduct(userID, productID uuid.UUID) (*models.Cart, error) {
	var cartItem models.Cart
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&cartItem).Error
	return &cartItem, err
}

func (r *cartRepository) Create(cartItem *models.Cart) error {
	return r.db.Create(cartItem).Error
}

func (r *cartRepository) Update(cartItem *models.Cart) error {
	return r.db.Save(cartItem).Error
}

func (r *cartRepository) Delete(userID, productID uuid.UUID) error {
	return r.db.Where("user_id = ? AND product_id = ?", userID, productID).Delete(&models.Cart{}).Error
}