// Package config handles local CLI configuration storage and retrieval.
package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
// Only stores platform_url and token - all other data is extracted from the JWT.
type Config struct {
	PlatformURL string `json:"platform_url"`
	Token       string `json:"token"`
}

// JWTClaims represents the claims extracted from the JWT token.
type JWTClaims struct {
	EmployeeID string
	OrgID      string
	ExpiresAt  time.Time
}

// ParseJWTClaims extracts claims from a JWT token without validating the signature.
// This is safe because the API server validates the token - we just need the cached claims.
func ParseJWTClaims(token string) (*JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// Decode the payload (middle part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims struct {
		EmployeeID string `json:"employee_id"`
		OrgID      string `json:"org_id"`
		Exp        int64  `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	return &JWTClaims{
		EmployeeID: claims.EmployeeID,
		OrgID:      claims.OrgID,
		ExpiresAt:  time.Unix(claims.Exp, 0),
	}, nil
}

// GetClaims parses and returns the JWT claims from the stored token.
func (c *Config) GetClaims() (*JWTClaims, error) {
	if c.Token == "" {
		return nil, fmt.Errorf("no token stored")
	}
	return ParseJWTClaims(c.Token)
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

	if config.Token == "" {
		return false, nil
	}

	// Verify we can parse the token and it has required claims
	claims, err := config.GetClaims()
	if err != nil {
		return false, nil
	}

	return claims.EmployeeID != "" && claims.OrgID != "", nil
}

// IsTokenValid checks if the stored token is valid (not expired).
func (m *Manager) IsTokenValid() (bool, error) {
	config, err := m.Load()
	if err != nil {
		return false, err
	}

	if config.Token == "" {
		return false, nil
	}

	claims, err := config.GetClaims()
	if err != nil {
		return false, nil
	}

	// Check if token has expired (with 5 minute buffer)
	return time.Now().Add(5 * time.Minute).Before(claims.ExpiresAt), nil
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
