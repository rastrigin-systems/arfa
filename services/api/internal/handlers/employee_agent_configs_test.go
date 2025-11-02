package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
// ListEmployeeAgentConfigs Tests
// ============================================================================

func TestListEmployeeAgentConfigs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()
	agentID1 := uuid.New()
	agentID2 := uuid.New()
	configID1 := uuid.New()
	configID2 := uuid.New()
	syncToken := "sync-token-123"
	lastSyncedAt := time.Now()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		Email:  "user@example.com",
		Status: "active",
	}

	configs := []db.ListEmployeeAgentConfigsRow{
		{
			ID:             configID1,
			EmployeeID:     employeeID,
			AgentID:        agentID1,
			ConfigOverride: []byte(`{"max_tokens":4000}`),
			IsEnabled:      true,
			SyncToken:      &syncToken,
			LastSyncedAt:   pgtype.Timestamp{Time: lastSyncedAt, Valid: true},
			CreatedAt:      pgtype.Timestamp{Valid: true},
			UpdatedAt:      pgtype.Timestamp{Valid: true},
			AgentName:      "Claude Code",
			AgentType:      "code",
			AgentProvider:  "anthropic",
		},
		{
			ID:             configID2,
			EmployeeID:     employeeID,
			AgentID:        agentID2,
			ConfigOverride: []byte(`{"temperature":0.7}`),
			IsEnabled:      false,
			SyncToken:      nil,
			LastSyncedAt:   pgtype.Timestamp{Valid: false},
			CreatedAt:      pgtype.Timestamp{Valid: true},
			UpdatedAt:      pgtype.Timestamp{Valid: true},
			AgentName:      "Cursor",
			AgentType:      "code",
			AgentProvider:  "openai",
		},
	}

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil)

	mockDB.EXPECT().
		ListEmployeeAgentConfigs(gomock.Any(), employeeID).
		Return(configs, nil)

	r := chi.NewRouter()
	r.Get("/employees/{employee_id}/agent-configs", handler.ListEmployeeAgentConfigs)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+employeeID.String()+"/agent-configs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeeAgentConfigsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.Total)
	assert.Len(t, response.Configs, 2)
	assert.Equal(t, "Claude Code", *response.Configs[0].AgentName)
	assert.True(t, response.Configs[0].IsEnabled)
	assert.NotNil(t, response.Configs[0].SyncToken)
	assert.NotNil(t, response.Configs[0].LastSyncedAt)
	assert.Equal(t, "Cursor", *response.Configs[1].AgentName)
	assert.False(t, response.Configs[1].IsEnabled)
	assert.Nil(t, response.Configs[1].SyncToken)
	assert.Nil(t, response.Configs[1].LastSyncedAt)
}

