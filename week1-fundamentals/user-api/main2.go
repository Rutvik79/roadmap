// package main

// import (
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"strconv"
// 	"strings"
// 	"sync"
// )

// type User struct {
// 	ID    int    `json:"id"`
// 	Name  string `json:"name"`
// 	Email string `json:"email"`
// }

// var (
// 	users   = make(map[int]User)
// 	nextID  = 1
// 	usersMu sync.RWMutex
// )

// func main() {
// 	http.HandleFunc("/users", usersHandler)
// 	http.HandleFunc("/users/", userHandler) // trailing slash for /users/{id}
// 	http.HandleFunc("/health", healthHandler)

// 	fmt.Println("Server Starting on :8000...")
// 	log.Fatal(http.ListenAndServe(":8000", nil))
// }

// func healthHandler(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
// }

// func usersHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	switch r.Method {
// 	case http.MethodGet:
// 		getAllUsers(w, r)
// 	case http.MethodPost:
// 		createUser(w, r)
// 	default:
// 		http.Error(w, "Method not alllowed", http.StatusMethodNotAllowed)
// 	}
// }

// func userHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	// Extract ID from path
// 	path := strings.TrimPrefix(r.URL.Path, "/users/")
// 	id, err := strconv.Atoi(path)
// 	if err != nil {
// 		http.Error(w, "Invalid user ID", http.StatusBadRequest)
// 		return
// 	}

// 	switch r.Method {
// 	case http.MethodGet:
// 		getUser(w, r, id)
// 	case http.MethodPut:
// 		updateUser(w, r, id)
// 	case http.MethodDelete:
// 		deleteUser(w, r, id)
// 	default:
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 	}
// }

// func getAllUsers(w http.ResponseWriter, r *http.Request) {
// 	usersMu.RLock()
// 	defer usersMu.RUnlock()

// 	userList := make([]User, 0, len(users))
// 	for _, user := range users {
// 		userList = append(userList, user)
// 	}

// 	json.NewEncoder(w).Encode(userList)
// }

// func getUser(w http.ResponseWriter, r *http.Request, id int) {
// 	usersMu.RLock()
// 	defer usersMu.RUnlock()

// 	user, exists := users[id]
// 	if !exists {
// 		http.Error(w, "User Not Found", http.StatusNotFound)
// 		return
// 	}

// 	json.NewEncoder(w).Encode(user)
// }

// func createUser(w http.ResponseWriter, r *http.Request) {
// 	var user User
// 	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	usersMu.Lock()
// 	user.ID = nextID
// 	nextID++
// 	users[user.ID] = user
// 	usersMu.Unlock()

// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(user)
// }

// func updateUser(w http.ResponseWriter, r *http.Request, id int) {
// 	usersMu.Lock()
// 	defer usersMu.Unlock()

// 	if _, exists := users[id]; !exists {
// 		http.Error(w, "User not found", http.StatusNotFound)
// 		return
// 	}

// 	var user User
// 	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
// 		http.Error(w, "Invlid Request Body", http.StatusBadRequest)
// 		return
// 	}

// 	user.ID = id
// 	users[id] = user

// 	json.NewEncoder(w).Encode(user)
// }

// func deleteUser(w http.ResponseWriter, r *http.Request, id int) {
// 	usersMu.Lock()
// 	defer usersMu.Unlock()

// 	if _, exists := users[id]; !exists {
// 		http.Error(w, "User not found", http.StatusNotFound)
// 		return
// 	}

// 	delete(users, id)
// 	w.WriteHeader(http.StatusNoContent)
// }