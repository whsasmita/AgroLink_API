package handlers

import (
	"github.com/whsasmita/AgroLink_API/services"
)

type AdminHandler struct {
	payoutService services.PayoutService
}

func NewAdminHandler(payoutService services.PayoutService) *AdminHandler {
	return &AdminHandler{payoutService: payoutService}
}

// func (h *AdminHandler) GetPendingPayouts(c *gin.Context) {
// 	payouts, err := h.payoutService.GetPendingPayouts()
// 	if err != nil {
// 		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve pending payouts", err)
// 		return
// 	}
// 	utils.SuccessResponse(c, http.StatusOK, "Pending payouts retrieved successfully", payouts)
// }