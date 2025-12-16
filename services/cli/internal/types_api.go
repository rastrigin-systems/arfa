// Package cli contains type aliases for backward compatibility.
// All types are now defined in their respective packages (api, config, auth).
package cli

import (
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
)

// ============================================================================
// Type Aliases - pointing to api package
// ============================================================================

// Authentication Types
type (
	LoginRequest      = api.LoginRequest
	LoginResponse     = api.LoginResponse
	LoginEmployeeInfo = api.LoginEmployeeInfo
	EmployeeInfo      = api.EmployeeInfo
)

// Agent Configuration Types
type (
	AgentConfigAPIResponse      = api.AgentConfigAPIResponse
	AgentConfig                 = api.AgentConfig
	MCPServerConfig             = api.MCPServerConfig
	ResolvedConfigsResponse     = api.ResolvedConfigsResponse
	OrgAgentConfigResponse      = api.OrgAgentConfigResponse
	TeamAgentConfigResponse     = api.TeamAgentConfigResponse
	EmployeeAgentConfigResponse = api.EmployeeAgentConfigResponse
)

// Token Types
type (
	ClaudeTokenStatusResponse    = api.ClaudeTokenStatusResponse
	EffectiveClaudeTokenResponse = api.EffectiveClaudeTokenResponse
)

// Sync Types
type (
	ClaudeCodeSyncResponse = api.ClaudeCodeSyncResponse
	AgentConfigSync        = api.AgentConfigSync
	SkillConfigSync        = api.SkillConfigSync
	MCPServerConfigSync    = api.MCPServerConfigSync
)

// Skill Types
type (
	Skill                      = api.Skill
	SkillFile                  = api.SkillFile
	EmployeeSkill              = api.EmployeeSkill
	ListSkillsResponse         = api.ListSkillsResponse
	ListEmployeeSkillsResponse = api.ListEmployeeSkillsResponse
)

// Logging Types
type (
	LogEntry         = api.LogEntry
	CreateLogRequest = api.CreateLogRequest
	LogEntryResponse = api.LogEntryResponse
	LogsResponse     = api.LogsResponse
	GetLogsParams    = api.GetLogsParams
)

// Legacy aliases for backward compatibility
type (
	APILogsResponse = api.LogsResponse
	APILogEntry     = api.LogEntryResponse
)
