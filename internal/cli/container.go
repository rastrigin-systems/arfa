package cli

import (
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
)

// ContainerManager manages Docker containers for agents and MCP servers
type ContainerManager struct {
	dockerClient *DockerClient
	networkName  string
}

// NewContainerManager creates a new container manager
func NewContainerManager(dockerClient *DockerClient) *ContainerManager {
	return &ContainerManager{
		dockerClient: dockerClient,
		networkName:  "ubik-network",
	}
}

// SetupNetwork creates the ubik network if it doesn't exist
func (cm *ContainerManager) SetupNetwork() error {
	exists, err := cm.dockerClient.NetworkExists(cm.networkName)
	if err != nil {
		return fmt.Errorf("failed to check network: %w", err)
	}

	if exists {
		fmt.Printf("✓ Network '%s' already exists\n", cm.networkName)
		return nil
	}

	fmt.Printf("Creating network '%s'...\n", cm.networkName)
	_, err = cm.dockerClient.CreateNetwork(cm.networkName)
	if err != nil {
		return fmt.Errorf("failed to create network: %w", err)
	}

	fmt.Printf("✓ Network '%s' created\n", cm.networkName)
	return nil
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

// AgentSpec defines configuration for an agent container
type AgentSpec struct {
	AgentID       string
	AgentName     string
	AgentType     string
	Image         string
	Configuration map[string]interface{}
	MCPServers    []MCPServerSpec
	APIKey        string // Anthropic API key or similar
}

// StartMCPServer starts an MCP server container
func (cm *ContainerManager) StartMCPServer(spec MCPServerSpec, workspacePath string) (string, error) {
	containerName := fmt.Sprintf("ubik-mcp-%s", spec.ServerID)

	fmt.Printf("  Starting %s (%s)...\n", spec.ServerName, spec.ServerType)

	// Pull image
	if err := cm.dockerClient.PullImage(spec.Image); err != nil {
		return "", fmt.Errorf("failed to pull MCP image: %w", err)
	}

	// Prepare container config
	config := &container.Config{
		Image: spec.Image,
		Env: []string{
			fmt.Sprintf("MCP_CONFIG=%s", toJSON(spec.Config)),
		},
		Labels: map[string]string{
			"com.ubik.managed":   "true",
			"com.ubik.type":      "mcp-server",
			"com.ubik.server-id": spec.ServerID,
		},
	}

	// Prepare host config with volume mount
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workspacePath,
				Target: "/workspace",
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// Add port binding if specified
	if spec.Port > 0 {
		config.ExposedPorts = nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", spec.Port)): struct{}{},
		}
		hostConfig.PortBindings = nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", spec.Port)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: fmt.Sprintf("%d", spec.Port),
				},
			},
		}
	}

	// Network config
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			cm.networkName: {},
		},
	}

	// Create container
	containerID, err := cm.dockerClient.CreateContainer(config, hostConfig, networkConfig, containerName)
	if err != nil {
		return "", err
	}

	// Start container
	if err := cm.dockerClient.StartContainer(containerID); err != nil {
		return "", err
	}

	fmt.Printf("  ✓ %s started (container: %s)\n", spec.ServerName, containerID[:12])
	return containerID, nil
}

