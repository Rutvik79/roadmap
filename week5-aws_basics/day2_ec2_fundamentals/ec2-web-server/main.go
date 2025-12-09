package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Hostname  string    `json:"hostname"`
}

type InfoResponse struct {
	Message string            `json:"message"`
	Headers map[string]string `json:"headers"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:    "health",
		Timestamp: time.Now(),
		Hostname:  hostname,
	})
}

func InfoHandler(w http.ResponseWriter, r *http.Request) {
	headers := make(map[string]string)
	for key, values := range r.Header {
		headers[key] = values[0]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(InfoResponse{
		Message: "Hello from EC2",
		Headers: headers,
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", HealthHandler)
	http.HandleFunc("/info", InfoHandler)

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
