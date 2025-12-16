package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncService_Sync_Success(t *testing.T) {
	tempDir := t.TempDir()

	// Setup mock platform server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Updated to use JWT-based /employees/me endpoint
		if r.URL.Path == "/api/v1/employees/me/agent-configs/resolved" {
			resp := api.ResolvedConfigsResponse{
				Configs: []api.AgentConfigAPIResponse{
					{
						AgentID:   "agent-1",
						AgentName: "Claude Code",
						AgentType: "claude-code",
						IsEnabled: true,
						Config:    map[string]interface{}{"model": "claude-3-5-sonnet"},
					},
				},
				Total: 1,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Setup config manager
	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))

	// Save authentication config
	cfg := &config.Config{
		PlatformURL: server.URL,
		Token:       "test-token",
		EmployeeID:  "emp-123",
	}
	err := cm.Save(cfg)
	require.NoError(t, err)

	// Setup auth service
	pc := api.NewClient(server.URL)
	pc.SetToken("test-token")
	authService := auth.NewService(cm, pc)

	// Setup sync service with HOME override
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	syncService := NewSyncService(cm, pc, authService)

	// Perform sync
	ctx := context.Background()
	result, err := syncService.Sync(ctx)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.AgentConfigs, 1)
	assert.Equal(t, "Claude Code", result.AgentConfigs[0].AgentName)
}

func TestSyncService_Sync_NotAuthenticated(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))

	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewSyncService(cm, pc, authService)

	// Sync should fail when not authenticated
	ctx := context.Background()
	result, err := syncService.Sync(ctx)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not authenticated")
}

func TestSyncService_Sync_NoConfigs(t *testing.T) {
	tempDir := t.TempDir()

	// Setup mock platform server that returns empty configs
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Updated to use JWT-based /employees/me endpoint
		if r.URL.Path == "/api/v1/employees/me/agent-configs/resolved" {
			resp := api.ResolvedConfigsResponse{
				Configs: []api.AgentConfigAPIResponse{},
				Total:   0,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Setup config manager
	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))

	// Save authentication config
	cfg := &config.Config{
		PlatformURL: server.URL,
		Token:       "test-token",
		EmployeeID:  "emp-123",
	}
	err := cm.Save(cfg)
	require.NoError(t, err)

	// Setup auth service
	pc := api.NewClient(server.URL)
	pc.SetToken("test-token")
	authService := auth.NewService(cm, pc)

	syncService := NewSyncService(cm, pc, authService)

	// Perform sync
	ctx := context.Background()
	result, err := syncService.Sync(ctx)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.AgentConfigs, 0)
}
