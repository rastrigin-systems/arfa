package docker

// ContainerInfo represents basic container information
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

// ProxyConfig defines configuration for the MITM proxy
type ProxyConfig struct {
	Host     string
	Port     int
	CertPath string
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
	ProxyConfig   *ProxyConfig
}

// RunnerConfig contains configuration for starting an agent
type RunnerConfig struct {
	AgentType   string
	AgentID     string
	AgentName   string
	Workspace   string
	APIKey      string
	ProxyPort   int
	CertPath    string
	SessionID   string
	Token       string            // JWT token for proxy authentication
	EmployeeID  string            // Employee ID (for backward compatibility, prefer Token)
	Environment map[string]string // Additional env vars from agent config
}

// AgentTypeMapping maps API agent types to CLI agent types
// This translates the database agent_type values to the binary names
var AgentTypeMapping = map[string]string{
	// API types -> CLI types
	"ide_assistant":   "claude-code",
	"code_completion": "cursor",
	"ai_editor":       "windsurf",
	"gemini_agent":    "gemini",
	"pair_programmer": "aider",
	// Also allow direct CLI types for backwards compatibility
	"claude-code": "claude-code",
	"cursor":      "cursor",
	"windsurf":    "windsurf",
	"gemini":      "gemini",
	"aider":       "aider",
}

// AgentBinaries maps CLI agent types to their binary names
var AgentBinaries = map[string][]string{
	"claude-code": {"claude"},
	"cursor":      {"cursor"},
	"windsurf":    {"windsurf"},
	"gemini":      {"gemini"},
	"aider":       {"aider"},
}

// AgentEnvVars maps agent types to their API key environment variable names
var AgentEnvVars = map[string]string{
	"claude-code": "ANTHROPIC_API_KEY",
	"cursor":      "ANTHROPIC_API_KEY",
	"windsurf":    "ANTHROPIC_API_KEY",
	"gemini":      "GEMINI_API_KEY",
	"aider":       "ANTHROPIC_API_KEY",
}
