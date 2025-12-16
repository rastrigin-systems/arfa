// Package container provides dependency injection for the CLI application.
// It manages the lifecycle of services and ensures proper dependency resolution.
package container

import (
	"sync"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
)

// Container manages dependencies for the CLI application.
// Services are lazily initialized on first access.
type Container struct {
	mu sync.RWMutex

	// Configuration
	configPath  string
	platformURL string

	// Lazily initialized services
	configManager  *cli.ConfigManager
	platformClient *cli.PlatformClient
	authService    *cli.AuthService
	syncService    *cli.SyncService
	agentService   *cli.AgentService
	skillsService  *cli.SkillsService
	dockerClient   *cli.DockerClient
}

// Option is a functional option for configuring the Container.
type Option func(*Container)

// New creates a new Container with the given options.
func New(opts ...Option) *Container {
	c := &Container{}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithConfigPath sets the config file path.
func WithConfigPath(path string) Option {
	return func(c *Container) {
		c.configPath = path
	}
}

// WithPlatformURL sets the default platform URL.
func WithPlatformURL(url string) Option {
	return func(c *Container) {
		c.platformURL = url
	}
}

// WithConfigManager sets a pre-configured ConfigManager.
// Use this for testing with mock implementations.
func WithConfigManager(cm *cli.ConfigManager) Option {
	return func(c *Container) {
		c.configManager = cm
	}
}

// WithPlatformClient sets a pre-configured PlatformClient.
// Use this for testing with mock implementations.
func WithPlatformClient(pc *cli.PlatformClient) Option {
	return func(c *Container) {
		c.platformClient = pc
	}
}

// WithAuthService sets a pre-configured AuthService.
// Use this for testing with mock implementations.
func WithAuthService(as *cli.AuthService) Option {
	return func(c *Container) {
		c.authService = as
	}
}

// ConfigManager returns the ConfigManager, creating it if necessary.
func (c *Container) ConfigManager() (*cli.ConfigManager, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.configManager != nil {
		return c.configManager, nil
	}

	var cm *cli.ConfigManager
	var err error

	if c.configPath != "" {
		// Use custom config path
		cm = cli.NewConfigManagerWithPath(c.configPath)
	} else {
		// Use default config manager
		cm, err = cli.NewConfigManager()
		if err != nil {
			return nil, err
		}
	}

	c.configManager = cm
	return c.configManager, nil
}

// PlatformClient returns the PlatformClient, creating it if necessary.
func (c *Container) PlatformClient() (*cli.PlatformClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.platformClient != nil {
		return c.platformClient, nil
	}

	platformURL := c.platformURL
	if platformURL == "" {
		platformURL = cli.DefaultPlatformURL
	}

	c.platformClient = cli.NewPlatformClient(platformURL)
	return c.platformClient, nil
}

// AuthService returns the AuthService, creating it if necessary.
func (c *Container) AuthService() (*cli.AuthService, error) {
	c.mu.RLock()
	if c.authService != nil {
		c.mu.RUnlock()
		return c.authService, nil
	}
	c.mu.RUnlock()

	// Need to create - acquire write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.authService != nil {
		return c.authService, nil
	}

	// Get dependencies (unlock temporarily to avoid deadlock)
	c.mu.Unlock()
	cm, err := c.ConfigManager()
	if err != nil {
		c.mu.Lock()
		return nil, err
	}
	pc, err := c.PlatformClient()
	if err != nil {
		c.mu.Lock()
		return nil, err
	}
	c.mu.Lock()

	c.authService = cli.NewAuthService(cm, pc)
	return c.authService, nil
}

// SyncService returns the SyncService, creating it if necessary.
func (c *Container) SyncService() (*cli.SyncService, error) {
	c.mu.RLock()
	if c.syncService != nil {
		c.mu.RUnlock()
		return c.syncService, nil
	}
	c.mu.RUnlock()

	// Need to create - acquire write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.syncService != nil {
		return c.syncService, nil
	}

	// Get dependencies (unlock temporarily to avoid deadlock)
	c.mu.Unlock()
	cm, err := c.ConfigManager()
	if err != nil {
		c.mu.Lock()
		return nil, err
	}
	pc, err := c.PlatformClient()
	if err != nil {
		c.mu.Lock()
		return nil, err
	}
	as, err := c.AuthService()
	if err != nil {
		c.mu.Lock()
		return nil, err
	}
	c.mu.Lock()

	c.syncService = cli.NewSyncService(cm, pc, as)
	return c.syncService, nil
}

// AgentService returns the AgentService, creating it if necessary.
func (c *Container) AgentService() (*cli.AgentService, error) {
	c.mu.RLock()
	if c.agentService != nil {
		c.mu.RUnlock()
		return c.agentService, nil
	}
	c.mu.RUnlock()

	// Need to create - acquire write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.agentService != nil {
		return c.agentService, nil
	}

	// Get dependencies (unlock temporarily to avoid deadlock)
	c.mu.Unlock()
	pc, err := c.PlatformClient()
	if err != nil {
		c.mu.Lock()
		return nil, err
	}
	cm, err := c.ConfigManager()
	if err != nil {
		c.mu.Lock()
		return nil, err
	}
	c.mu.Lock()

	c.agentService = cli.NewAgentService(pc, cm)
	return c.agentService, nil
}

// SkillsService returns the SkillsService, creating it if necessary.
func (c *Container) SkillsService() (*cli.SkillsService, error) {
	c.mu.RLock()
	if c.skillsService != nil {
		c.mu.RUnlock()
		return c.skillsService, nil
	}
	c.mu.RUnlock()

	// Need to create - acquire write lock
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after acquiring write lock
	if c.skillsService != nil {
		return c.skillsService, nil
	}

	// Get dependencies (unlock temporarily to avoid deadlock)
	c.mu.Unlock()
	pc, err := c.PlatformClient()
	if err != nil {
		c.mu.Lock()
		return nil, err
	}
	cm, err := c.ConfigManager()
	if err != nil {
		c.mu.Lock()
		return nil, err
	}
	c.mu.Lock()

	c.skillsService = cli.NewSkillsService(pc, cm)
	return c.skillsService, nil
}

// DockerClient returns the DockerClient, creating it if necessary.
// Note: This requires Docker to be available on the system.
func (c *Container) DockerClient() (*cli.DockerClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.dockerClient != nil {
		return c.dockerClient, nil
	}

	dc, err := cli.NewDockerClient()
	if err != nil {
		return nil, err
	}

	c.dockerClient = dc
	return c.dockerClient, nil
}

// Close cleans up any resources held by the container.
func (c *Container) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.dockerClient != nil {
		if err := c.dockerClient.Close(); err != nil {
			return err
		}
		c.dockerClient = nil
	}

	return nil
}
