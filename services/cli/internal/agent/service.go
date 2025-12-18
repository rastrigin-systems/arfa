package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
)

// Service handles agent management operations
type Service struct {
	client        APIClientInterface
	configManager ConfigManagerInterface
}

// NewService creates a new agent service with concrete types.
// This is the primary constructor for production use.
func NewService(client *api.Client, configManager *config.Manager) *Service {
	return &Service{
		client:        client,
		configManager: configManager,
	}
}

// NewServiceWithInterfaces creates a new agent service with interface types.
// This constructor enables dependency injection for testing with mocks.
func NewServiceWithInterfaces(client APIClientInterface, configManager ConfigManagerInterface) *Service {
	return &Service{
		client:        client,
		configManager: configManager,
	}
}

// ListAgents fetches all available agents from the platform
func (s *Service) ListAgents(ctx context.Context) ([]Agent, error) {
	var resp ListAgentsResponse
	if err := s.client.DoRequest(ctx, "GET", "/agents", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	return resp.Agents, nil
}

// GetAgent fetches details for a specific agent
func (s *Service) GetAgent(ctx context.Context, agentID string) (*Agent, error) {
	var agent Agent
	endpoint := fmt.Sprintf("/agents/%s", agentID)
	if err := s.client.DoRequest(ctx, "GET", endpoint, nil, &agent); err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	return &agent, nil
}

// ListEmployeeAgentConfigs fetches employee's assigned agent configs
func (s *Service) ListEmployeeAgentConfigs(ctx context.Context, employeeID string) ([]EmployeeAgentConfig, error) {
	var resp ListEmployeeAgentConfigsResponse
	endpoint := fmt.Sprintf("/employees/%s/agent-configs", employeeID)
	if err := s.client.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list employee agent configs: %w", err)
	}
	return resp.AgentConfigs, nil
}

// RequestAgent creates an employee agent configuration (request for access)
func (s *Service) RequestAgent(ctx context.Context, employeeID, agentID string) error {
	reqBody := CreateEmployeeAgentConfigRequest{
		AgentID:   agentID,
		Config:    nil, // Use default config
		IsEnabled: true,
	}

	endpoint := fmt.Sprintf("/employees/%s/agent-configs", employeeID)
	if err := s.client.DoRequest(ctx, "POST", endpoint, reqBody, nil); err != nil {
		return fmt.Errorf("failed to request agent: %w", err)
	}

	return nil
}

// CheckForUpdates checks if there are config updates available
func (s *Service) CheckForUpdates(ctx context.Context, employeeID string) (bool, error) {
	// Get local configs from ~/.ubik/agents/
	localConfigs, err := s.getLocalAgentConfigsInternal()
	if err != nil {
		return false, fmt.Errorf("failed to read local configs: %w", err)
	}

	// Get remote configs
	remoteConfigs, err := s.client.GetResolvedAgentConfigs(ctx, employeeID)
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
func (s *Service) GetLocalAgents() ([]api.AgentConfig, error) {
	return s.getLocalAgentConfigsInternal()
}

// getLocalAgentConfigsInternal reads agent configs from ~/.ubik/config/agents/ directory
func (s *Service) getLocalAgentConfigsInternal() ([]api.AgentConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	agentsDir := filepath.Join(homeDir, ".ubik", "config", "agents")

	// Check if agents directory exists
	if _, err := os.Stat(agentsDir); os.IsNotExist(err) {
		return []api.AgentConfig{}, nil
	}

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents directory: %w", err)
	}

	var configs []api.AgentConfig
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		agentDir := filepath.Join(agentsDir, entry.Name())

		// Load metadata
		metadataPath := filepath.Join(agentDir, "metadata.json")
		metadataData, err := os.ReadFile(metadataPath)
		if err != nil {
			// Skip if metadata file doesn't exist
			continue
		}

		// Parse metadata to get agent info
		var metadata struct {
			AgentID     string `json:"agent_id"`
			AgentName   string `json:"agent_name"`
			AgentType   string `json:"agent_type"`
			Provider    string `json:"provider"`
			DockerImage string `json:"docker_image"`
			IsEnabled   bool   `json:"is_enabled"`
		}
		if err := json.Unmarshal(metadataData, &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		// Load configuration
		configPath := filepath.Join(agentDir, "config.json")
		configData, err := os.ReadFile(configPath)
		if err != nil {
			// Skip if config file doesn't exist
			continue
		}

		var configuration map[string]interface{}
		if err := json.Unmarshal(configData, &configuration); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		// Load MCP servers if they exist
		var mcpServers []api.MCPServerConfig
		mcpPath := filepath.Join(agentDir, "mcp-servers.json")
		if mcpData, err := os.ReadFile(mcpPath); err == nil {
			if err := json.Unmarshal(mcpData, &mcpServers); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MCP servers: %w", err)
			}
		}

		configs = append(configs, api.AgentConfig{
			AgentID:       metadata.AgentID,
			AgentName:     metadata.AgentName,
			AgentType:     metadata.AgentType,
			Provider:      metadata.Provider,
			DockerImage:   metadata.DockerImage,
			IsEnabled:     metadata.IsEnabled,
			Configuration: configuration,
			MCPServers:    mcpServers,
		})
	}

	return configs, nil
}
