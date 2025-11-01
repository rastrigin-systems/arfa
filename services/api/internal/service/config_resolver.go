package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

// ConfigResolver handles hierarchical configuration resolution
// Merges org → team → employee configs with proper override semantics
type ConfigResolver struct {
	db db.Querier
}

// NewConfigResolver creates a new config resolver
func NewConfigResolver(database db.Querier) *ConfigResolver {
	return &ConfigResolver{
		db: database,
	}
}

// ResolvedAgentConfig represents a fully resolved agent configuration
type ResolvedAgentConfig struct {
	AgentID       uuid.UUID              `json:"agent_id"`
	AgentName     string                 `json:"agent_name"`
	AgentType     string                 `json:"agent_type"`
	Provider      string                 `json:"provider"`
	Config        map[string]interface{} `json:"config"`          // Merged config
	SystemPrompt  string                 `json:"system_prompt"`   // Concatenated prompts
	IsEnabled     bool                   `json:"is_enabled"`      // All levels must be enabled
	SyncToken     *string                `json:"sync_token"`      // For CLI sync
	LastSyncedAt  *string                `json:"last_synced_at"`  // ISO8601 timestamp
}

// ResolveEmployeeAgents resolves all agent configs for an employee
// This is the primary method used by the CLI sync endpoint
func (r *ConfigResolver) ResolveEmployeeAgents(ctx context.Context, employeeID uuid.UUID) ([]ResolvedAgentConfig, error) {
	// 1. Get employee to find org_id and team_id
	employee, err := r.db.GetEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	// 2. Get all org-level agent configs for this org
	orgConfigs, err := r.db.ListOrgAgentConfigs(ctx, employee.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list org agent configs: %w", err)
	}

	// 3. Resolve each agent config
	var resolved []ResolvedAgentConfig
	for _, orgConfig := range orgConfigs {
		resolvedConfig, err := r.ResolveAgentConfig(ctx, employeeID, orgConfig.AgentID)
		if err != nil {
			// Log error but continue with other agents
			continue
		}

		// Only include enabled agents
		if resolvedConfig.IsEnabled {
			resolved = append(resolved, *resolvedConfig)
		}
	}

	return resolved, nil
}

// ResolveAgentConfig resolves a single agent config for an employee
// Merges org → team → employee configs
func (r *ConfigResolver) ResolveAgentConfig(ctx context.Context, employeeID, agentID uuid.UUID) (*ResolvedAgentConfig, error) {
	// 1. Get employee to find org_id and team_id
	employee, err := r.db.GetEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	// 2. Get agent details
	agent, err := r.db.GetAgentByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	// 3. Get org-level config (base)
	orgConfig, err := r.db.GetOrgAgentConfig(ctx, db.GetOrgAgentConfigParams{
		OrgID:   employee.OrgID,
		AgentID: agentID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			// No org config = agent not enabled for this org
			return nil, fmt.Errorf("agent not configured at org level")
		}
		return nil, fmt.Errorf("failed to get org config: %w", err)
	}

	// Start with org config
	mergedConfig := make(map[string]interface{})
	if err := json.Unmarshal(orgConfig.Config, &mergedConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal org config: %w", err)
	}

	isEnabled := orgConfig.IsEnabled

	// 4. Get team-level config and merge (if employee has a team)
	if employee.TeamID.Valid {
		teamConfig, err := r.db.GetTeamAgentConfig(ctx, db.GetTeamAgentConfigParams{
			TeamID:  employee.TeamID.Bytes,
			AgentID: agentID,
		})
		if err == nil {
			// Team config exists, merge it
			var teamOverride map[string]interface{}
			if err := json.Unmarshal(teamConfig.ConfigOverride, &teamOverride); err != nil {
				return nil, fmt.Errorf("failed to unmarshal team config: %w", err)
			}
			mergedConfig = deepMerge(mergedConfig, teamOverride)
			isEnabled = isEnabled && teamConfig.IsEnabled
		} else if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("failed to get team config: %w", err)
		}
	}

	// 5. Get employee-level config and merge
	var syncToken *string
	var lastSyncedAt *string

	employeeConfig, err := r.db.GetEmployeeAgentConfigByAgent(ctx, db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agentID,
	})
	if err == nil {
		// Employee config exists, merge it
		var employeeOverride map[string]interface{}
		if err := json.Unmarshal(employeeConfig.ConfigOverride, &employeeOverride); err != nil {
			return nil, fmt.Errorf("failed to unmarshal employee config: %w", err)
		}
		mergedConfig = deepMerge(mergedConfig, employeeOverride)
		isEnabled = isEnabled && employeeConfig.IsEnabled

		// Set sync metadata (SyncToken is already *string)
		syncToken = employeeConfig.SyncToken

		// LastSyncedAt is pgtype.Timestamp
		if employeeConfig.LastSyncedAt.Valid {
			timestamp := employeeConfig.LastSyncedAt.Time.Format("2006-01-02T15:04:05Z07:00")
			lastSyncedAt = &timestamp
		}
	} else if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get employee config: %w", err)
	}

	// 6. Get and concatenate system prompts
	systemPrompt, err := r.resolveSystemPrompts(ctx, employee.OrgID, employee.TeamID, employeeID, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve system prompts: %w", err)
	}

	return &ResolvedAgentConfig{
		AgentID:      agentID,
		AgentName:    agent.Name,
		AgentType:    agent.Type,
		Provider:     agent.Provider,
		Config:       mergedConfig,
		SystemPrompt: systemPrompt,
		IsEnabled:    isEnabled,
		SyncToken:    syncToken,
		LastSyncedAt: lastSyncedAt,
	}, nil
}

