package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestListOrgAgentConfigs_Success tests successful retrieval of org agent configs
func TestListOrgAgentConfigs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	agent1ID := uuid.New()
	agent2ID := uuid.New()

	// Mock org configs
	configs := []db.ListOrgAgentConfigsRow{
		{
			ID:                 uuid.New(),
			OrgID:              orgID,
			AgentID:            agent1ID,
			Config:             []byte(`{"model":"claude-3-5-sonnet-20241022","temperature":0.2}`),
			IsEnabled:          true,
			AgentName:          "Claude Code",
			AgentType:          "claude-code",
			AgentProvider:      "anthropic",
			AgentDefaultConfig: []byte(`{}`),
			CreatedAt:          pgtype.Timestamp{Valid: true},
			UpdatedAt:          pgtype.Timestamp{Valid: true},
		},
		{
			ID:                 uuid.New(),
			OrgID:              orgID,
			AgentID:            agent2ID,
			Config:             []byte(`{"model":"gpt-4o","temperature":0.3}`),
			IsEnabled:          true,
			AgentName:          "Cursor",
			AgentType:          "cursor",
			AgentProvider:      "openai",
			AgentDefaultConfig: []byte(`{}`),
			CreatedAt:          pgtype.Timestamp{Valid: true},
			UpdatedAt:          pgtype.Timestamp{Valid: true},
		},
	}

	// Expect ListOrgAgentConfigs to be called
	mockDB.EXPECT().
		ListOrgAgentConfigs(gomock.Any(), orgID).
		Return(configs, nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations/current/agent-configs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListOrgAgentConfigs(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListOrgAgentConfigsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Configs))
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, "Claude Code", *response.Configs[0].AgentName)
	assert.Equal(t, "Cursor", *response.Configs[1].AgentName)
}

// TestListOrgAgentConfigs_EmptyResult tests when no configs exist
func TestListOrgAgentConfigs_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	orgID := uuid.New()

	// Return empty list
	mockDB.EXPECT().
		ListOrgAgentConfigs(gomock.Any(), orgID).
		Return([]db.ListOrgAgentConfigsRow{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations/current/agent-configs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListOrgAgentConfigs(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListOrgAgentConfigsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 0, len(response.Configs))
	assert.Equal(t, 0, response.Total)
}

// TestCreateOrgAgentConfig_Success tests successful creation of org agent config
func TestCreateOrgAgentConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	agentID := uuid.New()
	configID := uuid.New()

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}

	createdConfig := db.OrgAgentConfig{
		ID:        configID,
		OrgID:     orgID,
		AgentID:   agentID,
		Config:    []byte(`{"model":"claude-3-5-sonnet-20241022","temperature":0.2}`),
		IsEnabled: true,
		CreatedAt: pgtype.Timestamp{Valid: true},
		UpdatedAt: pgtype.Timestamp{Valid: true},
	}

	// Expect agent verification
	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	// Expect duplicate check
	mockDB.EXPECT().
		CheckOrgAgentConfigExists(gomock.Any(), db.CheckOrgAgentConfigExistsParams{
			OrgID:   orgID,
			AgentID: agentID,
		}).
		Return(false, nil)

	// Expect config creation
	mockDB.EXPECT().
		CreateOrgAgentConfig(gomock.Any(), gomock.Any()).
		Return(createdConfig, nil)

	// Expect agent fetch for response
	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	reqBody := map[string]interface{}{
		"agent_id": agentID.String(),
		"config": map[string]interface{}{
			"model":       "claude-3-5-sonnet-20241022",
			"temperature": 0.2,
		},
		"is_enabled": true,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/organizations/current/agent-configs", bytes.NewReader(body))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateOrgAgentConfig(rec, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response api.OrgAgentConfig
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.IsEnabled)
	assert.Equal(t, "Claude Code", *response.AgentName)
}

