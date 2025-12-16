package cli

import (
	"context"
	"fmt"
	"os"
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
func (cm *ContainerManager) SetupNetwork(ctx context.Context) error {
	exists, err := cm.dockerClient.NetworkExists(ctx, cm.networkName)
	if err != nil {
		return fmt.Errorf("failed to check network: %w", err)
	}

	if exists {
		fmt.Printf("✓ Network '%s' already exists\n", cm.networkName)
		return nil
	}

	fmt.Printf("Creating network '%s'...\n", cm.networkName)
	_, err = cm.dockerClient.CreateNetwork(ctx, cm.networkName)
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

// StartMCPServer starts an MCP server container
func (cm *ContainerManager) StartMCPServer(ctx context.Context, spec MCPServerSpec, workspacePath string) (string, error) {
	containerName := fmt.Sprintf("ubik-mcp-%s", spec.ServerID)

	fmt.Printf("  Starting %s (%s)...\n", spec.ServerName, spec.ServerType)

	// Pull image
	if err := cm.dockerClient.PullImage(ctx, spec.Image); err != nil {
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
	containerID, err := cm.dockerClient.CreateContainer(ctx, config, hostConfig, networkConfig, containerName)
	if err != nil {
		return "", err
	}

	// Start container
	if err := cm.dockerClient.StartContainer(ctx, containerID); err != nil {
		return "", err
	}

	fmt.Printf("  ✓ %s started (container: %s)\n", spec.ServerName, containerID[:12])
	return containerID, nil
}

// StartAgent starts an agent container
func (cm *ContainerManager) StartAgent(ctx context.Context, spec AgentSpec, workspacePath string) (string, error) {
	containerName := fmt.Sprintf("ubik-agent-%s", spec.AgentID)

	fmt.Printf("  Starting %s (%s)...\n", spec.AgentName, spec.AgentType)

	// Check if container with this name already exists and remove it
	if err := cm.dockerClient.RemoveContainerByName(ctx, containerName); err != nil {
		// Log warning but continue - container might not exist
		fmt.Printf("  Note: Cleaned up existing container\n")
	}

	// Pull image
	if err := cm.dockerClient.PullImage(ctx, spec.Image); err != nil {
		return "", fmt.Errorf("failed to pull agent image: %w", err)
	}

	// Prepare environment variables
	env := []string{
		fmt.Sprintf("AGENT_CONFIG=%s", toJSON(spec.Configuration)),
	}

	// Add API token based on agent type
	// Both ClaudeToken (centralized) and APIKey (legacy flag) map to the agent's expected env var
	token := spec.ClaudeToken
	if token == "" {
		token = spec.APIKey
	}

	if token != "" {
		envVar := "ANTHROPIC_API_KEY" // Default for claude-code
		if spec.AgentType == "gemini" {
			envVar = "GEMINI_API_KEY"
		}
		env = append(env, fmt.Sprintf("%s=%s", envVar, token))
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

	// Add Proxy config
	if spec.ProxyConfig != nil {
		proxyURL := fmt.Sprintf("http://%s:%d", spec.ProxyConfig.Host, spec.ProxyConfig.Port)
		env = append(env,
			fmt.Sprintf("HTTP_PROXY=%s", proxyURL),
			fmt.Sprintf("HTTPS_PROXY=%s", proxyURL),
			"NODE_EXTRA_CA_CERTS=/usr/local/share/ca-certificates/ubik-ca.pem", // Node.js
			"REQUESTS_CA_BUNDLE=/usr/local/share/ca-certificates/ubik-ca.pem",  // Python
			"SSL_CERT_FILE=/usr/local/share/ca-certificates/ubik-ca.pem",       // Generic
		)
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
	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: workspacePath,
			Target: "/workspace",
		},
	}

	// Mount CA certificate if proxy is enabled
	if spec.ProxyConfig != nil && spec.ProxyConfig.CertPath != "" {
		mounts = append(mounts, mount.Mount{
			Type:     mount.TypeBind,
			Source:   spec.ProxyConfig.CertPath,
			Target:   "/usr/local/share/ca-certificates/ubik-ca.pem",
			ReadOnly: true,
		})
	}

	// Mount Docker socket for Docker-in-Docker (DinD) support
	// This allows agents to run Docker commands on the host
	// Security note: This gives containers significant privileges
	mounts = append(mounts, mount.Mount{
		Type:   mount.TypeBind,
		Source: "/var/run/docker.sock",
		Target: "/var/run/docker.sock",
	})

	// Mount host git config (read-only) for git identity
	homeDir, err := os.UserHomeDir()
	if err == nil {
		gitConfigPath := filepath.Join(homeDir, ".gitconfig")
		if _, err := os.Stat(gitConfigPath); err == nil {
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   gitConfigPath,
				Target:   "/root/.gitconfig",
				ReadOnly: true,
			})
		}

		// Mount SSH keys (read-only) for git/gh authentication
		sshDir := filepath.Join(homeDir, ".ssh")
		if _, err := os.Stat(sshDir); err == nil {
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   sshDir,
				Target:   "/root/.ssh",
				ReadOnly: true,
			})
		}

		// Mount gh CLI config (read-only) for GitHub operations
		ghConfigDir := filepath.Join(homeDir, ".config", "gh")
		if _, err := os.Stat(ghConfigDir); err == nil {
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   ghConfigDir,
				Target:   "/root/.config/gh",
				ReadOnly: true,
			})
		}
	}

	// Note: We used to mount host binaries (detectHostTools),
	// but this caused issues with binary compatibility (macOS binaries → Linux containers).
	// Instead, we now:
	// 1. Mount Docker socket for docker/kubectl commands
	// 2. Mount git/gh configs for authentication
	// 3. Expect agent images to have their own tools installed (go, node, python, etc.)

	hostConfig := &container.HostConfig{
		Mounts: mounts,
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
	containerID, err := cm.dockerClient.CreateContainer(ctx, config, hostConfig, networkConfig, containerName)
	if err != nil {
		return "", err
	}

	// Start container
	if err := cm.dockerClient.StartContainer(ctx, containerID); err != nil {
		return "", err
	}

	fmt.Printf("  ✓ %s started (container: %s)\n", spec.AgentName, containerID[:12])
	return containerID, nil
}

