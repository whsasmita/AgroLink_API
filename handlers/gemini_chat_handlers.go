package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type GeminiChatHandler struct {
	service services.GeminiChatService
}

func NewGeminiChatHandler(service services.GeminiChatService) *GeminiChatHandler {
	return &GeminiChatHandler{service: service}
}

func (h *GeminiChatHandler) ChatPublic(c *gin.Context) {
	var req dto.GeminiChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	result, err := h.service.ChatPublic(c.ClientIP(), req)
	if err != nil {
		if errors.Is(err, services.ErrDailyLimitExceeded) {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "Public daily limit reached", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate Gemini reply", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Gemini reply generated successfully", result)
}

func (h *GeminiChatHandler) ChatPrivate(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user, ok := currentUser.(*models.User)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process user data from context", nil)
		return
	}

	var req dto.GeminiChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	result, err := h.service.ChatPrivate(user, req)
	if err != nil {
		if errors.Is(err, services.ErrDailyLimitExceeded) {
			utils.ErrorResponse(c, http.StatusTooManyRequests, "Daily chat limit reached", err)
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate Gemini reply", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Gemini reply generated successfully", gin.H{
		"reply": result.Reply,
		"meta":  result,
	})
}

func (h *GeminiChatHandler) PremiumStatus(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user, ok := currentUser.(*models.User)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process user data from context", nil)
		return
	}

	result, err := h.service.GetPremiumStatus(user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch premium status", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Premium status retrieved successfully", result)
}

func (h *GeminiChatHandler) InitiatePremiumCheckout(c *gin.Context) {
	currentUser, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user, ok := currentUser.(*models.User)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process user data from context", nil)
		return
	}

	result, err := h.service.InitiatePremiumCheckout(user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Failed to initiate premium checkout", err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Premium checkout created successfully", result)
}