func TestListEmployeeAgentConfigs_EmployeeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(db.Employee{}, pgx.ErrNoRows)

	r := chi.NewRouter()
	r.Get("/employees/{employee_id}/agent-configs", handler.ListEmployeeAgentConfigs)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+employeeID.String()+"/agent-configs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListEmployeeAgentConfigs_WrongOrg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	otherOrgID := uuid.New()
	employeeID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  otherOrgID, // Different org!
		Email:  "user@example.com",
		Status: "active",
	}

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil)

	r := chi.NewRouter()
	r.Get("/employees/{employee_id}/agent-configs", handler.ListEmployeeAgentConfigs)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+employeeID.String()+"/agent-configs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListEmployeeAgentConfigs_InvalidEmployeeID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()

	r := chi.NewRouter()
	r.Get("/employees/{employee_id}/agent-configs", handler.ListEmployeeAgentConfigs)

	req := httptest.NewRequest(http.MethodGet, "/employees/invalid-uuid/agent-configs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// CreateEmployeeAgentConfig Tests
// ============================================================================

func TestCreateEmployeeAgentConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()
	agentID := uuid.New()
	configID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		Email:  "user@example.com",
		Status: "active",
	}

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "code",
		Provider: "anthropic",
	}

	createdConfig := db.CreateEmployeeAgentConfigRow{
		ID:             configID,
		EmployeeID:     employeeID,
		AgentID:        agentID,
		ConfigOverride: []byte(`{"max_tokens":4000}`),
		IsEnabled:      true,
		SyncToken:      nil,
		LastSyncedAt:   pgtype.Timestamp{Valid: false},
		CreatedAt:      pgtype.Timestamp{Valid: true},
		UpdatedAt:      pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil)

	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	mockDB.EXPECT().
		CheckEmployeeAgentExists(gomock.Any(), db.CheckEmployeeAgentExistsParams{
			EmployeeID: employeeID,
			AgentID:    agentID,
		}).
		Return(false, nil)

	mockDB.EXPECT().
		CreateEmployeeAgentConfig(gomock.Any(), db.CreateEmployeeAgentConfigParams{
			EmployeeID:     employeeID,
			AgentID:        agentID,
			ConfigOverride: []byte(`{"max_tokens":4000}`),
			IsEnabled:      true,
		}).
		Return(createdConfig, nil)

	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	reqBody := api.CreateEmployeeAgentConfigRequest{
		AgentId: api.EmployeeId(agentID),
		ConfigOverride: map[string]interface{}{
			"max_tokens": float64(4000),
		},
		IsEnabled: boolPtr(true),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Post("/employees/{employee_id}/agent-configs", handler.CreateEmployeeAgentConfig)

	req := httptest.NewRequest(http.MethodPost, "/employees/"+employeeID.String()+"/agent-configs", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response api.EmployeeAgentConfig
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.Equal(t, "Claude Code", *response.AgentName)
	assert.True(t, response.IsEnabled)
}

func TestCreateEmployeeAgentConfig_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()
	agentID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		Email:  "user@example.com",
		Status: "active",
	}

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "code",
		Provider: "anthropic",
	}

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil)

	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	mockDB.EXPECT().
		CheckEmployeeAgentExists(gomock.Any(), db.CheckEmployeeAgentExistsParams{
			EmployeeID: employeeID,
			AgentID:    agentID,
		}).
		Return(true, nil)

	reqBody := api.CreateEmployeeAgentConfigRequest{
		AgentId: api.EmployeeId(agentID),
		ConfigOverride: map[string]interface{}{
			"max_tokens": float64(4000),
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Post("/employees/{employee_id}/agent-configs", handler.CreateEmployeeAgentConfig)

	req := httptest.NewRequest(http.MethodPost, "/employees/"+employeeID.String()+"/agent-configs", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestCreateEmployeeAgentConfig_MissingConfigOverride(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()
	agentID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		Email:  "user@example.com",
		Status: "active",
	}

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil)

	reqBody := api.CreateEmployeeAgentConfigRequest{
		AgentId: api.EmployeeId(agentID),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Post("/employees/{employee_id}/agent-configs", handler.CreateEmployeeAgentConfig)

	req := httptest.NewRequest(http.MethodPost, "/employees/"+employeeID.String()+"/agent-configs", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// GetEmployeeAgentConfig Tests
// ============================================================================

func TestGetEmployeeAgentConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()
	agentID := uuid.New()
	configID := uuid.New()
	syncToken := "sync-token-123"
	lastSyncedAt := time.Now()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		Email:  "user@example.com",
		Status: "active",
	}

	config := db.GetEmployeeAgentConfigRow{
		ID:             configID,
		EmployeeID:     employeeID,
		AgentID:        agentID,
		ConfigOverride: []byte(`{"max_tokens":4000}`),
		IsEnabled:      true,
		SyncToken:      &syncToken,
		LastSyncedAt:   pgtype.Timestamp{Time: lastSyncedAt, Valid: true},
		CreatedAt:      pgtype.Timestamp{Valid: true},
		UpdatedAt:      pgtype.Timestamp{Valid: true},
		AgentName:      "Claude Code",
		AgentType:      "code",
		AgentProvider:  "anthropic",
	}

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil)

	mockDB.EXPECT().
		GetEmployeeAgentConfig(gomock.Any(), configID).
		Return(config, nil)

	r := chi.NewRouter()
	r.Get("/employees/{employee_id}/agent-configs/{config_id}", handler.GetEmployeeAgentConfig)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+employeeID.String()+"/agent-configs/"+configID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.EmployeeAgentConfig
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.Equal(t, "Claude Code", *response.AgentName)
	assert.True(t, response.IsEnabled)
	assert.NotNil(t, response.SyncToken)
	assert.NotNil(t, response.LastSyncedAt)
}

func TestGetEmployeeAgentConfig_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()
	configID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		Email:  "user@example.com",
		Status: "active",
	}

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil)

	mockDB.EXPECT().
		GetEmployeeAgentConfig(gomock.Any(), configID).
		Return(db.GetEmployeeAgentConfigRow{}, pgx.ErrNoRows)

	r := chi.NewRouter()
	r.Get("/employees/{employee_id}/agent-configs/{config_id}", handler.GetEmployeeAgentConfig)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+employeeID.String()+"/agent-configs/"+configID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetEmployeeAgentConfig_WrongEmployee(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()
	otherEmployeeID := uuid.New()
	agentID := uuid.New()
	configID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		Email:  "user@example.com",
		Status: "active",
	}

	config := db.GetEmployeeAgentConfigRow{
		ID:             configID,
		EmployeeID:     otherEmployeeID, // Different employee!
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
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil)

	mockDB.EXPECT().
		GetEmployeeAgentConfig(gomock.Any(), configID).
		Return(config, nil)

	r := chi.NewRouter()
	r.Get("/employees/{employee_id}/agent-configs/{config_id}", handler.GetEmployeeAgentConfig)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+employeeID.String()+"/agent-configs/"+configID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// UpdateEmployeeAgentConfig Tests
