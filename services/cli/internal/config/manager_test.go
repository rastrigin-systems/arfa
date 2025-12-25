package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestJWT creates a JWT token for testing with the given claims.
// The signature is fake but that's fine since we don't validate signatures client-side.
func createTestJWT(employeeID, orgID string, expiresAt time.Time) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

	claims := map[string]interface{}{
		"employee_id": employeeID,
		"org_id":      orgID,
		"exp":         expiresAt.Unix(),
	}
	claimsJSON, _ := json.Marshal(claims)
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Fake signature - not validated client-side
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))

	return fmt.Sprintf("%s.%s.%s", header, payload, signature)
}

func TestManager_SaveAndLoad(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()

	m := &Manager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	// Create config with valid JWT
	token := createTestJWT("employee-456", "org-789", time.Now().Add(24*time.Hour))
	config := &Config{
		PlatformURL: "https://test.example.com",
		Token:       token,
	}

	// Save config
	err := m.Save(config)
	require.NoError(t, err)

	// Load config
	loaded, err := m.Load()
	require.NoError(t, err)

	// Verify
	assert.Equal(t, config.PlatformURL, loaded.PlatformURL)
	assert.Equal(t, config.Token, loaded.Token)

	// Verify claims can be extracted
	claims, err := loaded.GetClaims()
	require.NoError(t, err)
	assert.Equal(t, "employee-456", claims.EmployeeID)
	assert.Equal(t, "org-789", claims.OrgID)
}

func TestManager_LoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()

	m := &Manager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	// Load non-existent config should return empty config
	loaded, err := m.Load()
	require.NoError(t, err)
	assert.Equal(t, "", loaded.Token)
	assert.Equal(t, "", loaded.PlatformURL)
}

func TestManager_IsAuthenticated(t *testing.T) {
	tempDir := t.TempDir()

	m := &Manager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	// Initially not authenticated
	authenticated, err := m.IsAuthenticated()
	require.NoError(t, err)
	assert.False(t, authenticated)

	// Save authenticated config with valid JWT
	token := createTestJWT("employee-id", "org-id", time.Now().Add(24*time.Hour))
	config := &Config{
		PlatformURL: "https://test.example.com",
		Token:       token,
	}
	err = m.Save(config)
	require.NoError(t, err)

	// Now should be authenticated
	authenticated, err = m.IsAuthenticated()
	require.NoError(t, err)
	assert.True(t, authenticated)
}

func TestManager_IsAuthenticated_InvalidToken(t *testing.T) {
	tempDir := t.TempDir()

	m := &Manager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	// Save config with invalid token
	config := &Config{
		PlatformURL: "https://test.example.com",
		Token:       "not-a-valid-jwt",
	}
	err := m.Save(config)
	require.NoError(t, err)

	// Should not be authenticated due to invalid token
	authenticated, err := m.IsAuthenticated()
	require.NoError(t, err)
	assert.False(t, authenticated)
}

func TestManager_Clear(t *testing.T) {
	tempDir := t.TempDir()

	m := &Manager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	// Save config
	token := createTestJWT("employee-id", "org-id", time.Now().Add(24*time.Hour))
	config := &Config{
		PlatformURL: "https://test.example.com",
		Token:       token,
	}
	err := m.Save(config)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(m.configPath)
	require.NoError(t, err)

	// Clear config
	err = m.Clear()
	require.NoError(t, err)

	// Verify file is deleted
	_, err = os.Stat(m.configPath)
	assert.True(t, os.IsNotExist(err))

	// Clear again should not error
	err = m.Clear()
	require.NoError(t, err)
}

func TestManager_GetConfigPath(t *testing.T) {
	tempDir := t.TempDir()
	expectedPath := filepath.Join(tempDir, "config.json")

	m := &Manager{
		configPath: expectedPath,
	}

	assert.Equal(t, expectedPath, m.GetConfigPath())
}

