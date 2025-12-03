package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"week3-services+testing/api/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupAuth      func() string
		expectedStatus int
		shouldAbort    bool
	}{
		{
			name: "valid token",
			setupAuth: func() string {
				token, _ := auth.GenerateToken(1, "test@example.com")
				return "Bearer " + token
			},
			expectedStatus: http.StatusOK,
			shouldAbort:    false,
		},
		{
			name: "missing authorization header",
			setupAuth: func() string {
				return ""
			},
			expectedStatus: http.StatusUnauthorized,
			shouldAbort:    true,
		},
		{
			name: "invalid token format",
			setupAuth: func() string {
				return "invalidToken format"
			},
			expectedStatus: http.StatusUnauthorized,
			shouldAbort:    true,
		},
		{
			name: "invalid token",
			setupAuth: func() string {
				return "Bearer invalid.token.here"
			},
			expectedStatus: http.StatusUnauthorized,
			shouldAbort:    true,
		},
		{
			name: "missing bearer prefix",
			setupAuth: func() string {
				token, _ := auth.GenerateToken(1, "test@example.com")
				return token
			},
			expectedStatus: http.StatusUnauthorized,
			shouldAbort:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.Use(AuthRequired())
			router.GET("/protected", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "success",
				})
			})

			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			authHeader := tt.setupAuth()
			if authHeader != "" {
				req.Header.Set("Authorization", authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
