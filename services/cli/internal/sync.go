package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SyncService handles config synchronization
type SyncService struct {
	configManager    *ConfigManager
	platformClient   *PlatformClient
	authService      *AuthService
	dockerClient     *DockerClient
	containerManager *ContainerManager
}

// NewSyncService creates a new SyncService
func NewSyncService(configManager *ConfigManager, platformClient *PlatformClient, authService *AuthService) *SyncService {
	return &SyncService{
		configManager:  configManager,
		platformClient: platformClient,
		authService:    authService,
	}
}

// SetDockerClient sets the Docker client (optional for testing)
func (ss *SyncService) SetDockerClient(dockerClient *DockerClient) {
	ss.dockerClient = dockerClient
	if dockerClient != nil {
		ss.containerManager = NewContainerManager(dockerClient)
	}
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	AgentConfigs []AgentConfig
	UpdatedAt    time.Time
}

// Sync fetches configs from platform and stores them locally
func (ss *SyncService) Sync() (*SyncResult, error) {
	// Ensure user is authenticated
	config, err := ss.authService.RequireAuth()
	if err != nil {
		return nil, err
	}

	fmt.Println("✓ Fetching configs from platform...")

	// Fetch resolved agent configs
	agentConfigs, err := ss.platformClient.GetResolvedAgentConfigs(config.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agent configs: %w", err)
	}

	if len(agentConfigs) == 0 {
		fmt.Println("⚠ No agent configs found for your account")
		return &SyncResult{
			AgentConfigs: []AgentConfig{},
			UpdatedAt:    time.Now(),
		}, nil
	}

	// Save configs to local storage
	if err := ss.saveAgentConfigs(agentConfigs); err != nil {
		return nil, fmt.Errorf("failed to save agent configs: %w", err)
	}

	// Update last sync time
	config.LastSync = time.Now()
	if err := ss.configManager.Save(config); err != nil {
		return nil, fmt.Errorf("failed to update config: %w", err)
	}

	fmt.Printf("✓ Resolved configs for %d agent(s)\n", len(agentConfigs))
	for _, ac := range agentConfigs {
		if ac.IsEnabled {
			fmt.Printf("  • %s (%s)\n", ac.AgentName, ac.AgentType)
		}
	}

	return &SyncResult{
		AgentConfigs: agentConfigs,
		UpdatedAt:    time.Now(),
	}, nil
}

