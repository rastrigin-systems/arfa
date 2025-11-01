package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigManager_SaveAndLoad(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	// Create config
	config := &Config{
		PlatformURL:  "https://test.example.com",
		Token:        "test-token-123",
		EmployeeID:   "employee-456",
		DefaultAgent: "claude-code",
		LastSync:     time.Now(),
	}

	// Save config
	err := cm.Save(config)
	require.NoError(t, err)

	// Load config
	loaded, err := cm.Load()
	require.NoError(t, err)

	// Verify
	assert.Equal(t, config.PlatformURL, loaded.PlatformURL)
	assert.Equal(t, config.Token, loaded.Token)
	assert.Equal(t, config.EmployeeID, loaded.EmployeeID)
	assert.Equal(t, config.DefaultAgent, loaded.DefaultAgent)
	// Note: Time precision may vary due to JSON marshaling
}

func TestConfigManager_LoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	// Load non-existent config should return empty config
	loaded, err := cm.Load()
	require.NoError(t, err)
	assert.Equal(t, "", loaded.Token)
	assert.Equal(t, "", loaded.EmployeeID)
}

func TestConfigManager_IsAuthenticated(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	// Initially not authenticated
	authenticated, err := cm.IsAuthenticated()
	require.NoError(t, err)
	assert.False(t, authenticated)

	// Save authenticated config
	config := &Config{
		PlatformURL: "https://test.example.com",
		Token:       "test-token",
		EmployeeID:  "employee-id",
	}
	err = cm.Save(config)
	require.NoError(t, err)

	// Now should be authenticated
	authenticated, err = cm.IsAuthenticated()
	require.NoError(t, err)
	assert.True(t, authenticated)
}

func TestConfigManager_Clear(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	// Save config
	config := &Config{
		PlatformURL: "https://test.example.com",
		Token:       "test-token",
		EmployeeID:  "employee-id",
	}
	err := cm.Save(config)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(cm.configPath)
	require.NoError(t, err)

	// Clear config
	err = cm.Clear()
	require.NoError(t, err)

	// Verify file is deleted
	_, err = os.Stat(cm.configPath)
	assert.True(t, os.IsNotExist(err))

	// Clear again should not error
	err = cm.Clear()
	require.NoError(t, err)
}

func TestConfigManager_GetConfigPath(t *testing.T) {
	tempDir := t.TempDir()
	expectedPath := filepath.Join(tempDir, "config.json")

	cm := &ConfigManager{
		configPath: expectedPath,
	}

	assert.Equal(t, expectedPath, cm.GetConfigPath())
}

func TestNewConfigManager(t *testing.T) {
	// Test with temp HOME
	tempDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", oldHome)

	cm, err := NewConfigManager()
	require.NoError(t, err)
	assert.NotNil(t, cm)
	assert.Contains(t, cm.configPath, ".ubik")
	assert.Contains(t, cm.configPath, "config.json")

	// Verify config directory was created
	configDir := filepath.Join(tempDir, ".ubik")
	info, err := os.Stat(configDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestConfigManager_IsTokenValid(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: filepath.Join(tempDir, "config.json"),
	}

	t.Run("no token", func(t *testing.T) {
		// No config saved
		valid, err := cm.IsTokenValid()
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("valid token with no expiration", func(t *testing.T) {
		// Save config without expiration time (backwards compatibility)
		config := &Config{
			PlatformURL: "https://test.example.com",
			Token:       "test-token",
			EmployeeID:  "employee-id",
		}
		err := cm.Save(config)
		require.NoError(t, err)

		// Should be considered valid for backwards compatibility
		valid, err := cm.IsTokenValid()
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("valid token not expired", func(t *testing.T) {
		// Save config with future expiration
		config := &Config{
			PlatformURL:  "https://test.example.com",
			Token:        "test-token",
			TokenExpires: time.Now().Add(1 * time.Hour),
			EmployeeID:   "employee-id",
		}
		err := cm.Save(config)
		require.NoError(t, err)

		// Should be valid
		valid, err := cm.IsTokenValid()
		require.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("expired token", func(t *testing.T) {
		// Save config with past expiration
		config := &Config{
			PlatformURL:  "https://test.example.com",
			Token:        "test-token",
			TokenExpires: time.Now().Add(-1 * time.Hour),
			EmployeeID:   "employee-id",
		}
		err := cm.Save(config)
		require.NoError(t, err)

		// Should be invalid
		valid, err := cm.IsTokenValid()
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("token expiring soon within buffer", func(t *testing.T) {
		// Save config expiring in 3 minutes (within 5 minute buffer)
		config := &Config{
			PlatformURL:  "https://test.example.com",
			Token:        "test-token",
			TokenExpires: time.Now().Add(3 * time.Minute),
			EmployeeID:   "employee-id",
		}
		err := cm.Save(config)
		require.NoError(t, err)

		// Should be invalid (within 5 minute buffer)
		valid, err := cm.IsTokenValid()
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("token expiring outside buffer", func(t *testing.T) {
		// Save config expiring in 10 minutes (outside 5 minute buffer)
		config := &Config{
			PlatformURL:  "https://test.example.com",
			Token:        "test-token",
			TokenExpires: time.Now().Add(10 * time.Minute),
			EmployeeID:   "employee-id",
		}
		err := cm.Save(config)
		require.NoError(t, err)

		// Should be valid
		valid, err := cm.IsTokenValid()
		require.NoError(t, err)
		assert.True(t, valid)
	})
}
