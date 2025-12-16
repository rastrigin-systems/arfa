package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// AgentConfigCascade represents the full configuration cascade for an agent
type AgentConfigCascade struct {
	Agent          AgentInfo              `json:"agent"`
	OrgConfig      *ConfigLevel           `json:"org_config"`
	TeamConfig     *ConfigLevel           `json:"team_config"`
	EmployeeConfig *ConfigLevel           `json:"employee_config"`
	ResolvedConfig map[string]interface{} `json:"resolved_config"`
}

// AgentInfo contains basic agent information
type AgentInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Provider string `json:"provider"`
	Enabled  bool   `json:"enabled"`
}

// ConfigLevel represents configuration at a specific level (org/team/employee)
type ConfigLevel struct {
	Config    map[string]interface{} `json:"config"`
	IsEnabled bool                   `json:"is_enabled"`
	Source    string                 `json:"source"` // org name, team name, "You"
}

// NewShowCommand creates the 'agents show' command with dependencies from the container.
func NewShowCommand(c *container.Container) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show <agent-name|agent-id>",
		Short: "Show detailed agent configuration with cascade breakdown",
		Long: `Show detailed configuration for a specific agent, including:
- Organization-level base configuration
- Team-level overrides (if applicable)
- Personal (employee-level) overrides (if applicable)
- Final resolved configuration

This helps you understand where each configuration value comes from.

You can specify the agent by name (e.g., "Claude Code"), type (e.g., "ide_assistant"),
or ID (e.g., "a1111111-1111-1111-1111-111111111111").`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentName := args[0]

			configManager, err := c.ConfigManager()
			if err != nil {
				return fmt.Errorf("failed to get config manager: %w", err)
			}

			platformClient, err := c.APIClient()
			if err != nil {
				return fmt.Errorf("failed to get platform client: %w", err)
			}

			// Fetch cascade data
			cascade, err := fetchAgentCascade(agentName, configManager, platformClient)
			if err != nil {
				return fmt.Errorf("failed to fetch agent configuration: %w", err)
			}

			// Output based on format
			if jsonOutput {
				return outputJSON(cmd, cascade)
			}
			return outputFormatted(cmd, cascade)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

// fetchAgentCascade fetches configuration from all levels for the specified agent
func fetchAgentCascade(agentName string, configMgr cli.ConfigManagerInterface, client cli.APIClientInterface) (*AgentConfigCascade, error) {
	// Load config
	cfg, err := configMgr.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("not logged in. Run 'ubik login' first")
	}

	// Set token on client
	client.SetToken(cfg.Token)

	ctx := context.Background()

	// 1. Get current employee info
	employee, err := client.GetCurrentEmployee(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current employee: %w", err)
	}

	// 2. Get all agent configs to find matching agent
	// Get resolved config first to find the agent
	resolvedConfigs, err := client.GetResolvedAgentConfigs(ctx, employee.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get resolved configs: %w", err)
	}

	// Find matching agent by name, type, or ID
	var targetAgent *api.AgentConfig
	for i := range resolvedConfigs {
		if resolvedConfigs[i].AgentName == agentName ||
			resolvedConfigs[i].AgentType == agentName ||
			resolvedConfigs[i].AgentID == agentName {
			targetAgent = &resolvedConfigs[i]
			break
		}
	}

	if targetAgent == nil {
		return nil, fmt.Errorf("agent %q not found (tried matching by name, type, and ID)", agentName)
	}

	// 3. Get org-level config
	orgConfigs, err := client.GetOrgAgentConfigs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get org configs: %w", err)
	}

	var orgConfig *ConfigLevel
	for _, cfg := range orgConfigs {
		if cfg.AgentID == targetAgent.AgentID {
			orgConfig = &ConfigLevel{
				Config:    cfg.Config,
				IsEnabled: cfg.IsEnabled,
				Source:    "Organization", // TODO: Get actual org name
			}
			break
		}
	}

	// 4. Get team-level config (if employee has team)
	var teamConfig *ConfigLevel
	if employee.TeamID != nil && *employee.TeamID != "" {
		teamConfigs, err := client.GetTeamAgentConfigs(ctx, *employee.TeamID)
		if err == nil { // Ignore errors for team configs
			for _, cfg := range teamConfigs {
				if cfg.AgentID == targetAgent.AgentID {
					teamConfig = &ConfigLevel{
						Config:    cfg.ConfigOverride,
						IsEnabled: cfg.IsEnabled,
						Source:    "Team", // TODO: Get actual team name
					}
					break
				}
			}
		}
	}

	// 5. Get employee-level config
	var employeeConfig *ConfigLevel
	employeeConfigs, err := client.GetEmployeeAgentConfigs(ctx, employee.ID)
	if err == nil { // Ignore errors for employee configs
		for _, cfg := range employeeConfigs {
			if cfg.AgentID == targetAgent.AgentID {
				employeeConfig = &ConfigLevel{
					Config:    cfg.ConfigOverride,
					IsEnabled: cfg.IsEnabled,
					Source:    "You",
				}
				break
			}
		}
	}

	// Build cascade
	cascade := &AgentConfigCascade{
		Agent: AgentInfo{
			ID:       targetAgent.AgentID,
			Name:     targetAgent.AgentName,
			Type:     targetAgent.AgentType,
			Provider: targetAgent.Provider,
			Enabled:  targetAgent.IsEnabled,
		},
		OrgConfig:      orgConfig,
		TeamConfig:     teamConfig,
		EmployeeConfig: employeeConfig,
		ResolvedConfig: targetAgent.Configuration,
	}

	return cascade, nil
}

