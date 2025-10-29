package cli

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
