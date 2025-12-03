package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	users   = make(map[int]User)
	nextID  = 1
	usersMu sync.RWMutex
)

func main() {
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/users", handleUsers)
	http.HandleFunc("/users/", handleUser)

	fmt.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/users/")
	id, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid UserID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getUserById(w, r, id)
	case http.MethodPut:
		updateUserById(w, r, id)
	case http.MethodDelete:
		deleteUserById(w, r, id)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		getAllUsers(w, r)
	case http.MethodPost:
		createUser(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	usersMu.RLock()
	defer usersMu.RUnlock()

	usersList := make([]User, 0, len(users))
	for _, user := range users {
		usersList = append(usersList, user)
	}

	json.NewEncoder(w).Encode(usersList)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	usersMu.Lock()
	defer usersMu.Unlock()

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user.ID = nextID
	nextID++
	users[user.ID] = user

	json.NewEncoder(w).Encode(user)
}

func getUserById(w http.ResponseWriter, r *http.Request, id int) {
	usersMu.RLock()
	defer usersMu.RUnlock()

	user, exists := users[id]
	if !exists {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func updateUserById(w http.ResponseWriter, r *http.Request, id int) {
	usersMu.Lock()
	defer usersMu.Unlock()

	_, exists := users[id]
	if !exists {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user.ID = id
	users[id] = user

	json.NewEncoder(w).Encode(user)
}

func deleteUserById(w http.ResponseWriter, r *http.Request, id int) {
	usersMu.Lock()
	defer usersMu.Unlock()

	_, exists := users[id]
	if !exists {
		http.Error(w, "User not found", http.StatusBadRequest)
		return
	}

	delete(users, id)
	w.WriteHeader(http.StatusNoContent)
}
