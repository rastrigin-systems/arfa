package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sergeirastrigin/ubik-enterprise/internal/auth"
)

func TestGenerateJWT_ValidClaims(t *testing.T) {
	employeeID := uuid.New()
	orgID := uuid.New()

	token, err := auth.GenerateJWT(employeeID, orgID, 24*time.Hour)

	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Greater(t, len(token), 50, "JWT token should be reasonably long")
}

func TestVerifyJWT_ValidToken(t *testing.T) {
	employeeID := uuid.New()
	orgID := uuid.New()

	// Generate a valid token
	token, err := auth.GenerateJWT(employeeID, orgID, 24*time.Hour)
	require.NoError(t, err)

	// Verify the token
	claims, err := auth.VerifyJWT(token)

	require.NoError(t, err)
	assert.Equal(t, employeeID.String(), claims.EmployeeID)
	assert.Equal(t, orgID.String(), claims.OrgID)
	assert.WithinDuration(t, time.Now().Add(24*time.Hour), claims.ExpiresAt.Time, 2*time.Second)
}

func TestVerifyJWT_ExpiredToken(t *testing.T) {
	employeeID := uuid.New()
	orgID := uuid.New()

	// Generate an expired token (negative duration)
	token, err := auth.GenerateJWT(employeeID, orgID, -1*time.Hour)
	require.NoError(t, err)

	// Verify should fail for expired token
	claims, err := auth.VerifyJWT(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "expired")
}

func TestVerifyJWT_InvalidToken(t *testing.T) {
	// Verify should fail for malformed token
	claims, err := auth.VerifyJWT("invalid.token.here")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestVerifyJWT_EmptyToken(t *testing.T) {
	claims, err := auth.VerifyJWT("")

	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestHashPassword_ValidPassword(t *testing.T) {
	password := "SecurePass123!"

	hash, err := auth.HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash, "Hash should not equal plain password")
	assert.Greater(t, len(hash), 50, "bcrypt hash should be reasonably long")
	assert.Contains(t, hash, "$2a$", "Should be a bcrypt hash")
}

func TestHashPassword_EmptyPassword(t *testing.T) {
	hash, err := auth.HashPassword("")

	assert.Error(t, err)
	assert.Empty(t, hash)
}

func TestVerifyPassword_CorrectPassword(t *testing.T) {
	password := "SecurePass123!"
	hash, err := auth.HashPassword(password)
	require.NoError(t, err)

	// Verify with correct password
	isValid := auth.VerifyPassword(password, hash)

	assert.True(t, isValid)
}

func TestVerifyPassword_IncorrectPassword(t *testing.T) {
	password := "SecurePass123!"
	hash, err := auth.HashPassword(password)
	require.NoError(t, err)

	// Verify with wrong password
	isValid := auth.VerifyPassword("WrongPassword", hash)

	assert.False(t, isValid)
}

func TestVerifyPassword_EmptyPassword(t *testing.T) {
	hash, _ := auth.HashPassword("SomePassword")

	isValid := auth.VerifyPassword("", hash)

	assert.False(t, isValid)
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	isValid := auth.VerifyPassword("SomePassword", "invalid-hash")

	assert.False(t, isValid)
}

func TestHashToken_ValidToken(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0"

	hash := auth.HashToken(token)

	assert.NotEmpty(t, hash)
	assert.NotEqual(t, token, hash)
	assert.Len(t, hash, 64, "SHA256 hash should be 64 hex characters")
}

func TestHashToken_Deterministic(t *testing.T) {
	token := "test-token-123"

	hash1 := auth.HashToken(token)
	hash2 := auth.HashToken(token)

	assert.Equal(t, hash1, hash2, "Same token should produce same hash")
}

func TestHashToken_DifferentTokens(t *testing.T) {
	hash1 := auth.HashToken("token1")
	hash2 := auth.HashToken("token2")

	assert.NotEqual(t, hash1, hash2, "Different tokens should produce different hashes")
}
