package auth_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/commands/auth"
	"github.com/rastrigin-systems/arfa/services/cli/internal/config"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
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

// TestLoginCommand_NonInteractive tests the login command with email/password flags.
func TestLoginCommand_NonInteractive(t *testing.T) {
	t.Run("success - valid credentials", func(t *testing.T) {
		// Setup mock server that returns a valid JWT
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
				// Verify request body
				var req api.LoginRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, "test@example.com", req.Email)
				assert.Equal(t, "password123", req.Password)

				// Create a valid JWT token with employee/org claims
				token := createTestJWT("emp-123", "org-456", time.Now().Add(24*time.Hour))

				// Return success response
				resp := api.LoginResponse{
					Token:     token,
					ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
					Employee: api.LoginEmployeeInfo{
						ID:       "emp-123",
						OrgID:    "org-456",
						Email:    "test@example.com",
						FullName: "Test User",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		// Create temp config directory
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")

		// Create container with test config
		c := container.New(
			container.WithConfigPath(configPath),
			container.WithPlatformURL(server.URL),
		)
		defer c.Close()

		// Create command
		cmd := auth.NewLoginCommand(c)
		cmd.SetArgs([]string{
			"--url", server.URL,
			"--email", "test@example.com",
			"--password", "password123",
		})

		// Capture output
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)

		// Execute
		err := cmd.Execute()
		require.NoError(t, err)

		// Verify config was saved
		configData, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var cfg config.Config
		err = json.Unmarshal(configData, &cfg)
		require.NoError(t, err)

		assert.NotEmpty(t, cfg.Token)
		assert.Equal(t, server.URL, cfg.PlatformURL)

		// Verify claims can be extracted from the saved token
		claims, err := cfg.GetClaims()
		require.NoError(t, err)
		assert.Equal(t, "emp-123", claims.EmployeeID)
		assert.Equal(t, "org-456", claims.OrgID)
	})

	t.Run("failure - invalid credentials", func(t *testing.T) {
		// Setup mock server that returns 401
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error": "invalid credentials"}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		// Create temp config directory
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")

		// Create container
		c := container.New(
			container.WithConfigPath(configPath),
			container.WithPlatformURL(server.URL),
		)
		defer c.Close()

		// Create and execute command
		cmd := auth.NewLoginCommand(c)
		cmd.SetArgs([]string{
			"--url", server.URL,
			"--email", "test@example.com",
			"--password", "wrong-password",
		})

		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)

		// Execute - should fail
		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("failure - server error", func(t *testing.T) {
		// Setup mock server that returns 500
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "internal server error"}`))
		}))
		defer server.Close()

		// Create temp config directory
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")

		// Create container
		c := container.New(
			container.WithConfigPath(configPath),
			container.WithPlatformURL(server.URL),
		)
		defer c.Close()

		// Create and execute command
		cmd := auth.NewLoginCommand(c)
		cmd.SetArgs([]string{
			"--url", server.URL,
			"--email", "test@example.com",
			"--password", "password123",
		})

		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)

		// Execute - should fail
		err := cmd.Execute()
		require.Error(t, err)
	})
}

// TestLoginCommand_UsesExistingPlatformURL tests that login uses saved platform URL.
func TestLoginCommand_UsesExistingPlatformURL(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/login" {
			// Create a valid JWT token with employee/org claims
			token := createTestJWT("emp-123", "org-456", time.Now().Add(24*time.Hour))

			resp := api.LoginResponse{
				Token:     token,
				ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				Employee: api.LoginEmployeeInfo{
					ID:    "emp-123",
					OrgID: "org-456",
					Email: "test@example.com",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Create temp config with existing platform URL
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")

	// Create old config with valid JWT to test that it gets replaced
	oldToken := createTestJWT("old-emp", "old-org", time.Now().Add(24*time.Hour))
	existingConfig := config.Config{
		PlatformURL: server.URL,
		Token:       oldToken,
	}
	configData, _ := json.Marshal(existingConfig)
	os.WriteFile(configPath, configData, 0600)

	// Create container (no platform URL set - should use saved one)
	c := container.New(
		container.WithConfigPath(configPath),
	)
	defer c.Close()

	// Create command without --url flag
	cmd := auth.NewLoginCommand(c)
	cmd.SetArgs([]string{
		"--email", "test@example.com",
		"--password", "password123",
	})

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stdout)

	// Execute
	err := cmd.Execute()
	require.NoError(t, err)

	// Verify config was updated with new token
	updatedData, _ := os.ReadFile(configPath)
	var updatedConfig config.Config
	json.Unmarshal(updatedData, &updatedConfig)

	assert.NotEmpty(t, updatedConfig.Token)
	assert.Equal(t, server.URL, updatedConfig.PlatformURL)

	// Verify claims from new token
	claims, err := updatedConfig.GetClaims()
	require.NoError(t, err)
	assert.Equal(t, "emp-123", claims.EmployeeID)
}

// TestLoginCommand_Flags tests that command flags are properly configured.
func TestLoginCommand_Flags(t *testing.T) {
	c := container.New()
	defer c.Close()

	cmd := auth.NewLoginCommand(c)

	// Check flags exist
	assert.NotNil(t, cmd.Flags().Lookup("url"))
	assert.NotNil(t, cmd.Flags().Lookup("email"))
	assert.NotNil(t, cmd.Flags().Lookup("password"))
}
