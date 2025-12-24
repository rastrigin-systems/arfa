// Package api contains API request/response types for the Ubik platform.
package api

import "time"

// ============================================================================
// Authentication Types
// ============================================================================

// LoginRequest represents a login request to the platform API.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response from the platform API.
type LoginResponse struct {
	Token     string            `json:"token"`
	ExpiresAt string            `json:"expires_at"`
	Employee  LoginEmployeeInfo `json:"employee"`
}

// LoginEmployeeInfo contains employee info from login response.
type LoginEmployeeInfo struct {
	ID       string `json:"id"`
	OrgID    string `json:"org_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// EmployeeInfo contains detailed employee information.
type EmployeeInfo struct {
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	FullName string  `json:"full_name"`
	OrgID    string  `json:"org_id"`
	TeamID   *string `json:"team_id"`
}

// ============================================================================
// Agent Configuration Types
// ============================================================================

// AgentConfigAPIResponse represents an agent config as returned by the API.
type AgentConfigAPIResponse struct {
	AgentID      string                 `json:"agent_id"`
	AgentName    string                 `json:"agent_name"`
	AgentType    string                 `json:"agent_type"`
	IsEnabled    bool                   `json:"is_enabled"`
	Config       map[string]interface{} `json:"config"`
	Provider     string                 `json:"provider"`
	DockerImage  *string                `json:"docker_image"`
	SyncToken    string                 `json:"sync_token"`
	SystemPrompt string                 `json:"system_prompt"`
	LastSyncedAt *string                `json:"last_synced_at"`
}

// AgentConfig represents a resolved agent configuration (internal use).
type AgentConfig struct {
	AgentID       string                 `json:"agent_id"`
	AgentName     string                 `json:"agent_name"`
	AgentType     string                 `json:"agent_type"`
	Provider      string                 `json:"provider"`
	DockerImage   string                 `json:"docker_image"`
	IsEnabled     bool                   `json:"is_enabled"`
	Configuration map[string]interface{} `json:"configuration"`
	MCPServers    []MCPServerConfig      `json:"mcp_servers"`
}

// MCPServerConfig represents an MCP server configuration.
type MCPServerConfig struct {
	ServerID   string                 `json:"server_id"`
	ServerName string                 `json:"server_name"`
	ServerType string                 `json:"server_type"`
	IsEnabled  bool                   `json:"is_enabled"`
	Config     map[string]interface{} `json:"config"`
}

// ResolvedConfigsResponse represents the response from the resolved configs endpoint.
type ResolvedConfigsResponse struct {
	Configs []AgentConfigAPIResponse `json:"configs"`
	Total   int                      `json:"total"`
}

// OrgAgentConfigResponse represents an org-level agent config.
type OrgAgentConfigResponse struct {
	ID        string                 `json:"id"`
	AgentID   string                 `json:"agent_id"`
	AgentName string                 `json:"agent_name"`
	Config    map[string]interface{} `json:"config"`
	IsEnabled bool                   `json:"is_enabled"`
}

// TeamAgentConfigResponse represents a team-level agent config.
type TeamAgentConfigResponse struct {
	ID             string                 `json:"id"`
	AgentID        string                 `json:"agent_id"`
	AgentName      string                 `json:"agent_name"`
	ConfigOverride map[string]interface{} `json:"config_override"`
	IsEnabled      bool                   `json:"is_enabled"`
}

// EmployeeAgentConfigResponse represents an employee-level agent config from the API.
type EmployeeAgentConfigResponse struct {
	ID             string                 `json:"id"`
	AgentID        string                 `json:"agent_id"`
	AgentName      string                 `json:"agent_name"`
	ConfigOverride map[string]interface{} `json:"config_override"`
	IsEnabled      bool                   `json:"is_enabled"`
}

// ============================================================================
// Token Types
// ============================================================================

// ClaudeTokenStatusResponse represents the Claude token status response.
type ClaudeTokenStatusResponse struct {
	EmployeeID        string `json:"employee_id"`
	HasPersonalToken  bool   `json:"has_personal_token"`
	HasCompanyToken   bool   `json:"has_company_token"`
	ActiveTokenSource string `json:"active_token_source"`
}

// EffectiveClaudeTokenResponse represents the effective token response.
type EffectiveClaudeTokenResponse struct {
	Token      string `json:"token"`
	Source     string `json:"source"`
	OrgID      string `json:"org_id"`
	OrgName    string `json:"org_name"`
	EmployeeID string `json:"employee_id"`
}

// ============================================================================
// Sync Types
// ============================================================================

// ClaudeCodeSyncResponse represents the complete Claude Code configuration bundle.
type ClaudeCodeSyncResponse struct {
	Agents     []AgentConfigSync     `json:"agents"`
	Skills     []SkillConfigSync     `json:"skills"`
	MCPServers []MCPServerConfigSync `json:"mcp_servers"`
	Version    string                `json:"version"`
	SyncedAt   string                `json:"synced_at"`
}

// AgentConfigSync represents an agent configuration in the sync response.
type AgentConfigSync struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Filename  string                 `json:"filename"`
	Content   string                 `json:"content,omitempty"`
	Config    map[string]interface{} `json:"config"`
	Provider  string                 `json:"provider"`
	IsEnabled bool                   `json:"is_enabled"`
	Version   string                 `json:"version"`
}

// SkillConfigSync represents a skill configuration in the sync response.
type SkillConfigSync struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Category     string                 `json:"category,omitempty"`
	Version      string                 `json:"version"`
	Files        []map[string]string    `json:"files,omitempty"`
	Dependencies map[string]interface{} `json:"dependencies,omitempty"`
	IsEnabled    bool                   `json:"is_enabled"`
}

// MCPServerConfigSync represents an MCP server configuration in the sync response.
type MCPServerConfigSync struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Provider        string                 `json:"provider"`
	Version         string                 `json:"version"`
	Description     string                 `json:"description,omitempty"`
	DockerImage     string                 `json:"docker_image"`
	Config          map[string]interface{} `json:"config"`
	RequiredEnvVars []string               `json:"required_env_vars,omitempty"`
	IsEnabled       bool                   `json:"is_enabled"`
}

// ============================================================================
// Skill Types
// ============================================================================

// Skill represents a skill from the catalog.
type Skill struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Version      string                 `json:"version"`
	Files        []SkillFile            `json:"files"`
	Dependencies map[string]interface{} `json:"dependencies,omitempty"`
	IsActive     bool                   `json:"is_active"`
	CreatedAt    *time.Time             `json:"created_at,omitempty"`
	UpdatedAt    *time.Time             `json:"updated_at,omitempty"`
}

// SkillFile represents a file in a skill.
type SkillFile struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
}

// EmployeeSkill represents an employee's assigned skill.
type EmployeeSkill struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Version      string                 `json:"version"`
	Files        []SkillFile            `json:"files"`
	Dependencies map[string]interface{} `json:"dependencies,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
	IsActive     bool                   `json:"is_active"`
	IsEnabled    bool                   `json:"is_enabled"`
	InstalledAt  *time.Time             `json:"installed_at,omitempty"`
}

