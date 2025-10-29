package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
)

// ============================================================================
// ListTeamAgentConfigs Tests
// ============================================================================

func TestListTeamAgentConfigs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	agentID1 := uuid.New()
	agentID2 := uuid.New()
	configID1 := uuid.New()
	configID2 := uuid.New()

	team := db.Team{
		ID:    teamID,
		OrgID: orgID,
		Name:  "Engineering",
	}

	configs := []db.ListTeamAgentConfigsRow{
		{
			ID:             configID1,
			TeamID:         teamID,
			AgentID:        agentID1,
			ConfigOverride: []byte(`{"max_tokens":4000}`),
			IsEnabled:      true,
			CreatedAt:      pgtype.Timestamp{Valid: true},
			UpdatedAt:      pgtype.Timestamp{Valid: true},
			AgentName:      "Claude Code",
			AgentType:      "code",
			AgentProvider:  "anthropic",
		},
		{
			ID:             configID2,
			TeamID:         teamID,
			AgentID:        agentID2,
			ConfigOverride: []byte(`{"temperature":0.7}`),
			IsEnabled:      false,
			CreatedAt:      pgtype.Timestamp{Valid: true},
			UpdatedAt:      pgtype.Timestamp{Valid: true},
			AgentName:      "Cursor",
			AgentType:      "code",
			AgentProvider:  "openai",
		},
	}

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(team, nil)

	mockDB.EXPECT().
		ListTeamAgentConfigs(gomock.Any(), teamID).
		Return(configs, nil)

	r := chi.NewRouter()
	r.Get("/teams/{team_id}/agent-configs", handler.ListTeamAgentConfigs)

	req := httptest.NewRequest(http.MethodGet, "/teams/"+teamID.String()+"/agent-configs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListTeamAgentConfigsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Configs, 2)
	assert.Equal(t, "Claude Code", *response.Configs[0].AgentName)
	assert.True(t, response.Configs[0].IsEnabled)
	assert.Equal(t, "Cursor", *response.Configs[1].AgentName)
	assert.False(t, response.Configs[1].IsEnabled)
}

