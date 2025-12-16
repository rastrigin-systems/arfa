// Package cli contains type aliases for backward compatibility.
// Auth service types are now defined in the auth package.
package cli

import (
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/config"
)

// ============================================================================
// Type Aliases - pointing to auth package
// ============================================================================

// AuthService is an alias for auth.Service.
type AuthService = auth.Service

// NewAuthService creates a new auth.Service with concrete types.
func NewAuthService(configManager *config.Manager, apiClient *api.Client) *auth.Service {
	return auth.NewService(configManager, apiClient)
}

// NewAuthServiceWithInterfaces creates a new auth.Service with interface types.
func NewAuthServiceWithInterfaces(configManager auth.ConfigManagerInterface, apiClient auth.APIClientInterface) *auth.Service {
	return auth.NewServiceWithInterfaces(configManager, apiClient)
}
