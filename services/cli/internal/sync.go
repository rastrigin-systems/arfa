package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SyncService handles config synchronization
type SyncService struct {
	configManager    ConfigManagerInterface
	platformClient   PlatformClientInterface
	authService      AuthServiceInterface
	dockerClient     *DockerClient
	containerManager ContainerManagerInterface
}

// NewSyncService creates a new SyncService with concrete types.
// This is the primary constructor for production use.
func NewSyncService(configManager *ConfigManager, platformClient *PlatformClient, authService *AuthService) *SyncService {
	return &SyncService{
		configManager:  configManager,
		platformClient: platformClient,
		authService:    authService,
	}
}

// NewSyncServiceWithInterfaces creates a new SyncService with interface types.
// This constructor enables dependency injection for testing with mocks.
func NewSyncServiceWithInterfaces(configManager ConfigManagerInterface, platformClient PlatformClientInterface, authService AuthServiceInterface) *SyncService {
	return &SyncService{
		configManager:  configManager,
		platformClient: platformClient,
		authService:    authService,
	}
}

// SetDockerClient sets the Docker client and creates a ContainerManager.
// Deprecated: Use SetContainerManager instead for better dependency injection.
func (ss *SyncService) SetDockerClient(dockerClient *DockerClient) {
	ss.dockerClient = dockerClient
	if dockerClient != nil {
		ss.containerManager = NewContainerManager(dockerClient)
	}
}

// SetContainerManager sets the container manager directly.
// This is the preferred way to inject container management dependencies.
func (ss *SyncService) SetContainerManager(cm ContainerManagerInterface) {
	ss.containerManager = cm
}

// SyncResult represents the result of a sync operation
type SyncResult struct {
	AgentConfigs []AgentConfig
	UpdatedAt    time.Time
}

