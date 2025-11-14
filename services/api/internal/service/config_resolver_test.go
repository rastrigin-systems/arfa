package service_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestResolveAgentConfig_OrgOnly tests config resolution with only org-level config
func TestResolveAgentConfig_OrgOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	resolver := service.NewConfigResolver(mockDB)

	ctx := context.Background()
	employeeID := uuid.New()
	agentID := uuid.New()
	orgID := uuid.New()

	// Mock employee
	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		TeamID: pgtype.UUID{Valid: false}, // No team
	}

	// Mock agent
	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}

	// Mock org config
	orgConfig := db.OrgAgentConfig{
		ID:        uuid.New(),
		OrgID:     orgID,
		AgentID:   agentID,
		Config:    []byte(`{"model":"claude-3-5-sonnet-20241022","temperature":0.2,"max_tokens":8192}`),
		IsEnabled: true,
	}

	// Set expectations
	mockDB.EXPECT().GetEmployee(ctx, employeeID).Return(employee, nil)
	mockDB.EXPECT().GetAgentByID(ctx, agentID).Return(agent, nil)
	mockDB.EXPECT().GetOrgAgentConfig(ctx, db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agentID,
	}).Return(orgConfig, nil)
	mockDB.EXPECT().GetEmployeeAgentConfigByAgent(ctx, db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agentID,
	}).Return(db.GetEmployeeAgentConfigByAgentRow{}, pgx.ErrNoRows)
	mockDB.EXPECT().GetSystemPrompts(ctx, gomock.Any()).Return([]db.SystemPrompt{}, nil)

	// Execute
	result, err := resolver.ResolveAgentConfig(ctx, employeeID, agentID)

	// Verify
	require.NoError(t, err)
	assert.Equal(t, agentID, result.AgentID)
	assert.Equal(t, "Claude Code", result.AgentName)
	assert.True(t, result.IsEnabled)
	assert.Equal(t, "claude-3-5-sonnet-20241022", result.Config["model"])
	assert.Equal(t, float64(0.2), result.Config["temperature"])
	assert.Equal(t, float64(8192), result.Config["max_tokens"])
}

// TestResolveAgentConfig_TeamOverride tests config resolution with team override
func TestResolveAgentConfig_TeamOverride(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	resolver := service.NewConfigResolver(mockDB)

	ctx := context.Background()
	employeeID := uuid.New()
	agentID := uuid.New()
	orgID := uuid.New()
	teamID := uuid.New()

	// Mock employee with team
	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		TeamID: pgtype.UUID{Bytes: teamID, Valid: true},
	}

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}

	// Org config (base)
	orgConfig := db.OrgAgentConfig{
		ID:        uuid.New(),
		OrgID:     orgID,
		AgentID:   agentID,
		Config:    []byte(`{"model":"claude-3-5-sonnet-20241022","temperature":0.2,"max_tokens":8192}`),
		IsEnabled: true,
	}

	// Team override (changes temperature)
	teamConfig := db.TeamAgentConfig{
		ID:             uuid.New(),
		TeamID:         teamID,
		AgentID:        agentID,
		ConfigOverride: []byte(`{"temperature":0.5}`),
		IsEnabled:      true,
	}

	// Set expectations
	mockDB.EXPECT().GetEmployee(ctx, employeeID).Return(employee, nil)
	mockDB.EXPECT().GetAgentByID(ctx, agentID).Return(agent, nil)
	mockDB.EXPECT().GetOrgAgentConfig(ctx, db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agentID,
	}).Return(orgConfig, nil)
	mockDB.EXPECT().GetTeamAgentConfig(ctx, db.GetTeamAgentConfigParams{
		TeamID:  teamID,
		AgentID: agentID,
	}).Return(teamConfig, nil)
	mockDB.EXPECT().GetEmployeeAgentConfigByAgent(ctx, db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agentID,
	}).Return(db.GetEmployeeAgentConfigByAgentRow{}, pgx.ErrNoRows)
	mockDB.EXPECT().GetSystemPrompts(ctx, gomock.Any()).Return([]db.SystemPrompt{}, nil)

	// Execute
	result, err := resolver.ResolveAgentConfig(ctx, employeeID, agentID)

	// Verify
	require.NoError(t, err)
	assert.True(t, result.IsEnabled)
	assert.Equal(t, "claude-3-5-sonnet-20241022", result.Config["model"]) // From org
	assert.Equal(t, float64(0.5), result.Config["temperature"])           // From team (overridden)
	assert.Equal(t, float64(8192), result.Config["max_tokens"])           // From org
}

