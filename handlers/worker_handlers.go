// File: handlers/worker_handler.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/services" // Ubah import
)

// Sesuaikan struct agar memiliki field WorkerService
type WorkerHandler struct {
	service services.WorkerService // Ubah dependensi dari repo ke service
}

// Sesuaikan constructor agar menerima WorkerService
func NewWorkerHandler(service services.WorkerService) *WorkerHandler {
	return &WorkerHandler{service}
}

func (h *WorkerHandler) GetWorkers(c *gin.Context) {
	// 1. Parsing Query Parameters
	search := c.Query("search")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	sortBy := c.DefaultQuery("sort_by", "created_at")
	order := c.DefaultQuery("order", "desc")
	
	minDailyRate, _ := strconv.ParseFloat(c.DefaultQuery("min_daily_rate", "0"), 64)
	maxDailyRate, _ := strconv.ParseFloat(c.DefaultQuery("max_daily_rate", "0"), 64)
	minHourlyRate, _ := strconv.ParseFloat(c.DefaultQuery("min_hourly_rate", "0"), 64)
	maxHourlyRate, _ := strconv.ParseFloat(c.DefaultQuery("max_hourly_rate", "0"), 64)
	

	// 2. Memanggil Service (yang kini mengembalikan DTO)
	workers, total, err := h.service.GetWorkers(search, sortBy, order, limit, offset, minDailyRate, maxDailyRate, minHourlyRate, maxHourlyRate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workers"})
		return
	}

	// 3. Mengirim Respons yang sudah diformat
	c.JSON(http.StatusOK, gin.H{
		"data": workers,
		"pagination": gin.H{
			"limit":         limit,
			"offset":        offset,
			"total_records": total,
			"total_pages":   (total + int64(limit) - 1) / int64(limit),
		},
	})
}


func (h *WorkerHandler) GetWorker(c *gin.Context) {
    workerID := c.Param("id")

    worker, err := h.service.GetWorkerProfile(workerID)
    if err != nil {
        if err.Error() == "worker with ID " + workerID + " not found" {
            c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get worker profile"})
        return
    }

    c.JSON(http.StatusOK, worker)
}