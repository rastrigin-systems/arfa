package cli

import (
	"context"
	"io"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/agent"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/config"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/skill"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/sync"
)

// ============================================================================
// Core Interfaces - PR 1.1
// ============================================================================
// These interfaces enable dependency injection and testability by decoupling
// implementations from their consumers. All services should depend on interfaces
// rather than concrete types.
// ============================================================================

// ConfigManagerInterface defines the contract for local configuration management.
// Implementations handle reading/writing CLI configuration from ~/.ubik/config.json
type ConfigManagerInterface interface {
	// Load reads the configuration from disk.
	// Returns an empty Config if the file doesn't exist.
	Load() (*config.Config, error)

	// Save persists the configuration to disk.
	Save(cfg *config.Config) error

	// IsAuthenticated checks if the user has valid credentials stored.
	// Returns true if both token and employee_id are present.
	IsAuthenticated() (bool, error)

	// IsTokenValid checks if the stored token is valid (not expired).
	// Returns true if token exists and hasn't expired (with 5 minute buffer).
	IsTokenValid() (bool, error)

	// Clear removes the config file (logout).
	Clear() error

	// GetConfigPath returns the absolute path to the config file.
	GetConfigPath() string
}

// APIClientInterface defines the contract for API communication with the platform server.
// Implementations handle all HTTP requests to the backend API.
// All methods that make HTTP requests accept context.Context as the first parameter for cancellation and timeout support.
type APIClientInterface interface {
	// SetToken sets the authentication token for subsequent requests.
	SetToken(token string)

	// SetBaseURL sets the base URL for API requests.
	// This allows overriding the URL at runtime (e.g., during login).
	SetBaseURL(url string)

	// SetHTTPClient sets a custom HTTP client (primarily for testing).
	SetHTTPClient(client *http.Client)

	// -------------------------------------------------------------------------
	// Authentication
	// -------------------------------------------------------------------------

	// Login authenticates the user and returns a token.
	Login(ctx context.Context, email, password string) (*api.LoginResponse, error)

	// GetCurrentEmployee fetches information about the currently authenticated employee.
	GetCurrentEmployee(ctx context.Context) (*api.EmployeeInfo, error)

	// GetEmployeeInfo gets information about a specific employee by ID.
	GetEmployeeInfo(ctx context.Context, employeeID string) (*api.EmployeeInfo, error)

	// -------------------------------------------------------------------------
	// Agent Configuration
	// -------------------------------------------------------------------------

	// GetResolvedAgentConfigs fetches resolved agent configurations for an employee.
	GetResolvedAgentConfigs(ctx context.Context, employeeID string) ([]api.AgentConfig, error)

	// GetMyResolvedAgentConfigs fetches resolved agent configurations for the current employee.
	// Uses JWT token to identify the employee.
	GetMyResolvedAgentConfigs(ctx context.Context) ([]api.AgentConfig, error)

	// GetOrgAgentConfigs fetches organization-level agent configs.
	GetOrgAgentConfigs(ctx context.Context) ([]api.OrgAgentConfigResponse, error)

	// GetTeamAgentConfigs fetches team-level agent configs.
	GetTeamAgentConfigs(ctx context.Context, teamID string) ([]api.TeamAgentConfigResponse, error)

	// GetEmployeeAgentConfigs fetches employee-level agent configs.
	GetEmployeeAgentConfigs(ctx context.Context, employeeID string) ([]api.EmployeeAgentConfigResponse, error)

	// -------------------------------------------------------------------------
	// Claude Token Management
	// -------------------------------------------------------------------------

	// GetClaudeTokenStatus fetches the Claude token status for the current employee.
	GetClaudeTokenStatus(ctx context.Context) (*api.ClaudeTokenStatusResponse, error)

	// GetEffectiveClaudeToken fetches the effective Claude token value.
	GetEffectiveClaudeToken(ctx context.Context) (string, error)

	// GetEffectiveClaudeTokenInfo fetches the effective Claude token with full metadata.
	GetEffectiveClaudeTokenInfo(ctx context.Context) (*api.EffectiveClaudeTokenResponse, error)

	// -------------------------------------------------------------------------
	// Sync
	// -------------------------------------------------------------------------

	// GetClaudeCodeConfig fetches the complete Claude Code configuration bundle.
	GetClaudeCodeConfig(ctx context.Context) (*api.ClaudeCodeSyncResponse, error)

	// -------------------------------------------------------------------------
	// Skills
	// -------------------------------------------------------------------------

	// ListSkills fetches all available skills from the catalog.
	ListSkills(ctx context.Context) (*api.ListSkillsResponse, error)

	// GetSkill fetches details for a specific skill by ID.
	GetSkill(ctx context.Context, skillID string) (*api.Skill, error)

	// ListEmployeeSkills fetches skills assigned to the authenticated employee.
	ListEmployeeSkills(ctx context.Context) (*api.ListEmployeeSkillsResponse, error)

	// GetEmployeeSkill fetches a specific skill assigned to the authenticated employee.
	GetEmployeeSkill(ctx context.Context, skillID string) (*api.EmployeeSkill, error)

	// -------------------------------------------------------------------------
	// Logging
	// -------------------------------------------------------------------------

	// CreateLog sends a single log entry to the platform API.
	CreateLog(ctx context.Context, entry api.LogEntry) error

	// CreateLogBatch sends multiple log entries in a single request.
	CreateLogBatch(ctx context.Context, entries []api.LogEntry) error

	// GetLogs fetches logs from the API with optional filters.
	GetLogs(ctx context.Context, params api.GetLogsParams) (*api.LogsResponse, error)
}

