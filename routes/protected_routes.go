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
	// [DIREVISI] DEPENDENCY INJECTION
	// Diurutkan berdasarkan dependensi: Repositories -> Services -> Handlers
	// =================================================================

	// 1. Inisialisasi semua Repositories
	userRepo := repositories.NewUserRepository(db)
	farmRepo := repositories.NewFarmRepository(db)
	workerRepo := repositories.NewWorkerRepository(db)
	projectRepo := repositories.NewProjectRepository(db)
	appRepo := repositories.NewApplicationRepository(db)
	contractRepo := repositories.NewContractRepository(db)
	assignRepo := repositories.NewAssignmentRepository(db)
	invoiceRepo := repositories.NewInvoiceRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	payoutRepo := repositories.NewPayoutRepository(db)
	notifRepo := repositories.NewNotificationRepository(db)
	reviewRepo := repositories.NewReviewRepository(db)
	deliveryRepo := repositories.NewDeliveryRepository(db)
	locationTrackRepo := repositories.NewLocationTrackRepository(db)
	driverRepo := repositories.NewDriverRepository(db)
	// workerRepo dan projectRepo sudah ada

	// 2. Inisialisasi Services
	authService := services.NewAuthService(userRepo)
	profileService := services.NewProfileService(userRepo)
	farmService := services.NewFarmService(farmRepo)
	projectService := services.NewProjectService(projectRepo, assignRepo, invoiceRepo)
	contractService := services.NewContractService(contractRepo, projectService, invoiceRepo, deliveryRepo, db)
	emailService := services.NewEmailService()
	notificationService := services.NewNotificationService(notifRepo, emailService, userRepo)
	appService := services.NewApplicationService(appRepo, projectRepo, contractRepo, assignRepo, notificationService, db)
	paymentService := services.NewPaymentService(invoiceRepo, transactionRepo, payoutRepo, assignRepo, projectRepo, userRepo, deliveryRepo, db)
	reviewService := services.NewReviewService(reviewRepo, workerRepo, projectRepo, driverRepo, deliveryRepo, db)
	deliveryService := services.NewDeliveryService(deliveryRepo, driverRepo, contractRepo, db)
	offerService := services.NewOfferService(projectRepo, contractRepo, assignRepo, userRepo, db)
	trackingService := services.NewTrackingService(locationTrackRepo, deliveryRepo)

	notifHandler := handlers.NewNotificationHandler(notifRepo)

	// 3. Inisialisasi Handlers
	authHandler := handlers.NewAuthHandler(authService)
	profileHandler := handlers.NewProfileHandler(profileService)
	farmHandler := handlers.NewFarmHandler(farmService)
	projectHandler := handlers.NewProjectHandler(projectService)
	appHandler := handlers.NewApplicationHandler(appService)
	contractHandler := handlers.NewContractHandler(contractService)
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	offerHandler := handlers.NewOfferHandler(offerService)
	reviewHandler := handlers.NewReviewHandler(reviewService, deliveryService)
	deliveryHandler := handlers.NewDeliveryHandler(deliveryService)

	// deliveryRepo sudah diinisialisasi sebelumnya

	trackingHandler := handlers.NewTrackingHandler(trackingService)

	// =================================================================
	// [DIREVISI] ROUTE DEFINITIONS
	// Dikelompokkan berdasarkan sumber daya (resource)
	// =================================================================

	// Profile Routes
	router.GET("/profile", authHandler.GetProfile)
	router.PUT("/profile", profileHandler.UpdateProfile)
	router.POST("/profile/details", profileHandler.UpdateRoleDetails)
	router.POST("/profile/upload-photo", profileHandler.UploadProfilePhoto)
	// ... (rute profil lainnya)

	// Farm Routes (Hanya untuk Petani)
	farms := router.Group("/farms")
	farms.Use(middleware.RoleMiddleware("farmer"))
	{
		farms.POST("/", farmHandler.CreateFarm)
		farms.GET("/my", farmHandler.GetMyFarms)
		farms.GET("/:id", farmHandler.GetFarmByID)
		farms.PUT("/:id", farmHandler.UpdateFarm)
		farms.DELETE("/:id", farmHandler.DeleteFarm)
	}

	// Project Routes
	projects := router.Group("/projects")
	{
		projects.POST("/", middleware.RoleMiddleware("farmer"), projectHandler.CreateProject)
		projects.GET("/my", middleware.RoleMiddleware("farmer"), projectHandler.GetMyProjects)
		
		projects.GET("/:id/applications", middleware.RoleMiddleware("farmer"), appHandler.FindApplicationsByProjectID)
		projects.POST("/:id/apply", middleware.RoleMiddleware("worker"), appHandler.ApplyToProject)
		// Rute baru untuk melepaskan dana (payout)
		projects.POST("/:id/release-payment", middleware.RoleMiddleware("farmer"), paymentHandler.ReleaseProjectPayment)
		projects.POST("/:id/workers/:workerId/review", middleware.RoleMiddleware("farmer"), reviewHandler.CreateReview)
	}

	// Application Routes
	applications := router.Group("/applications")
	{
		applications.POST("/:id/accept", middleware.RoleMiddleware("farmer"), appHandler.AcceptApplication)
	}

	// Contract Routes
	contracts := router.Group("/contracts")
	{
		contracts.GET("/my", contractHandler.GetMyContracts)
		contracts.POST("/:id/sign", middleware.RoleMiddleware("worker", "driver"), contractHandler.SignContract)
		contracts.GET("/:id/download", contractHandler.DownloadContractPDF)
	}

	// Invoice Routes (untuk memulai pembayaran)
	invoices := router.Group("/invoices")
	{
		// Endpoint untuk petani memulai pembayaran via Midtrans
		invoices.POST("/:id/initiate-payment", middleware.RoleMiddleware("farmer"), paymentHandler.InitiateInvoicePayment)
		invoices.POST("/:id/release", middleware.RoleMiddleware("farmer"), paymentHandler.ReleaseProjectPayment)
		// Endpoint untuk melihat riwayat invoice
		// invoices.GET("/", paymentHandler.GetUserInvoices)
	}

	notifications := router.Group("/notifications")
	{
		notifications.GET("/", notifHandler.GetMyNotifications)
	}

	workers := router.Group("/workers")
	{
		// Rute untuk petani menawarkan proyek langsung ke pekerja
		workers.POST("/:workerId/direct-offer", middleware.RoleMiddleware("farmer"), offerHandler.CreateDirectOffer)
	}

	deliveries := router.Group("/deliveries")
    // Middleware di sini bisa disesuaikan jika ada endpoint yang bisa diakses kedua peran
    {
        // Petani membuat permintaan pengiriman baru
        deliveries.POST("/", middleware.RoleMiddleware("farmer"), deliveryHandler.CreateDelivery)
        // Petani mencari driver yang cocok
        deliveries.GET("/:id/find-drivers", middleware.RoleMiddleware("farmer"), deliveryHandler.FindDrivers)
        // Petani memilih driver dan menawarkan kontrak
        deliveries.POST("/:id/select-driver/:driverId", middleware.RoleMiddleware("farmer"), deliveryHandler.SelectDriver)
        
		 deliveries.GET("/my", deliveryHandler.GetMyDeliveries)
        // [RUTE BARU] Petani melacak pengiriman
        deliveries.GET("/:id/track", middleware.RoleMiddleware("farmer"), trackingHandler.GetLatestLocation)
        
        // [RUTE BARU] Driver mengirim update lokasi
        deliveries.POST("/:id/location", middleware.RoleMiddleware("driver"), trackingHandler.UpdateLocation)

        // [RUTE BARU] Petani melepaskan dana pengiriman
        deliveries.POST("/:id/release-payment", middleware.RoleMiddleware("farmer"), paymentHandler.ReleaseDeliveryPayment)
    }

}
