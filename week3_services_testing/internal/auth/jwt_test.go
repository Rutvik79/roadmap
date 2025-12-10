package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// os.Setenv("JWT_SECRET", "test-secret")

func TestGenerateToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	userID := 123
	email := "test@example.com"

	// Generate a valid token
	token, err := GenerateToken(userID, email)
	require.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	userID := 123
	email := "test@example.com"

	// Generate a valid token
	token, err := GenerateToken(userID, email)
	require.NoError(t, err)

	// Validate it
	claims, err := ValidateToken(token)

	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestValidateToken_Invalid(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "invalid format",
			token: "invalid.token.format",
		},
		{
			name:  "random string",
			token: "this-is-not-a-jwt-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateToken(tt.token)

			require.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestGenerateToken_TableDriven(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	tests := []struct {
		name   string
		userID int
		email  string
	}{
		{
			name:   "valid user 1",
			userID: 1,
			email:  "user1@example.com",
		},
		{
			name:   "valid user 2",
			userID: 999,
			email:  "user999@example.com",
		},
		{
			name:   "user with zero ID",
			userID: 0,
			email:  "zero@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateToken(tt.userID, tt.email)

			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// validate the token
			claims, err := ValidateToken(token)
			require.NoError(t, err)
			assert.NotNil(t, claims)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.email, claims.Email)
		})
	}
}
