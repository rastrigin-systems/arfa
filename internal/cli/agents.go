package cli

import (
	"fmt"
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
	// TODO: Implement agent config update checking
	// This requires integrating with SyncService
	return false, fmt.Errorf("not implemented yet")
}

// GetLocalAgents returns locally configured agents
func (as *AgentService) GetLocalAgents() ([]AgentConfig, error) {
	// TODO: Implement local agent listing
	// This requires integrating with SyncService
	return nil, fmt.Errorf("not implemented yet")
}
