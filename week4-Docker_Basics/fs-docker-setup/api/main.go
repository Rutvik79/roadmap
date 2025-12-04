package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var (
	db          *sql.DB
	redisClient *redis.Client
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Postgres  string    `json:"postgres"`
	Redis     string    `json:"redis"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	fmt.Println("Full Stack Docker Application")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize Postgres
	if err := initPostgres(); err != nil {
		log.Fatal("Failed to connect to Postgres:", err)
	}
	defer db.Close()

	// Initialize Redis
	if err := initRedis(); err != nil {
		log.Fatal("Faled to connect to Redis:", err)
	}
	defer redisClient.Close()

	// setup router
	router := setupRouter()

	// Graceful shutdown
	go func() {
		if err := router.Run(":3000"); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	}()

	log.Println("Api Server running on port 3000")

	// Wait for interrup signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server....")
}

func initPostgres() error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	// Test connection
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for database... (%d/30)", i+1)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return err
	}

	log.Println("Database connected")
	return nil
}

func initRedis() error {
	redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s",
			os.Getenv("REDIS_HOST"),
			os.Getenv("REDIS_POST"),
		),
	})

	if _, err := redisClient.Ping().Result(); err != nil {
		return err
	}

	log.Println("✅ Redis connected")
	return nil
}

func setupRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/health", healthCheckHandler)
	router.GET("/users", getUsersHandler)
	router.GET("/users/:id", getUserByIDHandler)
	router.POST("/users", createUserHandler)
	router.DELETE("/cache", clearCacheHandler)

	return router
}

// Health Check Endpoint
func healthCheckHandler(c *gin.Context) {
	pgHealth := "connected"
	if err := db.Ping(); err != nil {
		pgHealth = "disconnected"
	}

	redisHealth := "connected"
	if _, err := redisClient.Ping().Result(); err != nil {
		redisHealth = "disconnected"
	}

	c.JSON(200, HealthResponse{
		Status:    "Healthy",
		Postgres:  pgHealth,
		Redis:     redisHealth,
		Timestamp: time.Now(),
	})
}

// Get all users with Redis caching
func getUsersHandler(c *gin.Context) {
	// check cache first
	cached, err := redisClient.Get("users:all").Result()
	if err == nil {
		// Cache hit
		log.Print("🔥 Cache HIT")
		var users []User
		json.Unmarshal([]byte(cached), &users)
		c.Header("X-Cache", "HIT")
		c.Header("X-Source", "Redis")
		c.JSON(200, gin.H{
			"source": "cache",
			"data":   users,
		})
	}

	// Cache miss - query database
	log.Println("❄️ Cache MISS")
	c.Header("X-Cache", "MISS")
	c.Header("X-Source", "Postgress")
	rows, err := db.Query("SELECT id, name, email, created_at, updated_at FROM users ORDER BY created_at DESC")
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt); err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
		users = append(users, user)
	}

	// Store in cache for 60 seconds
	usersJSON, _ := json.Marshal(users)
	redisClient.Set("users:all", usersJSON, 60*time.Second)

	c.JSON(200, gin.H{
		"source": "database",
		"data":   users,
	})
}

// Get user by ID with Redis Caching
func getUserByIDHandler(c *gin.Context) {
	id := c.Param("id")
	cacheKey := fmt.Sprintf("user:%s", id)

	// Check cache
	cached, err := redisClient.Get(cacheKey).Result()
	if err != nil {
		log.Printf("🔥 Cache HIT for user %s\n", id)
		var user User
		json.Unmarshal([]byte(cached), &user)
		c.Header("X-Cache", "HIT")
		c.Header("X-Source", "Redis")
		c.JSON(200, gin.H{
			"source": "cache",
			"data":   user,
		})
		return
	}

	// Cache Miss - Query database
	c.Header("X-Cache", "MISS")
	c.Header("X-Source", "Postgress")
	log.Printf("❄️ Cache MISS for users %s\n", id)
	var user User
	err = db.QueryRow(
		"SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{
			"error": "User not found",
		})
		return
	} else if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	userJSON, _ := json.Marshal(user)
	redisClient.Set(cacheKey, userJSON, 60*time.Second)

	c.JSON(200, gin.H{
		"source": "database",
		"data":   user,
	})
}

// Create User
func createUserHandler(c *gin.Context) {
	var input struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{
			"error": "Name and Email required",
		})
		return
	}

	var user User
	err := db.QueryRow(
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at, updated_at",
		input.Name, input.Email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Invalidate cache
	redisClient.Del("users:all")

	c.JSON(201, user)
}

// Clear cache endpoint
func clearCacheHandler(c *gin.Context) {
	if err := redisClient.FlushAll().Err(); err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Cache cleared",
	})
}