// StartAgent starts an agent container
func (cm *ContainerManager) StartAgent(spec AgentSpec, workspacePath string) (string, error) {
	containerName := fmt.Sprintf("ubik-agent-%s", spec.AgentID)

	fmt.Printf("  Starting %s (%s)...\n", spec.AgentName, spec.AgentType)

	// Check if container with this name already exists and remove it
	if err := cm.dockerClient.RemoveContainerByName(containerName); err != nil {
		// Log warning but continue - container might not exist
		fmt.Printf("  Note: Cleaned up existing container\n")
	}

	// Pull image
	if err := cm.dockerClient.PullImage(spec.Image); err != nil {
		return "", fmt.Errorf("failed to pull agent image: %w", err)
	}

	// Prepare environment variables
	env := []string{
		fmt.Sprintf("AGENT_CONFIG=%s", toJSON(spec.Configuration)),
	}

	// Add API key if provided
	if spec.APIKey != "" {
		env = append(env, fmt.Sprintf("ANTHROPIC_API_KEY=%s", spec.APIKey))
	}

	// Add MCP config if there are MCP servers
	if len(spec.MCPServers) > 0 {
		mcpConfig := make(map[string]interface{})
		for _, mcp := range spec.MCPServers {
			mcpConfig[mcp.ServerType] = map[string]interface{}{
				"url": fmt.Sprintf("http://ubik-mcp-%s:%d", mcp.ServerID, mcp.Port),
			}
		}
		env = append(env, fmt.Sprintf("MCP_CONFIG=%s", toJSON(mcpConfig)))
	}

	// Prepare container config
	config := &container.Config{
		Image:        spec.Image,
		Env:          env,
		Tty:          true,
		OpenStdin:    true,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Labels: map[string]string{
			"com.ubik.managed":  "true",
			"com.ubik.type":     "agent",
			"com.ubik.agent-id": spec.AgentID,
		},
	}

	// Prepare host config with volume mount
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workspacePath,
				Target: "/workspace",
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	// Network config
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			cm.networkName: {},
		},
	}

	// Create container
	containerID, err := cm.dockerClient.CreateContainer(config, hostConfig, networkConfig, containerName)
	if err != nil {
		return "", err
	}

	// Start container
	if err := cm.dockerClient.StartContainer(containerID); err != nil {
		return "", err
	}

	fmt.Printf("  ✓ %s started (container: %s)\n", spec.AgentName, containerID[:12])
	return containerID, nil
}

// StopContainers stops all ubik-managed containers
func (cm *ContainerManager) StopContainers() error {
	containers, err := cm.dockerClient.ListContainers(false, map[string]string{
		"com.ubik.managed": "true",
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	if len(containers) == 0 {
		fmt.Println("No containers to stop")
		return nil
	}

	fmt.Printf("Stopping %d container(s)...\n", len(containers))
	timeout := 10 // seconds

	for _, c := range containers {
		fmt.Printf("  Stopping %s...\n", c.Name)
		if err := cm.dockerClient.StopContainer(c.ID, &timeout); err != nil {
			fmt.Printf("  ⚠ Failed to stop %s: %v\n", c.Name, err)
		} else {
			fmt.Printf("  ✓ %s stopped\n", c.Name)
		}
	}

	return nil
}

// CleanupContainers removes all stopped ubik-managed containers
func (cm *ContainerManager) CleanupContainers() error {
	containers, err := cm.dockerClient.ListContainers(true, map[string]string{
		"com.ubik.managed": "true",
	})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	stoppedCount := 0
	for _, c := range containers {
		if c.State != "running" {
			stoppedCount++
		}
	}

	if stoppedCount == 0 {
		fmt.Println("No stopped containers to clean up")
		return nil
	}

	fmt.Printf("Removing %d stopped container(s)...\n", stoppedCount)

	for _, c := range containers {
		if c.State != "running" {
			fmt.Printf("  Removing %s...\n", c.Name)
			if err := cm.dockerClient.RemoveContainer(c.ID, false); err != nil {
				fmt.Printf("  ⚠ Failed to remove %s: %v\n", c.Name, err)
			} else {
				fmt.Printf("  ✓ %s removed\n", c.Name)
			}
		}
	}

	return nil
}

// GetContainerStatus returns status of all ubik-managed containers
func (cm *ContainerManager) GetContainerStatus() ([]ContainerInfo, error) {
	return cm.dockerClient.ListContainers(true, map[string]string{
		"com.ubik.managed": "true",
	})
}

// Helper function to convert map to JSON string
func toJSON(v interface{}) string {
	// This is a simplified version - in production, use proper JSON marshaling
	return fmt.Sprintf("%v", v)
}

// GetWorkspacePath prompts user for workspace or uses default
func GetWorkspacePath(defaultPath string) (string, error) {
	// For now, just return the default
	// In Phase 3, we'll add interactive prompt
	absPath, err := filepath.Abs(defaultPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace path: %w", err)
	}
	return absPath, nil
}
