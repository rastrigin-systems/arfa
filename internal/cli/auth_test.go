package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a mock login server
func createMockLoginServer(t *testing.T, expectedEmail, expectedPassword string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			var reqBody LoginRequest
			json.NewDecoder(r.Body).Decode(&reqBody)

			// Return success response
			resp := LoginResponse{
				Token:     "test-token-abc123",
				ExpiresAt: "2024-12-31T23:59:59Z",
				Employee: LoginEmployeeInfo{
					ID:    "emp-123",
					OrgID: "org-456",
					Email: expectedEmail,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// Helper function to create a mock server that returns errors
func createMockLoginServerWithError(t *testing.T, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(`{"error":"Authentication failed"}`))
	}))
}

func TestAuthService_IsAuthenticated(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)

	// Initially not authenticated
	authenticated, err := authService.IsAuthenticated()
	require.NoError(t, err)
	assert.False(t, authenticated)

	// Manually save config
	config := &Config{
		PlatformURL: "https://test.example.com",
		Token:       "test-token",
		EmployeeID:  "employee-id",
	}
	err = cm.Save(config)
	require.NoError(t, err)

	// Now should be authenticated
	authenticated, err = authService.IsAuthenticated()
	require.NoError(t, err)
	assert.True(t, authenticated)
}

func TestAuthService_Logout(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)

	// Save config
	config := &Config{
		PlatformURL: "https://test.example.com",
		Token:       "test-token",
		EmployeeID:  "employee-id",
	}
	err := cm.Save(config)
	require.NoError(t, err)

	// Logout
	err = authService.Logout()
	require.NoError(t, err)

	// Verify not authenticated
	authenticated, err := authService.IsAuthenticated()
	require.NoError(t, err)
	assert.False(t, authenticated)
}

func TestAuthService_GetConfig(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)

	// Save config
	expectedConfig := &Config{
		PlatformURL: "https://test.example.com",
		Token:       "test-token",
		EmployeeID:  "employee-id",
	}
	err := cm.Save(expectedConfig)
	require.NoError(t, err)

	// Get config
	config, err := authService.GetConfig()
	require.NoError(t, err)
	assert.Equal(t, expectedConfig.PlatformURL, config.PlatformURL)
	assert.Equal(t, expectedConfig.Token, config.Token)
	assert.Equal(t, expectedConfig.EmployeeID, config.EmployeeID)
}

func TestAuthService_RequireAuth_NotAuthenticated(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)

	// RequireAuth should fail when not authenticated
	config, err := authService.RequireAuth()
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "not authenticated")
}

func TestAuthService_RequireAuth_Authenticated(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)

	// Save config
	expectedConfig := &Config{
		PlatformURL: "https://test.example.com",
		Token:       "test-token",
		EmployeeID:  "employee-id",
	}
	err := cm.Save(expectedConfig)
	require.NoError(t, err)

	// RequireAuth should succeed
	config, err := authService.RequireAuth()
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, expectedConfig.Token, config.Token)

	// Verify platform client was updated
	assert.Equal(t, expectedConfig.Token, pc.token)
	assert.Equal(t, expectedConfig.PlatformURL, pc.baseURL)
}

func TestAuthService_Login_Success(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	// Create mock server
	server := createMockLoginServer(t, "test@example.com", "password123")
	defer server.Close()

	pc := NewPlatformClient(server.URL)
	authService := NewAuthService(cm, pc)

	// Perform login
	err := authService.Login(server.URL, "test@example.com", "password123")
	require.NoError(t, err)

	// Verify config was saved
	config, err := cm.Load()
	require.NoError(t, err)
	assert.Equal(t, server.URL, config.PlatformURL)
	assert.Equal(t, "test-token-abc123", config.Token)
	assert.Equal(t, "emp-123", config.EmployeeID)
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	// Create mock server that returns 401
	server := createMockLoginServerWithError(t, 401)
	defer server.Close()

	pc := NewPlatformClient(server.URL)
	authService := NewAuthService(cm, pc)

	// Perform login - should fail
	err := authService.Login(server.URL, "test@example.com", "wrong-password")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")

	// Verify config was NOT saved (should return empty config)
	config, err := cm.Load()
	require.NoError(t, err) // Load returns empty config, not error
	assert.Empty(t, config.Token)
	assert.Empty(t, config.EmployeeID)
}