// StopContainers stops all ubik-managed containers
func (cm *ContainerManager) StopContainers(ctx context.Context) error {
	containers, err := cm.dockerClient.ListContainers(ctx, false, map[string]string{
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
		if err := cm.dockerClient.StopContainer(ctx, c.ID, &timeout); err != nil {
			fmt.Printf("  ⚠ Failed to stop %s: %v\n", c.Name, err)
		} else {
			fmt.Printf("  ✓ %s stopped\n", c.Name)
		}
	}

	return nil
}

// CleanupContainers removes all stopped ubik-managed containers
func (cm *ContainerManager) CleanupContainers(ctx context.Context) error {
	containers, err := cm.dockerClient.ListContainers(ctx, true, map[string]string{
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
			if err := cm.dockerClient.RemoveContainer(ctx, c.ID, false); err != nil {
				fmt.Printf("  ⚠ Failed to remove %s: %v\n", c.Name, err)
			} else {
				fmt.Printf("  ✓ %s removed\n", c.Name)
			}
		}
	}

	return nil
}

// GetContainerStatus returns status of all ubik-managed containers
func (cm *ContainerManager) GetContainerStatus(ctx context.Context) ([]ContainerInfo, error) {
	return cm.dockerClient.ListContainers(ctx, true, map[string]string{
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

// detectHostTools detects and returns mounts for host development tools
// This allows containers to use the host's installed runtimes (Go, Node, Python, etc.)
func detectHostTools() []mount.Mount {
	var mounts []mount.Mount

	// Tool detection map: executable name -> typical install paths
	tools := map[string][]string{
		"go":      {"/usr/local/go", "/opt/homebrew/opt/go/libexec"},
		"node":    {}, // Skip Node - already in base image
		"python3": {}, // Skip Python - already in base image
	}

	for tool, paths := range tools {
		if len(paths) == 0 {
			continue // Skip tools already in container
		}

		// Try to find the tool on the host
		toolPath := findTool(tool, paths)
		if toolPath != "" {
			mounts = append(mounts, mount.Mount{
				Type:     mount.TypeBind,
				Source:   toolPath,
				Target:   toolPath, // Mount at same path in container
				ReadOnly: true,
			})
			fmt.Printf("  ✓ Mounting host %s from %s\n", tool, toolPath)
		}
	}

	return mounts
}

// findTool attempts to locate a tool on the host system
func findTool(name string, candidatePaths []string) string {
	// First check candidate paths
	for _, path := range candidatePaths {
		if fileExists(path) {
			return path
		}
	}

	// No need to use exec.Command - just check known paths
	return ""
}

// fileExists checks if a file or directory exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && info != nil
}
