package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type CartHandler struct {
	cartService services.CartService
}

func NewCartHandler(s services.CartService) *CartHandler {
	return &CartHandler{cartService: s}
}

func (h *CartHandler) GetCart(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	cart, err := h.cartService.GetCart(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get cart", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Cart retrieved successfully", cart)
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	var input dto.AddToCartInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}
	currentUser := c.MustGet("user").(*models.User)
	_, err := h.cartService.AddToCart(currentUser.ID, input)
	if err != nil {
		utils.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "Item added to cart", nil)
}

func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err)
		return
	}
	var input dto.UpdateCartInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}
	currentUser := c.MustGet("user").(*models.User)
	_, err = h.cartService.UpdateCartItem(currentUser.ID, productID, input)
	if err != nil {
		utils.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Cart updated successfully", nil)
}

func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID", err)
		return
	}
	currentUser := c.MustGet("user").(*models.User)
	if err := h.cartService.RemoveFromCart(currentUser.ID, productID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove item from cart", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Item removed from cart", nil)
}