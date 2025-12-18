package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
)

// SyncHandler handles sync-related endpoints
type SyncHandler struct {
	db db.Querier
}

// NewSyncHandler creates a new sync handler
func NewSyncHandler(database db.Querier) *SyncHandler {
	return &SyncHandler{db: database}
}

// AgentConfig represents an agent configuration in the sync response
type AgentConfig struct {
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

// SkillConfig represents a skill configuration in the sync response
type SkillConfig struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Category     string                 `json:"category,omitempty"`
	Version      string                 `json:"version"`
	Files        []map[string]string    `json:"files,omitempty"`
	Dependencies map[string]interface{} `json:"dependencies,omitempty"`
	IsEnabled    bool                   `json:"is_enabled"`
}

// MCPServerConfig represents an MCP server configuration in the sync response
type MCPServerConfig struct {
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

// ClaudeCodeSyncResponse represents the complete sync response
type ClaudeCodeSyncResponse struct {
	Agents     []AgentConfig     `json:"agents"`
	Skills     []SkillConfig     `json:"skills"`
	MCPServers []MCPServerConfig `json:"mcp_servers"`
	Version    string            `json:"version"`
	SyncedAt   string            `json:"synced_at"`
}

// GetClaudeCodeSync returns complete Claude Code configuration bundle
// GET /api/v1/sync/claude-code
func (h *SyncHandler) GetClaudeCodeSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get employee ID from middleware context
	employeeID, err := GetEmployeeID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Missing authentication")
		return
	}

	// Fetch all agent configs
	agentRows, err := h.db.ListEmployeeAgentConfigs(ctx, employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch agent configurations")
		return
	}

	// Fetch all skills
	skillRows, err := h.db.ListEmployeeSkills(ctx, employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch skills")
		return
	}

	// Fetch all MCP configs
	mcpRows, err := h.db.ListEmployeeMCPConfigs(ctx, employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch MCP configurations")
		return
	}

	// Build response
	response := ClaudeCodeSyncResponse{
		Agents:     make([]AgentConfig, 0, len(agentRows)),
		Skills:     make([]SkillConfig, 0, len(skillRows)),
		MCPServers: make([]MCPServerConfig, 0, len(mcpRows)),
		Version:    "1.0.0",
		SyncedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	// Convert agent rows to response format
	for _, row := range agentRows {
		// Merge default config with override
		config := make(map[string]interface{})
		if row.AgentDefaultConfig != nil {
			json.Unmarshal(row.AgentDefaultConfig, &config)
		}
		if row.ConfigOverride != nil {
			var override map[string]interface{}
			json.Unmarshal(row.ConfigOverride, &override)
			for k, v := range override {
				config[k] = v
			}
		}

		response.Agents = append(response.Agents, AgentConfig{
			ID:        row.ID.String(),
			Name:      row.AgentName,
			Type:      row.AgentType,
			Filename:  row.AgentName + ".md",
			Config:    config,
			Provider:  row.AgentProvider,
			IsEnabled: row.IsEnabled,
			Version:   "1.0.0",
		})
	}

	// Convert skill rows to response format
	for _, row := range skillRows {
		var files []map[string]string
		if row.Files != nil {
			json.Unmarshal(row.Files, &files)
		}

		var deps map[string]interface{}
		if row.Dependencies != nil {
			json.Unmarshal(row.Dependencies, &deps)
		}

		var description, category string
		if row.Description != nil {
			description = *row.Description
		}
		if row.Category != nil {
			category = *row.Category
		}

		isEnabled := false
		if row.IsEnabled != nil {
			isEnabled = *row.IsEnabled
		}

		response.Skills = append(response.Skills, SkillConfig{
			ID:           row.ID.String(),
			Name:         row.Name,
			Description:  description,
			Category:     category,
			Version:      row.Version,
			Files:        files,
			Dependencies: deps,
			IsEnabled:    isEnabled,
		})
	}

	// Convert MCP rows to response format
	for _, row := range mcpRows {
		var config map[string]interface{}
		if row.ConnectionConfig != nil {
			json.Unmarshal(row.ConnectionConfig, &config)
		}

		var requiredVars []string
		if row.RequiredEnvVars != nil {
			json.Unmarshal(row.RequiredEnvVars, &requiredVars)
		}

		dockerImage := ""
		if row.DockerImage != nil {
			dockerImage = *row.DockerImage
		}

		isEnabled := false
		if row.IsEnabled != nil {
			isEnabled = *row.IsEnabled
		}

		response.MCPServers = append(response.MCPServers, MCPServerConfig{
			ID:              row.ID.String(),
			Name:            row.Name,
			Provider:        row.Provider,
			Version:         row.Version,
			Description:     row.Description,
			DockerImage:     dockerImage,
			Config:          config,
			RequiredEnvVars: requiredVars,
			IsEnabled:       isEnabled,
		})
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
