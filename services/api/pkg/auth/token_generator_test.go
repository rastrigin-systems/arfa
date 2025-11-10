package auth

import (
	"testing"
)

func TestGenerateSecureToken(t *testing.T) {
	t.Run("generates non-empty token", func(t *testing.T) {
		token, err := GenerateSecureToken()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if token == "" {
			t.Fatal("Expected non-empty token")
		}
	})

	t.Run("generates token of expected length (at least 40 characters for 256-bit base64)", func(t *testing.T) {
		token, err := GenerateSecureToken()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		// 32 bytes (256 bits) base64-encoded should be ~44 characters
		if len(token) < 40 {
			t.Fatalf("Expected token length >= 40, got %d", len(token))
		}
	})

	t.Run("generates unique tokens on multiple calls", func(t *testing.T) {
		token1, err := GenerateSecureToken()
		if err != nil {
			t.Fatalf("Expected no error on first call, got %v", err)
		}

		token2, err := GenerateSecureToken()
		if err != nil {
			t.Fatalf("Expected no error on second call, got %v", err)
		}

		if token1 == token2 {
			t.Fatal("Expected unique tokens, got identical ones")
		}
	})

	t.Run("generates tokens that are URL-safe", func(t *testing.T) {
		token, err := GenerateSecureToken()
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check for URL-unsafe characters that should not appear in base64 URL encoding
		unsafeChars := []rune{'+', '/', '='}
		for _, char := range unsafeChars {
			for _, tokenChar := range token {
				if tokenChar == char {
					t.Fatalf("Token contains URL-unsafe character: %c", char)
				}
			}
		}
	})

	t.Run("generates cryptographically random tokens (statistical test)", func(t *testing.T) {
		// Generate multiple tokens and verify they are statistically distinct
		tokens := make(map[string]bool)
		iterations := 100

		for i := 0; i < iterations; i++ {
			token, err := GenerateSecureToken()
			if err != nil {
				t.Fatalf("Expected no error on iteration %d, got %v", i, err)
			}
			if tokens[token] {
				t.Fatalf("Duplicate token generated at iteration %d", i)
			}
			tokens[token] = true
		}

		if len(tokens) != iterations {
			t.Fatalf("Expected %d unique tokens, got %d", iterations, len(tokens))
		}
	})
}
