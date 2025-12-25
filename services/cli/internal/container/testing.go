package container

import (
	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/auth"
	"github.com/rastrigin-systems/arfa/services/cli/internal/config"
)

// TestContainer provides utilities for creating containers in tests.
// It allows easy injection of mock dependencies.

// NewTestContainer creates a Container pre-configured for testing.
// Pass mocks through the provided options.
func NewTestContainer(opts ...Option) *Container {
	return New(opts...)
}

// WithMockConfigManager sets a config.Manager for testing.
// This allows injecting a mock or test implementation.
func WithMockConfigManager(cm *config.Manager) Option {
	return WithConfigManager(cm)
}

// WithMockAPIClient sets an api.Client for testing.
// This allows injecting a mock or test implementation.
func WithMockAPIClient(ac *api.Client) Option {
	return WithAPIClient(ac)
}

// WithMockAuthService sets an auth.Service for testing.
// This allows injecting a mock or test implementation.
func WithMockAuthService(as *auth.Service) Option {
	return WithAuthService(as)
}

// TestContainerBuilder provides a fluent API for building test containers.
type TestContainerBuilder struct {
	opts []Option
}

// NewTestContainerBuilder creates a new test container builder.
func NewTestContainerBuilder() *TestContainerBuilder {
	return &TestContainerBuilder{}
}

// WithConfigPath sets a custom config path for testing.
func (b *TestContainerBuilder) WithConfigPath(path string) *TestContainerBuilder {
	b.opts = append(b.opts, WithConfigPath(path))
	return b
}

// WithPlatformURL sets a custom platform URL for testing.
func (b *TestContainerBuilder) WithPlatformURL(url string) *TestContainerBuilder {
	b.opts = append(b.opts, WithPlatformURL(url))
	return b
}

// WithConfigManager injects a config.Manager for testing.
func (b *TestContainerBuilder) WithConfigManager(cm *config.Manager) *TestContainerBuilder {
	b.opts = append(b.opts, WithConfigManager(cm))
	return b
}

// WithAPIClient injects an api.Client for testing.
func (b *TestContainerBuilder) WithAPIClient(ac *api.Client) *TestContainerBuilder {
	b.opts = append(b.opts, WithAPIClient(ac))
	return b
}

// WithAuthService injects an auth.Service for testing.
func (b *TestContainerBuilder) WithAuthService(as *auth.Service) *TestContainerBuilder {
	b.opts = append(b.opts, WithAuthService(as))
	return b
}

// Build creates the configured Container.
func (b *TestContainerBuilder) Build() *Container {
	return New(b.opts...)
}