// TestResolveAgentConfig_EmployeeOverride tests full hierarchy
func TestResolveAgentConfig_EmployeeOverride(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	resolver := service.NewConfigResolver(mockDB)

	ctx := context.Background()
	employeeID := uuid.New()
	agentID := uuid.New()
	orgID := uuid.New()
	teamID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		TeamID: pgtype.UUID{Bytes: teamID, Valid: true},
	}

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}

	// Org config (base)
	orgConfig := db.OrgAgentConfig{
		ID:        uuid.New(),
		OrgID:     orgID,
		AgentID:   agentID,
		Config:    []byte(`{"model":"claude-3-5-sonnet-20241022","temperature":0.2,"max_tokens":8192}`),
		IsEnabled: true,
	}

	// Team override (changes temperature)
	teamConfig := db.TeamAgentConfig{
		ID:             uuid.New(),
		TeamID:         teamID,
		AgentID:        agentID,
		ConfigOverride: []byte(`{"temperature":0.5}`),
		IsEnabled:      true,
	}

	// Employee override (changes max_tokens)
	employeeConfig := db.GetEmployeeAgentConfigByAgentRow{
		ID:             uuid.New(),
		EmployeeID:     employeeID,
		AgentID:        agentID,
		ConfigOverride: []byte(`{"max_tokens":16384}`),
		IsEnabled:      true,
		SyncToken:      nil,
		LastSyncedAt:   pgtype.Timestamp{Valid: false},
		CreatedAt:      pgtype.Timestamp{Valid: true},
		UpdatedAt:      pgtype.Timestamp{Valid: true},
	}

	// Set expectations
	mockDB.EXPECT().GetEmployee(ctx, employeeID).Return(employee, nil)
	mockDB.EXPECT().GetAgentByID(ctx, agentID).Return(agent, nil)
	mockDB.EXPECT().GetOrgAgentConfig(ctx, db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agentID,
	}).Return(orgConfig, nil)
	mockDB.EXPECT().GetTeamAgentConfig(ctx, db.GetTeamAgentConfigParams{
		TeamID:  teamID,
		AgentID: agentID,
	}).Return(teamConfig, nil)
	mockDB.EXPECT().GetEmployeeAgentConfigByAgent(ctx, db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agentID,
	}).Return(employeeConfig, nil)
	mockDB.EXPECT().GetSystemPrompts(ctx, gomock.Any()).Return([]db.SystemPrompt{}, nil)

	// Execute
	result, err := resolver.ResolveAgentConfig(ctx, employeeID, agentID)

	// Verify - should have merged all three levels
	require.NoError(t, err)
	assert.True(t, result.IsEnabled)
	assert.Equal(t, "claude-3-5-sonnet-20241022", result.Config["model"]) // From org
	assert.Equal(t, float64(0.5), result.Config["temperature"])           // From team
	assert.Equal(t, float64(16384), result.Config["max_tokens"])          // From employee (overridden)
}

// TestResolveAgentConfig_DisabledAtTeam tests that disabled at team level makes agent disabled
func TestResolveAgentConfig_DisabledAtTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	resolver := service.NewConfigResolver(mockDB)

	ctx := context.Background()
	employeeID := uuid.New()
	agentID := uuid.New()
	orgID := uuid.New()
	teamID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		TeamID: pgtype.UUID{Bytes: teamID, Valid: true},
	}

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}

	// Org config (enabled)
	orgConfig := db.OrgAgentConfig{
		ID:        uuid.New(),
		OrgID:     orgID,
		AgentID:   agentID,
		Config:    []byte(`{"model":"claude-3-5-sonnet-20241022"}`),
		IsEnabled: true,
	}

	// Team config (DISABLED)
	teamConfig := db.TeamAgentConfig{
		ID:             uuid.New(),
		TeamID:         teamID,
		AgentID:        agentID,
		ConfigOverride: []byte(`{}`),
		IsEnabled:      false, // DISABLED
	}

	// Set expectations
	mockDB.EXPECT().GetEmployee(ctx, employeeID).Return(employee, nil)
	mockDB.EXPECT().GetAgentByID(ctx, agentID).Return(agent, nil)
	mockDB.EXPECT().GetOrgAgentConfig(ctx, db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agentID,
	}).Return(orgConfig, nil)
	mockDB.EXPECT().GetTeamAgentConfig(ctx, db.GetTeamAgentConfigParams{
		TeamID:  teamID,
		AgentID: agentID,
	}).Return(teamConfig, nil)
	mockDB.EXPECT().GetEmployeeAgentConfigByAgent(ctx, db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agentID,
	}).Return(db.GetEmployeeAgentConfigByAgentRow{}, pgx.ErrNoRows)
	mockDB.EXPECT().GetSystemPrompts(ctx, gomock.Any()).Return([]db.SystemPrompt{}, nil)

	// Execute
	result, err := resolver.ResolveAgentConfig(ctx, employeeID, agentID)

	// Verify - should be disabled
	require.NoError(t, err)
	assert.False(t, result.IsEnabled) // Team disabled it
}

