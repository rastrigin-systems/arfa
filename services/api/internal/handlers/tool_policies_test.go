package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/generated/mocks"
	"github.com/rastrigin-systems/arfa/services/api/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetEmployeeToolPolicies_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	employeeID := uuid.New()
	orgID := uuid.New()
	teamID := uuid.New()
	policyID := uuid.New()

	// Create session data
	sessionData := &db.GetSessionWithEmployeeRow{
		EmployeeID: employeeID,
		OrgID:      orgID,
		TeamID:     pgtype.UUID{Bytes: teamID, Valid: true},
	}

	// Mock policy data
	reason := "Shell commands are blocked"
	policies := []db.ToolPolicy{
		{
			ID:         policyID,
			OrgID:      orgID,
			TeamID:     pgtype.UUID{Valid: false},
			EmployeeID: pgtype.UUID{Valid: false},
			ToolName:   "Bash",
			Action:     "deny",
			Reason:     &reason,
			Conditions: nil,
		},
	}

	// Expect GetToolPoliciesForEmployee to be called
	mockDB.EXPECT().
		GetToolPoliciesForEmployee(gomock.Any(), db.GetToolPoliciesForEmployeeParams{
			OrgID:      orgID,
			TeamID:     sessionData.TeamID,
			EmployeeID: pgtype.UUID{Bytes: employeeID, Valid: true},
		}).
		Return(policies, nil)

	handler := handlers.NewToolPoliciesHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/tool-policies", nil)

	// Set up context with auth data
	ctx := req.Context()
	ctx = handlers.SetEmployeeIDInContext(ctx, employeeID)
	ctx = handlers.SetOrgIDInContext(ctx, orgID)
	ctx = handlers.SetSessionDataInContext(ctx, sessionData)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.GetEmployeeToolPolicies(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.EmployeeToolPoliciesResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify policies
	require.Len(t, response.Policies, 1)
	assert.Equal(t, "Bash", response.Policies[0].ToolName)
	assert.Equal(t, api.Deny, response.Policies[0].Action)
	assert.Equal(t, "Shell commands are blocked", *response.Policies[0].Reason)
	assert.Equal(t, api.ToolPolicyScopeOrganization, response.Policies[0].Scope)

	// Verify version and synced_at
	assert.Greater(t, response.Version, 0)
	assert.False(t, response.SyncedAt.IsZero())
}

func TestGetEmployeeToolPolicies_EmptyList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	employeeID := uuid.New()
	orgID := uuid.New()

	sessionData := &db.GetSessionWithEmployeeRow{
		EmployeeID: employeeID,
		OrgID:      orgID,
		TeamID:     pgtype.UUID{Valid: false}, // No team
	}

	// Expect empty result
	mockDB.EXPECT().
		GetToolPoliciesForEmployee(gomock.Any(), gomock.Any()).
		Return([]db.ToolPolicy{}, nil)

	handler := handlers.NewToolPoliciesHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/tool-policies", nil)
	ctx := req.Context()
	ctx = handlers.SetEmployeeIDInContext(ctx, employeeID)
	ctx = handlers.SetOrgIDInContext(ctx, orgID)
	ctx = handlers.SetSessionDataInContext(ctx, sessionData)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.GetEmployeeToolPolicies(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.EmployeeToolPoliciesResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Empty(t, response.Policies)
}

func TestGetEmployeeToolPolicies_MultipleScopeLevels(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	employeeID := uuid.New()
	orgID := uuid.New()
	teamID := uuid.New()

	sessionData := &db.GetSessionWithEmployeeRow{
		EmployeeID: employeeID,
		OrgID:      orgID,
		TeamID:     pgtype.UUID{Bytes: teamID, Valid: true},
	}

	orgReason := "Org-wide block"
	teamReason := "Team-specific block"
	empReason := "Employee-specific block"

	// Policies at different scope levels
	policies := []db.ToolPolicy{
		{
			ID:         uuid.New(),
			OrgID:      orgID,
			ToolName:   "Bash",
			Action:     "deny",
			Reason:     &orgReason,
			TeamID:     pgtype.UUID{Valid: false},
			EmployeeID: pgtype.UUID{Valid: false},
		},
		{
			ID:         uuid.New(),
			OrgID:      orgID,
			TeamID:     pgtype.UUID{Bytes: teamID, Valid: true},
			EmployeeID: pgtype.UUID{Valid: false},
			ToolName:   "Write",
			Action:     "deny",
			Reason:     &teamReason,
		},
		{
			ID:         uuid.New(),
			OrgID:      orgID,
			TeamID:     pgtype.UUID{Valid: false},
			EmployeeID: pgtype.UUID{Bytes: employeeID, Valid: true},
			ToolName:   "mcp__gcloud__%",
			Action:     "deny",
			Reason:     &empReason,
		},
	}

	mockDB.EXPECT().
		GetToolPoliciesForEmployee(gomock.Any(), gomock.Any()).
		Return(policies, nil)

	handler := handlers.NewToolPoliciesHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/tool-policies", nil)
	ctx := req.Context()
	ctx = handlers.SetEmployeeIDInContext(ctx, employeeID)
	ctx = handlers.SetOrgIDInContext(ctx, orgID)
	ctx = handlers.SetSessionDataInContext(ctx, sessionData)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	handler.GetEmployeeToolPolicies(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.EmployeeToolPoliciesResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	require.Len(t, response.Policies, 3)

	// Check scopes are correctly identified
	scopes := make(map[string]api.ToolPolicyScope)
	for _, p := range response.Policies {
		scopes[p.ToolName] = p.Scope
	}

	assert.Equal(t, api.ToolPolicyScopeOrganization, scopes["Bash"])
	assert.Equal(t, api.ToolPolicyScopeTeam, scopes["Write"])
	assert.Equal(t, api.ToolPolicyScopeEmployee, scopes["mcp__gcloud__%"])
}

func TestGetEmployeeToolPolicies_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	handler := handlers.NewToolPoliciesHandler(mockDB)

	// Request without auth context
	req := httptest.NewRequest(http.MethodGet, "/employees/me/tool-policies", nil)
	rec := httptest.NewRecorder()

	handler.GetEmployeeToolPolicies(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
