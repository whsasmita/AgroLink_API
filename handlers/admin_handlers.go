package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whsasmita/AgroLink_API/config"
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

	 appUrl := config.AppConfig_.App.APP_URL
	// Buat URL yang bisa diakses publik
	// Ganti "http://localhost:8080" dengan domain asli Anda saat production
	publicURL := fmt.Sprintf("%s/static/uploads/payouts/%s", appUrl, newFileName)
	// publicURL := fmt.Sprintf("http://localhost:8080/static/uploads/payouts/%s", newFileName)

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

func (h *AdminHandler) GetTransactions(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    response, err := h.adminService.GetCombinedTransactions(page, limit)
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get transactions", err)
        return
    }
    
    utils.SuccessResponse(c, http.StatusOK, "Transactions retrieved", response)
}

func (h *AdminHandler) GetAllUsers(c *gin.Context) {
	// Ambil query params dengan nilai default
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	roleFilter := c.Query("role") // filter opsional: farmer, worker, driver, general

	response, err := h.adminService.GetAllUsers(page, limit, search, roleFilter)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch users", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Users retrieved successfully", response)
}

func (h *AdminHandler) GetRevenueAnalytics(c *gin.Context) {
	// Default: 30 hari terakhir
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	// Parse query params jika ada (format YYYY-MM-DD)
	if startStr := c.Query("start_date"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = parsed
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			// Set ke akhir hari tersebut (23:59:59)
			endDate = parsed.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	stats, err := h.adminService.GetRevenueAnalytics(startDate, endDate)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get revenue analytics", err)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Revenue analytics retrieved", stats)
}

func (h *AdminHandler) ExportTransactions(c *gin.Context) {
    buffer, err := h.adminService.ExportTransactionsToExcel()
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate excel", err)
        return
    }

    // Set Headers untuk download file
    filename := fmt.Sprintf("transaksi_agrolink_%s.xlsx", time.Now().Format("20060102"))
    c.Header("Content-Disposition", "attachment; filename="+filename)
    c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    c.Header("Content-Length", fmt.Sprintf("%d", buffer.Len()))

    // Tulis buffer ke response body
    c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", buffer.Bytes())
}