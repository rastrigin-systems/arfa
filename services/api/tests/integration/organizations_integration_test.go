package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/services/api/internal/auth"
	"github.com/rastrigin-systems/arfa/services/api/internal/handlers"
	"github.com/rastrigin-systems/arfa/services/api/internal/middleware"
	"github.com/rastrigin-systems/arfa/services/api/tests/testutil"
)

// ============================================================================
// GetCurrentOrganization Integration Tests
// ============================================================================

// TDD Lesson: Integration test with real database
// Tests complete flow: authentication, org retrieval
func TestGetCurrentOrganization_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer func() { _ = conn.Close(testutil.GetContext(t)) }()
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@testcorp.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewOrganizationsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/organizations/current", handler.GetCurrentOrganization)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/organizations/current", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Organization
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, org.ID.String(), response.Id.String())
	assert.Equal(t, org.Name, response.Name)
	assert.Equal(t, org.Slug, response.Slug)
	assert.NotNil(t, response.MaxEmployees)
}

// TDD Lesson: Test authentication is required
func TestGetCurrentOrganization_Integration_Unauthorized(t *testing.T) {
	_, queries := testutil.SetupTestDB(t)

	handler := handlers.NewOrganizationsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/organizations/current", handler.GetCurrentOrganization)

	// Make request without token
	req := httptest.NewRequest(http.MethodGet, "/organizations/current", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ============================================================================
// UpdateCurrentOrganization Integration Tests
// ============================================================================

// TDD Lesson: Integration test for full update (all fields)
func TestUpdateCurrentOrganization_Integration_FullUpdate(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer func() { _ = conn.Close(testutil.GetContext(t)) }()
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@testcorp.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewOrganizationsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Patch("/organizations/current", handler.UpdateCurrentOrganization)

	// Prepare update request
	updateData := map[string]interface{}{
		"name":                    "Updated Corporation",
		"max_employees":           1000,
		"max_agents_per_employee": 20,
		"settings": map[string]interface{}{
			"features": []string{"sso", "audit_logs", "saml"},
			"theme":    "dark",
		},
	}
	bodyBytes, _ := json.Marshal(updateData)

	// Make request
	req := httptest.NewRequest(http.MethodPatch, "/organizations/current", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Organization
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Updated Corporation", response.Name)
	assert.Equal(t, 1000, *response.MaxEmployees)
	assert.NotNil(t, response.Settings)

	settings := *response.Settings
	assert.Equal(t, "dark", settings["theme"])
	features := settings["features"].([]interface{})
	assert.Contains(t, features, "sso")
	assert.Contains(t, features, "audit_logs")
	assert.Contains(t, features, "saml")
}

// TDD Lesson: Integration test for partial update (only some fields)
// This is CRITICAL - tests that NULLIF is working correctly in production
func TestUpdateCurrentOrganization_Integration_PartialUpdate(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer func() { _ = conn.Close(testutil.GetContext(t)) }()
	ctx := testutil.GetContext(t)

	// Create test organization with specific initial values
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Update org with known values first
	_, err := queries.UpdateOrganization(ctx, testutil.CreateUpdateOrgParams(
		org.ID,
		"Initial Corp Name",
		500,
		[]byte(`{"features":["sso","audit_logs"]}`),
	))
	require.NoError(t, err)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@testcorp.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewOrganizationsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Patch("/organizations/current", handler.UpdateCurrentOrganization)

	// Prepare partial update request (only name and max_employees)
	updateData := map[string]interface{}{
		"name":          "Partially Updated Corp",
		"max_employees": 1000,
		// NOT providing max_agents_per_employee or settings
	}
	bodyBytes, _ := json.Marshal(updateData)

	// Make request
	req := httptest.NewRequest(http.MethodPatch, "/organizations/current", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Organization
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Provided fields should be updated
	assert.Equal(t, "Partially Updated Corp", response.Name)
	assert.Equal(t, 1000, *response.MaxEmployees)

	// Unprovided fields should RETAIN their original values (NOT become 0 or null)
	assert.NotNil(t, response.Settings)
	settings := *response.Settings
	features := settings["features"].([]interface{})
	assert.Contains(t, features, "sso", "settings should retain original value")
	assert.Contains(t, features, "audit_logs", "settings should retain original value")
}

// TDD Lesson: Integration test for settings-only update
func TestUpdateCurrentOrganization_Integration_SettingsOnly(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer func() { _ = conn.Close(testutil.GetContext(t)) }()
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Update org with known values first
	_, err := queries.UpdateOrganization(ctx, testutil.CreateUpdateOrgParams(
		org.ID,
		"Test Corporation",
		500,
		[]byte(`{"features":["sso"]}`),
	))
	require.NoError(t, err)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@testcorp.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewOrganizationsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Patch("/organizations/current", handler.UpdateCurrentOrganization)

	// Prepare settings-only update
	updateData := map[string]interface{}{
		"settings": map[string]interface{}{
			"features": []string{"sso", "saml", "mfa"},
			"theme":    "dark",
		},
	}
	bodyBytes, _ := json.Marshal(updateData)

	// Make request
	req := httptest.NewRequest(http.MethodPatch, "/organizations/current", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Assertions
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Organization
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Name and numeric fields should retain original values
	assert.Equal(t, "Test Corporation", response.Name)
	assert.Equal(t, 500, *response.MaxEmployees)

	// Settings should be updated
	assert.NotNil(t, response.Settings)
	settings := *response.Settings
	assert.Equal(t, "dark", settings["theme"])
	features := settings["features"].([]interface{})
	assert.Contains(t, features, "sso")
	assert.Contains(t, features, "saml")
	assert.Contains(t, features, "mfa")
}

// TDD Lesson: Integration test for multi-tenancy isolation
// Ensures employees can only see/update their own organization
func TestOrganizations_Integration_MultiTenancyIsolation(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer func() { _ = conn.Close(testutil.GetContext(t)) }()
	ctx := testutil.GetContext(t)

	// Create two different organizations
	org1 := testutil.CreateTestOrg(t, queries, ctx)
	org2 := testutil.CreateTestOrg(t, queries, ctx)

	// Create roles
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	// Create employee in org1
	employee1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org1.ID,
		RoleID:   role.ID,
		Email:    "admin1@org1.com",
		FullName: "Admin Org1",
		Status:   "active",
	})

	// Create session for employee1
	token1, _ := auth.GenerateJWT(employee1.ID, org1.ID, 24*time.Hour)
	tokenHash1 := auth.HashToken(token1)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee1.ID, tokenHash1))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewOrganizationsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/organizations/current", handler.GetCurrentOrganization)

	// Employee from org1 should only see org1
	req := httptest.NewRequest(http.MethodGet, "/organizations/current", nil)
	req.Header.Set("Authorization", "Bearer "+token1)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Organization
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should see org1, NOT org2
	assert.Equal(t, org1.ID.String(), response.Id.String())
	assert.NotEqual(t, org2.ID.String(), response.Id.String())
}

// TDD Lesson: Integration test for invalid JSON
func TestUpdateCurrentOrganization_Integration_InvalidJSON(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer func() { _ = conn.Close(testutil.GetContext(t)) }()
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@testcorp.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewOrganizationsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Patch("/organizations/current", handler.UpdateCurrentOrganization)

	// Make request with invalid JSON
	req := httptest.NewRequest(http.MethodPatch, "/organizations/current", strings.NewReader("invalid json"))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
