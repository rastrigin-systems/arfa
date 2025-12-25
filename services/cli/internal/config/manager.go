// Package config handles local CLI configuration storage and retrieval.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DefaultPlatformURL returns the platform API URL.
// Can be overridden via ARFA_API_URL environment variable.
func DefaultPlatformURL() string {
	if url := os.Getenv("ARFA_API_URL"); url != "" {
		return url
	}
	return "http://localhost:8080"
}

// Config represents the local CLI configuration stored in ~/.arfa/config.json.
type Config struct {
	PlatformURL  string    `json:"platform_url"`
	Token        string    `json:"token"`
	TokenExpires time.Time `json:"token_expires"`
	EmployeeID   string    `json:"employee_id"`
	OrgID        string    `json:"org_id"`
	DefaultAgent string    `json:"default_agent"`
	LastSync     time.Time `json:"last_sync"`
}

// Manager handles local configuration storage and retrieval.
type Manager struct {
	configPath string
}

// NewManager creates a new Manager with default path (~/.arfa/config.json).
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".arfa")
	configPath := filepath.Join(configDir, "config.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &Manager{
		configPath: configPath,
	}, nil
}

// NewManagerWithPath creates a new Manager with a custom config path.
// Use this for testing or when you need a non-default config location.
func NewManagerWithPath(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// Load reads the config from disk.
func (m *Manager) Load() (*Config, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save writes the config to disk.
func (m *Manager) Save(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// IsAuthenticated checks if the user is authenticated.
func (m *Manager) IsAuthenticated() (bool, error) {
	config, err := m.Load()
	if err != nil {
		return false, err
	}

	return config.Token != "" && config.EmployeeID != "", nil
}

// IsTokenValid checks if the stored token is valid (not expired).
func (m *Manager) IsTokenValid() (bool, error) {
	config, err := m.Load()
	if err != nil {
		return false, err
	}

	// No token means not authenticated
	if config.Token == "" {
		return false, nil
	}

	// No expiration time means we can't validate (assume valid for backwards compatibility)
	if config.TokenExpires.IsZero() {
		return true, nil
	}

	// Check if token has expired (with 5 minute buffer)
	return time.Now().Add(5 * time.Minute).Before(config.TokenExpires), nil
}

// Clear removes the config file (logout).
func (m *Manager) Clear() error {
	if err := os.Remove(m.configPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already cleared
		}
		return fmt.Errorf("failed to remove config file: %w", err)
	}
	return nil
}

// GetConfigPath returns the path to the config file.
func (m *Manager) GetConfigPath() string {
	return m.configPath
}
