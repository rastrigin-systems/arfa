package api

import (
	"context"
	"net/http"
)

// ClientInterface defines the contract for API communication with the platform server.
// All methods that make HTTP requests accept context.Context for cancellation and timeout support.
type ClientInterface interface {
	// SetToken sets the authentication token for subsequent requests.
	SetToken(token string)

	// SetBaseURL sets the base URL for API requests.
	SetBaseURL(url string)

	// SetHTTPClient sets a custom HTTP client (primarily for testing).
	SetHTTPClient(client *http.Client)

	// BaseURL returns the current base URL.
	BaseURL() string

	// Token returns the current token.
	Token() string

	// -------------------------------------------------------------------------
	// Authentication
	// -------------------------------------------------------------------------

	// Login authenticates the user and returns a token.
	Login(ctx context.Context, email, password string) (*LoginResponse, error)

	// GetCurrentEmployee fetches information about the currently authenticated employee.
	GetCurrentEmployee(ctx context.Context) (*EmployeeInfo, error)

	// GetEmployeeInfo gets information about a specific employee by ID.
	GetEmployeeInfo(ctx context.Context, employeeID string) (*EmployeeInfo, error)

	// -------------------------------------------------------------------------
	// Agent Configuration
	// -------------------------------------------------------------------------

	// GetResolvedAgentConfigs fetches resolved agent configurations for an employee.
	GetResolvedAgentConfigs(ctx context.Context, employeeID string) ([]AgentConfig, error)

	// GetMyResolvedAgentConfigs fetches resolved agent configurations for the current employee.
	GetMyResolvedAgentConfigs(ctx context.Context) ([]AgentConfig, error)

	// GetOrgAgentConfigs fetches organization-level agent configs.
	GetOrgAgentConfigs(ctx context.Context) ([]OrgAgentConfigResponse, error)

	// GetTeamAgentConfigs fetches team-level agent configs.
	GetTeamAgentConfigs(ctx context.Context, teamID string) ([]TeamAgentConfigResponse, error)

	// GetEmployeeAgentConfigs fetches employee-level agent configs.
	GetEmployeeAgentConfigs(ctx context.Context, employeeID string) ([]EmployeeAgentConfigResponse, error)

	// -------------------------------------------------------------------------
	// Claude Token Management
	// -------------------------------------------------------------------------

	// GetClaudeTokenStatus fetches the Claude token status for the current employee.
	GetClaudeTokenStatus(ctx context.Context) (*ClaudeTokenStatusResponse, error)

	// GetEffectiveClaudeToken fetches the effective Claude token value.
	GetEffectiveClaudeToken(ctx context.Context) (string, error)

	// GetEffectiveClaudeTokenInfo fetches the effective Claude token with full metadata.
	GetEffectiveClaudeTokenInfo(ctx context.Context) (*EffectiveClaudeTokenResponse, error)

	// -------------------------------------------------------------------------
	// Sync
	// -------------------------------------------------------------------------

	// GetClaudeCodeConfig fetches the complete Claude Code configuration bundle.
	GetClaudeCodeConfig(ctx context.Context) (*ClaudeCodeSyncResponse, error)

	// -------------------------------------------------------------------------
	// Skills
	// -------------------------------------------------------------------------

	// ListSkills fetches all available skills from the catalog.
	ListSkills(ctx context.Context) (*ListSkillsResponse, error)

	// GetSkill fetches details for a specific skill by ID.
	GetSkill(ctx context.Context, skillID string) (*Skill, error)

	// ListEmployeeSkills fetches skills assigned to the authenticated employee.
	ListEmployeeSkills(ctx context.Context) (*ListEmployeeSkillsResponse, error)

	// GetEmployeeSkill fetches a specific skill assigned to the authenticated employee.
	GetEmployeeSkill(ctx context.Context, skillID string) (*EmployeeSkill, error)

	// -------------------------------------------------------------------------
	// Logging
	// -------------------------------------------------------------------------

	// CreateLog sends a single log entry to the platform API.
	CreateLog(ctx context.Context, entry LogEntry) error

	// CreateLogBatch sends multiple log entries in a single request.
	CreateLogBatch(ctx context.Context, entries []LogEntry) error

	// GetLogs fetches logs from the API with optional filters.
	GetLogs(ctx context.Context, params GetLogsParams) (*LogsResponse, error)
}

// Compile-time check that Client implements ClientInterface.
var _ ClientInterface = (*Client)(nil)