// ============================================================================

func TestUpdateEmployeeAgentConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()
	agentID := uuid.New()
	configID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		Email:  "user@example.com",
		Status: "active",
	}

	agent := db.Agent{
		ID:       agentID,
		Name:     "Claude Code",
		Type:     "code",
		Provider: "anthropic",
	}

	updatedConfig := db.UpdateEmployeeAgentConfigRow{
		ID:             configID,
		EmployeeID:     employeeID,
		AgentID:        agentID,
		ConfigOverride: []byte(`{"max_tokens":8000}`),
		IsEnabled:      false,
		SyncToken:      nil,
		LastSyncedAt:   pgtype.Timestamp{Valid: false},
		CreatedAt:      pgtype.Timestamp{Valid: true},
		UpdatedAt:      pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil)

	mockDB.EXPECT().
		UpdateEmployeeAgentConfig(gomock.Any(), db.UpdateEmployeeAgentConfigParams{
			ID:             configID,
			ConfigOverride: []byte(`{"max_tokens":8000}`),
			IsEnabled:      boolPtr(false),
		}).
		Return(updatedConfig, nil)

	mockDB.EXPECT().
		GetAgentByID(gomock.Any(), agentID).
		Return(agent, nil)

	isEnabled := false
	reqBody := api.UpdateEmployeeAgentConfigRequest{
		ConfigOverride: &map[string]interface{}{
			"max_tokens": float64(8000),
		},
		IsEnabled: &isEnabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Patch("/employees/{employee_id}/agent-configs/{config_id}", handler.UpdateEmployeeAgentConfig)

	req := httptest.NewRequest(http.MethodPatch, "/employees/"+employeeID.String()+"/agent-configs/"+configID.String(), strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.EmployeeAgentConfig
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.IsEnabled)
}

func TestUpdateEmployeeAgentConfig_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()
	configID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		Email:  "user@example.com",
		Status: "active",
	}

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil)

	mockDB.EXPECT().
		UpdateEmployeeAgentConfig(gomock.Any(), gomock.Any()).
		Return(db.UpdateEmployeeAgentConfigRow{}, pgx.ErrNoRows)

	isEnabled := false
	reqBody := api.UpdateEmployeeAgentConfigRequest{
		IsEnabled: &isEnabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	r := chi.NewRouter()
	r.Patch("/employees/{employee_id}/agent-configs/{config_id}", handler.UpdateEmployeeAgentConfig)

	req := httptest.NewRequest(http.MethodPatch, "/employees/"+employeeID.String()+"/agent-configs/"+configID.String(), strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// DeleteEmployeeAgentConfig Tests
// ============================================================================

func TestDeleteEmployeeAgentConfig_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()
	configID := uuid.New()

	employee := db.Employee{
		ID:     employeeID,
		OrgID:  orgID,
		Email:  "user@example.com",
		Status: "active",
	}

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), employeeID).
		Return(employee, nil)

	mockDB.EXPECT().
		DeleteEmployeeAgentConfig(gomock.Any(), configID).
		Return(nil)

	r := chi.NewRouter()
	r.Delete("/employees/{employee_id}/agent-configs/{config_id}", handler.DeleteEmployeeAgentConfig)

	req := httptest.NewRequest(http.MethodDelete, "/employees/"+employeeID.String()+"/agent-configs/"+configID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteEmployeeAgentConfig_InvalidConfigID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewEmployeeAgentConfigsHandler(mockDB)

	orgID := uuid.New()
	employeeID := uuid.New()

	r := chi.NewRouter()
	r.Delete("/employees/{employee_id}/agent-configs/{config_id}", handler.DeleteEmployeeAgentConfig)

	req := httptest.NewRequest(http.MethodDelete, "/employees/"+employeeID.String()+"/agent-configs/invalid-uuid", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