// Sync fetches configs from platform and stores them locally
func (ss *SyncService) Sync(ctx context.Context) (*SyncResult, error) {
	// Ensure user is authenticated
	config, err := ss.authService.RequireAuth()
	if err != nil {
		return nil, err
	}

	fmt.Println("✓ Fetching configs from platform...")

	// Fetch resolved agent configs (using JWT-based /employees/me endpoint)
	agentConfigs, err := ss.platformClient.GetMyResolvedAgentConfigs(ctx)
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

	// Clear default agent if it's no longer in the resolved configs
	if config.DefaultAgent != "" {
		defaultStillValid := false
		for _, ac := range agentConfigs {
			if ac.AgentID == config.DefaultAgent || ac.AgentName == config.DefaultAgent {
				defaultStillValid = true
				break
			}
		}
		if !defaultStillValid {
			config.DefaultAgent = ""
		}
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

// SyncClaudeCode fetches and installs complete Claude Code configuration
func (ss *SyncService) SyncClaudeCode(ctx context.Context, targetDir string) error {
	// Ensure user is authenticated
	_, err := ss.authService.RequireAuth()
	if err != nil {
		return err
	}

	fmt.Println("🔄 Syncing Claude Code configuration...")
	fmt.Println()

	// Fetch complete Claude Code config
	fmt.Println("📥 Downloading configurations...")
	config, err := ss.platformClient.GetClaudeCodeConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch Claude Code config: %w", err)
	}

	// Create directory structure
	claudeDir := filepath.Join(targetDir, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")
	skillsDir := filepath.Join(claudeDir, "skills")

	// Write agent files
	if len(config.Agents) > 0 {
		if err := WriteAgentFiles(agentsDir, config.Agents); err != nil {
			return fmt.Errorf("failed to write agent files: %w", err)
		}
		fmt.Printf("  ✓ %d agents (.claude/agents/)\n", len(config.Agents))
		for _, agent := range config.Agents {
			if agent.IsEnabled {
				fmt.Printf("    - %s\n", agent.Filename)
			}
		}
		fmt.Println()
	}

	// Write skill files
	if len(config.Skills) > 0 {
		if err := WriteSkillFiles(skillsDir, config.Skills); err != nil {
			return fmt.Errorf("failed to write skill files: %w", err)
		}
		fmt.Printf("  ✓ %d skills (.claude/skills/)\n", len(config.Skills))
		for _, skill := range config.Skills {
			if skill.IsEnabled {
				fmt.Printf("    - %s/\n", skill.Name)
			}
		}
		fmt.Println()
	}

	// Configure MCP servers
	if len(config.MCPServers) > 0 {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		claudeConfigPath := filepath.Join(homeDir, ".claude.json")
		if err := MergeMCPConfig(claudeConfigPath, config.MCPServers); err != nil {
			return fmt.Errorf("failed to configure MCP servers: %w", err)
		}
		fmt.Printf("  ✓ %d MCP servers (configured in ~/.claude.json)\n", len(config.MCPServers))
		for _, mcp := range config.MCPServers {
			if mcp.IsEnabled {
				fmt.Printf("    - %s\n", mcp.Name)
			}
		}
		fmt.Println()
	}

	fmt.Println("✅ Claude Code ready! Run 'claude' to start.")

	return nil
}

// agentMetadata stores metadata about an agent (separate from config)
type agentMetadata struct {
	AgentID     string `json:"agent_id"`
	AgentName   string `json:"agent_name"`
	AgentType   string `json:"agent_type"`
	Provider    string `json:"provider"`
	DockerImage string `json:"docker_image"`
	IsEnabled   bool   `json:"is_enabled"`
}

// saveAgentConfigs saves agent configs to ~/.ubik/config/agents/
// It also removes any agent directories that are not in the new config
func (ss *SyncService) saveAgentConfigs(configs []AgentConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	agentsDir := filepath.Join(homeDir, ".ubik", "config", "agents")
	if err := os.MkdirAll(agentsDir, 0700); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Build a set of agent IDs that should be kept
	activeAgentIDs := make(map[string]bool)
	for _, config := range configs {
		activeAgentIDs[config.AgentID] = true
	}

	// Clean up agents that are no longer in the resolved configs
	entries, err := os.ReadDir(agentsDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !activeAgentIDs[entry.Name()] {
				// This agent is no longer in resolved configs, remove it
				oldAgentDir := filepath.Join(agentsDir, entry.Name())
				if err := os.RemoveAll(oldAgentDir); err != nil {
					// Log but don't fail - best effort cleanup
					fmt.Printf("  ⚠ Warning: failed to remove old agent config %s: %v\n", entry.Name(), err)
				}
			}
		}
	}

	for _, config := range configs {
		agentDir := filepath.Join(agentsDir, config.AgentID)
		if err := os.MkdirAll(agentDir, 0700); err != nil {
			return fmt.Errorf("failed to create agent directory: %w", err)
		}

		// Save agent metadata (ID, name, type, etc.)
		metadataPath := filepath.Join(agentDir, "metadata.json")
		metadata := agentMetadata{
			AgentID:     config.AgentID,
			AgentName:   config.AgentName,
			AgentType:   config.AgentType,
			Provider:    config.Provider,
			DockerImage: config.DockerImage,
			IsEnabled:   config.IsEnabled,
		}
		metadataData, err := json.MarshalIndent(metadata, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		if err := os.WriteFile(metadataPath, metadataData, 0600); err != nil {
			return fmt.Errorf("failed to write metadata file: %w", err)
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

	agentsDir := filepath.Join(homeDir, ".ubik", "config", "agents")

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

		agentDir := filepath.Join(agentsDir, entry.Name())

		// Load metadata
		metadataPath := filepath.Join(agentDir, "metadata.json")
		metadataData, err := os.ReadFile(metadataPath)
		if err != nil {
			// Skip if metadata file doesn't exist
			continue
		}

		var metadata agentMetadata
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
		var mcpServers []MCPServerConfig
		mcpPath := filepath.Join(agentDir, "mcp-servers.json")
		if mcpData, err := os.ReadFile(mcpPath); err == nil {
			if err := json.Unmarshal(mcpData, &mcpServers); err != nil {
				return nil, fmt.Errorf("failed to unmarshal MCP config: %w", err)
			}
		}

		// Build full AgentConfig
		config := AgentConfig{
			AgentID:       metadata.AgentID,
			AgentName:     metadata.AgentName,
			AgentType:     metadata.AgentType,
			Provider:      metadata.Provider,
			DockerImage:   metadata.DockerImage,
			IsEnabled:     metadata.IsEnabled,
			Configuration: configuration,
			MCPServers:    mcpServers,
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
func (ss *SyncService) StartContainers(ctx context.Context, workspacePath string, apiKey string) error {
	if ss.dockerClient == nil || ss.containerManager == nil {
		return fmt.Errorf("Docker client not configured")
	}

	// Fetch effective Claude token from platform
	fmt.Println("\nFetching Claude API token...")
	claudeToken, err := ss.platformClient.GetEffectiveClaudeToken(ctx)
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
	if err := ss.dockerClient.Ping(ctx); err != nil {
		return fmt.Errorf("Docker is not running: %w", err)
	}
	fmt.Println("✓ Docker is running")

	// Get Docker version
	version, err := ss.dockerClient.GetVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Docker version: %w", err)
	}
	fmt.Printf("✓ Docker version: %s\n", version)

	// Setup network
	if err := ss.containerManager.SetupNetwork(ctx); err != nil {
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

			_, err := ss.containerManager.StartMCPServer(ctx, mcpSpec, workspacePath)
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

		// Start agent - use DockerImage from config (or fallback to constructed image)
		dockerImage := config.DockerImage
		if dockerImage == "" {
			dockerImage = fmt.Sprintf("ubik/%s:latest", config.AgentType)
		}
		agentSpec := AgentSpec{
			AgentID:       config.AgentID,
			AgentName:     config.AgentName,
			AgentType:     config.AgentType,
			Image:         dockerImage,
			Configuration: config.Configuration,
			MCPServers:    convertMCPServers(config.MCPServers),
			ClaudeToken:   agentAPIKey, // Use centralized token
			APIKey:        apiKey,      // Fallback for backward compatibility
		}

		_, err := ss.containerManager.StartAgent(ctx, agentSpec, workspacePath)
		if err != nil {
			fmt.Printf("  ⚠ Failed to start agent %s: %v\n", config.AgentName, err)
		}
	}

	return nil
}

// StopContainers stops all running containers
func (ss *SyncService) StopContainers(ctx context.Context) error {
	if ss.containerManager == nil {
		return fmt.Errorf("container manager not configured")
	}
	return ss.containerManager.StopContainers(ctx)
}

// GetContainerStatus returns the status of all containers
func (ss *SyncService) GetContainerStatus(ctx context.Context) ([]ContainerInfo, error) {
	if ss.containerManager == nil {
		return nil, fmt.Errorf("container manager not configured")
	}
	return ss.containerManager.GetContainerStatus(ctx)
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
