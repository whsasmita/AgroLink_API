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
	Role     string `json:"role" binding:"required,oneof=farmer worker expedition admin"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input", err)
		return
	}

	// user, err := h.AuthService.Register(req.Email, req.Password, req.Role, req.Name)
	user, err := h.AuthService.Register(req.Email, req.Password, req.Role, req.Name)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to register user", err)
		return
	}

	resp := dto.UserResponse{
		ID:    user.ID.String(),
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
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

	token, err := h.AuthService.Login(req.Email, req.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", gin.H{"token": token})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := h.AuthService.GetProfile(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user profile", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "User profile fetched", user)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert to *models.User and exclude sensitive fields
	u, ok := user.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        u.ID,
		"name":      u.Name,
		"email":     u.Email,
		"role":      u.Role,
		"createdAt": u.CreatedAt,
	})
}
