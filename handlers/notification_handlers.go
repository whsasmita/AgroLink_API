package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/utils"
)

type NotificationHandler struct {
	notifRepo repositories.NotificationRepository
}

func NewNotificationHandler(notifRepo repositories.NotificationRepository) *NotificationHandler {
	return &NotificationHandler{notifRepo: notifRepo}
}

func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	currentUser := c.MustGet("user").(*models.User)
	
	notifications, err := h.notifRepo.FindByUserID(currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve notifications", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Notifications retrieved successfully", notifications)
}