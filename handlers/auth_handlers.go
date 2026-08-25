package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type AuthHandler struct {
	AuthService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
	}
}

type RegisterRequest struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	Role        string `json:"role" binding:"required,oneof=farmer worker driver admin general mitra"`
	PhoneNumber string `json:"phone_number" binding:"required"`
}

type VerifyOTPRequest struct {
	Email   string `json:"email" binding:"required,email"`
	OTPCode string `json:"otp_code" binding:"required,len=6"`
}

type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data: "+err.Error(), err)
		return
	}

	user, err := h.AuthService.Register(req.Email, req.Password, req.Role, req.Name, req.PhoneNumber)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	resp := gin.H{
		"email":              user.Email,
		"name":               user.Name,
		"role":               user.Role,
		"email_verified":     user.EmailVerified,
		"expires_in_seconds": 600,
	}

	utils.SuccessResponse(c, http.StatusCreated, "Pendaftaran berhasil! Kode verifikasi OTP telah dikirimkan ke email Anda.", resp)
}

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data. Masukkan email dan 6-digit kode OTP.", err)
		return
	}

	user, token, err := h.AuthService.VerifyOTP(req.Email, req.OTPCode)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	userData := dto.UserResponse{
		ID:          user.ID.String(),
		Name:        user.Name,
		Email:       user.Email,
		Role:        user.Role,
		PhoneNumber: user.PhoneNumber,
	}

	resp := gin.H{
		"user":           userData,
		"token":          token,
		"email_verified": true,
	}

	utils.SuccessResponse(c, http.StatusOK, "Verifikasi email berhasil! Akun Anda telah aktif.", resp)
}

func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req ResendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Email wajib diisi dengan format valid.", err)
		return
	}

	if err := h.AuthService.ResendOTP(req.Email); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	resp := gin.H{
		"email":              req.Email,
		"expires_in_seconds": 600,
	}

	utils.SuccessResponse(c, http.StatusOK, "Kode verifikasi OTP baru telah berhasil dikirim ke email Anda.", resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	newUser, token, err := h.AuthService.Login(req.Email, req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", gin.H{
		"token":          token,
		"role":           newUser.Role,
		"email_verified": newUser.EmailVerified,
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userInterface, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}

	user, ok := userInterface.(*models.User)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process user data from context", nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User profile fetched successfully", user)
}
