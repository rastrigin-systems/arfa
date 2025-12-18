package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/auth"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoginCommand_NonInteractive tests the login command with email/password flags.
func TestLoginCommand_NonInteractive(t *testing.T) {
	t.Run("success - valid credentials", func(t *testing.T) {
		// Setup mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/auth/login" && r.Method == "POST" {
				// Verify request body
				var req api.LoginRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, "test@example.com", req.Email)
				assert.Equal(t, "password123", req.Password)

				// Return success response
				resp := api.LoginResponse{
					Token:     "test-token-12345",
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

		assert.Equal(t, "test-token-12345", cfg.Token)
		assert.Equal(t, "emp-123", cfg.EmployeeID)
		assert.Equal(t, server.URL, cfg.PlatformURL)
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
			resp := api.LoginResponse{
				Token:     "new-token",
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

	existingConfig := config.Config{
		PlatformURL: server.URL,
		Token:       "old-token",
		EmployeeID:  "old-emp",
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

	assert.Equal(t, "new-token", updatedConfig.Token)
	assert.Equal(t, server.URL, updatedConfig.PlatformURL)
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