func TestListTeamAgentConfigs_TeamNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(db.Team{}, pgx.ErrNoRows)

	r := chi.NewRouter()
	r.Get("/teams/{team_id}/agent-configs", handler.ListTeamAgentConfigs)

	req := httptest.NewRequest(http.MethodGet, "/teams/"+teamID.String()+"/agent-configs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListTeamAgentConfigs_InvalidTeamID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()

	r := chi.NewRouter()
	r.Get("/teams/{team_id}/agent-configs", handler.ListTeamAgentConfigs)

	req := httptest.NewRequest(http.MethodGet, "/teams/invalid-uuid/agent-configs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// CreateTeamAgentConfig Tests
// ============================================================================

func TestCreateTeamAgentConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	agentID := uuid.New()
	configID := uuid.New()

	team := db.Team{
		ID:    teamID,
		OrgID: orgID,
		Name:  "Engineering",
	}

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "code",
		Provider: "anthropic",
	}

	createdConfig := db.TeamAgentConfig{
		ID:             configID,
		TeamID:         teamID,
		AgentID:        agentID,
		ConfigOverride: []byte(`{"max_tokens":4000}`),
		IsEnabled:      true,
		CreatedAt:      pgtype.Timestamp{Valid: true},
		UpdatedAt:      pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(team, nil)

	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	mockDB.EXPECT().
		CheckTeamAgentConfigExists(gomock.Any(), db.CheckTeamAgentConfigExistsParams{
			TeamID:  teamID,
			AgentID: agentID,
		}).
		Return(false, nil)

	mockDB.EXPECT().
		CreateTeamAgentConfig(gomock.Any(), db.CreateTeamAgentConfigParams{
			TeamID:         teamID,
			AgentID:        agentID,
			ConfigOverride: []byte(`{"max_tokens":4000}`),
			IsEnabled:      true,
		}).
		Return(createdConfig, nil)

	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	reqBody := api.CreateTeamAgentConfigRequest{
		AgentId: api.TeamId(agentID),
		ConfigOverride: map[string]interface{}{
			"max_tokens": float64(4000),
		},
		IsEnabled: boolPtr(true),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Post("/teams/{team_id}/agent-configs", handler.CreateTeamAgentConfig)

	req := httptest.NewRequest(http.MethodPost, "/teams/"+teamID.String()+"/agent-configs", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response api.TeamAgentConfig
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.Equal(t, "Claude Code", *response.AgentName)
	assert.True(t, response.IsEnabled)
}

func TestCreateTeamAgentConfig_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	agentID := uuid.New()

	team := db.Team{
		ID:    teamID,
		OrgID: orgID,
		Name:  "Engineering",
	}

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "code",
		Provider: "anthropic",
	}

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(team, nil)

	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	mockDB.EXPECT().
		CheckTeamAgentConfigExists(gomock.Any(), db.CheckTeamAgentConfigExistsParams{
			TeamID:  teamID,
			AgentID: agentID,
		}).
		Return(true, nil)

	reqBody := api.CreateTeamAgentConfigRequest{
		AgentId: api.TeamId(agentID),
		ConfigOverride: map[string]interface{}{
			"max_tokens": float64(4000),
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Post("/teams/{team_id}/agent-configs", handler.CreateTeamAgentConfig)

	req := httptest.NewRequest(http.MethodPost, "/teams/"+teamID.String()+"/agent-configs", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestCreateTeamAgentConfig_MissingConfigOverride(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	agentID := uuid.New()

	team := db.Team{
		ID:    teamID,
		OrgID: orgID,
		Name:  "Engineering",
	}

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(team, nil)

	reqBody := api.CreateTeamAgentConfigRequest{
		AgentId: api.TeamId(agentID),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Post("/teams/{team_id}/agent-configs", handler.CreateTeamAgentConfig)

	req := httptest.NewRequest(http.MethodPost, "/teams/"+teamID.String()+"/agent-configs", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// GetTeamAgentConfig Tests
// ============================================================================

func TestGetTeamAgentConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	agentID := uuid.New()
	configID := uuid.New()

	team := db.Team{
		ID:    teamID,
		OrgID: orgID,
		Name:  "Engineering",
	}

	config := db.GetTeamAgentConfigByIDRow{
		ID:             configID,
		TeamID:         teamID,
		AgentID:        agentID,
		ConfigOverride: []byte(`{"max_tokens":4000}`),
		IsEnabled:      true,
		CreatedAt:      pgtype.Timestamp{Valid: true},
		UpdatedAt:      pgtype.Timestamp{Valid: true},
		AgentName:      "Claude Code",
		AgentType:      "code",
		AgentProvider:  "anthropic",
	}

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(team, nil)

	mockDB.EXPECT().
		GetTeamAgentConfigByID(gomock.Any(), configID).
		Return(config, nil)

	r := chi.NewRouter()
	r.Get("/teams/{team_id}/agent-configs/{config_id}", handler.GetTeamAgentConfig)

	req := httptest.NewRequest(http.MethodGet, "/teams/"+teamID.String()+"/agent-configs/"+configID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.TeamAgentConfig
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.Equal(t, "Claude Code", *response.AgentName)
	assert.True(t, response.IsEnabled)
}

func TestGetTeamAgentConfig_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	configID := uuid.New()

	team := db.Team{
		ID:    teamID,
		OrgID: orgID,
		Name:  "Engineering",
	}

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(team, nil)

	mockDB.EXPECT().
		GetTeamAgentConfigByID(gomock.Any(), configID).
		Return(db.GetTeamAgentConfigByIDRow{}, pgx.ErrNoRows)

	r := chi.NewRouter()
	r.Get("/teams/{team_id}/agent-configs/{config_id}", handler.GetTeamAgentConfig)

	req := httptest.NewRequest(http.MethodGet, "/teams/"+teamID.String()+"/agent-configs/"+configID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetTeamAgentConfig_WrongTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	otherTeamID := uuid.New()
	agentID := uuid.New()
	configID := uuid.New()

	team := db.Team{
		ID:    teamID,
		OrgID: orgID,
		Name:  "Engineering",
	}

	config := db.GetTeamAgentConfigByIDRow{
		ID:             configID,
		TeamID:         otherTeamID, // Different team!
		AgentID:        agentID,
		ConfigOverride: []byte(`{"max_tokens":4000}`),
		IsEnabled:      true,
		CreatedAt:      pgtype.Timestamp{Valid: true},
		UpdatedAt:      pgtype.Timestamp{Valid: true},
		AgentName:      "Claude Code",
		AgentType:      "code",
		AgentProvider:  "anthropic",
	}

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(team, nil)

	mockDB.EXPECT().
		GetTeamAgentConfigByID(gomock.Any(), configID).
		Return(config, nil)

	r := chi.NewRouter()
	r.Get("/teams/{team_id}/agent-configs/{config_id}", handler.GetTeamAgentConfig)

	req := httptest.NewRequest(http.MethodGet, "/teams/"+teamID.String()+"/agent-configs/"+configID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// UpdateTeamAgentConfig Tests
// ============================================================================

func TestUpdateTeamAgentConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	agentID := uuid.New()
	configID := uuid.New()

	team := db.Team{
		ID:    teamID,
		OrgID: orgID,
		Name:  "Engineering",
	}

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "code",
		Provider: "anthropic",
	}

	updatedConfig := db.TeamAgentConfig{
		ID:             configID,
		TeamID:         teamID,
		AgentID:        agentID,
		ConfigOverride: []byte(`{"max_tokens":8000}`),
		IsEnabled:      false,
		CreatedAt:      pgtype.Timestamp{Valid: true},
		UpdatedAt:      pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(team, nil)

	mockDB.EXPECT().
		UpdateTeamAgentConfig(gomock.Any(), db.UpdateTeamAgentConfigParams{
			ID:             configID,
			ConfigOverride: []byte(`{"max_tokens":8000}`),
			IsEnabled:      boolPtr(false),
		}).
		Return(updatedConfig, nil)

	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	isEnabled := false
	reqBody := api.UpdateTeamAgentConfigRequest{
		ConfigOverride: &map[string]interface{}{
			"max_tokens": float64(8000),
		},
		IsEnabled: &isEnabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Patch("/teams/{team_id}/agent-configs/{config_id}", handler.UpdateTeamAgentConfig)

	req := httptest.NewRequest(http.MethodPatch, "/teams/"+teamID.String()+"/agent-configs/"+configID.String(), strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.TeamAgentConfig
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.IsEnabled)
}

func TestUpdateTeamAgentConfig_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	configID := uuid.New()

	team := db.Team{
		ID:    teamID,
		OrgID: orgID,
		Name:  "Engineering",
	}

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(team, nil)

	mockDB.EXPECT().
		UpdateTeamAgentConfig(gomock.Any(), gomock.Any()).
		Return(db.TeamAgentConfig{}, pgx.ErrNoRows)

	isEnabled := false
	reqBody := api.UpdateTeamAgentConfigRequest{
		IsEnabled: &isEnabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Patch("/teams/{team_id}/agent-configs/{config_id}", handler.UpdateTeamAgentConfig)

	req := httptest.NewRequest(http.MethodPatch, "/teams/"+teamID.String()+"/agent-configs/"+configID.String(), strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// DeleteTeamAgentConfig Tests
// ============================================================================

func TestDeleteTeamAgentConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()
	configID := uuid.New()

	team := db.Team{
		ID:    teamID,
		OrgID: orgID,
		Name:  "Engineering",
	}

	mockDB.EXPECT().
		GetTeam(gomock.Any(), db.GetTeamParams{
			ID:    teamID,
			OrgID: orgID,
		}).
		Return(team, nil)

	mockDB.EXPECT().
		DeleteTeamAgentConfig(gomock.Any(), configID).
		Return(nil)

	r := chi.NewRouter()
	r.Delete("/teams/{team_id}/agent-configs/{config_id}", handler.DeleteTeamAgentConfig)

	req := httptest.NewRequest(http.MethodDelete, "/teams/"+teamID.String()+"/agent-configs/"+configID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteTeamAgentConfig_InvalidConfigID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewTeamAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	teamID := uuid.New()

	r := chi.NewRouter()
	r.Delete("/teams/{team_id}/agent-configs/{config_id}", handler.DeleteTeamAgentConfig)

	req := httptest.NewRequest(http.MethodDelete, "/teams/"+teamID.String()+"/agent-configs/invalid-uuid", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}
