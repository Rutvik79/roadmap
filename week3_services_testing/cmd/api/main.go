// @title User Management API
// @version 1.0
// @description A complete RESTful API for user management with JWT authentication
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

package main

import (
	"fmt"
	"log"
	"os"
	"week3_services_testing/api/internal/handlers"
	"week3_services_testing/api/internal/middleware"
	"week3_services_testing/api/internal/s3helper"

	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Gin Framework Basics")
	err := godotenv.Load()
	port := os.Getenv("PORT")
	if err != nil {
		log.Fatalf("Error Loading .env file")
	}

	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("S3_BUCKET_NAME environment variable is required")
	}

	// Initialize s3 client
	s3Client, err := s3helper.NewS3Client(bucketName)
	if err != nil {
		log.Fatalf("Failed to create S3 client: %v", err)
	}

	// Initialize handlers
	fileHandler := handlers.NewFileHandler(s3Client)

	// Create a Gin Router with Default middleware
	router := gin.Default()

	// Global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler())

	rateLimiter := middleware.NewRateLimiter(30, time.Minute)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "API is running",
		})
	})

	// // Simple GET endpoint
	// router.GET("/hello", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"status":  "ok",
	// 		"message": "Hello World",
	// 	})
	// })

	// // GET with parameter
	// router.GET("/hello/:name", func(c *gin.Context) {
	// 	name := c.Param("name")
	// 	c.JSON(http.StatusOK, gin.H{
	// 		// "status":  "ok",
	// 		"message": fmt.Sprintf("Hello %s!", name),
	// 	})
	// })

	// router.GET("/greet", func(c *gin.Context) {
	// 	name := c.DefaultQuery("name", "Guest")
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "Hello, " + name + "!",
	// 	})
	// })

	// Initialize UserHandler
	userHandler := handlers.NewUserHandler()
	authhandler := handlers.NewAuthHandler()

	api := router.Group("/api/v1")
	// API routes with rate limiting
	api.Use(rateLimiter.Middleware())
	{
		// public auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", authhandler.Register)
			auth.POST("/login", authhandler.Login)
		}

		files := api.Group("/files")
		{
			files.POST("/upload", fileHandler.UploadFile)
			files.GET("/list", fileHandler.ListFiles)
			files.GET("/download/*key", fileHandler.DownloadFile)
			files.DELETE("/*key", fileHandler.DeleteFile)
		}
		// Protected Routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// Profile
			protected.GET("/profile", authhandler.GetProfile)
			// change password
			protected.PUT("/change-password", authhandler.ChangePassword)
			// User routes (protected)
			userRoutes := protected.Group("/users")
			{
				userRoutes.GET("", userHandler.GetAllUsers)
				userRoutes.GET("/:id", userHandler.GetUser)
				userRoutes.POST("", userHandler.CreateUser)
				userRoutes.PUT("/:id", userHandler.UpdateUser)
				userRoutes.DELETE("/:id", userHandler.DeleteUser)
			}
		}
	}

	router.Run("0.0.0.0:" + port)
}
