package auth

import (
	"context"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
)

// ServiceInterface defines the contract for authentication operations.
type ServiceInterface interface {
	// LoginInteractive performs interactive login with user prompts.
	LoginInteractive(ctx context.Context) error

	// Login performs non-interactive login with provided credentials.
	Login(ctx context.Context, platformURL, email, password string) error

	// Logout removes stored credentials.
	Logout() error

	// IsAuthenticated checks if user is authenticated.
	IsAuthenticated() (bool, error)

	// GetConfig returns the current config.
	GetConfig() (*config.Config, error)

	// RequireAuth ensures user is authenticated, returns error if not.
	RequireAuth() (*config.Config, error)
}

// Compile-time check that Service implements ServiceInterface.
var _ ServiceInterface = (*Service)(nil)

// ConfigManagerInterface defines what the auth service needs from config management.
// This is a subset of config.ManagerInterface to enable testing with mocks.
type ConfigManagerInterface interface {
	// Load reads the config from disk.
	Load() (*config.Config, error)

	// Save writes the config to disk.
	Save(config *config.Config) error

	// IsAuthenticated checks if the user is authenticated.
	IsAuthenticated() (bool, error)

	// IsTokenValid checks if the stored token is valid (not expired).
	IsTokenValid() (bool, error)

	// Clear removes the config file (logout).
	Clear() error
}

// APIClientInterface defines what the auth service needs from the API client.
// This is a subset of api.ClientInterface to enable testing with mocks.
type APIClientInterface interface {
	// SetToken sets the authentication token.
	SetToken(token string)

	// SetBaseURL sets the base URL for API requests.
	SetBaseURL(url string)

	// Login authenticates the user and returns a token.
	Login(ctx context.Context, email, password string) (*api.LoginResponse, error)
}
