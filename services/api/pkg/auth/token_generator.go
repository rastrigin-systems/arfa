package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateSecureToken generates a cryptographically secure random token
// suitable for password reset links. The token is 256-bit (32 bytes) and
// base64 URL-encoded for safe use in URLs.
func GenerateSecureToken() (string, error) {
	// Generate 32 bytes (256 bits) of random data
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	// Encode to base64 URL-safe format (no padding)
	token := base64.RawURLEncoding.EncodeToString(bytes)
	return token, nil
}