// ListSkillsResponse represents the response from list skills endpoint.
type ListSkillsResponse struct {
	Skills []Skill `json:"skills"`
	Total  int     `json:"total"`
}

// ListEmployeeSkillsResponse represents the response from list employee skills endpoint.
type ListEmployeeSkillsResponse struct {
	Skills []EmployeeSkill `json:"skills"`
	Total  int             `json:"total"`
}

// ============================================================================
// Logging Types
// ============================================================================

// LogEntry represents a log entry to send to the API.
type LogEntry struct {
	SessionID     string                 `json:"session_id,omitempty"`
	ClientName    string                 `json:"client_name,omitempty"`
	ClientVersion string                 `json:"client_version,omitempty"`
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category"`
	Content       string                 `json:"content,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
}

// CreateLogRequest represents a single log creation request.
type CreateLogRequest struct {
	SessionID     *string                 `json:"session_id,omitempty"`
	ClientName    *string                 `json:"client_name,omitempty"`
	ClientVersion *string                 `json:"client_version,omitempty"`
	EventType     string                  `json:"event_type"`
	EventCategory string                  `json:"event_category"`
	Content       *string                 `json:"content,omitempty"`
	Payload       *map[string]interface{} `json:"payload,omitempty"`
}

// LogEntryResponse represents a log entry from the API response.
type LogEntryResponse struct {
	ID            string                 `json:"id"`
	SessionID     string                 `json:"session_id"`
	ClientName    string                 `json:"client_name,omitempty"`
	ClientVersion string                 `json:"client_version,omitempty"`
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category"`
	Content       string                 `json:"content"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// LogsResponse represents the paginated logs response from the API.
type LogsResponse struct {
	Logs       []LogEntryResponse `json:"logs"`
	TotalCount int                `json:"total_count"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
}

