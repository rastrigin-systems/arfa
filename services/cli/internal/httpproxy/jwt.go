package httpproxy

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

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

// AuthMeResponse is the response from /api/v1/auth/me
type AuthMeResponse struct {
	ID    string `json:"id"`
	OrgID string `json:"org_id"`
	Email string `json:"email"`
	Name  string `json:"full_name"`
}

// ValidateTokenWithPlatform validates a token by calling the platform API
// This is the preferred method as it validates against the same secret
// used by the platform API server.
func ValidateTokenWithPlatform(token, platformURL string) (*TokenClaims, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}
	if platformURL == "" {
		return nil, errors.New("platform URL not configured")
	}

	// Call /api/v1/auth/me to validate token and get employee info
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", platformURL+"/api/v1/auth/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call platform API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidToken
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("platform API returned status %d", resp.StatusCode)
	}

	var authResp AuthMeResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Return claims extracted from API response
	return &TokenClaims{
		EmployeeID: authResp.ID,
		OrgID:      authResp.OrgID,
	}, nil
}