// TestCreateOrgAgentConfig_DuplicateAgent tests duplicate agent error
func TestCreateOrgAgentConfig_DuplicateAgent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	agentID := uuid.New()

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}

	// Expect agent verification
	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	// Expect duplicate check - returns true
	mockDB.EXPECT().
		CheckOrgAgentConfigExists(gomock.Any(), db.CheckOrgAgentConfigExistsParams{
			OrgID:   orgID,
			AgentID: agentID,
		}).
		Return(true, nil)

	reqBody := map[string]interface{}{
		"agent_id": agentID.String(),
		"config": map[string]interface{}{
			"model": "claude-3-5-sonnet-20241022",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/organizations/current/agent-configs", bytes.NewReader(body))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateOrgAgentConfig(rec, req)

	// Verify response
	assert.Equal(t, http.StatusConflict, rec.Code)
}

// TestCreateOrgAgentConfig_AgentNotFound tests agent not found error
func TestCreateOrgAgentConfig_AgentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	agentID := uuid.New()

	// Expect agent verification - returns not found
	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(db.Agent{}, pgx.ErrNoRows)

	reqBody := map[string]interface{}{
		"agent_id": agentID.String(),
		"config": map[string]interface{}{
			"model": "claude-3-5-sonnet-20241022",
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/organizations/current/agent-configs", bytes.NewReader(body))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateOrgAgentConfig(rec, req)

	// Verify response
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TestCreateOrgAgentConfig_InvalidJSON tests invalid JSON error
func TestCreateOrgAgentConfig_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	orgID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/organizations/current/agent-configs", bytes.NewReader([]byte("invalid json")))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateOrgAgentConfig(rec, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestUpdateOrgAgentConfig_Success tests successful update
func TestUpdateOrgAgentConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	configID := uuid.New()
	orgID := uuid.New()
	agentID := uuid.New()

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}

	updatedConfig := db.OrgAgentConfig{
		ID:        configID,
		OrgID:     orgID,
		AgentID:   agentID,
		Config:    []byte(`{"model":"claude-3-5-sonnet-20241022","temperature":0.5}`),
		IsEnabled: true,
		CreatedAt: pgtype.Timestamp{Valid: true},
		UpdatedAt: pgtype.Timestamp{Valid: true},
	}

	// Expect config update
	mockDB.EXPECT().
		UpdateOrgAgentConfig(gomock.Any(), gomock.Any()).
		Return(updatedConfig, nil)

	// Expect agent fetch for response
	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	reqBody := map[string]interface{}{
		"config": map[string]interface{}{
			"temperature": 0.5,
		},
	}
	body, _ := json.Marshal(reqBody)

	// Use chi router to properly set URL params
	r := chi.NewRouter()
	r.Patch("/organizations/current/agent-configs/{config_id}", handler.UpdateOrgAgentConfig)

	req := httptest.NewRequest(http.MethodPatch, "/organizations/current/agent-configs/"+configID.String(), bytes.NewReader(body))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestUpdateOrgAgentConfig_NotFound tests config not found error
func TestUpdateOrgAgentConfig_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	configID := uuid.New()
	orgID := uuid.New()

	// Expect config update - returns not found
	mockDB.EXPECT().
		UpdateOrgAgentConfig(gomock.Any(), gomock.Any()).
		Return(db.OrgAgentConfig{}, pgx.ErrNoRows)

	reqBody := map[string]interface{}{
		"config": map[string]interface{}{
			"temperature": 0.5,
		},
	}
	body, _ := json.Marshal(reqBody)

	// Use chi router to properly set URL params
	r := chi.NewRouter()
	r.Patch("/organizations/current/agent-configs/{config_id}", handler.UpdateOrgAgentConfig)

	req := httptest.NewRequest(http.MethodPatch, "/organizations/current/agent-configs/"+configID.String(), bytes.NewReader(body))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TestDeleteOrgAgentConfig_Success tests successful deletion
func TestDeleteOrgAgentConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	configID := uuid.New()
	orgID := uuid.New()

	// Expect config deletion
	mockDB.EXPECT().
		DeleteOrgAgentConfig(gomock.Any(), configID).
		Return(nil)

	// Use chi router to properly set URL params
	r := chi.NewRouter()
	r.Delete("/organizations/current/agent-configs/{config_id}", handler.DeleteOrgAgentConfig)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/current/agent-configs/"+configID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// TestGetEmployeeResolvedAgentConfigs_Success tests successful config resolution
func TestGetEmployeeResolvedAgentConfigs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	employeeID := uuid.New()
	orgID := uuid.New()
	agent1ID := uuid.New()
	agent2ID := uuid.New()

	employee := db.GetEmployeeRow{
		ID:     employeeID,
		OrgID:  orgID,
		TeamID: pgtype.UUID{Valid: false},
	}

	// Mock for verification
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil).
		Times(4) // Once for verify, once for ResolveEmployeeAgents, twice for ResolveAgentConfig

	// Mock org configs
	orgConfigs := []db.ListOrgAgentConfigsRow{
		{
			ID:                 uuid.New(),
			OrgID:              orgID,
			AgentID:            agent1ID,
			Config:             []byte(`{"model":"claude-3-5-sonnet-20241022"}`),
			IsEnabled:          true,
			AgentName:          "Claude Code",
			AgentType:          "claude-code",
			AgentProvider:      "anthropic",
			AgentDefaultConfig: []byte(`{}`),
		},
		{
			ID:                 uuid.New(),
			OrgID:              orgID,
			AgentID:            agent2ID,
			Config:             []byte(`{"model":"gpt-4o"}`),
			IsEnabled:          true,
			AgentName:          "Cursor",
			AgentType:          "cursor",
			AgentProvider:      "openai",
			AgentDefaultConfig: []byte(`{}`),
		},
	}

	mockDB.EXPECT().
		ListOrgAgentConfigs(gomock.Any(), orgID).
		Return(orgConfigs, nil)

	// Mock agent 1 resolution
	agent1 := db.Agent{
		ID:       agent1ID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}
	orgConfig1 := db.OrgAgentConfig{
		ID:        uuid.New(),
		OrgID:     orgID,
		AgentID:   agent1ID,
		Config:    []byte(`{"model":"claude-3-5-sonnet-20241022"}`),
		IsEnabled: true,
	}

	mockDB.EXPECT().GetAgentByID(gomock.Any(), agent1ID).Return(agent1, nil)
	mockDB.EXPECT().GetOrgAgentConfig(gomock.Any(), db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agent1ID,
	}).Return(orgConfig1, nil)
	mockDB.EXPECT().GetEmployeeAgentConfigByAgent(gomock.Any(), db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agent1ID,
	}).Return(db.GetEmployeeAgentConfigByAgentRow{}, pgx.ErrNoRows)
	mockDB.EXPECT().GetSystemPrompts(gomock.Any(), gomock.Any()).Return([]db.SystemPrompt{}, nil)

	// Mock agent 2 resolution
	agent2 := db.Agent{
		ID:       agent2ID,
		Name:     "Cursor",
		Type:     "cursor",
		Provider: "openai",
	}
	orgConfig2 := db.OrgAgentConfig{
		ID:        uuid.New(),
		OrgID:     orgID,
		AgentID:   agent2ID,
		Config:    []byte(`{"model":"gpt-4o"}`),
		IsEnabled: true,
	}

	mockDB.EXPECT().GetAgentByID(gomock.Any(), agent2ID).Return(agent2, nil)
	mockDB.EXPECT().GetOrgAgentConfig(gomock.Any(), db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agent2ID,
	}).Return(orgConfig2, nil)
	mockDB.EXPECT().GetEmployeeAgentConfigByAgent(gomock.Any(), db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agent2ID,
	}).Return(db.GetEmployeeAgentConfigByAgentRow{}, pgx.ErrNoRows)
	mockDB.EXPECT().GetSystemPrompts(gomock.Any(), gomock.Any()).Return([]db.SystemPrompt{}, nil)

	// Use chi router to properly set URL params
	r := chi.NewRouter()
	r.Get("/employees/{employee_id}/agent-configs/resolved", handler.GetEmployeeResolvedAgentConfigs)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+employeeID.String()+"/agent-configs/resolved", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListResolvedAgentConfigsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Configs))
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, "Claude Code", response.Configs[0].AgentName)
	assert.Equal(t, "Cursor", response.Configs[1].AgentName)
}

