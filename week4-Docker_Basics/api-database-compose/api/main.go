package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

var (
	db          *sql.DB
	redisClient *redis.Client
	ctx         = context.Background()
)

type HealthStatus struct {
	Status   string `json:"status"`
	Database string `json:"database"`
	Redis    string `json:"redis"`
	Version  string `json:"version"`
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func initDB() error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "secret"),
		getEnv("DB_NAME", "myapp"),
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	// Retry connection with timeout
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

	// Create table
	createTable := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        email VARCHAR(100) UNIQUE NOT NULL,
        created_at TIMESTAMP DEFAULT NOW()
    )`
	_, err = db.Exec(createTable)
	return err
}

func initRedis() error {
	redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s",
			getEnv("REDIS_HOST", "localhost"),
			getEnv("REDIS_PORT", "6379"),
		),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})

	// Test connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Redis not available: %v", err)
		return err
	}

	log.Println("Redis connected")
	return nil
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	health := HealthStatus{
		Status:  "healthy",
		Version: "1.0.0",
	}

	// Check database
	if err := db.Ping(); err != nil {
		health.Database = "unhealthy"
		health.Status = "degraded"
	} else {
		health.Database = "healthy"
	}

	// Check Redis
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		health.Redis = "unhealthy"
		health.Status = "degraded"
	} else {
		health.Redis = "healthy"
	}

	w.Header().Set("Content-Type", "application/json")
	if health.Status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(health)
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Try cache
	cached, err := redisClient.Get(ctx, "users:all").Result()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write([]byte(cached))
		return
	}

	// Query database
	rows, err := db.Query("SELECT id, name, email FROM users ORDER BY id")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	// Cache for 30 seconds
	data, _ := json.Marshal(users)
	redisClient.Set(ctx, "users:all", data, 30*time.Second)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(data)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := db.QueryRow(
		"INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id",
		user.Name, user.Email,
	).Scan(&user.ID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Invalidate cache
	redisClient.Del(ctx, "users:all")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func main() {
	if err := initDB(); err != nil {
		log.Fatal("Database initialization failed:", err)
	}
	defer db.Close()

	if err := initRedis(); err != nil {
		log.Println("Redis initialization failed:", err)
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getUsersHandler(w, r)
		case http.MethodPost:
			createUserHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
