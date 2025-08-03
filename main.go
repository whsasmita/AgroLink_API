package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/whsasmita/AgroLink_API/config"
	"github.com/whsasmita/AgroLink_API/routes"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db := config.ConnectDatabase()

	// Run migrations
	// config.AutoMigrate()
	// config.CreateIndexes()
	// config.SeedDefaultData()

	// Graceful shutdown
	defer config.CloseDatabase()

	// Handle graceful shutdown
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

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Agri Platform API is running",
		})
	})

	// API version grouping
	v1 := r.Group("/api/v1")
	{
		// Auth routes
		authGroup := v1.Group("/auth")
		routes.RegisterRoutes(authGroup, db)

		// Profile routes
		profile := v1.Group("/profile")
		{
			profile.GET("/", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Get profile - coming soon"})
			})
			profile.PUT("/", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Update profile - coming soon"})
			})
			profile.POST("/upload-photo", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Upload photo - coming soon"})
			})
		}

		// Search routes
		search := v1.Group("/search")
		{
			search.GET("/workers", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Search workers - coming soon"})
			})
			search.GET("/expeditions", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Search expeditions - coming soon"})
			})
		}

		// Project routes
		projects := v1.Group("/projects")
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

		// Contract routes
		contracts := v1.Group("/contracts")
		{
			contracts.GET("/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Get contract - coming soon"})
			})
			contracts.POST("/sign", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Sign contract - coming soon"})
			})
		}

		// Payment routes
		payments := v1.Group("/payments")
		{
			payments.POST("/proof", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Upload payment proof - coming soon"})
			})
			payments.PUT("/release", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Release payment - coming soon"})
			})
		}

		// Review routes
		reviews := v1.Group("/reviews")
		{
			reviews.POST("/", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Create review - coming soon"})
			})
			reviews.GET("/user/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Get user reviews - coming soon"})
			})
		}

		// Notification routes
		notifications := v1.Group("/notifications")
		{
			notifications.GET("/", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Get notifications - coming soon"})
			})
			notifications.PUT("/read", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Mark as read - coming soon"})
			})
		}
	}

	// Get port from configuration
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