// outputJSON outputs the cascade in JSON format
func outputJSON(cmd *cobra.Command, cascade *AgentConfigCascade) error {
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetIndent("", "  ")
	return encoder.Encode(cascade)
}

// outputFormatted outputs the cascade in a human-readable formatted view
func outputFormatted(cmd *cobra.Command, cascade *AgentConfigCascade) error {
	out := cmd.OutOrStdout()

	// Colors
	blue := color.New(color.FgBlue, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)
	gray := color.New(color.FgHiBlack)

	// Header
	fmt.Fprintf(out, "\n")
	blue.Fprintf(out, "Agent: %s\n", cascade.Agent.Name)
	fmt.Fprintf(out, "Type: %s\n", cascade.Agent.Type)
	fmt.Fprintf(out, "Provider: %s\n", cascade.Agent.Provider)
	if cascade.Agent.Enabled {
		green.Fprintf(out, "Status: ✓ Enabled\n")
	} else {
		color.New(color.FgRed).Fprintf(out, "Status: ✗ Disabled\n")
	}
	fmt.Fprintf(out, "\n")

	// Configuration Cascade
	blue.Fprintf(out, "Configuration Cascade:\n")
	fmt.Fprintf(out, "\n")

	// Level 1: Organization
	if cascade.OrgConfig != nil {
		fmt.Fprintf(out, "┌─────────────────────────────────────────────────────────────┐\n")
		cyan.Fprintf(out, "│ 📋 Level 1: Organization (%s)", cascade.OrgConfig.Source)
		padding := 60 - len(cascade.OrgConfig.Source) - 23
		fmt.Fprintf(out, "%s│\n", strings.Repeat(" ", padding))
		fmt.Fprintf(out, "├─────────────────────────────────────────────────────────────┤\n")

		printConfigMap(out, cascade.OrgConfig.Config, "  ", nil)

		fmt.Fprintf(out, "└─────────────────────────────────────────────────────────────┘\n")
		fmt.Fprintf(out, "\n")
	}

	// Level 2: Team
	if cascade.TeamConfig != nil {
		fmt.Fprintf(out, "┌─────────────────────────────────────────────────────────────┐\n")
		yellow.Fprintf(out, "│ 📋 Level 2: Team Override (%s)", cascade.TeamConfig.Source)
		padding := 60 - len(cascade.TeamConfig.Source) - 28
		fmt.Fprintf(out, "%s│\n", strings.Repeat(" ", padding))
		fmt.Fprintf(out, "├─────────────────────────────────────────────────────────────┤\n")

		printConfigMap(out, cascade.TeamConfig.Config, "  ", cascade.OrgConfig.Config)

		fmt.Fprintf(out, "└─────────────────────────────────────────────────────────────┘\n")
		fmt.Fprintf(out, "\n")
	}

	// Level 3: Employee
	if cascade.EmployeeConfig != nil {
		fmt.Fprintf(out, "┌─────────────────────────────────────────────────────────────┐\n")
		green.Fprintf(out, "│ 📋 Level 3: Personal Override (You)")
		fmt.Fprintf(out, "%s│\n", strings.Repeat(" ", 23))
		fmt.Fprintf(out, "├─────────────────────────────────────────────────────────────┤\n")

		baseConfig := cascade.OrgConfig.Config
		if cascade.TeamConfig != nil {
			baseConfig = mergeConfigs(baseConfig, cascade.TeamConfig.Config)
		}
		printConfigMap(out, cascade.EmployeeConfig.Config, "  ", baseConfig)

		fmt.Fprintf(out, "└─────────────────────────────────────────────────────────────┘\n")
		fmt.Fprintf(out, "\n")
	}

	// Final Resolved Configuration
	blue.Fprintf(out, "Final Resolved Configuration:\n")
	fmt.Fprintf(out, "{\n")

	// Determine source for each field
	sources := determineConfigSources(cascade)

	// Sort keys for consistent output
	keys := make([]string, 0, len(cascade.ResolvedConfig))
	for k := range cascade.ResolvedConfig {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, key := range keys {
		value := cascade.ResolvedConfig[key]
		source := sources[key]

		// Format value
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = fmt.Sprintf("%q", v)
		case float64:
			if v == float64(int64(v)) {
				valueStr = fmt.Sprintf("%.0f", v)
			} else {
				valueStr = fmt.Sprintf("%.1f", v)
			}
		default:
			valueStr = fmt.Sprintf("%v", v)
		}

		// Print with source annotation
		fmt.Fprintf(out, "  %q: %s", key, valueStr)
		if i < len(keys)-1 {
			fmt.Fprintf(out, ",")
		}
		gray.Fprintf(out, "  // from %s", source)
		if source == "employee" {
			green.Fprintf(out, " ⭐")
		} else if source == "team" {
			yellow.Fprintf(out, " ⬆️")
		}
		fmt.Fprintf(out, "\n")
	}

	fmt.Fprintf(out, "}\n")
	fmt.Fprintf(out, "\n")

	return nil
}

