package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WriteAgentFiles writes agent .md files to the specified directory
func WriteAgentFiles(agentsDir string, agents []AgentConfigSync) error {
	// Create agents directory
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return fmt.Errorf("failed to create agents directory: %w", err)
	}

	// Write each agent file
	for _, agent := range agents {
		if !agent.IsEnabled {
			continue
		}

		agentPath := filepath.Join(agentsDir, agent.Filename)
		if err := os.WriteFile(agentPath, []byte(agent.Content), 0644); err != nil {
			return fmt.Errorf("failed to write agent file %s: %w", agent.Filename, err)
		}
	}

	return nil
}

// WriteSkillFiles writes skill files to the specified directory
func WriteSkillFiles(skillsDir string, skills []SkillConfigSync) error {
	// Create skills directory
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return fmt.Errorf("failed to create skills directory: %w", err)
	}

	// Write each skill
	for _, skill := range skills {
		if !skill.IsEnabled {
			continue
		}

		skillDir := filepath.Join(skillsDir, skill.Name)
		if err := os.MkdirAll(skillDir, 0755); err != nil {
			return fmt.Errorf("failed to create skill directory %s: %w", skill.Name, err)
		}

		// Write skill files
		for _, file := range skill.Files {
			path := file["path"]
			content := file["content"]

			filePath := filepath.Join(skillDir, path)

			// Create parent directory if needed
			fileDir := filepath.Dir(filePath)
			if err := os.MkdirAll(fileDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", fileDir, err)
			}

			// Write file
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", filePath, err)
			}
		}
	}

	return nil
}

// MergeMCPConfig merges MCP servers into ~/.claude.json
func MergeMCPConfig(configPath string, mcpServers []MCPServerConfigSync) error {
	// Read existing config (if exists)
	var config map[string]interface{}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		// File doesn't exist, create new config
		config = map[string]interface{}{
			"mcpServers": make(map[string]interface{}),
		}
	} else {
		// Parse existing config
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}

		// Ensure mcpServers exists
		if _, ok := config["mcpServers"]; !ok {
			config["mcpServers"] = make(map[string]interface{})
		}
	}

	// Get mcpServers map
	servers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		servers = make(map[string]interface{})
		config["mcpServers"] = servers
	}

	// Add/update MCP servers
	for _, mcp := range mcpServers {
		if !mcp.IsEnabled {
			continue
		}

		servers[mcp.Name] = map[string]interface{}{
			"image":  mcp.DockerImage,
			"config": mcp.Config,
		}
	}

	// Write config back
	data, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
