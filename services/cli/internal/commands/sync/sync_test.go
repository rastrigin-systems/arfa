package sync_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/sync"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/config"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupAuthenticatedConfig creates a config file with valid authentication.
func setupAuthenticatedConfig(t *testing.T, tempDir, platformURL string) string {
	configPath := filepath.Join(tempDir, "config.json")
	cfg := config.Config{
		PlatformURL:  platformURL,
		Token:        "valid-token",
		TokenExpires: time.Now().Add(24 * time.Hour),
		EmployeeID:   "emp-123",
	}
	configData, err := json.Marshal(cfg)
	require.NoError(t, err)
	err = os.WriteFile(configPath, configData, 0600)
	require.NoError(t, err)
	return configPath
}

// TestSyncCommand_Success tests successful sync operation.
func TestSyncCommand_Success(t *testing.T) {
	t.Run("syncs agent configs", func(t *testing.T) {
		// Setup mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			switch r.URL.Path {
			case "/api/v1/employees/me/agent-configs/resolved":
				resp := api.ResolvedConfigsResponse{
					Configs: []api.AgentConfigAPIResponse{
						{
							AgentID:   "agent-1",
							AgentName: "Claude Code",
							AgentType: "claude-code",
							Provider:  "anthropic",
							IsEnabled: true,
							Config:    map[string]interface{}{"model": "claude-3"},
						},
						{
							AgentID:   "agent-2",
							AgentName: "Cursor",
							AgentType: "cursor",
							Provider:  "anthropic",
							IsEnabled: true,
							Config:    map[string]interface{}{},
						},
					},
					Total: 2,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
			default:
				http.NotFound(w, r)
			}
		}))
		defer server.Close()

		// Setup
		tempDir := t.TempDir()
		configPath := setupAuthenticatedConfig(t, tempDir, server.URL)

		// Create container
		c := container.New(
			container.WithConfigPath(configPath),
			container.WithPlatformURL(server.URL),
		)
		defer c.Close()

		// Create and execute command
		cmd := sync.NewSyncCommand(c)
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)

		err := cmd.Execute()
		require.NoError(t, err)
		// Command executed successfully - configs were synced
	})

	t.Run("handles empty configs", func(t *testing.T) {
		// Setup mock server that returns empty configs
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if r.URL.Path == "/api/v1/employees/me/agent-configs/resolved" {
				resp := api.ResolvedConfigsResponse{
					Configs: []api.AgentConfigAPIResponse{},
					Total:   0,
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(resp)
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		// Setup
		tempDir := t.TempDir()
		configPath := setupAuthenticatedConfig(t, tempDir, server.URL)

		// Create container
		c := container.New(
			container.WithConfigPath(configPath),
			container.WithPlatformURL(server.URL),
		)
		defer c.Close()

		// Create and execute command
		cmd := sync.NewSyncCommand(c)
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)

		err := cmd.Execute()
		require.NoError(t, err)
		// Should succeed even with empty configs
	})
}

// TestSyncCommand_NotAuthenticated tests sync when user is not authenticated.
func TestSyncCommand_NotAuthenticated(t *testing.T) {
	t.Run("fails when not authenticated", func(t *testing.T) {
		// Setup mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		// Setup with empty config (not authenticated)
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")
		emptyConfig := config.Config{}
		configData, _ := json.Marshal(emptyConfig)
		os.WriteFile(configPath, configData, 0600)

		// Create container
		c := container.New(
			container.WithConfigPath(configPath),
			container.WithPlatformURL(server.URL),
		)
		defer c.Close()

		// Create and execute command
		cmd := sync.NewSyncCommand(c)
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
	})

	t.Run("fails when token expired", func(t *testing.T) {
		// Setup mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		// Setup with expired token
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, "config.json")
		expiredConfig := config.Config{
			PlatformURL:  server.URL,
			Token:        "expired-token",
			TokenExpires: time.Now().Add(-1 * time.Hour), // Expired
			EmployeeID:   "emp-123",
		}
		configData, _ := json.Marshal(expiredConfig)
		os.WriteFile(configPath, configData, 0600)

		// Create container
		c := container.New(
			container.WithConfigPath(configPath),
			container.WithPlatformURL(server.URL),
		)
		defer c.Close()

		// Create and execute command
		cmd := sync.NewSyncCommand(c)
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})
}

// TestSyncCommand_APIError tests sync when API returns an error.
func TestSyncCommand_APIError(t *testing.T) {
	t.Run("handles API error", func(t *testing.T) {
		// Setup mock server that returns error
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if r.URL.Path == "/api/v1/employees/me/agent-configs/resolved" {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "database error"}`))
				return
			}
			http.NotFound(w, r)
		}))
		defer server.Close()

		// Setup
		tempDir := t.TempDir()
		configPath := setupAuthenticatedConfig(t, tempDir, server.URL)

		// Create container
		c := container.New(
			container.WithConfigPath(configPath),
			container.WithPlatformURL(server.URL),
		)
		defer c.Close()

		// Create and execute command
		cmd := sync.NewSyncCommand(c)
		var stdout bytes.Buffer
		cmd.SetOut(&stdout)
		cmd.SetErr(&stdout)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch agent configs")
	})
}

// TestSyncCommand_Flags tests that command flags are properly configured.
func TestSyncCommand_Flags(t *testing.T) {
	c := container.New()
	defer c.Close()

	cmd := sync.NewSyncCommand(c)

	// Check flags exist
	assert.NotNil(t, cmd.Flags().Lookup("start-containers"))
	assert.NotNil(t, cmd.Flags().Lookup("workspace"))
	assert.NotNil(t, cmd.Flags().Lookup("api-key"))

	// Check default values
	startContainers, _ := cmd.Flags().GetBool("start-containers")
	assert.False(t, startContainers)

	workspace, _ := cmd.Flags().GetString("workspace")
	assert.Equal(t, ".", workspace)
}
