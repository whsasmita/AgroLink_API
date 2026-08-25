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

type MitraProfileHandler struct {
	profileService services.MitraProfileService
}

func NewMitraProfileHandler(profileService services.MitraProfileService) *MitraProfileHandler {
	return &MitraProfileHandler{profileService: profileService}
}

func (h *MitraProfileHandler) CreateProfile(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user := userVal.(*models.User)

	var req dto.CreateMitraProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.profileService.CreateProfile(user.ID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Mitra profile created successfully", resp)
}

func (h *MitraProfileHandler) GetMyProfile(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user := userVal.(*models.User)

	resp, err := h.profileService.GetMyProfile(user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Mitra profile retrieved successfully", resp)
}

func (h *MitraProfileHandler) GetAllVerified(c *gin.Context) {
	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		pagination.Page = 1
		pagination.Limit = 10
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.Limit <= 0 {
		pagination.Limit = 10
	}

	resp, err := h.profileService.FindAllVerified(pagination)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch verified mitra list", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Verified mitra list fetched successfully", resp)
}

func (h *MitraProfileHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid mitra ID format", err)
		return
	}

	resp, err := h.profileService.FindByID(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Mitra profile details fetched successfully", resp)
}

func (h *MitraProfileHandler) GetPendingVerifications(c *gin.Context) {
	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		pagination.Page = 1
		pagination.Limit = 10
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.Limit <= 0 {
		pagination.Limit = 10
	}

	resp, err := h.profileService.GetPendingVerifications(pagination)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch pending verifications", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pending mitra verifications fetched successfully", resp)
}

func (h *MitraProfileHandler) ReviewVerification(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	admin := userVal.(*models.User)

	idParam := c.Param("id")
	mitraUserID, err := uuid.Parse(idParam)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid mitra user ID format", err)
		return
	}

	var req dto.ReviewMitraVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data", err)
		return
	}

	if err := h.profileService.ReviewVerification(admin.ID, mitraUserID, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Mitra verification reviewed successfully", nil)
}

// ---------------------------------------------------------------------
// MitraCooperationHandler
// ---------------------------------------------------------------------

type MitraCooperationHandler struct {
	coopService    services.MitraCooperationService
	paymentService services.PaymentService
}

func NewMitraCooperationHandler(coopService services.MitraCooperationService, paymentService services.PaymentService) *MitraCooperationHandler {
	return &MitraCooperationHandler{
		coopService:    coopService,
		paymentService: paymentService,
	}
}

func (h *MitraCooperationHandler) CreateOffer(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user := userVal.(*models.User)

	var req dto.CreateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid offer input", err)
		return
	}

	resp, err := h.coopService.CreateOffer(user.ID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Cooperation offer created successfully", resp)
}

func (h *MitraCooperationHandler) CreateApplication(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user := userVal.(*models.User)

	var req dto.CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid application input", err)
		return
	}

	resp, err := h.coopService.CreateApplication(user.ID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Cooperation application created successfully", resp)
}

func (h *MitraCooperationHandler) GetMyCooperations(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user := userVal.(*models.User)

	var pagination dto.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		pagination.Page = 1
		pagination.Limit = 10
	}
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.Limit <= 0 {
		pagination.Limit = 10
	}

	resp, err := h.coopService.FindMyCooperations(user.ID, user.Role, pagination)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch cooperations", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cooperations fetched successfully", resp)
}

func (h *MitraCooperationHandler) GetByID(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user := userVal.(*models.User)
	id := c.Param("id")

	resp, err := h.coopService.FindByID(id, user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cooperation details fetched successfully", resp)
}

func (h *MitraCooperationHandler) ReviewCooperation(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user := userVal.(*models.User)
	id := c.Param("id")

	var req dto.ReviewCooperationRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.coopService.ReviewCooperation(id, user.ID, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cooperation status updated to reviewed", nil)
}

func (h *MitraCooperationHandler) ApproveCooperation(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user := userVal.(*models.User)
	id := c.Param("id")

	var req dto.ApproveCooperationRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.coopService.ApproveCooperation(id, user.ID, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cooperation approved successfully, invoice generated", nil)
}

func (h *MitraCooperationHandler) RejectCooperation(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user := userVal.(*models.User)
	id := c.Param("id")

	var req dto.RejectCooperationRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.coopService.RejectCooperation(id, user.ID, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cooperation rejected", nil)
}

func (h *MitraCooperationHandler) InitiatePayment(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user := userVal.(*models.User)
	id := c.Param("id")

	resp, err := h.paymentService.InitiateCooperationPayment(id, user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payment initiated successfully", resp)
}

func (h *MitraCooperationHandler) ReleasePayment(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	admin := userVal.(*models.User)
	id := c.Param("id")

	if err := h.paymentService.ReleaseCooperationFunds(id, admin.ID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Cooperation funds released to farmer successfully", nil)
}

func (h *MitraCooperationHandler) CreateReview(c *gin.Context) {
	userVal, exists := c.Get("user")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User context not found", nil)
		return
	}
	user := userVal.(*models.User)
	id := c.Param("id")

	var req dto.CreateMitraReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid review input", err)
		return
	}

	if err := h.coopService.CreateReview(id, user.ID, user.Role, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Review created successfully", nil)
}