// ============================================================================
// Service Interfaces - PR 1.2
// ============================================================================
// Higher-level service interfaces that orchestrate business logic using the
// core interfaces above.
// ============================================================================

// AuthServiceInterface defines the contract for authentication operations.
// Implementations handle user login/logout and credential management.
type AuthServiceInterface interface {
	// LoginInteractive performs interactive login with prompts for URL, email, and password.
	LoginInteractive(ctx context.Context) error

	// Login performs non-interactive login with provided credentials.
	Login(ctx context.Context, platformURL, email, password string) error

	// Logout removes stored credentials.
	Logout() error

	// IsAuthenticated checks if user is authenticated.
	IsAuthenticated() (bool, error)

	// GetConfig returns the current configuration.
	GetConfig() (*config.Config, error)

	// RequireAuth ensures user is authenticated and token is valid.
	// Returns the current config with platform client configured.
	RequireAuth() (*config.Config, error)
}

// SyncServiceInterface defines the contract for configuration synchronization.
// Implementations handle fetching configs from platform and storing locally.
type SyncServiceInterface interface {
	// Sync fetches configs from platform and stores them locally.
	Sync(ctx context.Context) (*sync.Result, error)

	// SyncClaudeCode fetches and installs complete Claude Code configuration to targetDir.
	SyncClaudeCode(ctx context.Context, targetDir string) error

	// GetLocalAgentConfigs loads agent configs from local storage.
	GetLocalAgentConfigs() ([]api.AgentConfig, error)

	// GetAgentConfig retrieves a specific agent config by ID or name.
	GetAgentConfig(idOrName string) (*api.AgentConfig, error)

	// SetDockerClient sets the Docker client for container operations.
	// Deprecated: Use SetContainerManager instead for better dependency injection.
	SetDockerClient(dockerClient sync.DockerClientInterface)

	// SetContainerManager sets the container manager directly.
	// This is the preferred way to inject container management dependencies.
	SetContainerManager(cm sync.ContainerManagerInterface)

	// StartContainers starts Docker containers for synced agent configs.
	StartContainers(ctx context.Context, workspacePath string, apiKey string) error

	// StopContainers stops all running containers.
	StopContainers(ctx context.Context) error

	// GetContainerStatus returns the status of all containers.
	GetContainerStatus(ctx context.Context) ([]sync.ContainerInfo, error)
}

// AgentServiceInterface defines the contract for agent management operations.
// Implementations handle listing agents, managing configs, and checking updates.
type AgentServiceInterface interface {
	// ListAgents fetches all available agents from the platform.
	ListAgents(ctx context.Context) ([]agent.Agent, error)

	// GetAgent fetches details for a specific agent.
	GetAgent(ctx context.Context, agentID string) (*agent.Agent, error)

	// ListEmployeeAgentConfigs fetches employee's assigned agent configs.
	ListEmployeeAgentConfigs(ctx context.Context, employeeID string) ([]agent.EmployeeAgentConfig, error)

	// RequestAgent creates an employee agent configuration (request for access).
	RequestAgent(ctx context.Context, employeeID, agentID string) error

	// CheckForUpdates checks if there are config updates available.
	CheckForUpdates(ctx context.Context, employeeID string) (bool, error)

	// GetLocalAgents returns locally configured agents.
	GetLocalAgents() ([]api.AgentConfig, error)
}

// SkillsServiceInterface defines the contract for skill management operations.
// Implementations handle listing skills from catalog and local storage.
type SkillsServiceInterface interface {
	// ListCatalogSkills fetches all available skills from the platform catalog.
	ListCatalogSkills(ctx context.Context) ([]api.Skill, error)

	// GetSkill fetches details for a specific skill from the catalog.
	GetSkill(ctx context.Context, skillID string) (*api.Skill, error)

	// GetSkillByName fetches a skill by name (searches catalog).
	GetSkillByName(ctx context.Context, name string) (*api.Skill, error)

	// ListEmployeeSkills fetches skills assigned to the authenticated employee.
	ListEmployeeSkills(ctx context.Context) ([]api.EmployeeSkill, error)

	// GetEmployeeSkill fetches a specific skill assigned to the employee.
	GetEmployeeSkill(ctx context.Context, skillID string) (*api.EmployeeSkill, error)

	// GetEmployeeSkillByName fetches an employee skill by name.
	GetEmployeeSkillByName(ctx context.Context, name string) (*api.EmployeeSkill, error)

	// GetLocalSkills returns locally installed skills from .claude/skills/.
	GetLocalSkills() ([]skill.LocalSkillInfo, error)

	// GetLocalSkill returns details for a specific locally installed skill.
	GetLocalSkill(name string) (*skill.LocalSkillInfo, error)
}

