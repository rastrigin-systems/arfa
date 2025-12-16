package agent

import "time"

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
