package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"week3_services_testing/api/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthHandler_Register(t *testing.T) {
	router := setupTestRouter()
	handler := NewAuthHandler()

	router.POST("/register", handler.Register)

	tests := []struct {
		name           string
		requestBody    models.RegisterRequest
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful registration",
			requestBody: models.RegisterRequest{
				Name:     "Alice",
				Email:    "alice@example.com",
				Age:      25,
				Password: "password123",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var response models.AuthResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Token)
				assert.Equal(t, "Alice", response.User.Name)
				assert.Equal(t, "alice@example.com", response.User.Email)
			},
		},
		{
			name: "duplicate email",
			requestBody: models.RegisterRequest{
				Name:     "Bob",
				Email:    "alice@example.com",
				Age:      30,
				Password: "password123",
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]string
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "already exists")
			},
		},
		{
			name: "invalid email",
			requestBody: models.RegisterRequest{
				Name:     "Charlie",
				Email:    "invalid-email",
				Password: "password123",
				Age:      28,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
		{
			name: "password too short",
			requestBody: models.RegisterRequest{
				Name:     "Dave",
				Email:    "dave@example.com",
				Password: "123",
				Age:      36,
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal request body
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// create Request
			req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Record Response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Check response body if needed
			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	router := setupTestRouter()
	handler := NewAuthHandler()

	router.POST("/register", handler.Register)
	router.POST("/login", handler.Login)

	// First, register a user
	registerReq := models.RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
		Age:      25,
	}

	// create request
	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Require because register test already cover this test
	require.Equal(t, http.StatusCreated, w.Code)

	// now test login
	tests := []struct {
		name           string
		requestBody    models.LoginRequest
		expectedStatus int
		checkToken     bool
	}{
		{
			name: "successful login",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			checkToken:     true,
		},
		{
			name: "wrong password",
			requestBody: models.LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			checkToken:     false,
		},
		{
			name: "non-existent user",
			requestBody: models.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			checkToken:     false,
		},
		{
			name: "invalid email format",
			requestBody: models.LoginRequest{
				Email:    "invalid email",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			checkToken:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// construct request body
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// construct request
			req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// send request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// check response
			fmt.Println("w.Code !!!!!!!!", w.Code)
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkToken {
				var response models.AuthResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Token)
			}
		})
	}
}
