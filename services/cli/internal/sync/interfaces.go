package sync

import (
	"context"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
)

// ============================================================================
// Dependency Interfaces
// ============================================================================
// These interfaces define what the sync package needs from other packages.
// This enables dependency injection and testability.
// ============================================================================

// ConfigManagerInterface defines what sync needs from config.Manager
type ConfigManagerInterface interface {
	Load() (*config.Config, error)
	Save(cfg *config.Config) error
	IsAuthenticated() (bool, error)
	IsTokenValid() (bool, error)
	Clear() error
	GetConfigPath() string
}

// APIClientInterface defines what sync needs from api.Client
type APIClientInterface interface {
	GetMyResolvedAgentConfigs(ctx context.Context) ([]api.AgentConfig, error)
	GetClaudeCodeConfig(ctx context.Context) (*api.ClaudeCodeSyncResponse, error)
	GetEffectiveClaudeToken(ctx context.Context) (string, error)
	GetMyToolPolicies(ctx context.Context) (*api.EmployeeToolPoliciesResponse, error)
}

// AuthServiceInterface defines what sync needs from auth.Service
type AuthServiceInterface interface {
	RequireAuth() (*config.Config, error)
}

// DockerClientInterface defines what sync needs from DockerClient
type DockerClientInterface interface {
	Ping(ctx context.Context) error
	GetVersion(ctx context.Context) (string, error)
}

// ContainerManagerInterface defines what sync needs from ContainerManager
type ContainerManagerInterface interface {
	SetupNetwork(ctx context.Context) error
	StartMCPServer(ctx context.Context, spec MCPServerSpec, workspacePath string) (string, error)
	StartAgent(ctx context.Context, spec AgentSpec, workspacePath string) (string, error)
	StopContainers(ctx context.Context) error
	GetContainerStatus(ctx context.Context) ([]ContainerInfo, error)
}

// ============================================================================
// Types used by sync package
// ============================================================================

// ContainerInfo represents information about a Docker container
type ContainerInfo struct {
	ID      string
	Name    string
	Image   string
	State   string
	Status  string
	Created int64
}

// MCPServerSpec defines configuration for an MCP server container
type MCPServerSpec struct {
	ServerID   string
	ServerName string
	ServerType string
	Image      string
	Port       int
	Config     map[string]interface{}
}

// AgentSpec defines configuration for an agent container
type AgentSpec struct {
	AgentID       string
	AgentName     string
	AgentType     string
	Image         string
	Configuration map[string]interface{}
	MCPServers    []MCPServerSpec
	APIKey        string // Deprecated: Use ClaudeToken instead
	ClaudeToken   string // Claude API token (from hybrid auth)
}

// ============================================================================
// Service Interface
// ============================================================================

// ServiceInterface defines the contract for configuration synchronization.
// Implementations handle fetching configs from platform and storing locally.
type ServiceInterface interface {
	// Sync fetches configs from platform and stores them locally.
	Sync(ctx context.Context) (*Result, error)

	// SyncClaudeCode fetches and installs complete Claude Code configuration to targetDir.
	SyncClaudeCode(ctx context.Context, targetDir string) error

	// GetLocalAgentConfigs loads agent configs from local storage.
	GetLocalAgentConfigs() ([]api.AgentConfig, error)

	// GetAgentConfig retrieves a specific agent config by ID or name.
	GetAgentConfig(idOrName string) (*api.AgentConfig, error)

	// GetLocalToolPolicies loads tool policies from local storage (~/.ubik/policies.json).
	GetLocalToolPolicies() ([]api.ToolPolicy, error)

	// SetDockerClient sets the Docker client for container operations.
	// Deprecated: Use SetContainerManager instead for better dependency injection.
	SetDockerClient(dockerClient DockerClientInterface)

	// SetContainerManager sets the container manager directly.
	// This is the preferred way to inject container management dependencies.
	SetContainerManager(cm ContainerManagerInterface)

	// StartContainers starts Docker containers for synced agent configs.
	StartContainers(ctx context.Context, workspacePath string, apiKey string) error

	// StopContainers stops all running containers.
	StopContainers(ctx context.Context) error

	// GetContainerStatus returns the status of all containers.
	GetContainerStatus(ctx context.Context) ([]ContainerInfo, error)
}

// Compile-time interface implementation check
var _ ServiceInterface = (*Service)(nil)
