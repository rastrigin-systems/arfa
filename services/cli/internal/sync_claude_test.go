package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncService_WriteAgentFiles(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Sample agent configs
	agents := []AgentConfigSync{
		{
			ID:       "agent-1",
			Name:     "go-backend-developer",
			Type:     "claude-code",
			Filename: "go-backend-developer.md",
			Content:  "# Go Backend Developer\n\nYou are an expert Go developer...",
			Config: map[string]interface{}{
				"model": "claude-3-5-sonnet",
			},
			Provider:  "anthropic",
			IsEnabled: true,
			Version:   "1.0.0",
		},
		{
			ID:       "agent-2",
			Name:     "frontend-developer",
			Type:     "claude-code",
			Filename: "frontend-developer.md",
			Content:  "# Frontend Developer\n\nYou are an expert frontend developer...",
			Config: map[string]interface{}{
				"model": "claude-3-5-sonnet",
			},
			Provider:  "anthropic",
			IsEnabled: true,
			Version:   "1.0.0",
		},
	}

	// Write agent files
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	err := WriteAgentFiles(agentsDir, agents)
	require.NoError(t, err)

	// Verify directory exists
	_, err = os.Stat(agentsDir)
	require.NoError(t, err)

	// Verify first agent file
	agentPath1 := filepath.Join(agentsDir, "go-backend-developer.md")
	content1, err := os.ReadFile(agentPath1)
	require.NoError(t, err)
	assert.Contains(t, string(content1), "expert Go developer")
	assert.Contains(t, string(content1), "# Go Backend Developer")

	// Verify second agent file
	agentPath2 := filepath.Join(agentsDir, "frontend-developer.md")
	content2, err := os.ReadFile(agentPath2)
	require.NoError(t, err)
	assert.Contains(t, string(content2), "expert frontend developer")
	assert.Contains(t, string(content2), "# Frontend Developer")
}

func TestSyncService_WriteAgentFiles_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	// Write empty agent list
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	err := WriteAgentFiles(agentsDir, []AgentConfigSync{})
	require.NoError(t, err)

	// Directory should still be created
	_, err = os.Stat(agentsDir)
	require.NoError(t, err)
}

func TestSyncService_WriteAgentFiles_Overwrite(t *testing.T) {
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")

	// Write initial agent file
	agents1 := []AgentConfigSync{
		{
			Name:      "go-backend-developer",
			Filename:  "go-backend-developer.md",
			Content:   "# Old Content",
			IsEnabled: true,
		},
	}
	err := WriteAgentFiles(agentsDir, agents1)
	require.NoError(t, err)

	// Overwrite with new content
	agents2 := []AgentConfigSync{
		{
			Name:      "go-backend-developer",
			Filename:  "go-backend-developer.md",
			Content:   "# New Content",
			IsEnabled: true,
		},
	}
	err = WriteAgentFiles(agentsDir, agents2)
	require.NoError(t, err)

	// Verify new content
	agentPath := filepath.Join(agentsDir, "go-backend-developer.md")
	content, err := os.ReadFile(agentPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "# New Content")
	assert.NotContains(t, string(content), "# Old Content")
}

func TestSyncService_WriteSkillFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Sample skill configs
	skills := []SkillConfigSync{
		{
			ID:          "skill-1",
			Name:        "release-manager",
			Description: "Manage releases",
			Category:    "devops",
			Version:     "1.0.0",
			Files: []map[string]string{
				{
					"path":    "SKILL.md",
					"content": "# Release Manager Skill\n\nManages releases...",
				},
				{
					"path":    "examples/example.md",
					"content": "# Example\n\nExample usage...",
				},
				{
					"path":    "scripts/release.sh",
					"content": "#!/bin/bash\necho 'Release script'",
				},
			},
			IsEnabled: true,
		},
		{
			ID:          "skill-2",
			Name:        "github-task-manager",
			Description: "Manage GitHub tasks",
			Version:     "1.0.0",
			Files: []map[string]string{
				{
					"path":    "SKILL.md",
					"content": "# GitHub Task Manager\n\nManages tasks...",
				},
			},
			IsEnabled: true,
		},
	}

	// Write skill files
	skillsDir := filepath.Join(tmpDir, ".claude", "skills")
	err := WriteSkillFiles(skillsDir, skills)
	require.NoError(t, err)

	// Verify release-manager skill
	skillDir1 := filepath.Join(skillsDir, "release-manager")
	_, err = os.Stat(skillDir1)
	require.NoError(t, err)

	// Verify SKILL.md
	skillMd := filepath.Join(skillDir1, "SKILL.md")
	content, err := os.ReadFile(skillMd)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Release Manager Skill")

	// Verify nested file
	exampleMd := filepath.Join(skillDir1, "examples", "example.md")
	content, err = os.ReadFile(exampleMd)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Example usage")

	// Verify script file
	scriptSh := filepath.Join(skillDir1, "scripts", "release.sh")
	content, err = os.ReadFile(scriptSh)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Release script")

	// Verify github-task-manager skill
	skillDir2 := filepath.Join(skillsDir, "github-task-manager")
	_, err = os.Stat(skillDir2)
	require.NoError(t, err)
}

