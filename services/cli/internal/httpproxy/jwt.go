package httpproxy

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// TokenClaims represents the decoded JWT claims
type TokenClaims struct {
	EmployeeID string `json:"employee_id"`
	OrgID      string `json:"org_id"`
	jwt.RegisteredClaims
}

var (
	// ErrInvalidToken indicates the token is malformed or has invalid signature
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken indicates the token has expired
	ErrExpiredToken = errors.New("token has expired")
)

// getJWTSecret returns the JWT secret from environment or default
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-key-change-in-production"
	}
	return []byte(secret)
}

// DecodeToken decodes and validates a JWT token, extracting employee and org IDs
func DecodeToken(tokenString string) (*TokenClaims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJWTSecret(), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