// TestGetEmployeeResolvedAgentConfigs_EmployeeNotFound tests employee not found error
func TestGetEmployeeResolvedAgentConfigs_EmployeeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	employeeID := uuid.New()
	orgID := uuid.New()

	// Mock employee verification - returns not found
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(db.GetEmployeeRow{}, pgx.ErrNoRows)

	// Use chi router to properly set URL params
	r := chi.NewRouter()
	r.Get("/employees/{employee_id}/agent-configs/resolved", handler.GetEmployeeResolvedAgentConfigs)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+employeeID.String()+"/agent-configs/resolved", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TestGetMyResolvedAgentConfigs_Success tests successful config resolution using JWT employee_id
func TestGetMyResolvedAgentConfigs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	employeeID := uuid.New()
	orgID := uuid.New()
	agent1ID := uuid.New()
	agent2ID := uuid.New()

	employee := db.GetEmployeeRow{
		ID:     employeeID,
		OrgID:  orgID,
		TeamID: pgtype.UUID{Valid: false},
	}

	// Mock for verification
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil).
		Times(4) // Once for verify, once for ResolveEmployeeAgents, twice for ResolveAgentConfig

	// Mock org configs
	orgConfigs := []db.ListOrgAgentConfigsRow{
		{
			ID:                 uuid.New(),
			OrgID:              orgID,
			AgentID:            agent1ID,
			Config:             []byte(`{"model":"claude-3-5-sonnet-20241022"}`),
			IsEnabled:          true,
			AgentName:          "Claude Code",
			AgentType:          "claude-code",
			AgentProvider:      "anthropic",
			AgentDefaultConfig: []byte(`{}`),
		},
		{
			ID:                 uuid.New(),
			OrgID:              orgID,
			AgentID:            agent2ID,
			Config:             []byte(`{"model":"gpt-4o"}`),
			IsEnabled:          true,
			AgentName:          "Cursor",
			AgentType:          "cursor",
			AgentProvider:      "openai",
			AgentDefaultConfig: []byte(`{}`),
		},
	}

	mockDB.EXPECT().
		ListOrgAgentConfigs(gomock.Any(), orgID).
		Return(orgConfigs, nil)

	// Mock agent 1 resolution
	agent1 := db.Agent{
		ID:       agent1ID,
		Name:     "Claude Code",
		Type:     "claude-code",
		Provider: "anthropic",
	}
	orgConfig1 := db.OrgAgentConfig{
		ID:        uuid.New(),
		OrgID:     orgID,
		AgentID:   agent1ID,
		Config:    []byte(`{"model":"claude-3-5-sonnet-20241022"}`),
		IsEnabled: true,
	}

	mockDB.EXPECT().GetAgentByID(gomock.Any(), agent1ID).Return(agent1, nil)
	mockDB.EXPECT().GetOrgAgentConfig(gomock.Any(), db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agent1ID,
	}).Return(orgConfig1, nil)
	mockDB.EXPECT().GetEmployeeAgentConfigByAgent(gomock.Any(), db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agent1ID,
	}).Return(db.GetEmployeeAgentConfigByAgentRow{}, pgx.ErrNoRows)
	mockDB.EXPECT().GetSystemPrompts(gomock.Any(), gomock.Any()).Return([]db.SystemPrompt{}, nil)

	// Mock agent 2 resolution
	agent2 := db.Agent{
		ID:       agent2ID,
		Name:     "Cursor",
		Type:     "cursor",
		Provider: "openai",
	}
	orgConfig2 := db.OrgAgentConfig{
		ID:        uuid.New(),
		OrgID:     orgID,
		AgentID:   agent2ID,
		Config:    []byte(`{"model":"gpt-4o"}`),
		IsEnabled: true,
	}

	mockDB.EXPECT().GetAgentByID(gomock.Any(), agent2ID).Return(agent2, nil)
	mockDB.EXPECT().GetOrgAgentConfig(gomock.Any(), db.GetOrgAgentConfigParams{
		OrgID:   orgID,
		AgentID: agent2ID,
	}).Return(orgConfig2, nil)
	mockDB.EXPECT().GetEmployeeAgentConfigByAgent(gomock.Any(), db.GetEmployeeAgentConfigByAgentParams{
		EmployeeID: employeeID,
		AgentID:    agent2ID,
	}).Return(db.GetEmployeeAgentConfigByAgentRow{}, pgx.ErrNoRows)
	mockDB.EXPECT().GetSystemPrompts(gomock.Any(), gomock.Any()).Return([]db.SystemPrompt{}, nil)

	// Use chi router to properly set context
	r := chi.NewRouter()
	r.Get("/employees/me/agent-configs/resolved", handler.GetMyResolvedAgentConfigs)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/agent-configs/resolved", nil)
	// Set both org_id and employee_id in context (as JWT middleware would)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListResolvedAgentConfigsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Configs))
	assert.Equal(t, 2, response.Total)
}

// TestGetMyResolvedAgentConfigs_Unauthorized tests when employee_id is not in context
func TestGetMyResolvedAgentConfigs_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrgAgentConfigsHandler(mockDB)

	orgID := uuid.New()

	// Use chi router
	r := chi.NewRouter()
	r.Get("/employees/me/agent-configs/resolved", handler.GetMyResolvedAgentConfigs)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/agent-configs/resolved", nil)
	// Only set org_id, not employee_id (simulating missing JWT claim)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
