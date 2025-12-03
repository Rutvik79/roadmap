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
	"week3-services+testing/api/internal/handlers"
	"week3-services+testing/api/internal/middleware"

	"net/http"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Gin Framework Basics")
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error Loading .env file")
	}

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

	router.Run(":8080")
}
