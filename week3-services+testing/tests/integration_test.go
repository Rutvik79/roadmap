package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"week3-services+testing/api/internal/handlers"
	"week3-services+testing/api/internal/middleware"
	"week3-services+testing/api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SetupIntegrationTestServer() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error Loading .env file")
	}
	// Add Middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.ErrorHandler())

	// Intialize Handlers
	authHandler := handlers.NewAuthHandler()
	userHandler := handlers.NewUserHandler()

	// Routes
	api := router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		// protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			protected.GET("/profile", authHandler.GetProfile)

			users := protected.Group("/users")
			{
				users.GET("", userHandler.GetAllUsers)
				users.GET("/:id", userHandler.GetUser)
				users.POST("", userHandler.CreateUser)
				users.PUT("/:id", userHandler.UpdateUser)
				users.DELETE("/:id", userHandler.DeleteUser)
			}
		}
	}

	return router
}

func TestCompleteUserFlow(t *testing.T) {
	router := SetupIntegrationTestServer()

	// Step 1: Register a User
	registerReq := models.RegisterRequest{
		Name:     "Integration Test User",
		Email:    "integration@example.com",
		Password: "password123",
		Age:      26,
	}

	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, w.Code, http.StatusCreated)

	var registerResponse models.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &registerResponse)

	require.NoError(t, err)
	token := registerResponse.Token
	assert.NotEmpty(t, token)

	// Step 2: Login with same credentials
	loginRequest := models.LoginRequest{
		Email:    "integration@example.com",
		Password: "password123",
	}
	body, _ = json.Marshal(loginRequest)
	req, _ = http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, w.Code, http.StatusOK)

	var loginResponse models.AuthResponse
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	require.NoError(t, err)
	assert.NotEmpty(t, loginResponse.Token)
	token = loginResponse.Token

	// Step 3: Get profile with token
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Step 4: Create additional users
	for i := 1; i <= 5; i++ {
		user := models.User{
			Name:  "User" + string(rune('0'+i)),
			Email: "user" + string(rune('0'+i)) + "@example.com",
			Age:   20 + i,
		}

		body, _ = json.Marshal(user)
		req, _ = http.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
	}

	// Step 5: Get all users with pagination
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/users?page1&page_size=3", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, w.Code, http.StatusOK)

	var paginatedResponse models.PaginatedResponse
	err = json.Unmarshal(w.Body.Bytes(), &paginatedResponse)
	require.NoError(t, err)
	assert.Equal(t, 1, paginatedResponse.Page)
	assert.Equal(t, 3, paginatedResponse.PageSize)
	assert.Greater(t, paginatedResponse.TotalItems, 0)

	// Step 6: Update a user
	updateUser := models.User{
		Name:  "Updated User",
		Email: "updated@example.com",
		Age:   30,
	}

	body, _ = json.Marshal(updateUser)
	req, _ = http.NewRequest(http.MethodPut, "/api/v1/users/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "Application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	// Step 7: Delete a User
	req, _ = http.NewRequest(http.MethodDelete, "/api/v1/users/2", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, w.Code, http.StatusOK)

	// Step 8: Try to access protected route without token (should fail)
	req, _ = http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, w.Code, http.StatusUnauthorized)
}

func TestCompleteAuthFlow(t *testing.T) {
	router := SetupIntegrationTestServer()

	tests := []struct {
		name           string
		email          string
		password       string
		registerStatus int
		loginStatus    int
	}{
		{
			name:           "successful flow",
			email:          "successful@example.com",
			password:       "password123",
			registerStatus: http.StatusCreated,
			loginStatus:    http.StatusOK,
		},
		{
			name:           "duplicate registration",
			email:          "successful@example.com",
			password:       "password123",
			registerStatus: http.StatusConflict,
			loginStatus:    http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registerReq := models.RegisterRequest{
				Name:     tt.name,
				Email:    tt.email,
				Password: tt.password,
				Age:      25,
			}

			body, _ := json.Marshal(registerReq)
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.registerStatus, w.Code)

			if tt.registerStatus == http.StatusCreated {
				loginReq := models.LoginRequest{
					Email:    tt.email,
					Password: tt.password,
				}
				body, _ := json.Marshal(loginReq)
				req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, w.Code, tt.loginStatus)

			}
		})
	}
}
