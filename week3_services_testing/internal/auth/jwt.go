package auth

import (
	"errors"
	"fmt"
	"os"

	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GetJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		fmt.Println("before panic")
		panic("JWT_SECRET environement variable not set")
	}
	return secret
}

// var jwtSecret = []byte(GetJWTSecret()) // TODO: CHANGE IN PRODUCTION

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token
func GenerateToken(userID int, email string) (string, error) {
	// 1. Create claims (payload)
	claims := Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 2. Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 3. Sign with secret key
	return token.SignedString([]byte(GetJWTSecret()))
}

func ValidateToken(tokenString string) (*Claims, error) {
	// parse token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(GetJWTSecret()), nil // Provide secret for verification
	})

	if err != nil {
		return nil, err // parse error or expired
	}

	// Extract claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// token = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJlbWFpbCI6ImFsaWNlQGV4YW1wbGUuY29tIiwiZXhwIjoxNzY0MDQxMzY4LCJpYXQiOjE3NjM5NTQ5Njh9.eh7Smh9-n8JpF04uzy8bdd15FCavf-4KIAdDzAy-k0s
