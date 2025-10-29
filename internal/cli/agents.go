package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AgentService handles agent management operations
type AgentService struct {
	client        *PlatformClient
	configManager *ConfigManager
}

// NewAgentService creates a new agent service
func NewAgentService(client *PlatformClient, configManager *ConfigManager) *AgentService {
	return &AgentService{
		client:        client,
		configManager: configManager,
	}
}

// Agent represents an agent from the catalog
type Agent struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Provider           string                 `json:"provider"`
	Description        string                 `json:"description"`
	LogoURL            string                 `json:"logo_url"`
	SupportedPlatforms []string               `json:"supported_platforms"`
	PricingTier        string                 `json:"pricing_tier"`
	DefaultConfig      map[string]interface{} `json:"default_config"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// ListAgentsResponse represents the response from list agents endpoint
type ListAgentsResponse struct {
	Agents []Agent `json:"agents"`
	Total  int     `json:"total"`
}

// EmployeeAgentConfig represents an employee's agent configuration
type EmployeeAgentConfig struct {
	ID         string                 `json:"id"`
	EmployeeID string                 `json:"employee_id"`
	AgentID    string                 `json:"agent_id"`
	AgentName  string                 `json:"agent_name"`
	Config     map[string]interface{} `json:"config"`
	IsEnabled  bool                   `json:"is_enabled"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// ListEmployeeAgentConfigsResponse represents employee agent configs
type ListEmployeeAgentConfigsResponse struct {
	AgentConfigs []EmployeeAgentConfig `json:"agent_configs"`
	Total        int                   `json:"total"`
}

// CreateEmployeeAgentConfigRequest represents a request to create an agent config
type CreateEmployeeAgentConfigRequest struct {
	AgentID   string                 `json:"agent_id"`
	Config    map[string]interface{} `json:"config,omitempty"`
	IsEnabled bool                   `json:"is_enabled"`
}

// ListAgents fetches all available agents from the platform
func (as *AgentService) ListAgents() ([]Agent, error) {
	var resp ListAgentsResponse
	if err := as.client.doRequest("GET", "/agents", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	return resp.Agents, nil
}

// GetAgent fetches details for a specific agent
func (as *AgentService) GetAgent(agentID string) (*Agent, error) {
	var agent Agent
	endpoint := fmt.Sprintf("/agents/%s", agentID)
	if err := as.client.doRequest("GET", endpoint, nil, &agent); err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	return &agent, nil
}

// ListEmployeeAgentConfigs fetches employee's assigned agent configs
func (as *AgentService) ListEmployeeAgentConfigs(employeeID string) ([]EmployeeAgentConfig, error) {
	var resp ListEmployeeAgentConfigsResponse
	endpoint := fmt.Sprintf("/employees/%s/agent-configs", employeeID)
	if err := as.client.doRequest("GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list employee agent configs: %w", err)
	}
	return resp.AgentConfigs, nil
}

// RequestAgent creates an employee agent configuration (request for access)
func (as *AgentService) RequestAgent(employeeID, agentID string) error {
	reqBody := CreateEmployeeAgentConfigRequest{
		AgentID:   agentID,
		Config:    nil, // Use default config
		IsEnabled: true,
	}

	endpoint := fmt.Sprintf("/employees/%s/agent-configs", employeeID)
	if err := as.client.doRequest("POST", endpoint, reqBody, nil); err != nil {
		return fmt.Errorf("failed to request agent: %w", err)
	}

	return nil
}

// CheckForUpdates checks if there are config updates available
func (as *AgentService) CheckForUpdates(employeeID string) (bool, error) {
	// Get local configs from ~/.ubik/agents/
	localConfigs, err := as.getLocalAgentConfigsInternal()
	if err != nil {
		return false, fmt.Errorf("failed to read local configs: %w", err)
	}

	// Get remote configs
	remoteConfigs, err := as.client.GetResolvedAgentConfigs(employeeID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch remote configs: %w", err)
	}

	// Simple check: if counts differ, there are updates
	if len(localConfigs) != len(remoteConfigs) {
		return true, nil
	}

	// Check for config changes (compare agent IDs)
	localAgentIDs := make(map[string]bool)
	for _, agent := range localConfigs {
		localAgentIDs[agent.AgentID] = true
	}

	for _, remoteAgent := range remoteConfigs {
		if !localAgentIDs[remoteAgent.AgentID] {
			return true, nil // New agent found
		}
	}

	// TODO: Deep comparison of config content
	// For now, just checking presence/absence

	return false, nil
}

// GetLocalAgents returns locally configured agents
func (as *AgentService) GetLocalAgents() ([]AgentConfig, error) {
	return as.getLocalAgentConfigsInternal()
}

// getLocalAgentConfigsInternal reads agent configs from ~/.ubik/agents/ directory
func (as *AgentService) getLocalAgentConfigsInternal() ([]AgentConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	agentsDir := filepath.Join(homeDir, ".ubik", "agents")

	// Check if agents directory exists
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return []AgentConfig{}, nil
	}

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents directory: %w", err)
	}

	var configs []AgentConfig
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		configPath := filepath.Join(agentsDir, entry.Name(), "config.json")
		configData, err := os.ReadFile(configPath)
		if err != nil {
			// Skip if config file doesn't exist
			continue
		}

		var config AgentConfig
		if err := json.Unmarshal(configData, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		configs = append(configs, config)
	}

	return configs, nil
}
