package main

// TODO pertimbangkan untuk menggunakan cloud storage
import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/handlers"
	"github.com/whsasmita/AgroLink_API/middleware"
	"github.com/whsasmita/AgroLink_API/pkg/chat"
	"github.com/whsasmita/AgroLink_API/repositories"
	"github.com/whsasmita/AgroLink_API/routes"
	"github.com/whsasmita/AgroLink_API/services"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	cfg := config.LoadConfig()

	// Connect to database
	db := config.ConnectDatabase()

	// Run migration 
	// config.RunMigrationWithReset()
	// config.AutoMigrate()
	// config.CreateIndexes()

	// Graceful shutdown 
	defer config.CloseDatabase()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("🛑 Shutting down server...")
		config.CloseDatabase()
		os.Exit(0)
	}()

	// Set Gin mode based on environment
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	r := gin.Default()

	config.InitMidtrans()

	// Menyajikan file statis dari folder 'public' di URL '/static'
	// Contoh: file di public/images/logo.png bisa diakses di http://localhost:8080/static/images/logo.png
	r.Static("/static", "./public")

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080", "http://localhost:5173", "https://goagrolink.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	chatHub := chat.NewHub()
	go chatHub.Run()


	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Agri Platform API is running",
		})
	})
	userRepo := repositories.NewUserRepository(db)
	invoiceRepo := repositories.NewInvoiceRepository(db)
	transactionRepo := repositories.NewTransactionRepository(db)
	payoutRepo := repositories.NewPayoutRepository(db)
	assignRepo := repositories.NewAssignmentRepository(db)
	projectRepo := repositories.NewProjectRepository(db)
	webhookRepo := repositories.NewWebhookLogRepository(db)
	deliveryRepo := repositories.NewDeliveryRepository(db)
	orderRepo := repositories.NewOrderRepository(db)
	productRepo := repositories.NewProductRepository(db)
	ecommPaymentRepo := repositories.NewECommercePaymentRepository(db)
	eCommercePaymentService := services.NewECommercePaymentService(
		ecommPaymentRepo, orderRepo, userRepo,productRepo, db,
	)
	paymentService := services.NewPaymentService(
		invoiceRepo,
		transactionRepo,
		payoutRepo,
		assignRepo,
		projectRepo,
		userRepo,
		deliveryRepo,
		db,
	)
	webhookHandler := handlers.NewWebhookHandler(
		paymentService,
		webhookRepo,
		eCommercePaymentService,
		invoiceRepo,
		ecommPaymentRepo,
	)
	chatHandler := handlers.NewChatHandler(chatHub)
	api := r.Group("/api")
	{
		api.POST("/webhooks/midtrans-notification", webhookHandler.HandleMidtransNotification)
		v1 := api.Group("/v1")
		{
			v1Public := v1.Group("/public")
			{
				routes.PublicRoutes(v1Public, db)
			}
			protectedGroup := v1.Group("/")
			protectedGroup.Use(middleware.AuthMiddleware(userRepo))
			{
				routes.ProtectedRoutes(protectedGroup, db, chatHandler)
				// routes.ProtectedRoutes(protectedGroup, db)
			}
		}
	}
	port := cfg.App.Port

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📍 Health check: http://localhost:%s/health", port)
	log.Printf("📖 API docs: http://localhost:%s/api/v1", port)
	log.Printf("🗄️  Database: %s@%s:%s/%s", cfg.Database.User, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	// Start server
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
