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
}

func NewCartService(cartRepo repositories.CartRepository, productRepo repositories.ProductRepository) CartService {
	return &cartService{cartRepo: cartRepo, productRepo: productRepo}
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
	product, err := s.productRepo.FindByID(input.ProductID)
	if err != nil {
		return nil, errors.New("product not found")
	}
	if product.Stock < input.Quantity {
		return nil, errors.New("insufficient stock")
	}
	cartItem, err := s.cartRepo.FindByUserAndProduct(userID, input.ProductID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if cartItem.ID != uuid.Nil {
		cartItem.Quantity += input.Quantity
		if product.Stock < cartItem.Quantity {
			return nil, errors.New("insufficient stock for updated quantity")
		}
		err = s.cartRepo.Update(cartItem)
	} else {
		cartItem = &models.Cart{
			UserID: userID, ProductID: input.ProductID, Quantity: input.Quantity,
		}
		err = s.cartRepo.Create(cartItem)
	}
	return cartItem, err
}

func (s *cartService) UpdateCartItem(userID, productID uuid.UUID, input dto.UpdateCartInput) (*models.Cart, error) {
	if input.Quantity == 0 {
		return nil, s.RemoveFromCart(userID, productID)
	}
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return nil, errors.New("product not found")
	}
	if product.Stock < input.Quantity {
		return nil, errors.New("insufficient stock")
	}
	cartItem, err := s.cartRepo.FindByUserAndProduct(userID, productID)
	if err != nil {
		return nil, errors.New("item not found in cart")
	}
	cartItem.Quantity = input.Quantity
	err = s.cartRepo.Update(cartItem)
	return cartItem, err
}

func (s *cartService) RemoveFromCart(userID, productID uuid.UUID) error {
	return s.cartRepo.Delete(userID, productID)
}