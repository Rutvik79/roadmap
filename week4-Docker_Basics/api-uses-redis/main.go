package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

type Response struct {
	Message string `json:"message"`
	Cached  bool   `json:"cached"`
}

func initRedis() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", redisHost),
		Password: "",
		DB:       0,
	})

	// Test Connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Could not connect to Redis: %v\n", err)
	} else {
		log.Println("Connected to Redis successfully")
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

func cacheHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		key = "demo"
	}

	// Ṭry to get from cache
	cached, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		// Not in cache, generate new value
		value := fmt.Sprintf("Generated at %s", time.Now().Format(time.RFC3339))

		// Store in cache for 60 seconds
		err = redisClient.Set(ctx, key, value, 60*time.Second).Err()
		if err != nil {
			log.Printf("Error setting cache: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{
			Message: value,
			Cached:  false,
		})
	} else if err != nil {
		http.Error(w, "Error accessing cache", http.StatusInternalServerError)
		return
	} else {
		// Found in cache
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{
			Message: cached,
			Cached:  true,
		})
	}
}

func main() {
	fmt.Println("API using Redis over docker network")
	initRedis()

	http.HandleFunc("/health", HealthHandler)
	http.HandleFunc("/cache", cacheHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
