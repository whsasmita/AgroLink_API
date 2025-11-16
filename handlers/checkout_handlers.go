package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type CheckoutHandler struct {
	checkoutService services.CheckoutService
}

func NewCheckoutHandler(s services.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{checkoutService: s}
}

// CreateOrders adalah handler untuk proses checkout.
// Ini akan mengambil keranjang pengguna dan mengubahnya menjadi pesanan.
func (h *CheckoutHandler) CreateOrders(c *gin.Context) {
	// 1. Ambil pengguna yang sedang login dari context
	currentUser, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}
	user := currentUser.(*models.User)

	// 2. Panggil CheckoutService untuk memproses keranjang
	// Service ini akan mengembalikan DTO respons pembayaran
	paymentResponse, err := h.checkoutService.CreateOrdersFromCart(user.ID)
	if err != nil {
		// Tangani error, seperti "cart is empty" atau "insufficient stock"
		utils.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		return
	}

	// 3. Kembalikan snap_token ke frontend
	utils.SuccessResponse(c, http.StatusCreated, "Checkout successful, please proceed to payment", paymentResponse)
}

func (h *CheckoutHandler) DirectCheckout(c *gin.Context) {
	var input services.DirectCheckoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input: product_id and quantity are required", err)
		return
	}
	currentUser := c.MustGet("user").(*models.User)
	paymentResponse, err := h.checkoutService.CreateDirectCheckout(currentUser.ID, input)
	if err != nil {
		utils.ErrorResponse(c, http.StatusConflict, err.Error(), nil)
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, "Direct checkout successful, please proceed to payment", paymentResponse)
}