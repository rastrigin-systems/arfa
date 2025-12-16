// Package cli contains type aliases for backward compatibility.
// API client types are now defined in the api package.
package cli

import (
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
)

// ============================================================================
// Type Aliases - pointing to api package
// ============================================================================

// APIClient is an alias for api.Client.
type APIClient = api.Client

// NewAPIClient creates a new api.Client.
func NewAPIClient(baseURL string) *api.Client {
	return api.NewClient(baseURL)
}
