package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/config"
)

func TestNewAgentPicker(t *testing.T) {
	// Create temp config dir
	tmpDir, err := os.MkdirTemp("", "ubik-test-picker")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config manager with temp dir
	configPath := filepath.Join(tmpDir, "config.json")
	configManager := config.NewManagerWithPath(configPath)

	picker := NewAgentPicker(configManager)
	if picker == nil {
		t.Error("expected non-nil picker")
	}
	if picker.configManager != configManager {
		t.Error("expected configManager to be set")
	}
}

func TestAgentPicker_GetDefaultAgent_NoConfig(t *testing.T) {
	// Create temp config dir
	tmpDir, err := os.MkdirTemp("", "ubik-test-picker")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	configManager := config.NewManagerWithPath(configPath)

	picker := NewAgentPicker(configManager)
	defaultAgent := picker.GetDefaultAgent()

	if defaultAgent != "" {
		t.Errorf("expected empty default agent, got %q", defaultAgent)
	}
}

func TestAgentPicker_GetDefaultAgent_WithConfig(t *testing.T) {
	// Create temp config dir
	tmpDir, err := os.MkdirTemp("", "ubik-test-picker")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	configManager := config.NewManagerWithPath(configPath)

	// Save config with default agent
	cfg := &config.Config{
		DefaultAgent: "test-agent-123",
	}
	if err := configManager.Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	picker := NewAgentPicker(configManager)
	defaultAgent := picker.GetDefaultAgent()

	if defaultAgent != "test-agent-123" {
		t.Errorf("expected default agent 'test-agent-123', got %q", defaultAgent)
	}
}

func TestAgentPicker_ClearDefault(t *testing.T) {
	// Create temp config dir
	tmpDir, err := os.MkdirTemp("", "ubik-test-picker")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	configManager := config.NewManagerWithPath(configPath)

	// Save config with default agent
	cfg := &config.Config{
		DefaultAgent: "test-agent-123",
	}
	if err := configManager.Save(cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	picker := NewAgentPicker(configManager)

	// Clear default
	if err := picker.ClearDefault(); err != nil {
		t.Fatalf("failed to clear default: %v", err)
	}

	// Verify cleared
	defaultAgent := picker.GetDefaultAgent()
	if defaultAgent != "" {
		t.Errorf("expected empty default agent after clear, got %q", defaultAgent)
	}
}

func TestAgentPicker_SelectAgent_NoAgents(t *testing.T) {
	// Create temp config dir
	tmpDir, err := os.MkdirTemp("", "ubik-test-picker")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	configManager := config.NewManagerWithPath(configPath)

	picker := NewAgentPicker(configManager)

	_, err = picker.SelectAgent([]api.AgentConfig{}, false, false)
	if err == nil {
		t.Error("expected error for empty agents list")
	}
	if err.Error() != "no agents available" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgentPicker_SelectAgent_NoEnabledAgents(t *testing.T) {
	// Create temp config dir
	tmpDir, err := os.MkdirTemp("", "ubik-test-picker")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	configManager := config.NewManagerWithPath(configPath)

	picker := NewAgentPicker(configManager)

	agents := []api.AgentConfig{
		{
			AgentID:   "agent-1",
			AgentName: "Test Agent",
			IsEnabled: false, // disabled
		},
	}

	_, err = picker.SelectAgent(agents, false, false)
	if err == nil {
		t.Error("expected error for no enabled agents")
	}
	if err.Error() != "no enabled agents available" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgentPicker_SelectAgent_SingleAgent(t *testing.T) {
	// Create temp config dir
	tmpDir, err := os.MkdirTemp("", "ubik-test-picker")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, "config.json")
	configManager := config.NewManagerWithPath(configPath)

	// Save empty config
	if err := configManager.Save(&config.Config{}); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	picker := NewAgentPicker(configManager)

	agents := []api.AgentConfig{
		{
			AgentID:     "agent-1",
			AgentName:   "Test Agent",
			AgentType:   "claude-code",
			Provider:    "Anthropic",
			DockerImage: "ubik/claude-code:latest",
			IsEnabled:   true,
		},
	}

	// Single enabled agent should be returned directly (no picker shown) when not forcing interactive
	selected, err := picker.SelectAgent(agents, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if selected == nil {
		t.Fatal("expected non-nil selected agent")
	}

	if selected.AgentID != "agent-1" {
		t.Errorf("expected agent-1, got %s", selected.AgentID)
	}

	// Check that default was saved
	savedDefault := picker.GetDefaultAgent()
	if savedDefault != "agent-1" {
		t.Errorf("expected default to be saved as 'agent-1', got %q", savedDefault)
	}
}

func TestAgentPickerItem(t *testing.T) {
	item := AgentPickerItem{
		Name:        "Claude Code",
		Type:        "claude-code",
		Provider:    "Anthropic",
		DockerImage: "ubik/claude-code:v1.0.0",
		ID:          "test-id",
		IsDefault:   true,
	}

	if item.Name != "Claude Code" {
		t.Errorf("unexpected name: %s", item.Name)
	}
	if item.Type != "claude-code" {
		t.Errorf("unexpected type: %s", item.Type)
	}
	if item.Provider != "Anthropic" {
		t.Errorf("unexpected provider: %s", item.Provider)
	}
	if item.DockerImage != "ubik/claude-code:v1.0.0" {
		t.Errorf("unexpected docker image: %s", item.DockerImage)
	}
	if item.ID != "test-id" {
		t.Errorf("unexpected id: %s", item.ID)
	}
	if !item.IsDefault {
		t.Error("expected IsDefault to be true")
	}
}