// ============================================================================
// Docker/Container Interfaces - PR 1.3
// ============================================================================
// Low-level interfaces for Docker daemon communication and container lifecycle.
// ============================================================================

// DockerClientInterface defines the contract for Docker daemon communication.
// Implementations wrap the Docker SDK client for container operations.
// All methods accept context.Context as the first parameter for cancellation and timeout support.
type DockerClientInterface interface {
	// Close closes the Docker client connection.
	Close() error

	// Ping checks if Docker daemon is running.
	Ping(ctx context.Context) error

	// GetVersion returns Docker version information.
	GetVersion(ctx context.Context) (string, error)

	// -------------------------------------------------------------------------
	// Image Operations
	// -------------------------------------------------------------------------

	// PullImage pulls a Docker image (or uses local if available).
	PullImage(ctx context.Context, imageName string) error

	// -------------------------------------------------------------------------
	// Container Operations
	// -------------------------------------------------------------------------

	// CreateContainer creates a Docker container.
	CreateContainer(ctx context.Context, config *DockerContainerConfig, hostConfig *DockerHostConfig, networkConfig *DockerNetworkConfig, containerName string) (string, error)

	// StartContainer starts a Docker container.
	StartContainer(ctx context.Context, containerID string) error

	// StopContainer stops a Docker container.
	StopContainer(ctx context.Context, containerID string, timeout *int) error

	// RemoveContainer removes a Docker container.
	RemoveContainer(ctx context.Context, containerID string, force bool) error

	// RemoveContainerByName finds and removes a container by name.
	RemoveContainerByName(ctx context.Context, name string) error

	// ListContainers lists Docker containers with optional filters.
	ListContainers(ctx context.Context, all bool, labelFilter map[string]string) ([]ContainerInfo, error)

	// -------------------------------------------------------------------------
	// Container Logs
	// -------------------------------------------------------------------------

	// GetContainerLogs retrieves logs from a container.
	GetContainerLogs(ctx context.Context, containerID string, follow bool) (io.ReadCloser, error)

	// StreamContainerLogs streams container logs to stdout/stderr.
	StreamContainerLogs(ctx context.Context, containerID string) error

	// -------------------------------------------------------------------------
	// Network Operations
	// -------------------------------------------------------------------------

	// CreateNetwork creates a Docker network.
	CreateNetwork(ctx context.Context, name string) (string, error)

	// NetworkExists checks if a network exists.
	NetworkExists(ctx context.Context, name string) (bool, error)

	// RemoveNetwork removes a Docker network.
	RemoveNetwork(ctx context.Context, name string) error
}

// ContainerManagerInterface defines the contract for managing containers.
// Implementations orchestrate container lifecycle for agents and MCP servers.
// All methods accept context.Context as the first parameter for cancellation and timeout support.
type ContainerManagerInterface interface {
	// SetupNetwork creates the ubik network if it doesn't exist.
	SetupNetwork(ctx context.Context) error

	// StartMCPServer starts an MCP server container.
	StartMCPServer(ctx context.Context, spec MCPServerSpec, workspacePath string) (string, error)

	// StartAgent starts an agent container.
	StartAgent(ctx context.Context, spec AgentSpec, workspacePath string) (string, error)

	// StopContainers stops all ubik-managed containers.
	StopContainers(ctx context.Context) error

	// CleanupContainers removes all stopped ubik-managed containers.
	CleanupContainers(ctx context.Context) error

	// GetContainerStatus returns status of all ubik-managed containers.
	GetContainerStatus(ctx context.Context) ([]ContainerInfo, error)
}

// DockerContainerConfig is an alias for container.Config from Docker SDK.
// This allows the interface to not directly depend on Docker types.
type DockerContainerConfig = container.Config

// DockerHostConfig is an alias for container.HostConfig from Docker SDK.
type DockerHostConfig = container.HostConfig

// DockerNetworkConfig is an alias for network.NetworkingConfig from Docker SDK.
type DockerNetworkConfig = network.NetworkingConfig

// Compile-time interface implementation checks.
// These ensure that the concrete types implement their respective interfaces.
var (
	_ ConfigManagerInterface    = (*config.Manager)(nil)
	_ APIClientInterface        = (*api.Client)(nil)
	_ AuthServiceInterface      = (*auth.Service)(nil)
	_ SyncServiceInterface      = (*sync.Service)(nil)
	_ AgentServiceInterface     = (*agent.Service)(nil)
	_ SkillsServiceInterface    = (*skill.Service)(nil)
	_ DockerClientInterface     = (*DockerClient)(nil)
	_ ContainerManagerInterface = (*ContainerManager)(nil)
)
