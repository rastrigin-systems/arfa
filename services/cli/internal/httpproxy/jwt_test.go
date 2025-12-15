package httpproxy

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateTestToken creates a JWT token for testing
func generateTestToken(t *testing.T, employeeID, orgID string, expiresIn time.Duration) string {
	t.Helper()

	claims := &TokenClaims{
		EmployeeID: employeeID,
		OrgID:      orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := getJWTSecret()
	tokenString, err := token.SignedString(secret)
	require.NoError(t, err)

	return tokenString
}

func TestDecodeToken_Valid(t *testing.T) {
	token := generateTestToken(t, "emp-123", "org-456", time.Hour)

	claims, err := DecodeToken(token)
	require.NoError(t, err)
	assert.Equal(t, "emp-123", claims.EmployeeID)
	assert.Equal(t, "org-456", claims.OrgID)
}

func TestDecodeToken_Expired(t *testing.T) {
	token := generateTestToken(t, "emp-123", "org-456", -time.Hour)

	_, err := DecodeToken(token)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestDecodeToken_Invalid(t *testing.T) {
	_, err := DecodeToken("invalid.token.here")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestDecodeToken_Empty(t *testing.T) {
	_, err := DecodeToken("")
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestDecodeToken_WrongSigningMethod(t *testing.T) {
	// Create token with different signing method (RS256 instead of HS256)
	claims := &TokenClaims{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	// Use none signing method (unsigned)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := DecodeToken(tokenString)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestDecodeToken_CustomSecret(t *testing.T) {
	// Set custom JWT secret
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "custom-test-secret")
	defer os.Setenv("JWT_SECRET", originalSecret)

	// Generate token with custom secret
	token := generateTestToken(t, "emp-789", "org-012", time.Hour)

	claims, err := DecodeToken(token)
	require.NoError(t, err)
	assert.Equal(t, "emp-789", claims.EmployeeID)
	assert.Equal(t, "org-012", claims.OrgID)
}

func TestDecodeToken_MismatchedSecret(t *testing.T) {
	// Generate token with one secret
	token := generateTestToken(t, "emp-123", "org-456", time.Hour)

	// Change secret and try to decode
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "different-secret-for-validation")
	defer os.Setenv("JWT_SECRET", originalSecret)

	_, err := DecodeToken(token)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestGetJWTSecret_Default(t *testing.T) {
	// Ensure no JWT_SECRET is set
	originalSecret := os.Getenv("JWT_SECRET")
	os.Unsetenv("JWT_SECRET")
	defer os.Setenv("JWT_SECRET", originalSecret)

	secret := getJWTSecret()
	assert.Equal(t, []byte("dev-secret-key-change-in-production"), secret)
}

func TestGetJWTSecret_FromEnv(t *testing.T) {
	originalSecret := os.Getenv("JWT_SECRET")
	os.Setenv("JWT_SECRET", "my-custom-secret")
	defer os.Setenv("JWT_SECRET", originalSecret)

	secret := getJWTSecret()
	assert.Equal(t, []byte("my-custom-secret"), secret)
}
