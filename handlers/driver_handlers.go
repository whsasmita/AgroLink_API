// File: handlers/driver_handler.go
package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/services"
)

type DriverHandler struct {
	service services.DriverService
}

func NewDriverHandler(service services.DriverService) *DriverHandler {
	return &DriverHandler{service}
}


func (h *DriverHandler) GetDrivers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	sortBy := c.DefaultQuery("sort_by", "rating")
	order := c.DefaultQuery("order", "desc")

	drivers, total, err := h.service.GetDrivers(sortBy, order, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get drivers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": drivers,
		"pagination": gin.H{
			"limit":         limit,
			"offset":        offset,
			"total_records": total,
			"total_pages":   (total + int64(limit) - 1) / int64(limit),
		},
	})
}

func (h *DriverHandler) GetDriver(c *gin.Context) {
	driverID := c.Param("id")

	driver, err := h.service.GetDriverProfile(driverID)
	if err != nil {
		if err.Error() == "driver with ID " + driverID + " not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get driver profile"})
		return
	}

	c.JSON(http.StatusOK, driver)
}
