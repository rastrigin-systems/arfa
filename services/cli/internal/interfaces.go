package cli

import (
	"io"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
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
	Load() (*Config, error)

	// Save persists the configuration to disk.
	Save(config *Config) error

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

// PlatformClientInterface defines the contract for API communication with the platform server.
// Implementations handle all HTTP requests to the backend API.
type PlatformClientInterface interface {
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
	Login(email, password string) (*LoginResponse, error)

	// GetCurrentEmployee fetches information about the currently authenticated employee.
	GetCurrentEmployee() (*EmployeeInfo, error)

	// GetEmployeeInfo gets information about a specific employee by ID.
	GetEmployeeInfo(employeeID string) (*EmployeeInfo, error)

	// -------------------------------------------------------------------------
	// Agent Configuration
	// -------------------------------------------------------------------------

	// GetResolvedAgentConfigs fetches resolved agent configurations for an employee.
	GetResolvedAgentConfigs(employeeID string) ([]AgentConfig, error)

	// GetMyResolvedAgentConfigs fetches resolved agent configurations for the current employee.
	// Uses JWT token to identify the employee.
	GetMyResolvedAgentConfigs() ([]AgentConfig, error)

	// GetOrgAgentConfigs fetches organization-level agent configs.
	GetOrgAgentConfigs() ([]OrgAgentConfigResponse, error)

	// GetTeamAgentConfigs fetches team-level agent configs.
	GetTeamAgentConfigs(teamID string) ([]TeamAgentConfigResponse, error)

	// GetEmployeeAgentConfigs fetches employee-level agent configs.
	GetEmployeeAgentConfigs(employeeID string) ([]EmployeeAgentConfigResponse, error)

	// -------------------------------------------------------------------------
	// Claude Token Management
	// -------------------------------------------------------------------------

	// GetClaudeTokenStatus fetches the Claude token status for the current employee.
	GetClaudeTokenStatus() (*ClaudeTokenStatusResponse, error)

	// GetEffectiveClaudeToken fetches the effective Claude token value.
	GetEffectiveClaudeToken() (string, error)

	// GetEffectiveClaudeTokenInfo fetches the effective Claude token with full metadata.
	GetEffectiveClaudeTokenInfo() (*EffectiveClaudeTokenResponse, error)

	// -------------------------------------------------------------------------
	// Sync
	// -------------------------------------------------------------------------

	// GetClaudeCodeConfig fetches the complete Claude Code configuration bundle.
	GetClaudeCodeConfig() (*ClaudeCodeSyncResponse, error)

	// -------------------------------------------------------------------------
	// Skills
	// -------------------------------------------------------------------------

	// ListSkills fetches all available skills from the catalog.
	ListSkills() (*ListSkillsResponse, error)

	// GetSkill fetches details for a specific skill by ID.
	GetSkill(skillID string) (*Skill, error)

	// ListEmployeeSkills fetches skills assigned to the authenticated employee.
	ListEmployeeSkills() (*ListEmployeeSkillsResponse, error)

	// GetEmployeeSkill fetches a specific skill assigned to the authenticated employee.
	GetEmployeeSkill(skillID string) (*EmployeeSkill, error)

	// -------------------------------------------------------------------------
	// Logging
	// -------------------------------------------------------------------------

	// CreateLog sends a single log entry to the platform API.
	CreateLog(entry LogEntry) error

	// CreateLogBatch sends multiple log entries in a single request.
	CreateLogBatch(entries []LogEntry) error
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
	LoginInteractive() error

	// Login performs non-interactive login with provided credentials.
	Login(platformURL, email, password string) error

	// Logout removes stored credentials.
	Logout() error

	// IsAuthenticated checks if user is authenticated.
	IsAuthenticated() (bool, error)

	// GetConfig returns the current configuration.
	GetConfig() (*Config, error)

	// RequireAuth ensures user is authenticated and token is valid.
	// Returns the current config with platform client configured.
	RequireAuth() (*Config, error)
}

// SyncServiceInterface defines the contract for configuration synchronization.
// Implementations handle fetching configs from platform and storing locally.
type SyncServiceInterface interface {
	// Sync fetches configs from platform and stores them locally.
	Sync() (*SyncResult, error)

	// SyncClaudeCode fetches and installs complete Claude Code configuration to targetDir.
	SyncClaudeCode(targetDir string) error

	// GetLocalAgentConfigs loads agent configs from local storage.
	GetLocalAgentConfigs() ([]AgentConfig, error)

	// GetAgentConfig retrieves a specific agent config by ID or name.
	GetAgentConfig(idOrName string) (*AgentConfig, error)

	// SetDockerClient sets the Docker client for container operations.
	SetDockerClient(dockerClient *DockerClient)

	// StartContainers starts Docker containers for synced agent configs.
	StartContainers(workspacePath string, apiKey string) error

	// StopContainers stops all running containers.
	StopContainers() error

	// GetContainerStatus returns the status of all containers.
	GetContainerStatus() ([]ContainerInfo, error)
}

// AgentServiceInterface defines the contract for agent management operations.
// Implementations handle listing agents, managing configs, and checking updates.
type AgentServiceInterface interface {
	// ListAgents fetches all available agents from the platform.
	ListAgents() ([]Agent, error)

	// GetAgent fetches details for a specific agent.
	GetAgent(agentID string) (*Agent, error)

	// ListEmployeeAgentConfigs fetches employee's assigned agent configs.
	ListEmployeeAgentConfigs(employeeID string) ([]EmployeeAgentConfig, error)

	// RequestAgent creates an employee agent configuration (request for access).
	RequestAgent(employeeID, agentID string) error

	// CheckForUpdates checks if there are config updates available.
	CheckForUpdates(employeeID string) (bool, error)

	// GetLocalAgents returns locally configured agents.
	GetLocalAgents() ([]AgentConfig, error)
}

// SkillsServiceInterface defines the contract for skill management operations.
// Implementations handle listing skills from catalog and local storage.
type SkillsServiceInterface interface {
	// ListCatalogSkills fetches all available skills from the platform catalog.
	ListCatalogSkills() ([]Skill, error)

	// GetSkill fetches details for a specific skill from the catalog.
	GetSkill(skillID string) (*Skill, error)

	// GetSkillByName fetches a skill by name (searches catalog).
	GetSkillByName(name string) (*Skill, error)

	// ListEmployeeSkills fetches skills assigned to the authenticated employee.
	ListEmployeeSkills() ([]EmployeeSkill, error)

	// GetEmployeeSkill fetches a specific skill assigned to the employee.
	GetEmployeeSkill(skillID string) (*EmployeeSkill, error)

	// GetEmployeeSkillByName fetches an employee skill by name.
	GetEmployeeSkillByName(name string) (*EmployeeSkill, error)

	// GetLocalSkills returns locally installed skills from .claude/skills/.
	GetLocalSkills() ([]LocalSkillInfo, error)

	// GetLocalSkill returns details for a specific locally installed skill.
	GetLocalSkill(name string) (*LocalSkillInfo, error)
}

// ============================================================================
// Docker/Container Interfaces - PR 1.3
// ============================================================================
// Low-level interfaces for Docker daemon communication and container lifecycle.
// ============================================================================

// DockerClientInterface defines the contract for Docker daemon communication.
// Implementations wrap the Docker SDK client for container operations.
// NOTE: This interface uses the current signatures. Phase 2 will add context.Context.
type DockerClientInterface interface {
	// Close closes the Docker client connection.
	Close() error

	// Ping checks if Docker daemon is running.
	Ping() error

	// GetVersion returns Docker version information.
	GetVersion() (string, error)

	// -------------------------------------------------------------------------
	// Image Operations
	// -------------------------------------------------------------------------

	// PullImage pulls a Docker image (or uses local if available).
	PullImage(imageName string) error

	// -------------------------------------------------------------------------
	// Container Operations
	// -------------------------------------------------------------------------

	// CreateContainer creates a Docker container.
	CreateContainer(config *DockerContainerConfig, hostConfig *DockerHostConfig, networkConfig *DockerNetworkConfig, containerName string) (string, error)

	// StartContainer starts a Docker container.
	StartContainer(containerID string) error

	// StopContainer stops a Docker container.
	StopContainer(containerID string, timeout *int) error

	// RemoveContainer removes a Docker container.
	RemoveContainer(containerID string, force bool) error

	// RemoveContainerByName finds and removes a container by name.
	RemoveContainerByName(name string) error

	// ListContainers lists Docker containers with optional filters.
	ListContainers(all bool, labelFilter map[string]string) ([]ContainerInfo, error)

	// -------------------------------------------------------------------------
	// Container Logs
	// -------------------------------------------------------------------------

	// GetContainerLogs retrieves logs from a container.
	GetContainerLogs(containerID string, follow bool) (io.ReadCloser, error)

	// StreamContainerLogs streams container logs to stdout/stderr.
	StreamContainerLogs(containerID string) error

	// -------------------------------------------------------------------------
	// Network Operations
	// -------------------------------------------------------------------------

	// CreateNetwork creates a Docker network.
	CreateNetwork(name string) (string, error)

	// NetworkExists checks if a network exists.
	NetworkExists(name string) (bool, error)

	// RemoveNetwork removes a Docker network.
	RemoveNetwork(name string) error
}

// ContainerManagerInterface defines the contract for managing containers.
// Implementations orchestrate container lifecycle for agents and MCP servers.
type ContainerManagerInterface interface {
	// SetupNetwork creates the ubik network if it doesn't exist.
	SetupNetwork() error

	// StartMCPServer starts an MCP server container.
	StartMCPServer(spec MCPServerSpec, workspacePath string) (string, error)

	// StartAgent starts an agent container.
	StartAgent(spec AgentSpec, workspacePath string) (string, error)

	// StopContainers stops all ubik-managed containers.
	StopContainers() error

	// CleanupContainers removes all stopped ubik-managed containers.
	CleanupContainers() error

	// GetContainerStatus returns status of all ubik-managed containers.
	GetContainerStatus() ([]ContainerInfo, error)
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
	_ ConfigManagerInterface     = (*ConfigManager)(nil)
	_ PlatformClientInterface    = (*PlatformClient)(nil)
	_ AuthServiceInterface       = (*AuthService)(nil)
	_ SyncServiceInterface       = (*SyncService)(nil)
	_ AgentServiceInterface      = (*AgentService)(nil)
	_ SkillsServiceInterface     = (*SkillsService)(nil)
	_ DockerClientInterface      = (*DockerClient)(nil)
	_ ContainerManagerInterface  = (*ContainerManager)(nil)
)
