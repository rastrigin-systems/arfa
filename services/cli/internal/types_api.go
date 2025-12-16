// Package cli contains API request/response types for the Ubik platform.
package cli

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
	TeamID   *string `json:"team_id"` // nullable
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
	DockerImage  *string                `json:"docker_image"` // Docker image reference (nullable)
	SyncToken    string                 `json:"sync_token"`
	SystemPrompt string                 `json:"system_prompt"`
	LastSyncedAt *string                `json:"last_synced_at"` // nullable timestamp
}

// AgentConfig represents a resolved agent configuration (internal use).
type AgentConfig struct {
	AgentID       string                 `json:"agent_id"`
	AgentName     string                 `json:"agent_name"`
	AgentType     string                 `json:"agent_type"`
	Provider      string                 `json:"provider"`
	DockerImage   string                 `json:"docker_image"` // Docker image reference
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
// Note: This is different from EmployeeAgentConfig in agents.go which includes
// additional fields like EmployeeID, CreatedAt, UpdatedAt for local representation.
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
	ActiveTokenSource string `json:"active_token_source"` // "personal", "company", or "none"
}

// EffectiveClaudeTokenResponse represents the effective token response.
type EffectiveClaudeTokenResponse struct {
	Token      string `json:"token"`
	Source     string `json:"source"` // "personal" or "company"
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
	AgentID       string                 `json:"agent_id,omitempty"`
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category"`
	Content       string                 `json:"content,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
}

// CreateLogRequest represents a single log creation request.
type CreateLogRequest struct {
	SessionID     *string                 `json:"session_id,omitempty"`
	AgentID       *string                 `json:"agent_id,omitempty"`
	EventType     string                  `json:"event_type"`
	EventCategory string                  `json:"event_category"`
	Content       *string                 `json:"content,omitempty"`
	Payload       *map[string]interface{} `json:"payload,omitempty"`
}
