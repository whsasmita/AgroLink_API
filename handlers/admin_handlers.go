package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/dto"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/services"
	"github.com/whsasmita/AgroLink_API/utils"
)

type AdminHandler struct {
	adminService services.AdminService
}

func NewAdminHandler(s services.AdminService) *AdminHandler {
	return &AdminHandler{adminService: s}
}

// --- Dashboard ---

func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get dashboard stats", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Dashboard stats retrieved", stats)
}

// --- Payout Handlers ---

func (h *AdminHandler) GetPendingPayouts(c *gin.Context) {
	payouts, err := h.adminService.GetPendingPayouts()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get pending payouts", err)
		return
	}
	utils.SuccessResponse(c, http.StatusOK, "Pending payouts retrieved", payouts)
}

func (h *AdminHandler) MarkPayoutAsCompleted(c *gin.Context) {
	payoutID := c.Param("id")
	currentUser := c.MustGet("user").(*models.User)

	// [PERBAIKAN] Ambil file dari form-data
	// "transfer_proof_file" adalah nama 'key' yang harus dikirim dari frontend
	file, err := c.FormFile("transfer_proof_file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "File 'transfer_proof_file' not provided", err)
		return
	}

	// === Validasi File (Opsional tapi direkomendasikan) ===
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid file type. Only JPG, JPEG, PNG are allowed.", nil)
		return
	}

	// === Simpan File ke Server Lokal ===
	// Buat nama file yang unik
	newFileName := fmt.Sprintf("payout-%s-%d%s", payoutID, time.Now().UnixNano(), ext)
	
	// Pastikan folder "public/uploads/payouts" sudah ada
	savePath := filepath.Join("public", "uploads", "payouts", newFileName)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to save file", err)
		return
	}

	// Buat URL yang bisa diakses publik
	// Ganti "http://localhost:8080" dengan domain asli Anda saat production
	publicURL := fmt.Sprintf("http://localhost:8080/static/uploads/payouts/%s", newFileName)

	// Panggil service dengan URL yang baru dibuat
	err = h.adminService.MarkPayoutAsCompleted(payoutID, currentUser.ID, publicURL)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Payout marked as completed", gin.H{
		"new_status":  "completed",
		"proof_url": publicURL,
	})
}

// --- Verification Handlers ---
func (h *AdminHandler) GetPendingVerifications(c *gin.Context) {
	verifications, err := h.adminService.GetPendingVerifications()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get pending verifications", err)
		return
	}
	// Di sini kita kirim model lengkap, termasuk data user
	utils.SuccessResponse(c, http.StatusOK, "Pending verifications retrieved", verifications)
}

func (h *AdminHandler) ReviewVerification(c *gin.Context) {
	// 1. Ambil ID dari URL
	verificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid verification ID format", err)
		return
	}

	// 2. Bind body JSON
	var input dto.ReviewVerificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid input data: status (approved/rejected) is required", err)
		return
	}

	// 3. Ambil ID admin dari token
	currentUser := c.MustGet("user").(*models.User)

	// 4. Panggil service
	err = h.adminService.ReviewVerification(verificationID, input, currentUser.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	
	utils.SuccessResponse(c, http.StatusOK, "Verification reviewed successfully", nil)
}