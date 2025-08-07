package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/utils"
)

func AuthMiddleware(userRepo repositories.UserRepository) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        fmt.Printf("data Header %s\n", authHeader)
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            utils.Unauthorized(c, "Authorization header missing or malformed")
            c.Abort()
            return
        }

        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := config.ValidateToken(tokenStr)
        if err != nil {
            utils.Unauthorized(c, "Invalid token")
            c.Abort()
            return
        }

        user, err := userRepo.FindByID(claims.Subject)
        if err != nil {
            utils.Unauthorized(c, "User not found")
            c.Abort()
            return
        }

        c.Set("user", user)
        c.Next()
    }
}
