package handlers

import (
	"net/http"
	"sync"
	"week3_services_testing/api/internal/auth"
	"week3_services_testing/api/internal/models"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	users  map[string]models.User
	nextID int
	mu     sync.RWMutex
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		users:  make(map[string]models.User),
		nextID: 1,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest

	// 1. Validate Input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// 2. check if user already exists
	if _, exists := h.users[req.Email]; exists {
		c.JSON(http.StatusConflict, gin.H{
			"error": "User already exists",
		})
		return
	}

	// 3. Hash Password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	// 4. create user
	user := models.User{
		ID:       h.nextID,
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword, //Stored Hash
		Age:      req.Age,
	}
	h.nextID++
	h.users[user.Email] = user

	// 5. Generate token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate token",
			"message": err.Error(),
		})
		return
	}

	// 6. Return token + user
	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user, // Password not in JSON (json:"-" tag)
	})
}

// Login handles user Login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	// 1. Validate Input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 2. Find User by Email
	h.mu.RLock()
	user, exists := h.users[req.Email]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// 3. check password
	if !auth.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	// 4. Generate Token
	token, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	// 5. Return token + user
	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// GetProfile returns authenticated user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	email, _ := c.Get("email")

	h.mu.RLock()
	user, exists := h.users[email.(string)]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"id":   userID,
	})
}
