package container

import (
	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
)

// TestContainer provides utilities for creating containers in tests.
// It allows easy injection of mock dependencies.

// NewTestContainer creates a Container pre-configured for testing.
// Pass mocks through the provided options.
func NewTestContainer(opts ...Option) *Container {
	return New(opts...)
}

// WithMockConfigManager sets a ConfigManager for testing.
// This allows injecting a mock or test implementation.
func WithMockConfigManager(cm *cli.ConfigManager) Option {
	return WithConfigManager(cm)
}

// WithMockAPIClient sets an APIClient for testing.
// This allows injecting a mock or test implementation.
func WithMockAPIClient(ac *cli.APIClient) Option {
	return WithAPIClient(ac)
}

// WithMockAuthService sets an AuthService for testing.
// This allows injecting a mock or test implementation.
func WithMockAuthService(as *cli.AuthService) Option {
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

// WithConfigManager injects a ConfigManager for testing.
func (b *TestContainerBuilder) WithConfigManager(cm *cli.ConfigManager) *TestContainerBuilder {
	b.opts = append(b.opts, WithConfigManager(cm))
	return b
}

// WithAPIClient injects an APIClient for testing.
func (b *TestContainerBuilder) WithAPIClient(ac *cli.APIClient) *TestContainerBuilder {
	b.opts = append(b.opts, WithAPIClient(ac))
	return b
}

// WithAuthService injects an AuthService for testing.
func (b *TestContainerBuilder) WithAuthService(as *cli.AuthService) *TestContainerBuilder {
	b.opts = append(b.opts, WithAuthService(as))
	return b
}

// Build creates the configured Container.
func (b *TestContainerBuilder) Build() *Container {
	return New(b.opts...)
}