// printConfigMap prints a configuration map with optional override indicators
func printConfigMap(out interface{ Write([]byte) (int, error) }, config map[string]interface{}, indent string, baseConfig map[string]interface{}) {
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := config[key]

		// Check if this is an override
		isOverride := false
		if baseConfig != nil {
			if baseVal, exists := baseConfig[key]; exists && baseVal != value {
				isOverride = true
			}
		}

		// Format value
		var valueStr string
		switch v := value.(type) {
		case string:
			valueStr = fmt.Sprintf("%q", v)
		case float64:
			if v == float64(int64(v)) {
				valueStr = fmt.Sprintf("%.0f", v)
			} else {
				valueStr = fmt.Sprintf("%.1f", v)
			}
		default:
			valueStr = fmt.Sprintf("%v", v)
		}

		fmt.Fprintf(out, "│ %s%s: %s", indent, key, valueStr)

		// Add override indicator
		if isOverride {
			padding := 54 - len(indent) - len(key) - len(valueStr)
			if padding < 0 {
				padding = 0
			}
			fmt.Fprintf(out, "%s⬆️ overrides", strings.Repeat(" ", padding))
		}

		fmt.Fprintf(out, "%s│\n", strings.Repeat(" ", maxInt(0, 55-len(indent)-len(key)-len(valueStr))))
	}
}

// determineConfigSources determines which level each config value comes from
func determineConfigSources(cascade *AgentConfigCascade) map[string]string {
	sources := make(map[string]string)

	// Start with org
	if cascade.OrgConfig != nil {
		for key := range cascade.OrgConfig.Config {
			sources[key] = "org"
		}
	}

	// Override with team
	if cascade.TeamConfig != nil {
		for key := range cascade.TeamConfig.Config {
			sources[key] = "team"
		}
	}

	// Override with employee
	if cascade.EmployeeConfig != nil {
		for key := range cascade.EmployeeConfig.Config {
			sources[key] = "employee"
		}
	}

	return sources
}

// mergeConfigs merges two config maps (used for determining overrides)
func mergeConfigs(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range base {
		result[k] = v
	}
	for k, v := range override {
		result[k] = v
	}
	return result
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
