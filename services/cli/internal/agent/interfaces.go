package agent

import (
	"context"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/config"
)

// ============================================================================
// Dependency Interfaces
// ============================================================================
// These interfaces define what the agent package needs from other packages.
// This enables dependency injection and testability.
// ============================================================================

// APIClientInterface defines what agent needs from api.Client
type APIClientInterface interface {
	// DoRequest performs an HTTP request to the API
	DoRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error
	// GetResolvedAgentConfigs fetches resolved agent configurations for an employee
	GetResolvedAgentConfigs(ctx context.Context, employeeID string) ([]api.AgentConfig, error)
}

// ConfigManagerInterface defines what agent needs from config.Manager
type ConfigManagerInterface interface {
	Load() (*config.Config, error)
	Save(cfg *config.Config) error
	GetConfigPath() string
}

// ============================================================================
// Service Interface
// ============================================================================

// ServiceInterface defines the contract for agent management operations.
// Implementations handle listing agents, managing configs, and checking updates.
type ServiceInterface interface {
	// ListAgents fetches all available agents from the platform.
	ListAgents(ctx context.Context) ([]Agent, error)

	// GetAgent fetches details for a specific agent.
	GetAgent(ctx context.Context, agentID string) (*Agent, error)

	// ListEmployeeAgentConfigs fetches employee's assigned agent configs.
	ListEmployeeAgentConfigs(ctx context.Context, employeeID string) ([]EmployeeAgentConfig, error)

	// RequestAgent creates an employee agent configuration (request for access).
	RequestAgent(ctx context.Context, employeeID, agentID string) error

	// CheckForUpdates checks if there are config updates available.
	CheckForUpdates(ctx context.Context, employeeID string) (bool, error)

	// GetLocalAgents returns locally configured agents.
	GetLocalAgents() ([]api.AgentConfig, error)
}

// Compile-time interface implementation check
var _ ServiceInterface = (*Service)(nil)