func TestNewManager(t *testing.T) {
	// Test with temp HOME
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	m, err := NewManager()
	require.NoError(t, err)
	assert.NotNil(t, m)
	assert.Contains(t, m.configPath, ".arfa")
	assert.Contains(t, m.configPath, "config.json")

	// Verify config directory was created
	configDir := filepath.Join(tempDir, ".arfa")
	info, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestNewManagerWithPath(t *testing.T) {
	customPath := "/tmp/custom/config.json"
	m := NewManagerWithPath(customPath)
	assert.Equal(t, customPath, m.configPath)
}

func TestManager_IsTokenValid(t *testing.T) {
	tempDir := t.TempDir()

	m := &Manager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	t.Run("no token", func(t *testing.T) {
		// Clear any existing config
		m.Clear()

		// No config saved
		valid, err := m.IsTokenValid()
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("invalid token format", func(t *testing.T) {
		// Save config with invalid token
		config := &Config{
			PlatformURL: "https://test.example.com",
			Token:       "not-a-jwt",
		}
		err := m.Save(config)
		require.NoError(t, err)

		// Should be invalid due to parse error
		valid, err := m.IsTokenValid()
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("valid token not expired", func(t *testing.T) {
		// Save config with future expiration
		token := createTestJWT("employee-id", "org-id", time.Now().Add(1*time.Hour))
		config := &Config{
			PlatformURL: "https://test.example.com",
			Token:       token,
		}
		err := m.Save(config)
		require.NoError(t, err)

		// Should be valid
		valid, err := m.IsTokenValid()
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("expired token", func(t *testing.T) {
		// Save config with past expiration
		token := createTestJWT("employee-id", "org-id", time.Now().Add(-1*time.Hour))
		config := &Config{
			PlatformURL: "https://test.example.com",
			Token:       token,
		}
		err := m.Save(config)
		require.NoError(t, err)

		// Should be invalid
		valid, err := m.IsTokenValid()
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("token expiring soon within buffer", func(t *testing.T) {
		// Save config expiring in 3 minutes (within 5 minute buffer)
		token := createTestJWT("employee-id", "org-id", time.Now().Add(3*time.Minute))
		config := &Config{
			PlatformURL: "https://test.example.com",
			Token:       token,
		}
		err := m.Save(config)
		require.NoError(t, err)

		// Should be invalid (within 5 minute buffer)
		valid, err := m.IsTokenValid()
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("token expiring outside buffer", func(t *testing.T) {
		// Save config expiring in 10 minutes (outside 5 minute buffer)
		token := createTestJWT("employee-id", "org-id", time.Now().Add(10*time.Minute))
		config := &Config{
			PlatformURL: "https://test.example.com",
			Token:       token,
		}
		err := m.Save(config)
		require.NoError(t, err)

		// Should be valid
		valid, err := m.IsTokenValid()
		require.NoError(t, err)
		assert.True(t, valid)
	})
}

func TestParseJWTClaims(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		token := createTestJWT("emp-123", "org-456", time.Unix(1735689600, 0))
		claims, err := ParseJWTClaims(token)
		require.NoError(t, err)
		assert.Equal(t, "emp-123", claims.EmployeeID)
		assert.Equal(t, "org-456", claims.OrgID)
		assert.Equal(t, time.Unix(1735689600, 0), claims.ExpiresAt)
	})

	t.Run("invalid format - not enough parts", func(t *testing.T) {
		_, err := ParseJWTClaims("only.two")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JWT format")
	})

	t.Run("invalid format - bad base64", func(t *testing.T) {
		_, err := ParseJWTClaims("header.!!!invalid!!!.signature")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode JWT payload")
	})

	t.Run("invalid format - bad json", func(t *testing.T) {
		badPayload := base64.RawURLEncoding.EncodeToString([]byte("not json"))
		_, err := ParseJWTClaims("header." + badPayload + ".signature")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JWT claims")
	})
}

func TestConfig_GetClaims(t *testing.T) {
	t.Run("with token", func(t *testing.T) {
		token := createTestJWT("emp-123", "org-456", time.Now().Add(1*time.Hour))
		cfg := &Config{Token: token}
		claims, err := cfg.GetClaims()
		require.NoError(t, err)
		assert.Equal(t, "emp-123", claims.EmployeeID)
		assert.Equal(t, "org-456", claims.OrgID)
	})

	t.Run("without token", func(t *testing.T) {
		cfg := &Config{}
		_, err := cfg.GetClaims()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no token stored")
	})
}
