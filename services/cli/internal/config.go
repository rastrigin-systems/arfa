package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the local CLI configuration stored in ~/.ubik/config.json
type Config struct {
	PlatformURL  string    `json:"platform_url"`
	Token        string    `json:"token"`
	TokenExpires time.Time `json:"token_expires"`
	EmployeeID   string    `json:"employee_id"`
	DefaultAgent string    `json:"default_agent"`
	LastSync     time.Time `json:"last_sync"`
}

// ConfigManager handles local configuration storage and retrieval
type ConfigManager struct {
	configPath string
}

// NewConfigManager creates a new ConfigManager
func NewConfigManager() (*ConfigManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".ubik")
	configPath := filepath.Join(configDir, "config.json")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &ConfigManager{
		configPath: configPath,
	}, nil
}

// Load reads the config from disk
func (cm *ConfigManager) Load() (*Config, error) {
	data, err := os.ReadFile(cm.configPath)
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

// Save writes the config to disk
func (cm *ConfigManager) Save(config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(cm.configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// IsAuthenticated checks if the user is authenticated
func (cm *ConfigManager) IsAuthenticated() (bool, error) {
	config, err := cm.Load()
	if err != nil {
		return false, err
	}

	return config.Token != "" && config.EmployeeID != "", nil
}

// IsTokenValid checks if the stored token is valid (not expired)
func (cm *ConfigManager) IsTokenValid() (bool, error) {
	config, err := cm.Load()
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

// Clear removes the config file (logout)
func (cm *ConfigManager) Clear() error {
	if err := os.Remove(cm.configPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already cleared
		}
		return fmt.Errorf("failed to remove config file: %w", err)
	}
	return nil
}

// GetConfigPath returns the path to the config file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}
