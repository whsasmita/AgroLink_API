package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SuccessResponse(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(code, gin.H{
		"status":  "success",
		"message": message,
		"data":    data,
	})
}

func ErrorResponse(c *gin.Context, code int, message string, err error) {
	c.JSON(code, gin.H{
		"status":  "error",
		"message": message,
		"error":   err.Error(),
	})
}


func JSON(c *gin.Context, status int, data gin.H) {
    c.JSON(status, data)
}

func Success(c *gin.Context, message string, data interface{}) {
    JSON(c, http.StatusOK, gin.H{
        "success": true,
        "message": message,
        "data":    data,
    })
}

func Created(c *gin.Context, message string, data interface{}) {
    JSON(c, http.StatusCreated, gin.H{
        "success": true,
        "message": message,
        "data":    data,
    })
}

func BadRequest(c *gin.Context, message string) {
    JSON(c, http.StatusBadRequest, gin.H{
        "success": false,
        "error":   message,
    })
}

func Unauthorized(c *gin.Context, message string) {
    JSON(c, http.StatusUnauthorized, gin.H{
        "success": false,
        "error":   message,
    })
}

func Forbidden(c *gin.Context, message string) {
    JSON(c, http.StatusForbidden, gin.H{
        "success": false,
        "error":   message,
    })
}

func NotFound(c *gin.Context, message string) {
    JSON(c, http.StatusNotFound, gin.H{
        "success": false,
        "error":   message,
    })
}

func InternalError(c *gin.Context, message string) {
    JSON(c, http.StatusInternalServerError, gin.H{
        "success": false,
        "error":   message,
    })
}
