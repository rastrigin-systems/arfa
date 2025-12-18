package container

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
)

func TestNew(t *testing.T) {
	c := New()
	assert.NotNil(t, c)
}

func TestWithConfigPath(t *testing.T) {
	customPath := "/tmp/test/config.json"
	c := New(WithConfigPath(customPath))
	assert.Equal(t, customPath, c.configPath)
}

func TestWithPlatformURL(t *testing.T) {
	customURL := "https://custom.example.com"
	c := New(WithPlatformURL(customURL))
	assert.Equal(t, customURL, c.platformURL)
}

func TestContainer_ConfigManager_WithCustomPath(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	c := New(WithConfigPath(configPath))
	cm, err := c.ConfigManager()

	require.NoError(t, err)
	assert.NotNil(t, cm)

	// Second call should return same instance
	cm2, err := c.ConfigManager()
	require.NoError(t, err)
	assert.Same(t, cm, cm2)
}

func TestContainer_APIClient_WithCustomURL(t *testing.T) {
	customURL := "https://custom.example.com"
	c := New(WithPlatformURL(customURL))

	ac, err := c.APIClient()
	require.NoError(t, err)
	assert.NotNil(t, ac)

	// Second call should return same instance
	ac2, err := c.APIClient()
	require.NoError(t, err)
	assert.Same(t, ac, ac2)
}

func TestContainer_APIClient_WithDefaultURL(t *testing.T) {
	c := New()

	ac, err := c.APIClient()
	require.NoError(t, err)
	assert.NotNil(t, ac)
}

func TestContainer_WithPreConfiguredServices(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Pre-configure services
	cm := config.NewManagerWithPath(configPath)
	ac := api.NewClient("https://test.example.com")

	c := New(
		WithConfigManager(cm),
		WithAPIClient(ac),
	)

	// Should return pre-configured instances
	gotCm, err := c.ConfigManager()
	require.NoError(t, err)
	assert.Same(t, cm, gotCm)

	gotAc, err := c.APIClient()
	require.NoError(t, err)
	assert.Same(t, ac, gotAc)
}

func TestContainer_AuthService(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	c := New(
		WithConfigPath(configPath),
		WithPlatformURL("https://test.example.com"),
	)

	as, err := c.AuthService()
	require.NoError(t, err)
	assert.NotNil(t, as)

	// Second call should return same instance
	as2, err := c.AuthService()
	require.NoError(t, err)
	assert.Same(t, as, as2)
}

func TestContainer_SyncService(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	c := New(
		WithConfigPath(configPath),
		WithPlatformURL("https://test.example.com"),
	)

	ss, err := c.SyncService()
	require.NoError(t, err)
	assert.NotNil(t, ss)

	// Second call should return same instance
	ss2, err := c.SyncService()
	require.NoError(t, err)
	assert.Same(t, ss, ss2)
}

func TestContainer_AgentService(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	c := New(
		WithConfigPath(configPath),
		WithPlatformURL("https://test.example.com"),
	)

	as, err := c.AgentService()
	require.NoError(t, err)
	assert.NotNil(t, as)

	// Second call should return same instance
	as2, err := c.AgentService()
	require.NoError(t, err)
	assert.Same(t, as, as2)
}

func TestContainer_SkillsService(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	c := New(
		WithConfigPath(configPath),
		WithPlatformURL("https://test.example.com"),
	)

	ss, err := c.SkillsService()
	require.NoError(t, err)
	assert.NotNil(t, ss)

	// Second call should return same instance
	ss2, err := c.SkillsService()
	require.NoError(t, err)
	assert.Same(t, ss, ss2)
}

func TestContainer_Close(t *testing.T) {
	c := New()

	// Close without any services created should not error
	err := c.Close()
	assert.NoError(t, err)
}

func TestTestContainerBuilder(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	c := NewTestContainerBuilder().
		WithConfigPath(configPath).
		WithPlatformURL("https://test.example.com").
		Build()

	assert.NotNil(t, c)
	assert.Equal(t, configPath, c.configPath)
	assert.Equal(t, "https://test.example.com", c.platformURL)
}
