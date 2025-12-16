package config

// ManagerInterface defines the contract for configuration management.
// All configuration storage and retrieval operations go through this interface.
type ManagerInterface interface {
	// Load reads the config from disk.
	Load() (*Config, error)

	// Save writes the config to disk.
	Save(config *Config) error

	// IsAuthenticated checks if the user is authenticated.
	IsAuthenticated() (bool, error)

	// IsTokenValid checks if the stored token is valid (not expired).
	IsTokenValid() (bool, error)

	// Clear removes the config file (logout).
	Clear() error

	// GetConfigPath returns the path to the config file.
	GetConfigPath() string
}

// Compile-time check that Manager implements ManagerInterface.
var _ ManagerInterface = (*Manager)(nil)