// ============================================================================
// Tool Policy Types
// ============================================================================

// ToolPolicyScope indicates the level at which a policy is applied.
type ToolPolicyScope string

const (
	ToolPolicyScopeOrganization ToolPolicyScope = "organization"
	ToolPolicyScopeTeam         ToolPolicyScope = "team"
	ToolPolicyScopeEmployee     ToolPolicyScope = "employee"
)

// ToolPolicyAction indicates what happens when a policy matches.
type ToolPolicyAction string

const (
	ToolPolicyActionDeny  ToolPolicyAction = "deny"
	ToolPolicyActionAudit ToolPolicyAction = "audit"
)

// ToolPolicy represents a policy that controls tool access for an employee.
type ToolPolicy struct {
	ID         string                 `json:"id,omitempty"`
	ToolName   string                 `json:"tool_name"`
	Action     ToolPolicyAction       `json:"action"`
	Reason     *string                `json:"reason,omitempty"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	Scope      *ToolPolicyScope       `json:"scope,omitempty"`
}

// EmployeeToolPoliciesResponse represents the response from GET /employees/me/tool-policies.
type EmployeeToolPoliciesResponse struct {
	Policies []ToolPolicy `json:"policies"`
	Version  int          `json:"version"`
	SyncedAt string       `json:"synced_at"`
}

// ============================================================================
// Webhook Types
// ============================================================================

// WebhookDestination represents a webhook destination for log export.
type WebhookDestination struct {
	ID          string            `json:"id"`
	OrgID       string            `json:"org_id,omitempty"`
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	AuthType    string            `json:"auth_type"`
	AuthConfig  map[string]string `json:"auth_config,omitempty"`
	EventTypes  []string          `json:"event_types"`
	EventFilter map[string]string `json:"event_filter,omitempty"`
	Enabled     bool              `json:"enabled"`
	BatchSize   int               `json:"batch_size"`
	TimeoutMs   int               `json:"timeout_ms"`
	RetryMax    int               `json:"retry_max"`
	CreatedBy   string            `json:"created_by,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// CreateWebhookRequest represents the request body for creating a webhook.
type CreateWebhookRequest struct {
	Name        string            `json:"name"`
	URL         string            `json:"url"`
	AuthType    string            `json:"auth_type,omitempty"`
	AuthConfig  map[string]string `json:"auth_config,omitempty"`
	EventTypes  []string          `json:"event_types,omitempty"`
	EventFilter map[string]string `json:"event_filter,omitempty"`
	Enabled     *bool             `json:"enabled,omitempty"`
}

// UpdateWebhookRequest represents the request body for updating a webhook.
type UpdateWebhookRequest struct {
	Name        *string           `json:"name,omitempty"`
	URL         *string           `json:"url,omitempty"`
	AuthType    *string           `json:"auth_type,omitempty"`
	AuthConfig  map[string]string `json:"auth_config,omitempty"`
	EventTypes  []string          `json:"event_types,omitempty"`
	EventFilter map[string]string `json:"event_filter,omitempty"`
	Enabled     *bool             `json:"enabled,omitempty"`
}

// ListWebhooksResponse represents the response from listing webhooks.
type ListWebhooksResponse struct {
	Destinations []WebhookDestination `json:"destinations"`
}

// WebhookTestResult represents the result of testing a webhook.
type WebhookTestResult struct {
	Success        bool   `json:"success"`
	ResponseStatus int    `json:"response_status"`
	ResponseTimeMs int    `json:"response_time_ms"`
	ErrorMessage   string `json:"error_message,omitempty"`
}
