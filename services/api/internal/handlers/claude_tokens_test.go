package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/generated/mocks"
	"github.com/rastrigin-systems/arfa/services/api/internal/handlers"
)

// ============================================================================
// SetOrganizationClaudeToken Tests
// ============================================================================

func TestSetOrganizationClaudeToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	orgID := uuid.New()
	token := "test-claude-token-abc123"

	mockDB.EXPECT().
		SetOrganizationClaudeToken(gomock.Any(), db.SetOrganizationClaudeTokenParams{
			ID:             orgID,
			ClaudeApiToken: &token,
		}).
		Return(nil)

	requestBody := api.SetClaudeTokenRequest{
		Token: token,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPut, "/organizations/current/claude-token", bytes.NewBuffer(body))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetOrganizationClaudeToken(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ClaudeTokenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "updated successfully")
}

func TestSetOrganizationClaudeToken_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	requestBody := api.SetClaudeTokenRequest{
		Token: "test-token",
	}
	body, _ := json.Marshal(requestBody)

	// No org_id in context
	req := httptest.NewRequest(http.MethodPut, "/organizations/current/claude-token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetOrganizationClaudeToken(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestSetOrganizationClaudeToken_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	orgID := uuid.New()

	req := httptest.NewRequest(http.MethodPut, "/organizations/current/claude-token", bytes.NewBufferString("invalid json"))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetOrganizationClaudeToken(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSetOrganizationClaudeToken_EmptyToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	orgID := uuid.New()

	requestBody := api.SetClaudeTokenRequest{
		Token: "",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPut, "/organizations/current/claude-token", bytes.NewBuffer(body))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetOrganizationClaudeToken(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestSetOrganizationClaudeToken_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	orgID := uuid.New()
	token := "test-token"

	mockDB.EXPECT().
		SetOrganizationClaudeToken(gomock.Any(), gomock.Any()).
		Return(pgx.ErrTxClosed)

	requestBody := api.SetClaudeTokenRequest{
		Token: token,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPut, "/organizations/current/claude-token", bytes.NewBuffer(body))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetOrganizationClaudeToken(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ============================================================================
// DeleteOrganizationClaudeToken Tests
// ============================================================================

func TestDeleteOrganizationClaudeToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	orgID := uuid.New()

	mockDB.EXPECT().
		DeleteOrganizationClaudeToken(gomock.Any(), orgID).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/current/claude-token", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.DeleteOrganizationClaudeToken(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ClaudeTokenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "deleted successfully")
}

func TestDeleteOrganizationClaudeToken_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	// No org_id in context
	req := httptest.NewRequest(http.MethodDelete, "/organizations/current/claude-token", nil)
	rec := httptest.NewRecorder()

	handler.DeleteOrganizationClaudeToken(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestDeleteOrganizationClaudeToken_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	orgID := uuid.New()

	mockDB.EXPECT().
		DeleteOrganizationClaudeToken(gomock.Any(), orgID).
		Return(pgx.ErrTxClosed)

	req := httptest.NewRequest(http.MethodDelete, "/organizations/current/claude-token", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.DeleteOrganizationClaudeToken(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ============================================================================
// SetEmployeeClaudeToken Tests
// ============================================================================

func TestSetEmployeeClaudeToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	employeeID := uuid.New()
	token := "employee-claude-token-xyz789"

	mockDB.EXPECT().
		SetEmployeePersonalToken(gomock.Any(), db.SetEmployeePersonalTokenParams{
			ID:                  employeeID,
			PersonalClaudeToken: &token,
		}).
		Return(nil)

	requestBody := api.SetClaudeTokenRequest{
		Token: token,
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPut, "/employees/me/claude-token", bytes.NewBuffer(body))
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetEmployeeClaudeToken(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ClaudeTokenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "updated successfully")
}

func TestSetEmployeeClaudeToken_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	requestBody := api.SetClaudeTokenRequest{
		Token: "test-token",
	}
	body, _ := json.Marshal(requestBody)

	// No employee_id in context
	req := httptest.NewRequest(http.MethodPut, "/employees/me/claude-token", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.SetEmployeeClaudeToken(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ============================================================================
// DeleteEmployeeClaudeToken Tests
// ============================================================================

func TestDeleteEmployeeClaudeToken_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	employeeID := uuid.New()

	mockDB.EXPECT().
		DeleteEmployeePersonalToken(gomock.Any(), employeeID).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/employees/me/claude-token", nil)
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	rec := httptest.NewRecorder()

	handler.DeleteEmployeeClaudeToken(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ClaudeTokenResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Contains(t, response.Message, "deleted successfully")
}

func TestDeleteEmployeeClaudeToken_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	// No employee_id in context
	req := httptest.NewRequest(http.MethodDelete, "/employees/me/claude-token", nil)
	rec := httptest.NewRecorder()

	handler.DeleteEmployeeClaudeToken(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ============================================================================
// GetClaudeTokenStatus Tests
// ============================================================================

func TestGetClaudeTokenStatus_BothTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	employeeID := uuid.New()

	mockDB.EXPECT().
		GetEmployeeTokenStatus(gomock.Any(), employeeID).
		Return(db.GetEmployeeTokenStatusRow{
			EmployeeID:        employeeID,
			FullName:          "Alice Smith",
			HasPersonalToken:  true,
			HasCompanyToken:   true,
			ActiveTokenSource: "personal",
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/claude-token/status", nil)
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	rec := httptest.NewRecorder()

	handler.GetClaudeTokenStatus(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ClaudeTokenStatusResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, employeeID, *response.EmployeeId)
	assert.True(t, response.HasPersonalToken)
	assert.True(t, response.HasCompanyToken)
	assert.Equal(t, api.ClaudeTokenStatusResponseActiveTokenSourcePersonal, response.ActiveTokenSource)
}

func TestGetClaudeTokenStatus_CompanyTokenOnly(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	employeeID := uuid.New()

	mockDB.EXPECT().
		GetEmployeeTokenStatus(gomock.Any(), employeeID).
		Return(db.GetEmployeeTokenStatusRow{
			EmployeeID:        employeeID,
			FullName:          "Bob Jones",
			HasPersonalToken:  false,
			HasCompanyToken:   true,
			ActiveTokenSource: "company",
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/claude-token/status", nil)
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	rec := httptest.NewRecorder()

	handler.GetClaudeTokenStatus(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ClaudeTokenStatusResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.HasPersonalToken)
	assert.True(t, response.HasCompanyToken)
	assert.Equal(t, api.ClaudeTokenStatusResponseActiveTokenSourceCompany, response.ActiveTokenSource)
}

func TestGetClaudeTokenStatus_NoTokens(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	employeeID := uuid.New()

	mockDB.EXPECT().
		GetEmployeeTokenStatus(gomock.Any(), employeeID).
		Return(db.GetEmployeeTokenStatusRow{
			EmployeeID:        employeeID,
			FullName:          "Charlie Brown",
			HasPersonalToken:  false,
			HasCompanyToken:   false,
			ActiveTokenSource: "none",
		}, nil)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/claude-token/status", nil)
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	rec := httptest.NewRecorder()

	handler.GetClaudeTokenStatus(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ClaudeTokenStatusResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.HasPersonalToken)
	assert.False(t, response.HasCompanyToken)
	assert.Equal(t, api.ClaudeTokenStatusResponseActiveTokenSourceNone, response.ActiveTokenSource)
}

func TestGetClaudeTokenStatus_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	// No employee_id in context
	req := httptest.NewRequest(http.MethodGet, "/employees/me/claude-token/status", nil)
	rec := httptest.NewRecorder()

	handler.GetClaudeTokenStatus(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetClaudeTokenStatus_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewClaudeTokensHandler(mockDB)

	employeeID := uuid.New()

	mockDB.EXPECT().
		GetEmployeeTokenStatus(gomock.Any(), employeeID).
		Return(db.GetEmployeeTokenStatusRow{}, pgx.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/claude-token/status", nil)
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	rec := httptest.NewRecorder()

	handler.GetClaudeTokenStatus(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
