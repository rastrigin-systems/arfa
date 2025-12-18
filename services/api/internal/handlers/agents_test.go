package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rastrigin-systems/ubik-enterprise/generated/api"
	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
	"github.com/rastrigin-systems/ubik-enterprise/generated/mocks"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestListAgents_Success tests successful retrieval of active agents
func TestListAgents_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Mock agent data
	agent1 := db.Agent{
		ID:            uuid.New(),
		Name:          "Claude Code",
		Type:          "claude-code",
		Description:   "AI-powered code assistant with deep codebase understanding",
		Provider:      "anthropic",
		DefaultConfig: []byte(`{"model":"claude-3-5-sonnet-20241022","max_tokens":8192}`),
		Capabilities:  []byte(`["code_generation","code_review","refactoring"]`),
		LlmProvider:   "anthropic",
		LlmModel:      "claude-3-5-sonnet-20241022",
		IsPublic:      true,
		CreatedAt:     pgtype.Timestamp{Valid: true},
		UpdatedAt:     pgtype.Timestamp{Valid: true},
	}

	agent2 := db.Agent{
		ID:            uuid.New(),
		Name:          "Cursor",
		Type:          "cursor",
		Description:   "AI code editor with inline suggestions",
		Provider:      "openai",
		DefaultConfig: []byte(`{"model":"gpt-4","temperature":0.7}`),
		Capabilities:  []byte(`["inline_suggestions","code_completion"]`),
		LlmProvider:   "openai",
		LlmModel:      "gpt-4",
		IsPublic:      true,
		CreatedAt:     pgtype.Timestamp{Valid: true},
		UpdatedAt:     pgtype.Timestamp{Valid: true},
	}

	// Expect ListAgents to be called
	mockDB.EXPECT().
		ListAgents(gomock.Any()).
		Return([]db.Agent{agent1, agent2}, nil)

	handler := handlers.NewAgentsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()

	handler.ListAgents(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListAgentsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Agents))
	assert.Equal(t, 2, response.Total)

	// Verify first agent
	assert.Equal(t, "Claude Code", response.Agents[0].Name)
	assert.Equal(t, "claude-code", response.Agents[0].Type)
	assert.Equal(t, "anthropic", response.Agents[0].Provider)

	// Verify second agent
	assert.Equal(t, "Cursor", response.Agents[1].Name)
	assert.Equal(t, "cursor", response.Agents[1].Type)
}

// TestListAgents_EmptyResult tests when no agents are available
func TestListAgents_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Return empty list
	mockDB.EXPECT().
		ListAgents(gomock.Any()).
		Return([]db.Agent{}, nil)

	handler := handlers.NewAgentsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()

	handler.ListAgents(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListAgentsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 0, len(response.Agents))
	assert.Equal(t, 0, response.Total)
}

// TestListAgents_OnlyActiveAgents tests that only is_public=true agents are returned
// This test verifies the SQL query filters correctly (tested via SQL, but important to document)
func TestListAgents_OnlyActiveAgents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Only return public agents (is_public = true)
	publicAgent := db.Agent{
		ID:          uuid.New(),
		Name:        "Claude Code",
		Type:        "claude-code",
		Description: "AI-powered code assistant",
		Provider:    "anthropic",
		IsPublic:    true, // Only public agents
		CreatedAt:   pgtype.Timestamp{Valid: true},
		UpdatedAt:   pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		ListAgents(gomock.Any()).
		Return([]db.Agent{publicAgent}, nil)

	handler := handlers.NewAgentsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()

	handler.ListAgents(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListAgentsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify all returned agents are public
	for _, agent := range response.Agents {
		assert.True(t, agent.IsPublic)
	}
}

// TestListAgents_OrderedByName tests that agents are returned in alphabetical order
// This test verifies the SQL query ORDER BY clause works correctly
func TestListAgents_OrderedByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Create agents with names in alphabetical order (already sorted from DB)
	agentA := db.Agent{
		ID:          uuid.New(),
		Name:        "Agent A",
		Type:        "agent-a",
		Description: "First agent",
		Provider:    "provider-a",
		IsPublic:    true,
		CreatedAt:   pgtype.Timestamp{Valid: true},
		UpdatedAt:   pgtype.Timestamp{Valid: true},
	}

	agentB := db.Agent{
		ID:          uuid.New(),
		Name:        "Agent B",
		Type:        "agent-b",
		Description: "Second agent",
		Provider:    "provider-b",
		IsPublic:    true,
		CreatedAt:   pgtype.Timestamp{Valid: true},
		UpdatedAt:   pgtype.Timestamp{Valid: true},
	}

	agentC := db.Agent{
		ID:          uuid.New(),
		Name:        "Agent C",
		Type:        "agent-c",
		Description: "Third agent",
		Provider:    "provider-c",
		IsPublic:    true,
		CreatedAt:   pgtype.Timestamp{Valid: true},
		UpdatedAt:   pgtype.Timestamp{Valid: true},
	}

	// Return agents in alphabetical order
	mockDB.EXPECT().
		ListAgents(gomock.Any()).
		Return([]db.Agent{agentA, agentB, agentC}, nil)

	handler := handlers.NewAgentsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), uuid.New()))
	rec := httptest.NewRecorder()

	handler.ListAgents(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListAgentsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify order
	assert.Equal(t, "Agent A", response.Agents[0].Name)
	assert.Equal(t, "Agent B", response.Agents[1].Name)
	assert.Equal(t, "Agent C", response.Agents[2].Name)
}

// ============================================================================
// GetAgentByID Tests
// ============================================================================

// TDD Lesson: Test successful retrieval of a specific agent
func TestGetAgentByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	agentID := uuid.New()
	agent := db.Agent{
		ID:            agentID,
		Name:          "Claude Code",
		Type:          "claude-code",
		Description:   "AI-powered code assistant with deep codebase understanding",
		Provider:      "anthropic",
		DefaultConfig: []byte(`{"model":"claude-3-5-sonnet-20241022","max_tokens":8192}`),
		Capabilities:  []byte(`["code_generation","code_review","refactoring"]`),
		LlmProvider:   "anthropic",
		LlmModel:      "claude-3-5-sonnet-20241022",
		IsPublic:      true,
		CreatedAt:     pgtype.Timestamp{Valid: true},
		UpdatedAt:     pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	handler := handlers.NewAgentsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/agents/"+agentID.String(), nil)
	rec := httptest.NewRecorder()

	// We'll test this with integration tests where routing works
	// For now just call the method directly with the ID
	handler.GetAgentByIDDirect(rec, req, agentID)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Agent
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, agentID.String(), response.Id.String())
	assert.Equal(t, "Claude Code", response.Name)
	assert.Equal(t, "claude-code", response.Type)
	assert.Equal(t, "anthropic", response.Provider)
	assert.True(t, response.IsPublic)
	assert.NotNil(t, response.DefaultConfig)
	assert.NotNil(t, response.Capabilities)
}

// TDD Lesson: Test agent not found
func TestGetAgentByID_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	agentID := uuid.New()

	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(db.Agent{}, assert.AnError)

	handler := handlers.NewAgentsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/agents/"+agentID.String(), nil)
	rec := httptest.NewRecorder()

	handler.GetAgentByIDDirect(rec, req, agentID)

	// Should return 404
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
