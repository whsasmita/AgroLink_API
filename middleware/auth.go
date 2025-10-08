package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/utils"
)

func AuthMiddleware(userRepo repositories.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1) Authorization: Bearer <token>
		if ah := c.GetHeader("Authorization"); strings.HasPrefix(strings.ToLower(ah), "bearer ") {
			tokenString = strings.TrimSpace(ah[7:])
		}
		// 2) Jika handshake WebSocket, izinkan token via query / subprotocol
		if tokenString == "" && strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
			// a) query ?token=...
			tokenString = c.Query("token")
			// b) (opsional) Sec-WebSocket-Protocol: "Bearer, <token>"
			if tokenString == "" {
				if sp := c.GetHeader("Sec-WebSocket-Protocol"); sp != "" {
					parts := strings.Split(sp, ",")
					for _, p := range parts {
						p = strings.TrimSpace(p)
						if strings.HasPrefix(strings.ToLower(p), "bearer ") {
							tokenString = strings.TrimSpace(p[7:])
							break
						}
					}
				}
			}
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		claims, err := config.ValidateToken(tokenString)
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
