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
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)

	// transactionRepo := repositories.NewTransactionRepository(db)
	// userRepo sudah diinisialisasi di atas

	// 2. Inisialisasi Service Pembayaran
	// PaymentService membutuhkan TransactionRepository dan UserRepository
	// paymentService := services.NewPaymentService(transactionRepo, userRepo)

	// 3. Inisialisasi Handler Pembayaran (untuk rute terproteksi)
	// webhookHandler := handlers.NewWebhookHandler(paymentService)

	// =================================================================
	// ROUTE DEFINITIONS (Daftarkan semua endpoint di sini)
	// =================================================================

	// Auth Routes
	authGroup := router.Group("/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)

	// Worker Routes
	workerGroup := router.Group("/workers")
	workerGroup.GET("/", workerHandler.GetWorkers)
	workerGroup.GET("/:id", workerHandler.GetWorker)

	// Driver Routes
	driverGroup := router.Group("/drivers")
	driverGroup.GET("/", driverHandler.GetDrivers)
	driverGroup.GET("/:id", driverHandler.GetDriver)

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
