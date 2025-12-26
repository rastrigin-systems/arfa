// Package container provides dependency injection for the CLI application.
// It manages the lifecycle of services and ensures proper dependency resolution.
package container

import (
	gosync "sync"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/auth"
	"github.com/rastrigin-systems/arfa/services/cli/internal/config"
)

// Container manages dependencies for the CLI application.
// Services are lazily initialized on first access.
type Container struct {
	mu gosync.RWMutex

	// Configuration
	configPath  string
	platformURL string

	// Lazily initialized services
	configManager *config.Manager
	apiClient     *api.Client
	authService   *auth.Service
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

// WithConfigManager sets a pre-configured config.Manager.
// Use this for testing with mock implementations.
func WithConfigManager(cm *config.Manager) Option {
	return func(c *Container) {
		c.configManager = cm
	}
}

// WithAPIClient sets a pre-configured api.Client.
// Use this for testing with mock implementations.
func WithAPIClient(ac *api.Client) Option {
	return func(c *Container) {
		c.apiClient = ac
	}
}

// WithAuthService sets a pre-configured auth.Service.
// Use this for testing with mock implementations.
func WithAuthService(as *auth.Service) Option {
	return func(c *Container) {
		c.authService = as
	}
}

// ConfigManager returns the config.Manager, creating it if necessary.
func (c *Container) ConfigManager() (*config.Manager, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.configManager != nil {
		return c.configManager, nil
	}

	var cm *config.Manager
	var err error

	if c.configPath != "" {
		// Use custom config path
		cm = config.NewManagerWithPath(c.configPath)
	} else {
		// Use default config manager
		cm, err = config.NewManager()
		if err != nil {
			return nil, err
		}
	}

	c.configManager = cm
	return c.configManager, nil
}

// APIClient returns the api.Client, creating it if necessary.
// It loads the platform URL and token from config if available.
func (c *Container) APIClient() (*api.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.apiClient != nil {
		return c.apiClient, nil
	}

	// Get or create config manager
	if c.configManager == nil {
		if c.configPath != "" {
			c.configManager = config.NewManagerWithPath(c.configPath)
		} else {
			var err error
			c.configManager, err = config.NewManager()
			if err != nil {
				// Fall back to default URL if we can't create config manager
				platformURL := c.platformURL
				if platformURL == "" {
					platformURL = config.DefaultPlatformURL()
				}
				c.apiClient = api.NewClient(platformURL)
				return c.apiClient, nil
			}
		}
	}

	// Load config to get platform URL and token
	cfg, err := c.configManager.Load()
	if err != nil {
		// Config doesn't exist yet, use defaults
		platformURL := c.platformURL
		if platformURL == "" {
			platformURL = config.DefaultPlatformURL()
		}
		c.apiClient = api.NewClient(platformURL)
		return c.apiClient, nil
	}

	// Use config values
	platformURL := cfg.PlatformURL
	if platformURL == "" {
		platformURL = c.platformURL
	}
	if platformURL == "" {
		platformURL = config.DefaultPlatformURL()
	}

	c.apiClient = api.NewClient(platformURL)
	if cfg.Token != "" {
		c.apiClient.SetToken(cfg.Token)
	}

	return c.apiClient, nil
}

// AuthService returns the auth.Service, creating it if necessary.
func (c *Container) AuthService() (*auth.Service, error) {
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
	ac, err := c.APIClient()
	if err != nil {
		c.mu.Lock()
		return nil, err
	}
	c.mu.Lock()

	c.authService = auth.NewService(cm, ac)
	return c.authService, nil
}

// Close cleans up any resources held by the container.
func (c *Container) Close() error {
	// No resources to clean up currently
	return nil
}