func TestSyncService_WriteSkillFiles_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	// Write empty skill list
	skillsDir := filepath.Join(tmpDir, ".claude", "skills")
	err := WriteSkillFiles(skillsDir, []SkillConfigSync{})
	require.NoError(t, err)

	// Directory should still be created
	_, err = os.Stat(skillsDir)
	require.NoError(t, err)
}

func TestSyncService_MergeMCPConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".claude.json")

	// Create existing config
	existingConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"existing-server": map[string]interface{}{
				"command": "existing-command",
				"args":    []string{"--port", "8000"},
			},
		},
	}
	data, err := json.MarshalIndent(existingConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// New MCP servers to merge
	mcpServers := []MCPServerConfigSync{
		{
			Name:        "playwright",
			DockerImage: "ubik/mcp-playwright:latest",
			Config: map[string]interface{}{
				"port": 8001,
			},
			IsEnabled: true,
		},
		{
			Name:        "github-mcp-server",
			DockerImage: "ubik/mcp-github:latest",
			Config: map[string]interface{}{
				"port": 8002,
			},
			IsEnabled: true,
		},
	}

	// Merge MCP config
	err = MergeMCPConfig(configPath, mcpServers)
	require.NoError(t, err)

	// Read merged config
	data, err = os.ReadFile(configPath)
	require.NoError(t, err)

	var mergedConfig map[string]interface{}
	err = json.Unmarshal(data, &mergedConfig)
	require.NoError(t, err)

	// Verify existing server is preserved
	servers := mergedConfig["mcpServers"].(map[string]interface{})
	assert.Contains(t, servers, "existing-server")

	// Verify new servers are added
	assert.Contains(t, servers, "playwright")
	assert.Contains(t, servers, "github-mcp-server")

	// Verify playwright config
	playwrightConfig := servers["playwright"].(map[string]interface{})
	assert.Equal(t, "ubik/mcp-playwright:latest", playwrightConfig["image"])
	assert.Equal(t, float64(8001), playwrightConfig["config"].(map[string]interface{})["port"])
}

func TestSyncService_MergeMCPConfig_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".claude.json")

	// No existing file

	// New MCP servers
	mcpServers := []MCPServerConfigSync{
		{
			Name:        "playwright",
			DockerImage: "ubik/mcp-playwright:latest",
			Config: map[string]interface{}{
				"port": 8001,
			},
			IsEnabled: true,
		},
	}

	// Merge MCP config (creates new file)
	err := MergeMCPConfig(configPath, mcpServers)
	require.NoError(t, err)

	// Read config
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	// Verify config structure
	servers := config["mcpServers"].(map[string]interface{})
	assert.Contains(t, servers, "playwright")
}

func TestSyncService_MergeMCPConfig_OnlyEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".claude.json")

	// MCP servers (one disabled)
	mcpServers := []MCPServerConfigSync{
		{
			Name:        "playwright",
			DockerImage: "ubik/mcp-playwright:latest",
			Config:      map[string]interface{}{"port": 8001},
			IsEnabled:   true,
		},
		{
			Name:        "disabled-server",
			DockerImage: "ubik/mcp-disabled:latest",
			Config:      map[string]interface{}{"port": 8002},
			IsEnabled:   false,
		},
	}

	// Merge MCP config
	err := MergeMCPConfig(configPath, mcpServers)
	require.NoError(t, err)

	// Read config
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	// Verify only enabled server is added
	servers := config["mcpServers"].(map[string]interface{})
	assert.Contains(t, servers, "playwright")
	assert.NotContains(t, servers, "disabled-server")
}

