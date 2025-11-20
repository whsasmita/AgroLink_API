package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/whsasmita/AgroLink_API/services"
)

type ProfitHandler struct {
	profitService services.ProfitService
}

func NewProfitHandler(profitService services.ProfitService) *ProfitHandler {
	return &ProfitHandler{profitService: profitService}
}

// GET /api/v1/admin/reports/platform-profit?start_date=2025-09-01&end_date=2025-11-30&source_type=utama
func (h *ProfitHandler) GetPlatformProfitReport(c *gin.Context) {
	const layout = "2006-01-02"

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")
	sourceType := strings.ToLower(strings.TrimSpace(c.Query("source_type"))) // "" / "utama" / "ecommerce"

	// validasi source_type
	if sourceType != "" && sourceType != "utama" && sourceType != "ecommerce" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid source_type, allowed values: '', 'utama', 'ecommerce'",
		})
		return
	}

	// default periode: 30 hari terakhir
	now := time.Now()
	var (
		start time.Time
		end   time.Time
		err   error
	)

	if startStr == "" || endStr == "" {
		end = now
		start = now.AddDate(0, 0, -30)
	} else {
		start, err = time.Parse(layout, startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid start_date format, expected YYYY-MM-DD",
			})
			return
		}

		end, err = time.Parse(layout, endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "invalid end_date format, expected YYYY-MM-DD",
			})
			return
		}
	}

	// jadikan end ke akhir hari
	end = end.AddDate(0, 0, 1).Add(-time.Nanosecond)

	total, daily, err := h.profitService.GetPlatformProfitReport(start, end, sourceType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to fetch platform profit report",
			"error":   err.Error(),
		})
		return
	}

	// untuk response, tampilkan kembali tanggal end asli (tanpa +1 hari)
	originalEnd := end.Add(time.Nanosecond).AddDate(0, 0, -1)

	c.JSON(http.StatusOK, gin.H{
		"message": "platform profit report fetched successfully",
		"data": gin.H{
			"period": gin.H{
				"start_date": start.Format(layout),
				"end_date":   originalEnd.Format(layout),
			},
			"filter": gin.H{
				"source_type": sourceType, // "" berarti semua
			},
			"total_summary": total,
			"daily_summary": daily,
		},
	})
}
