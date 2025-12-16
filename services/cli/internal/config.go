// Package cli contains type aliases for backward compatibility.
// Config types are now defined in the config package.
package cli

import (
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/config"
)

// ============================================================================
// Type Aliases - pointing to config package
// ============================================================================

// Config is an alias for config.Config.
type Config = config.Config

// ConfigManager is an alias for config.Manager.
type ConfigManager = config.Manager

// DefaultPlatformURL is the default platform API URL.
const DefaultPlatformURL = config.DefaultPlatformURL

// NewConfigManager creates a new config.Manager with default path.
func NewConfigManager() (*config.Manager, error) {
	return config.NewManager()
}

// NewConfigManagerWithPath creates a new config.Manager with a custom path.
func NewConfigManagerWithPath(path string) *config.Manager {
	return config.NewManagerWithPath(path)
}
