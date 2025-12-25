package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/generated/mocks"
	"github.com/rastrigin-systems/arfa/services/api/internal/handlers"
)

// ============================================================================
// GetCurrentOrganization Tests
// ============================================================================

func TestGetCurrentOrganization_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrganizationsHandler(mockDB)

	orgID := uuid.New()
	settings := []byte(`{"features":["sso","audit_logs"]}`)

	org := db.Organization{
		ID:                   orgID,
		Name:                 "Acme Corporation",
		Slug:                 "acme-corp",
		Plan:                 "enterprise",
		Settings:             settings,
		MaxEmployees:         500,
		MaxAgentsPerEmployee: 10,
		CreatedAt:            pgtype.Timestamp{Valid: true},
		UpdatedAt:            pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		GetOrganization(gomock.Any(), orgID).
		Return(org, nil)

	req := httptest.NewRequest(http.MethodGet, "/organizations/current", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetCurrentOrganization(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Organization
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, orgID.String(), response.Id.String())
	assert.Equal(t, "Acme Corporation", response.Name)
	assert.Equal(t, "acme-corp", response.Slug)
	assert.Equal(t, api.OrganizationPlan("enterprise"), response.Plan)
	assert.Equal(t, 500, *response.MaxEmployees)
	assert.Equal(t, 10, *response.MaxAgentsPerEmployee)
}

func TestGetCurrentOrganization_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrganizationsHandler(mockDB)

	// No org_id in context
	req := httptest.NewRequest(http.MethodGet, "/organizations/current", nil)
	rec := httptest.NewRecorder()

	handler.GetCurrentOrganization(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetCurrentOrganization_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrganizationsHandler(mockDB)

	orgID := uuid.New()

	mockDB.EXPECT().
		GetOrganization(gomock.Any(), orgID).
		Return(db.Organization{}, pgx.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/organizations/current", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetCurrentOrganization(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// UpdateCurrentOrganization Tests
// ============================================================================

func TestUpdateCurrentOrganization_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrganizationsHandler(mockDB)

	orgID := uuid.New()
	newName := "Acme Corp Updated"
	newMaxEmployees := int32(1000)
	newMaxAgents := int32(15)
	newSettings := map[string]interface{}{
		"features": []string{"sso", "audit_logs", "saml"},
	}

	reqBody := handlers.UpdateOrganizationRequest{
		Name:                 &newName,
		MaxEmployees:         &newMaxEmployees,
		MaxAgentsPerEmployee: &newMaxAgents,
		Settings:             &newSettings,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	mockDB.EXPECT().
		UpdateOrganization(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx interface{}, params db.UpdateOrganizationParams) (db.Organization, error) {
			assert.Equal(t, orgID, params.ID)
			assert.Equal(t, newName, params.Name)
			assert.Equal(t, newMaxEmployees, params.MaxEmployees)
			assert.Equal(t, newMaxAgents, params.MaxAgentsPerEmployee)
			return db.Organization{
				ID:                   orgID,
				Name:                 params.Name.(string),
				Slug:                 "acme-corp",
				Plan:                 "enterprise",
				Settings:             params.Settings,
				MaxEmployees:         params.MaxEmployees.(int32),
				MaxAgentsPerEmployee: params.MaxAgentsPerEmployee.(int32),
				CreatedAt:            pgtype.Timestamp{Valid: true},
				UpdatedAt:            pgtype.Timestamp{Valid: true},
			}, nil
		})

	req := httptest.NewRequest(http.MethodPatch, "/organizations/current", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.UpdateCurrentOrganization(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Organization
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, newName, response.Name)
	assert.Equal(t, int(newMaxEmployees), *response.MaxEmployees)
	assert.Equal(t, int(newMaxAgents), *response.MaxAgentsPerEmployee)
}

func TestUpdateCurrentOrganization_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrganizationsHandler(mockDB)

	newName := "Updated Name"
	reqBody := handlers.UpdateOrganizationRequest{
		Name: &newName,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// No org_id in context
	req := httptest.NewRequest(http.MethodPatch, "/organizations/current", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.UpdateCurrentOrganization(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUpdateCurrentOrganization_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrganizationsHandler(mockDB)

	orgID := uuid.New()

	req := httptest.NewRequest(http.MethodPatch, "/organizations/current", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.UpdateCurrentOrganization(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateCurrentOrganization_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrganizationsHandler(mockDB)

	orgID := uuid.New()
	newName := "Updated Name"
	reqBody := handlers.UpdateOrganizationRequest{
		Name: &newName,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	mockDB.EXPECT().
		UpdateOrganization(gomock.Any(), gomock.Any()).
		Return(db.Organization{}, pgx.ErrNoRows)

	req := httptest.NewRequest(http.MethodPatch, "/organizations/current", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.UpdateCurrentOrganization(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateCurrentOrganization_PartialUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewOrganizationsHandler(mockDB)

	orgID := uuid.New()
	newName := "Just Update Name"
	reqBody := handlers.UpdateOrganizationRequest{
		Name: &newName,
		// Other fields omitted - should use existing values
	}
	bodyBytes, _ := json.Marshal(reqBody)

	mockDB.EXPECT().
		UpdateOrganization(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx interface{}, params db.UpdateOrganizationParams) (db.Organization, error) {
			assert.Equal(t, orgID, params.ID)
			assert.Equal(t, newName, params.Name)
			return db.Organization{
				ID:                   orgID,
				Name:                 newName,
				Slug:                 "acme-corp",
				Plan:                 "enterprise",
				Settings:             []byte(`{}`),
				MaxEmployees:         500, // unchanged
				MaxAgentsPerEmployee: 10,  // unchanged
				CreatedAt:            pgtype.Timestamp{Valid: true},
				UpdatedAt:            pgtype.Timestamp{Valid: true},
			}, nil
		})

	req := httptest.NewRequest(http.MethodPatch, "/organizations/current", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.UpdateCurrentOrganization(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Organization
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, newName, response.Name)
	assert.Equal(t, 500, *response.MaxEmployees)
	assert.Equal(t, 10, *response.MaxAgentsPerEmployee)
}
