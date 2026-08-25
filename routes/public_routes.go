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
	otpRepo := repositories.NewOTPRepository(db)
	emailService := services.NewEmailService()
	geminiRepo := repositories.NewGeminiChatRepository(db)

	// Komponen untuk Autentikasi & Profil (Get)
	authService := services.NewAuthService(userRepo, otpRepo, emailService)
	authHandler := handlers.NewAuthHandler(authService)
	geminiService := services.NewGeminiChatService(geminiRepo)
	geminiHandler := handlers.NewGeminiChatHandler(geminiService)

	// (Nantinya, inisialisasi untuk Project, dll. juga di sini)
	projectRepo := repositories.NewProjectRepository(db)
	assignRepo := repositories.NewAssignmentRepository(db)
	invoiceRepo := repositories.NewInvoiceRepository(db)
	projectService := services.NewProjectService(projectRepo, assignRepo, invoiceRepo)
	projectHandler := handlers.NewProjectHandler(projectService)

	//Komponen Worker
	workerRepo := repositories.NewWorkerRepository(db)
	workerService := services.NewWorkerService(workerRepo)
	workerHandler := handlers.NewWorkerHandler(workerService)

	// Komponen Driver/Ekspedisi
	driverRepo := repositories.NewDriverRepository(db)
	driverService := services.NewDriverService(driverRepo)
	driverHandler := handlers.NewDriverHandler(driverService)

	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo, db)
	productHandler := handlers.NewProductHandler(productService)

	// =================================================================
	// ROUTE DEFINITIONS (Daftarkan semua endpoint di sini)
	// =================================================================

	// Auth Routes
	authGroup := router.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/verify-otp", authHandler.VerifyOTP)
	authGroup.POST("/resend-otp", authHandler.ResendOTP)
	authGroup.POST("/login", authHandler.Login)

	// Worker Routes
	workerGroup := router.Group("/workers")
	workerGroup.GET("/", workerHandler.GetWorkers)
	workerGroup.GET("/:id", workerHandler.GetWorker)

	// Driver Routes
	driverGroup := router.Group("/drivers")
	driverGroup.GET("/", driverHandler.GetDrivers)
	driverGroup.GET("/:id", driverHandler.GetDriver)

	aiGroup := router.Group("/ai")
	{
		aiGroup.POST("/chat", geminiHandler.ChatPublic)
	}

	// paymentRoute := router.Group("/transactions")
	// paymentRoute.POST("/webhooks/midtrans-notification", webhookHandler.HandleMidtransNotification)

	projects := router.Group("/projects")
	{
		projects.GET("/:id", projectHandler.GetProjectByID)
		projects.GET("/", projectHandler.FindAllProjects)
	}

	products := router.Group("/products")
	{
		products.GET("/", productHandler.GetAllProducts)
		products.GET("/:id", productHandler.GetProductByID)
	}

	// Tambahkan juga routes lain seperti: search, contracts, payments, reviews, notifications ke sini.
}
