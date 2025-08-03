package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/models"
	"github.com/whsasmita/AgroLink_API/utils"
)

func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user, exists := c.Get("user")
        if !exists {
            utils.Forbidden(c, "Unauthorized")
            c.Abort()
            return
        }

        role := user.(*models.User).Role
        for _, allowed := range allowedRoles {
            if role == allowed {
                c.Next()
                return
            }
        }

        utils.Forbidden(c, "You do not have permission to access this resource")
        c.Abort()
    }
}