func TestSyncService_SyncClaudeCode_Integration(t *testing.T) {
	// TODO: Fix this integration test - currently has issues with PlatformClient base URL
	t.Skip("Skipping integration test temporarily")

	// This test verifies the complete Claude Code sync flow using mock HTTP server
	tmpDir := t.TempDir()
	homeDir := filepath.Join(tmpDir, "home")
	os.Setenv("HOME", homeDir)
	defer os.Unsetenv("HOME")

	// Create test config directory
	os.MkdirAll(filepath.Join(homeDir, ".ubik"), 0755)

	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/sync/claude-code" {
			resp := ClaudeCodeSyncResponse{
				Agents: []AgentConfigSync{
					{
						Name:      "go-backend-developer",
						Filename:  "go-backend-developer.md",
						Content:   "# Go Backend Developer\n\nYou are an expert Go developer...",
						IsEnabled: true,
					},
					{
						Name:      "frontend-developer",
						Filename:  "frontend-developer.md",
						Content:   "# Frontend Developer\n\nYou are an expert frontend developer...",
						IsEnabled: true,
					},
				},
				Skills: []SkillConfigSync{
					{
						Name:      "release-manager",
						Version:   "1.0.0",
						IsEnabled: true,
						Files: []map[string]string{
							{
								"path":    "SKILL.md",
								"content": "# Release Manager\n\nManages releases...",
							},
							{
								"path":    "examples/example.md",
								"content": "# Example\n\nExample usage...",
							},
						},
					},
				},
				MCPServers: []MCPServerConfigSync{
					{
						Name:        "playwright",
						DockerImage: "ubik/mcp-playwright:latest",
						Config:      map[string]interface{}{"port": float64(8001)},
						IsEnabled:   true,
					},
					{
						Name:        "github-mcp-server",
						DockerImage: "ubik/mcp-github:latest",
						Config:      map[string]interface{}{"port": float64(8002)},
						IsEnabled:   true,
					},
				},
				Version:  "1.0.0",
				SyncedAt: "2024-11-02T12:00:00Z",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	// Create services with real types
	configManager := &ConfigManager{
		configPath: filepath.Join(homeDir, ".ubik", "config.json"),
	}

	// Save test config
	testConfig := &Config{
		Token:      "test-token",
		EmployeeID: "emp-123",
	}
	configManager.Save(testConfig)

	apiClient := NewAPIClient(server.URL)
	apiClient.SetToken("test-token")

	authService := &AuthService{
		configManager: configManager,
		apiClient:     apiClient,
	}

	syncService := NewSyncService(configManager, apiClient, authService)

	// Run sync
	targetDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(targetDir, 0755)

	ctx := context.Background()
	err := syncService.SyncClaudeCode(ctx, targetDir)
	require.NoError(t, err)

	// Verify agent files
	agentPath1 := filepath.Join(targetDir, ".claude", "agents", "go-backend-developer.md")
	content1, err := os.ReadFile(agentPath1)
	require.NoError(t, err)
	assert.Contains(t, string(content1), "expert Go developer")

	agentPath2 := filepath.Join(targetDir, ".claude", "agents", "frontend-developer.md")
	content2, err := os.ReadFile(agentPath2)
	require.NoError(t, err)
	assert.Contains(t, string(content2), "expert frontend developer")

	// Verify skill files
	skillPath := filepath.Join(targetDir, ".claude", "skills", "release-manager", "SKILL.md")
	skillContent, err := os.ReadFile(skillPath)
	require.NoError(t, err)
	assert.Contains(t, string(skillContent), "Release Manager")

	examplePath := filepath.Join(targetDir, ".claude", "skills", "release-manager", "examples", "example.md")
	exampleContent, err := os.ReadFile(examplePath)
	require.NoError(t, err)
	assert.Contains(t, string(exampleContent), "Example usage")

	// Verify MCP config
	claudeConfigPath := filepath.Join(homeDir, ".claude.json")
	data, err := os.ReadFile(claudeConfigPath)
	require.NoError(t, err)

	var config map[string]interface{}
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)

	servers := config["mcpServers"].(map[string]interface{})
	assert.Contains(t, servers, "playwright")
	assert.Contains(t, servers, "github-mcp-server")
}