// resolveSystemPrompts fetches and concatenates prompts from all hierarchy levels
// Order: org prompts → team prompts → employee prompts (by priority within each level)
func (r *ConfigResolver) resolveSystemPrompts(ctx context.Context, orgID uuid.UUID, teamID pgtype.UUID, employeeID uuid.UUID, agentID uuid.UUID) (string, error) {
	teamUUID := uuid.Nil
	if teamID.Valid {
		teamUUID = teamID.Bytes
	}

	prompts, err := r.db.GetSystemPrompts(ctx, db.GetSystemPromptsParams{
		ScopeID:   orgID,
		AgentID:   pgtype.UUID{Bytes: agentID, Valid: true},
		ScopeID_2: teamUUID,
		ScopeID_3: employeeID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get system prompts: %w", err)
	}

	// Concatenate prompts with newlines
	var parts []string
	for _, prompt := range prompts {
		if prompt.Prompt != "" {
			parts = append(parts, prompt.Prompt)
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

// deepMerge performs a deep merge of two JSON objects
// Values in 'override' take precedence over 'base'
func deepMerge(base, override map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy base
	for k, v := range base {
		result[k] = v
	}

	// Merge override
	for k, v := range override {
		if existingVal, exists := result[k]; exists {
			// If both are maps, recursively merge
			if existingMap, ok := existingVal.(map[string]interface{}); ok {
				if overrideMap, ok := v.(map[string]interface{}); ok {
					result[k] = deepMerge(existingMap, overrideMap)
					continue
				}
			}
		}
		// Otherwise, override wins
		result[k] = v
	}

	return result
}

// ToAPIAgentConfig converts ResolvedAgentConfig to API format
// TODO: Uncomment once api.AgentConfig is added to OpenAPI spec
// func (r *ResolvedAgentConfig) ToAPIAgentConfig() api.AgentConfig {
// 	return api.AgentConfig{
// 		AgentId:      (*openapi_types.UUID)(&r.AgentID),
// 		AgentName:    r.AgentName,
// 		AgentType:    r.AgentType,
// 		Provider:     r.Provider,
// 		Config:       r.Config,
// 		SystemPrompt: &r.SystemPrompt,
// 		IsEnabled:    r.IsEnabled,
// 		SyncToken:    r.SyncToken,
// 		LastSyncedAt: r.LastSyncedAt,
// 	}
// }
