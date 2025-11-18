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
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=farmer worker driver admin general"`
	PhoneNumber string 	`json:"phone_number" binding:"required"`
}

// TODO pertimbangkan untuk menggunakan email verifikasi
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	// user, err := h.AuthService.Register(req.Email, req.Password, req.Role, req.Name)
	user, token ,err := h.AuthService.Register(req.Email, req.Password, req.Role, req.Name, req.PhoneNumber)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to register user", err)
		return
	}

	userData := dto.UserResponse{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
		PhoneNumber: user.PhoneNumber,
	}
	resp := gin.H{
		"user":  userData,
		"token": token,
	}

	utils.SuccessResponse(c, http.StatusCreated, "User registered successfully", resp)
}


type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	newUser,token, err := h.AuthService.Login(req.Email, req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", gin.H{"token": token, "role" : newUser.Role })
}

// Di dalam file handlers/auth_handlers.go

func (h *AuthHandler) GetProfile(c *gin.Context) {
    // 1. Ambil data dengan kunci "user" yang sudah di-set oleh middleware Anda.
    userInterface, exists := c.Get("user")
    if !exists {
        // Ini seharusnya tidak terjadi jika middleware berjalan dengan benar,
        // tapi ini adalah pemeriksaan keamanan yang baik.
        utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
        return
    }

    // 2. Lakukan type assertion dari interface{} ke *models.User
    user, ok := userInterface.(*models.User)
    if !ok {
        // Error ini terjadi jika data yang disimpan di konteks bukan *models.User
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process user data from context", nil)
        return
    }

    utils.SuccessResponse(c, http.StatusOK, "User profile fetched successfully", user)
}

