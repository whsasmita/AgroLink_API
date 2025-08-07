package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/handlers"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/services"
	"gorm.io/gorm"
)

// ProtectedRoutes mendaftarkan semua endpoint yang memerlukan autentikasi.
func ProtectedRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// =================================================================
	// DEPENDENCY INJECTION (Inisialisasi semua komponen di sini)
	// =================================================================
	userRepo := repositories.NewUserRepository(db)

	// Komponen untuk Autentikasi & Profil (Get)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// Komponen untuk Profil (Update)
	profileService := services.NewProfileService(userRepo)
	profileHandler := handlers.NewProfileHandler(profileService)

	// (Nantinya, inisialisasi untuk Project, dll. juga di sini)
	// projectRepo := repositories.NewProjectRepository(db)
	// projectService := services.NewProjectService(projectRepo)
	// projectHandler := handlers.NewProjectHandler(projectService)

	// =================================================================
	// ROUTE DEFINITIONS (Daftarkan semua endpoint di sini)
	// =================================================================

	// Profile Routes
	// GET /api/v1/profile -> Mengambil profil pengguna yang sedang login.
	router.GET("/profile", authHandler.GetProfile)
	// PUT /api/v1/profile -> Memperbarui profil pengguna yang sedang login.
	router.PUT("/profile", profileHandler.UpdateProfile)
	// POST /api/v1/profile/upload-photo -> (Placeholder)
	router.POST("/profile/upload-photo", profileHandler.UploadProfilePhoto)
	router.POST("/profile/details", profileHandler.UpdateRoleDetails)

	// Project Routes (Placeholder)
	projects := router.Group("/projects")
	{
		projects.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Get projects - coming soon"})
		})
		projects.POST("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Create project - coming soon"})
		})
		projects.GET("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Get project detail - coming soon"})
		})
		projects.PUT("/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Update project - coming soon"})
		})
		projects.POST("/:id/assign", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Assign worker - coming soon"})
		})
	}

	// Tambahkan juga routes lain seperti: search, contracts, payments, reviews, notifications ke sini.
}
