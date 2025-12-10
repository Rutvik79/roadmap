package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"week3_services_testing/api/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_CreateUser(t *testing.T) {
	router := setupTestRouter()
	handler := NewUserHandler()

	router.POST("/users", handler.CreateUser)

	tests := []struct {
		name           string
		user           models.User
		expectedStatus int
	}{
		{
			name: "valid user",
			user: models.User{
				Name:  "Alice",
				Email: "alice@example.com",
				Age:   25,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing name",
			user: models.User{
				Email: "test@example.com",
				Age:   25,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "age out of range",
			user: models.User{
				Name:  "Bob",
				Email: "bob@example.com",
				Age:   200,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			user: models.User{
				Name:  "Charlie",
				Email: "invalid-email",
				Age:   30,
			}, expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// construct request body
			body, err := json.Marshal(tt.user)
			require.NoError(t, err)

			// construct request
			req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// create recorder to record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var createdUser models.User
				err := json.Unmarshal(w.Body.Bytes(), &createdUser)
				require.NoError(t, err)
				assert.NotZero(t, createdUser.ID)
				assert.Equal(t, tt.user.Name, createdUser.Name)
			}
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	router := setupTestRouter()
	handler := NewUserHandler()

	router.POST("/users", handler.CreateUser)
	router.GET("/users/:id", handler.GetUser)

	// create a user first
	user := models.User{
		Name:  "Test User",
		Email: "test@example.com",
		Age:   25,
	}

	body, _ := json.Marshal(user)
	req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var createdUser models.User
	json.Unmarshal(w.Body.Bytes(), &createdUser)

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "existing user",
			userID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "non-existent user",
			userID:         "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid user ID",
			userID:         "abc",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var fetchedUser models.User
				err := json.Unmarshal(w.Body.Bytes(), &fetchedUser)
				require.NoError(t, err)
				assert.Equal(t, createdUser.ID, fetchedUser.ID)
			}
		})
	}
}

func TestUserHandler_GetAllUser(t *testing.T) {
	router := setupTestRouter()
	handler := NewUserHandler()

	router.POST("/users", handler.CreateUser)
	router.GET("/users", handler.GetAllUsers)

	users := []models.User{
		{Name: "Alice", Email: "alice@example.com", Age: 25},
		{Name: "Bob", Email: "bob@example.com", Age: 30},
		{Name: "Charlie", Email: "charlie@example.com", Age: 35},
	}

	for _, user := range users {
		body, _ := json.Marshal(user)
		req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// Get all users
	req, err := http.NewRequest(http.MethodGet, "/users", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, float64(3), response["count"])

	usersList := response["users"].([]interface{})
	assert.Len(t, usersList, 3)
}
