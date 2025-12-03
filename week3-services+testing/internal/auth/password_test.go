package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	password := "mySecurePassword123"

	hash, err := HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, hash, password)

	// Hash should be different each time
	hash2, err := HashPassword(password)
	require.NoError(t, err)
	assert.NotEqual(t, hash, hash2)
}

func TestCheckPassword(t *testing.T) {
	password := "mySecurePassword123"

	hash, err := HashPassword(password)
	require.NoError(t, err)

	// Correct Password should match
	assert.True(t, CheckPassword(password, hash))

	// Wrong Password should not match
	assert.False(t, CheckPassword("wrongPassword", hash))
	assert.False(t, CheckPassword("", hash))
}

func TestCheckPassword_TableDriven(t *testing.T) {
	correctPassword := "correct123"
	hash, _ := HashPassword(correctPassword)

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: correctPassword,
			hash:     hash,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrong123",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty hash",
			password: correctPassword,
			hash:     "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPassword(tt.password, tt.hash)
			assert.Equal(t, result, tt.want)
		})
	}
}
