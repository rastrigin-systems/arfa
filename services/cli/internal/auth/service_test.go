package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create a mock login server.
func createMockLoginServer(t *testing.T, expectedEmail, expectedPassword string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			var reqBody api.LoginRequest
			json.NewDecoder(r.Body).Decode(&reqBody)

			// Return success response
			resp := api.LoginResponse{
				Token:     "test-token-abc123",
				ExpiresAt: "2024-12-31T23:59:59Z",
				Employee: api.LoginEmployeeInfo{
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

// Helper function to create a mock server that returns errors.
func createMockLoginServerWithError(t *testing.T, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(`{"error":"Authentication failed"}`))
	}))
}

func TestService_IsAuthenticated(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := NewService(cm, pc)

	// Initially not authenticated
	authenticated, err := authService.IsAuthenticated()
	require.NoError(t, err)
	assert.False(t, authenticated)

	// Manually save config
	cfg := &config.Config{
		PlatformURL: "https://test.example.com",
		Token:       "test-token",
		EmployeeID:  "employee-id",
	}
	err = cm.Save(cfg)
	require.NoError(t, err)

	// Now should be authenticated
	authenticated, err = authService.IsAuthenticated()
	require.NoError(t, err)
	assert.True(t, authenticated)
}

func TestService_Logout(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := NewService(cm, pc)

	// Save config
	cfg := &config.Config{
		PlatformURL: "https://test.example.com",
		Token:       "test-token",
		EmployeeID:  "employee-id",
	}
	err := cm.Save(cfg)
	require.NoError(t, err)

	// Logout
	err = authService.Logout()
	require.NoError(t, err)

	// Verify not authenticated
	authenticated, err := authService.IsAuthenticated()
	require.NoError(t, err)
	assert.False(t, authenticated)
}

func TestService_GetConfig(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := NewService(cm, pc)

	// Save config
	expectedConfig := &config.Config{
		PlatformURL: "https://test.example.com",
		Token:       "test-token",
		EmployeeID:  "employee-id",
	}
	err := cm.Save(expectedConfig)
	require.NoError(t, err)

	// Get config
	cfg, err := authService.GetConfig()
	require.NoError(t, err)
	assert.Equal(t, expectedConfig.PlatformURL, cfg.PlatformURL)
	assert.Equal(t, expectedConfig.Token, cfg.Token)
	assert.Equal(t, expectedConfig.EmployeeID, cfg.EmployeeID)
}

func TestService_RequireAuth_NotAuthenticated(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := NewService(cm, pc)

	// RequireAuth should fail when not authenticated
	cfg, err := authService.RequireAuth()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "not authenticated")
}

func TestService_RequireAuth_Authenticated(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := NewService(cm, pc)

	// Save config
	expectedConfig := &config.Config{
		PlatformURL: "https://test.example.com",
		Token:       "test-token",
		EmployeeID:  "employee-id",
	}
	err := cm.Save(expectedConfig)
	require.NoError(t, err)

	// RequireAuth should succeed
	cfg, err := authService.RequireAuth()
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, expectedConfig.Token, cfg.Token)

	// Verify platform client was updated
	assert.Equal(t, expectedConfig.Token, pc.Token())
	assert.Equal(t, expectedConfig.PlatformURL, pc.BaseURL())
}

func TestService_Login_Success(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))

	// Create mock server
	server := createMockLoginServer(t, "test@example.com", "password123")
	defer server.Close()

	pc := api.NewClient(server.URL)
	authService := NewService(cm, pc)

	// Perform login
	ctx := context.Background()
	err := authService.Login(ctx, server.URL, "test@example.com", "password123")
	require.NoError(t, err)

	// Verify config was saved
	cfg, err := cm.Load()
	require.NoError(t, err)
	assert.Equal(t, server.URL, cfg.PlatformURL)
	assert.Equal(t, "test-token-abc123", cfg.Token)
	assert.Equal(t, "emp-123", cfg.EmployeeID)
}

func TestService_Login_InvalidCredentials(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))

	// Create mock server that returns 401
	server := createMockLoginServerWithError(t, 401)
	defer server.Close()

	pc := api.NewClient(server.URL)
	authService := NewService(cm, pc)

	// Perform login - should fail
	ctx := context.Background()
	err := authService.Login(ctx, server.URL, "test@example.com", "wrong-password")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication failed")

	// Verify config was NOT saved (should return empty config)
	cfg, err := cm.Load()
	require.NoError(t, err) // Load returns empty config, not error
	assert.Empty(t, cfg.Token)
	assert.Empty(t, cfg.EmployeeID)
}

func TestParseExpiresAt(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		checkFunc func(t *testing.T, result interface{})
	}{
		{
			name:    "empty string defaults to 24 hours",
			input:   "",
			wantErr: false,
		},
		{
			name:    "RFC3339 format",
			input:   "2024-12-31T23:59:59Z",
			wantErr: false,
		},
		{
			name:    "RFC3339Nano format",
			input:   "2024-12-31T23:59:59.123456789Z",
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseExpiresAt(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.False(t, result.IsZero())
			}
		})
	}
}
