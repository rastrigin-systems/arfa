package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/config"
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

// Helper function to create a mock login server that returns a valid JWT.
func createMockLoginServer(t *testing.T, expectedEmail, expectedPassword string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
			var reqBody api.LoginRequest
			_ = json.NewDecoder(r.Body).Decode(&reqBody)

			// Create a valid JWT token with employee/org claims
			token := createTestJWT("emp-123", "org-456", time.Now().Add(24*time.Hour))

			// Return success response
			resp := api.LoginResponse{
				Token:     token,
				ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				Employee: api.LoginEmployeeInfo{
					ID:    "emp-123",
					OrgID: "org-456",
					Email: expectedEmail,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// Helper function to create a mock server that returns errors.
func createMockLoginServerWithError(t *testing.T, statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(`{"error":"Authentication failed"}`))
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

	// Manually save config with valid JWT
	token := createTestJWT("employee-id", "org-id", time.Now().Add(24*time.Hour))
	cfg := &config.Config{
		PlatformURL: "https://test.example.com",
		Token:       token,
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

	// Save config with valid JWT
	token := createTestJWT("employee-id", "org-id", time.Now().Add(24*time.Hour))
	cfg := &config.Config{
		PlatformURL: "https://test.example.com",
		Token:       token,
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

	// Save config with valid JWT
	token := createTestJWT("employee-id", "org-id", time.Now().Add(24*time.Hour))
	expectedConfig := &config.Config{
		PlatformURL: "https://test.example.com",
		Token:       token,
	}
	err := cm.Save(expectedConfig)
	require.NoError(t, err)

	// Get config
	cfg, err := authService.GetConfig()
	require.NoError(t, err)
	assert.Equal(t, expectedConfig.PlatformURL, cfg.PlatformURL)
	assert.Equal(t, expectedConfig.Token, cfg.Token)

	// Verify claims can be extracted
	claims, err := cfg.GetClaims()
	require.NoError(t, err)
	assert.Equal(t, "employee-id", claims.EmployeeID)
	assert.Equal(t, "org-id", claims.OrgID)
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

	// Save config with valid JWT
	token := createTestJWT("employee-id", "org-id", time.Now().Add(24*time.Hour))
	expectedConfig := &config.Config{
		PlatformURL: "https://test.example.com",
		Token:       token,
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

func TestService_RequireAuth_ExpiredToken(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := NewService(cm, pc)

	// Save config with expired JWT
	token := createTestJWT("employee-id", "org-id", time.Now().Add(-1*time.Hour))
	cfg := &config.Config{
		PlatformURL: "https://test.example.com",
		Token:       token,
	}
	err := cm.Save(cfg)
	require.NoError(t, err)

	// RequireAuth should fail with expired token
	resultCfg, err := authService.RequireAuth()
	assert.Error(t, err)
	assert.Nil(t, resultCfg)
	assert.Contains(t, err.Error(), "expired")
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
	assert.NotEmpty(t, cfg.Token)

	// Verify claims can be extracted from saved token
	claims, err := cfg.GetClaims()
	require.NoError(t, err)
	assert.Equal(t, "emp-123", claims.EmployeeID)
	assert.Equal(t, "org-456", claims.OrgID)
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
}
