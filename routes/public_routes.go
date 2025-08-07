package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/handlers"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/services"
	"gorm.io/gorm"
)

// PublicRoutes mendaftarkan semua endpoint yang bisa diakses secara publik.
func PublicRoutes(router *gin.RouterGroup, db *gorm.DB) {
	// =================================================================
	// DEPENDENCY INJECTION (Inisialisasi semua komponen di sini)
	// =================================================================
	userRepo := repositories.NewUserRepository(db)

	// Komponen untuk Autentikasi & Profil (Get)
	authService := services.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService)

	// (Nantinya, inisialisasi untuk Project, dll. juga di sini)
	// projectRepo := repositories.NewProjectRepository(db)
	// projectService := services.NewProjectService(projectRepo)
	// projectHandler := handlers.NewProjectHandler(projectService)

	// =================================================================
	// ROUTE DEFINITIONS (Daftarkan semua endpoint di sini)
	// =================================================================
	
	// Auth Routes
	authGroup := router.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)


	// Tambahkan juga routes lain seperti: search, contracts, payments, reviews, notifications ke sini.
}