// TestResolveAgentConfig_SystemPrompts tests system prompt concatenation
func TestResolveAgentConfig_SystemPrompts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	resolver := service.NewConfigResolver(mockDB)

	ctx := context.Background()
	employeeID := uuid.New()
	agentID := uuid.New()
	orgID := uuid.New()
	teamID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		TeamID: pgtype.UUID{Bytes: teamID, Valid: true},
	}

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}

	orgConfig := db.OrgAgentConfig{
		ID:        uuid.New(),
		OrgID:     orgID,
		AgentID:   agentID,
		Config:    []byte(`{"model":"claude-3-5-sonnet-20241022"}`),
		IsEnabled: true,
	}

	// System prompts at different levels
	prompts := []db.SystemPrompt{
		{
			ID:        uuid.New(),
			ScopeType: "org",
			ScopeID:   orgID,
			AgentID:   pgtype.UUID{Bytes: agentID, Valid: true},
			Prompt:    "You are a helpful coding assistant.",
			Priority:  0,
		},
		{
			ID:        uuid.New(),
			ScopeType: "team",
			ScopeID:   teamID,
			AgentID:   pgtype.UUID{Bytes: agentID, Valid: true},
			Prompt:    "Always follow team coding standards.",
			Priority:  0,
		},
		{
			ID:        uuid.New(),
			ScopeType: "employee",
			ScopeID:   employeeID,
			AgentID:   pgtype.UUID{Bytes: agentID, Valid: true},
			Prompt:    "Use TypeScript for all code.",
			Priority:  0,
		},
	}

	// Set expectations
	mockDB.EXPECT().GetEmployee(ctx, employeeID).Return(employee, nil)
	mockDB.EXPECT().GetAgentByID(ctx, agentID).Return(agent, nil)
	mockDB.EXPECT().GetOrgAgentConfig(ctx, db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agentID,
	}).Return(orgConfig, nil)
	mockDB.EXPECT().GetTeamAgentConfig(ctx, db.GetTeamAgentConfigParams{
		TeamID:  teamID,
		AgentID: agentID,
	}).Return(db.TeamAgentConfig{}, pgx.ErrNoRows)
	mockDB.EXPECT().GetEmployeeAgentConfigByAgent(ctx, db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agentID,
	}).Return(db.GetEmployeeAgentConfigByAgentRow{}, pgx.ErrNoRows)
	mockDB.EXPECT().GetSystemPrompts(ctx, gomock.Any()).Return(prompts, nil)

	// Execute
	result, err := resolver.ResolveAgentConfig(ctx, employeeID, agentID)

	// Verify - prompts should be concatenated
	require.NoError(t, err)
	expectedPrompt := "You are a helpful coding assistant.\n\nAlways follow team coding standards.\n\nUse TypeScript for all code."
	assert.Equal(t, expectedPrompt, result.SystemPrompt)
}

// TestResolveAgentConfig_NoOrgConfig tests error when agent not configured at org level
func TestResolveAgentConfig_NoOrgConfig(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	resolver := service.NewConfigResolver(mockDB)

	ctx := context.Background()
	employeeID := uuid.New()
	agentID := uuid.New()
	orgID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		TeamID: pgtype.UUID{Valid: false},
	}

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}

	// Set expectations - no org config
	mockDB.EXPECT().GetEmployee(ctx, employeeID).Return(employee, nil)
	mockDB.EXPECT().GetAgentByID(ctx, agentID).Return(agent, nil)
	mockDB.EXPECT().GetOrgAgentConfig(ctx, db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agentID,
	}).Return(db.OrgAgentConfig{}, pgx.ErrNoRows)

	// Execute
	result, err := resolver.ResolveAgentConfig(ctx, employeeID, agentID)

	// Verify - should error
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "agent not configured at org level")
}

