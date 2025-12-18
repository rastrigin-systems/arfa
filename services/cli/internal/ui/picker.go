package ui

import (
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
)

// AgentPicker provides interactive agent selection
type AgentPicker struct {
	configManager ConfigManagerInterface
}

// NewAgentPicker creates a new agent picker
func NewAgentPicker(configManager *config.Manager) *AgentPicker {
	return &AgentPicker{
		configManager: configManager,
	}
}

// NewAgentPickerWithInterface creates a new agent picker with a config manager interface
func NewAgentPickerWithInterface(configManager ConfigManagerInterface) *AgentPicker {
	return &AgentPicker{
		configManager: configManager,
	}
}

// SelectAgent shows an interactive picker and returns the selected agent
// If saveAsDefault is true, the selection will be saved as the default agent
// If forceInteractive is true, always show the picker even with a single agent
func (p *AgentPicker) SelectAgent(agents []api.AgentConfig, saveAsDefault bool, forceInteractive bool) (*api.AgentConfig, error) {
	if len(agents) == 0 {
		return nil, fmt.Errorf("no agents available")
	}

	// Get current default
	cfg, _ := p.configManager.Load()
	currentDefault := ""
	if cfg != nil {
		currentDefault = cfg.DefaultAgent
	}

	// Build picker items
	items := make([]AgentPickerItem, 0, len(agents))
	for _, agent := range agents {
		if !agent.IsEnabled {
			continue
		}
		isDefault := agent.AgentID == currentDefault || agent.AgentName == currentDefault
		items = append(items, AgentPickerItem{
			Name:        agent.AgentName,
			Type:        agent.AgentType,
			Provider:    agent.Provider,
			DockerImage: agent.DockerImage,
			ID:          agent.AgentID,
			IsDefault:   isDefault,
		})
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("no enabled agents available")
	}

	// If only one agent and not forcing interactive, use it directly
	if len(items) == 1 && !forceInteractive {
		for i := range agents {
			if agents[i].AgentID == items[0].ID {
				if saveAsDefault {
					p.saveDefault(agents[i].AgentID)
				}
				return &agents[i], nil
			}
		}
	}

	// Custom template for the picker
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "▸ {{ .Name | cyan }} ({{ .Provider }}){{ if .IsDefault }} ← default{{ end }}",
		Inactive: "  {{ .Name | white }} ({{ .Provider }}){{ if .IsDefault }} ← default{{ end }}",
		Selected: "✓ {{ .Name | green }}",
		Details: `
--------- Agent Details ----------
{{ "Name:" | faint }}	{{ .Name }}
{{ "Type:" | faint }}	{{ .Type }}
{{ "Provider:" | faint }}	{{ .Provider }}
{{ "Image:" | faint }}	{{ .DockerImage }}`,
	}

	// Search function
	searcher := func(input string, index int) bool {
		item := items[index]
		name := strings.ToLower(item.Name)
		input = strings.ToLower(input)
		return strings.Contains(name, input)
	}

	prompt := promptui.Select{
		Label:     "Select an agent to run",
		Items:     items,
		Templates: templates,
		Size:      10,
		Searcher:  searcher,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			return nil, fmt.Errorf("selection cancelled")
		}
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	selectedItem := items[idx]

	// Find the full agent config
	var selectedAgent *api.AgentConfig
	for i := range agents {
		if agents[i].AgentID == selectedItem.ID {
			selectedAgent = &agents[i]
			break
		}
	}

	if selectedAgent == nil {
		return nil, fmt.Errorf("selected agent not found")
	}

	// Save as default if requested
	if saveAsDefault {
		if err := p.saveDefault(selectedAgent.AgentID); err != nil {
			fmt.Printf("⚠ Warning: failed to save default: %v\n", err)
		} else {
			fmt.Printf("✓ Saved as default agent\n")
		}
	}

	return selectedAgent, nil
}

// saveDefault saves the agent ID as the default
func (p *AgentPicker) saveDefault(agentID string) error {
	cfg, err := p.configManager.Load()
	if err != nil {
		return err
	}

	cfg.DefaultAgent = agentID
	return p.configManager.Save(cfg)
}

// ConfirmSaveDefault asks the user if they want to save the selection as default
func (p *AgentPicker) ConfirmSaveDefault() bool {
	prompt := promptui.Prompt{
		Label:     "Save as default agent",
		IsConfirm: true,
		Default:   "y",
	}

	result, err := prompt.Run()
	if err != nil {
		return false
	}

	return strings.ToLower(result) == "y" || result == ""
}

// GetDefaultAgent returns the currently configured default agent
func (p *AgentPicker) GetDefaultAgent() string {
	cfg, err := p.configManager.Load()
	if err != nil || cfg == nil {
		return ""
	}
	return cfg.DefaultAgent
}

// ClearDefault removes the default agent setting
func (p *AgentPicker) ClearDefault() error {
	cfg, err := p.configManager.Load()
	if err != nil {
		return err
	}

	cfg.DefaultAgent = ""
	return p.configManager.Save(cfg)
}