// saveAgentConfigs saves agent configs to ~/.ubik/agents/
func (ss *SyncService) saveAgentConfigs(configs []AgentConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	agentsDir := filepath.Join(homeDir, ".ubik", "agents")
	if err := os.MkdirAll(agentsDir, 0700); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	for _, config := range configs {
		agentDir := filepath.Join(agentsDir, config.AgentID)
		if err := os.MkdirAll(agentDir, 0700); err != nil {
			return fmt.Errorf("failed to create agent directory: %w", err)
		}

		// Save agent config (only the configuration map, not the whole struct)
		configPath := filepath.Join(agentDir, "config.json")
		configData, err := json.MarshalIndent(config.Configuration, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		if err := os.WriteFile(configPath, configData, 0600); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		// Save MCP servers config separately
		if len(config.MCPServers) > 0 {
			mcpPath := filepath.Join(agentDir, "mcp-servers.json")
			mcpData, err := json.MarshalIndent(config.MCPServers, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal MCP config: %w", err)
			}
			if err := os.WriteFile(mcpPath, mcpData, 0600); err != nil {
				return fmt.Errorf("failed to write MCP config file: %w", err)
			}
		}
	}

	return nil
}

// GetLocalAgentConfigs loads agent configs from local storage
func (ss *SyncService) GetLocalAgentConfigs() ([]AgentConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
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

// GetAgentConfig retrieves a specific agent config by ID or name
func (ss *SyncService) GetAgentConfig(idOrName string) (*AgentConfig, error) {
	configs, err := ss.GetLocalAgentConfigs()
	if err != nil {
		return nil, err
	}

	for _, config := range configs {
		if config.AgentID == idOrName || config.AgentName == idOrName {
			return &config, nil
		}
	}

	return nil, fmt.Errorf("agent config not found: %s", idOrName)
}

// StartContainers starts Docker containers for synced agent configs
func (ss *SyncService) StartContainers(workspacePath string, apiKey string) error {
	if ss.dockerClient == nil || ss.containerManager == nil {
		return fmt.Errorf("Docker client not configured")
	}

	// Fetch effective Claude token from platform
	fmt.Println("\nFetching Claude API token...")
	claudeToken, err := ss.platformClient.GetEffectiveClaudeToken()
	if err != nil {
		fmt.Printf("⚠ Warning: Could not fetch Claude token: %v\n", err)
		fmt.Printf("  Falling back to provided API key (if any)\n")
		// Continue anyway - we'll use the legacy apiKey if provided
		claudeToken = ""
	} else {
		fmt.Println("✓ Claude API token retrieved")
	}

	// Check Docker is running
	fmt.Println("\nChecking Docker...")
	if err := ss.dockerClient.Ping(); err != nil {
		return fmt.Errorf("Docker is not running: %w", err)
	}
	fmt.Println("✓ Docker is running")

	// Get Docker version
	version, err := ss.dockerClient.GetVersion()
	if err != nil {
		return fmt.Errorf("failed to get Docker version: %w", err)
	}
	fmt.Printf("✓ Docker version: %s\n", version)

	// Setup network
	if err := ss.containerManager.SetupNetwork(); err != nil {
		return fmt.Errorf("failed to setup network: %w", err)
	}

	// Get local agent configs
	configs, err := ss.GetLocalAgentConfigs()
	if err != nil {
		return fmt.Errorf("failed to get agent configs: %w", err)
	}

	if len(configs) == 0 {
		fmt.Println("⚠ No agent configs to start")
		return nil
	}

	// Start containers for each enabled agent
	fmt.Println("\n✓ Starting containers...")
	for _, config := range configs {
		if !config.IsEnabled {
			fmt.Printf("  Skipping %s (disabled)\n", config.AgentName)
			continue
		}

		// Start MCP servers first
		for _, mcp := range config.MCPServers {
			if !mcp.IsEnabled {
				continue
			}

			mcpSpec := MCPServerSpec{
				ServerID:   mcp.ServerID,
				ServerName: mcp.ServerName,
				ServerType: mcp.ServerType,
				Image:      fmt.Sprintf("ubik/mcp-%s:latest", mcp.ServerType),
				Port:       8001, // TODO: Get from config
				Config:     mcp.Config,
			}

			_, err := ss.containerManager.StartMCPServer(mcpSpec, workspacePath)
			if err != nil {
				fmt.Printf("  ⚠ Failed to start MCP server %s: %v\n", mcp.ServerName, err)
			}
		}

		// Use Claude token from organization/employee (hybrid auth)
		// This is centrally managed in the Settings page
		agentAPIKey := claudeToken
		if agentAPIKey != "" {
			fmt.Printf("  Using centralized Claude API token (organization/employee)\n")
		} else if apiKey != "" {
			// Fallback to legacy API key parameter (for backward compatibility)
			agentAPIKey = apiKey
			fmt.Printf("  Using legacy API key parameter\n")
		} else {
			fmt.Printf("  ⚠ Warning: No API token configured. Configure via Settings → Security tab\n")
		}

		// Start agent
		agentSpec := AgentSpec{
			AgentID:       config.AgentID,
			AgentName:     config.AgentName,
			AgentType:     config.AgentType,
			Image:         fmt.Sprintf("ubik/%s:latest", config.AgentType),
			Configuration: config.Configuration,
			MCPServers:    convertMCPServers(config.MCPServers),
			ClaudeToken:   agentAPIKey, // Use centralized token
			APIKey:        apiKey,      // Fallback for backward compatibility
		}

		_, err := ss.containerManager.StartAgent(agentSpec, workspacePath)
		if err != nil {
			fmt.Printf("  ⚠ Failed to start agent %s: %v\n", config.AgentName, err)
		}
	}

	return nil
}

// StopContainers stops all running containers
func (ss *SyncService) StopContainers() error {
	if ss.containerManager == nil {
		return fmt.Errorf("container manager not configured")
	}
	return ss.containerManager.StopContainers()
}

// GetContainerStatus returns the status of all containers
func (ss *SyncService) GetContainerStatus() ([]ContainerInfo, error) {
	if ss.containerManager == nil {
		return nil, fmt.Errorf("container manager not configured")
	}
	return ss.containerManager.GetContainerStatus()
}

// Helper to convert MCPServerConfig to MCPServerSpec
func convertMCPServers(configs []MCPServerConfig) []MCPServerSpec {
	specs := make([]MCPServerSpec, len(configs))
	for i, cfg := range configs {
		specs[i] = MCPServerSpec{
			ServerID:   cfg.ServerID,
			ServerName: cfg.ServerName,
			ServerType: cfg.ServerType,
			Image:      fmt.Sprintf("ubik/mcp-%s:latest", cfg.ServerType),
			Port:       8001 + i, // Simple port allocation
			Config:     cfg.Config,
		}
	}
	return specs
}