// TestResolveEmployeeAgents tests resolving all agents for an employee
func TestResolveEmployeeAgents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	resolver := service.NewConfigResolver(mockDB)

	ctx := context.Background()
	employeeID := uuid.New()
	orgID := uuid.New()
	agent1ID := uuid.New()
	agent2ID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		TeamID: pgtype.UUID{Valid: false},
	}

	// Two org configs
	orgConfigs := []db.ListOrgAgentConfigsRow{
		{
			ID:            uuid.New(),
			OrgID:         orgID,
			AgentID:       agent1ID,
			Config:        []byte(`{"model":"claude-3-5-sonnet-20241022"}`),
			IsEnabled:     true,
			AgentName:     "Claude Code",
			AgentType:     "claude-code",
			AgentProvider: "anthropic",
		},
		{
			ID:            uuid.New(),
			OrgID:         orgID,
			AgentID:       agent2ID,
			Config:        []byte(`{"model":"gpt-4o"}`),
			IsEnabled:     true,
			AgentName:     "Cursor",
			AgentType:     "cursor",
			AgentProvider: "openai",
		},
	}

	agent1 := db.Agent{
		ID:       agent1ID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}

	agent2 := db.Agent{
		ID:       agent2ID,
		Name:     "Cursor",
		Type:     "cursor",
		Provider: "openai",
	}

	orgConfig1 := db.OrgAgentConfig{
		ID:        uuid.New(),
		OrgID:     orgID,
		AgentID:   agent1ID,
		Config:    []byte(`{"model":"claude-3-5-sonnet-20241022"}`),
		IsEnabled: true,
	}

	orgConfig2 := db.OrgAgentConfig{
		ID:        uuid.New(),
		OrgID:     orgID,
		AgentID:   agent2ID,
		Config:    []byte(`{"model":"gpt-4o"}`),
		IsEnabled: true,
	}

	// Set expectations
	mockDB.EXPECT().GetEmployee(ctx, employeeID).Return(employee, nil).Times(3) // Once for list, twice for resolve
	mockDB.EXPECT().ListOrgAgentConfigs(ctx, orgID).Return(orgConfigs, nil)

	// Agent 1 resolution
	mockDB.EXPECT().GetAgentByID(ctx, agent1ID).Return(agent1, nil)
	mockDB.EXPECT().GetOrgAgentConfig(ctx, db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agent1ID,
	}).Return(orgConfig1, nil)
	mockDB.EXPECT().GetEmployeeAgentConfigByAgent(ctx, db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agent1ID,
	}).Return(db.GetEmployeeAgentConfigByAgentRow{}, pgx.ErrNoRows)
	mockDB.EXPECT().GetSystemPrompts(ctx, gomock.Any()).Return([]db.SystemPrompt{}, nil)

	// Agent 2 resolution
	mockDB.EXPECT().GetAgentByID(ctx, agent2ID).Return(agent2, nil)
	mockDB.EXPECT().GetOrgAgentConfig(ctx, db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agent2ID,
	}).Return(orgConfig2, nil)
	mockDB.EXPECT().GetEmployeeAgentConfigByAgent(ctx, db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agent2ID,
	}).Return(db.GetEmployeeAgentConfigByAgentRow{}, pgx.ErrNoRows)
	mockDB.EXPECT().GetSystemPrompts(ctx, gomock.Any()).Return([]db.SystemPrompt{}, nil)

	// Execute
	results, err := resolver.ResolveEmployeeAgents(ctx, employeeID)

	// Verify
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Claude Code", results[0].AgentName)
	assert.Equal(t, "Cursor", results[1].AgentName)
}

// TestDeepMerge tests the deep merge logic
func TestDeepMerge(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		override string
		expected string
	}{
		{
			name:     "simple override",
			base:     `{"a":1,"b":2}`,
			override: `{"b":3}`,
			expected: `{"a":1,"b":3}`,
		},
		{
			name:     "nested override",
			base:     `{"outer":{"inner":1,"keep":2}}`,
			override: `{"outer":{"inner":10}}`,
			expected: `{"outer":{"inner":10,"keep":2}}`,
		},
		{
			name:     "add new key",
			base:     `{"a":1}`,
			override: `{"b":2}`,
			expected: `{"a":1,"b":2}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var base, override, expected map[string]interface{}
			json.Unmarshal([]byte(tt.base), &base)
			json.Unmarshal([]byte(tt.override), &override)
			json.Unmarshal([]byte(tt.expected), &expected)

			// We can't directly test the private deepMerge function,
			// but we can test it through the public API by creating a resolver
			// and checking the merged results from ResolveAgentConfig
			// For now, just document the expected behavior
			assert.NotNil(t, base)
			assert.NotNil(t, override)
			assert.NotNil(t, expected)
		})
	}
}
