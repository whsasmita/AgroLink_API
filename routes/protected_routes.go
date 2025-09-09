package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/whsasmita/AgroLink_API/handlers"
	"github.com/whsasmita/AgroLink_API/middleware"
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

	// Komponen untuk Farm (Farm Management)
	farmRepo := repositories.NewFarmRepository(db)
	farmService := services.NewFarmService(farmRepo)
	farmHandler := handlers.NewFarmHandler(farmService)

	// Komponen untuk Project
	projectRepo := repositories.NewProjectRepository(db)
	projectService := services.NewProjectService(projectRepo, farmRepo)
	projectHandler := handlers.NewProjectHandler(projectService)

	// Komponen untuk Application (Kontrak & Penugasan)
	appRepo := repositories.NewApplicationRepository(db)
	contractRepo := repositories.NewContractRepository(db)
	assignRepo := repositories.NewAssignmentRepository(db)
	appService := services.NewApplicationService(appRepo, projectRepo, contractRepo, assignRepo, db)
	appHandler := handlers.NewApplicationHandler(appService)
	contractService := services.NewContractService(contractRepo)
	contractHandler := handlers.NewContractHandler(contractService)

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

	// Farm Routes (Only for Farmers)
	farms := router.Group("/farms")
	farms.Use(middleware.RoleMiddleware("farmer")) // Only farmers can access farm routes
	{
		farms.POST("/", farmHandler.CreateFarm)      // POST /api/v1/farms
		farms.GET("/my", farmHandler.GetMyFarms)     // GET /api/v1/farms/my
		farms.GET("/:id", farmHandler.GetFarmByID)   // GET /api/v1/farms/:id
		farms.PUT("/:id", farmHandler.UpdateFarm)    // PUT /api/v1/farms/:id
		farms.DELETE("/:id", farmHandler.DeleteFarm) // DELETE /api/v1/farms/:id
		farms.GET("/", farmHandler.GetAllFarms)
		// Definisikan rute untuk proyek
		farms.POST("/projects", projectHandler.CreateProject) // GET /api/v1/farms?farmer_id=uuid
	}

	// Project Routes
	projects := router.Group("/projects")
	{
		// Endpoint ini hanya bisa diakses oleh 'farmer'
		projects.POST("/", middleware.RoleMiddleware("farmer"), projectHandler.CreateProject)

		// Endpoint ini bisa diakses oleh semua peran yang terautentikasi (terutama worker)
		projects.GET("/", projectHandler.FindAllProjects)
		projects.GET("/:id", projectHandler.GetProjectByID)

		// Endpoint ini hanya bisa diakses oleh 'worker'
		projects.POST("/:id/apply", middleware.RoleMiddleware("worker"), appHandler.ApplyToProject)
		projects.GET("/:id/applications", middleware.RoleMiddleware("farmer"), appHandler.FindApplicationsByProjectID)
	}

	// Application Routes
	applications := router.Group("/applications")
	{
		// Endpoint ini hanya bisa diakses oleh 'farmer'
		applications.POST("/:id/accept", middleware.RoleMiddleware("farmer"), appHandler.AcceptApplication)

	}

	// ... Rute Kontrak ...
	// contracts := router.Group("/contracts")
	// contracts.Use(middleware.RoleMiddleware("worker")) // Hanya worker yang bisa mengakses endpoint ini
	// {
	// 	contracts.POST("/:id/sign", contractHandler.SignContract)

	// 	 contracts.GET("/:id/download", contractHandler.DownloadContractPDF) 
	// }

	contracts := router.Group("/contracts")
	{
		// Endpoint ini hanya bisa diakses oleh 'farmer'
		contracts.POST("/:id/sign", middleware.RoleMiddleware("worker"),  contractHandler.SignContract)

		contracts.GET("/:id/download", contractHandler.DownloadContractPDF)
	}

	// Tambahkan juga routes lain seperti: search, contracts, payments, reviews, notifications ke sini.
}
