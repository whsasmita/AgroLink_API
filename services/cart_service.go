package services

import (
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"gorm.io/gorm"
)

type CartService interface {
	GetCart(userID uuid.UUID) (*dto.CartResponse, error)
	AddToCart(userID uuid.UUID, input dto.AddToCartInput) (*models.Cart, error)
	UpdateCartItem(userID, productID uuid.UUID, input dto.UpdateCartInput) (*models.Cart, error)
	RemoveFromCart(userID, productID uuid.UUID) error
}

type cartService struct {
	cartRepo    repositories.CartRepository
	productRepo repositories.ProductRepository
	db          *gorm.DB
}

func NewCartService(cartRepo repositories.CartRepository, productRepo repositories.ProductRepository, db *gorm.DB) CartService {
	return &cartService{cartRepo: cartRepo, productRepo: productRepo, db: db}
}

func (s *cartService) GetCart(userID uuid.UUID) (*dto.CartResponse, error) {
	cartItems, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	var itemsResponse []dto.CartItemResponse
	var totalPrice float64
	for _, item := range cartItems {
		var firstImage string
		if item.Product.ImageURLs != nil {
			var images []string
			if err := json.Unmarshal(item.Product.ImageURLs, &images); err == nil && len(images) > 0 {
				firstImage = images[0]
			}
		}
		subtotal := float64(item.Quantity) * item.Product.Price
		itemsResponse = append(itemsResponse, dto.CartItemResponse{
			ProductID: item.ProductID, Title: item.Product.Title,
			Price: item.Product.Price, ImageURL: firstImage,
			Quantity: item.Quantity, Subtotal: subtotal,
		})
		totalPrice += subtotal
	}
	return &dto.CartResponse{Items: itemsResponse, TotalPrice: totalPrice}, nil
}

func (s *cartService) AddToCart(userID uuid.UUID, input dto.AddToCartInput) (*models.Cart, error) {
	var finalCartItem *models.Cart
	err := s.db.Transaction(func(tx *gorm.DB) error {
		product, err := s.productRepo.FindByIDForUpdate(tx, input.ProductID)
		if err != nil { return errors.New("product not found") }

		availableStock := product.Stock - product.ReservedStock
		if availableStock < input.Quantity { return errors.New("insufficient stock") }

		cartItem, err := s.cartRepo.FindByUserAndProductWithTx(tx, userID, input.ProductID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) { return err }

		if cartItem.ID != uuid.Nil {
			cartItem.Quantity += input.Quantity
			if err := s.cartRepo.Update(tx, cartItem); err != nil { return err }
			finalCartItem = cartItem
		} else {
			newCartItem := &models.Cart{UserID: userID, ProductID: input.ProductID, Quantity: input.Quantity}
			if err := s.cartRepo.Create(tx, newCartItem); err != nil { return err }
			finalCartItem = newCartItem
		}
		
		product.ReservedStock += input.Quantity
		if err := s.productRepo.UpdateStock(tx, product); err != nil { return err }
		
		return nil
	})
	return finalCartItem, err
}

func (s *cartService) UpdateCartItem(userID, productID uuid.UUID, input dto.UpdateCartInput) (*models.Cart, error) {
	var updatedCartItem *models.Cart
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if input.Quantity == 0 {
			// Jika kuantitas 0, panggil fungsi Remove (yang juga transaksional)
			// Kita perlu membuat varian RemoveFromCart yang menerima 'tx'
			return s.removeFromCartTx(tx, userID, productID)
		}

		product, err := s.productRepo.FindByIDForUpdate(tx, productID)
		if err != nil { return errors.New("product not found") }

		cartItem, err := s.cartRepo.FindByUserAndProductWithTx(tx, userID, productID)
		if err != nil { return errors.New("item not found in cart") }

		qtyDifference := input.Quantity - cartItem.Quantity
		availableStock := product.Stock - product.ReservedStock

		if qtyDifference > availableStock {
			return errors.New("insufficient stock")
		}

		cartItem.Quantity = input.Quantity
		if err := s.cartRepo.Update(tx, cartItem); err != nil { return err }

		product.ReservedStock += qtyDifference
		if err := s.productRepo.UpdateStock(tx, product); err != nil { return err }

		updatedCartItem = cartItem
		return nil
	})
	if updatedCartItem == nil && err == nil {
		// Kasus jika item dihapus (kuantitas 0)
		return nil, nil
	}
	return updatedCartItem, err
}

func (s *cartService) RemoveFromCart(userID, productID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return s.removeFromCartTx(tx, userID, productID)
	})
}

// Fungsi helper internal untuk dipanggil di dalam transaksi
func (s *cartService) removeFromCartTx(tx *gorm.DB, userID, productID uuid.UUID) error {
	cartItem, err := s.cartRepo.FindByUserAndProductWithTx(tx, userID, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return errors.New("item not found in cart") }
		return err
	}
	
	product, err := s.productRepo.FindByIDForUpdate(tx, productID)
	if err != nil { return errors.New("product not found") }
	
	if err := s.cartRepo.Delete(tx, userID, productID); err != nil { return err }
	
	product.ReservedStock -= cartItem.Quantity
	if product.ReservedStock < 0 { product.ReservedStock = 0 }
	
	return s.productRepo.UpdateStock(tx, product)
}