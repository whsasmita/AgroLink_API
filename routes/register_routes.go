package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/handlers"
	"github.com/whsasmita/AgroLink_API/middleware"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/services"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.RouterGroup, db *gorm.DB) {
	userRepo := repositories.NewUserRepository(db)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// Auth
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	// Protected
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(userRepo))

	{
		protected.GET("/me", authHandler.Me)

		// Role-based routes
		protected.GET("/farmer-only", middleware.RoleMiddleware("farmer"), func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Halo Petani!"})
		})
		protected.GET("/worker-only", middleware.RoleMiddleware("worker"), func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Halo Pekerja!"})
		})
	}
}
