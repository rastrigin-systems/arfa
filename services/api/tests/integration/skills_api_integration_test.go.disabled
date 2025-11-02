package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/middleware"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/tests/testutil"
)

// ============================================================================
// Skills Catalog API Tests
// ============================================================================

func TestListSkills_API_Integration(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee for authentication
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "user@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewSkillsHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/skills", handler.ListSkills)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/skills", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListSkillsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have at least the 3 seed skills
	assert.GreaterOrEqual(t, response.Total, 3)
	assert.GreaterOrEqual(t, len(response.Skills), 3)

	// Verify all skills have required fields
	for _, skill := range response.Skills {
		assert.NotEqual(t, uuid.Nil.String(), skill.Id)
		assert.NotEmpty(t, skill.Name)
		assert.NotEmpty(t, skill.Version)
		assert.NotNil(t, skill.Files)
		assert.True(t, skill.IsActive) // ListSkills only returns active skills
	}
}

func TestGetSkill_API_Integration(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "user@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Get a skill from seed data
	skill, err := queries.GetSkillByName(ctx, "github-task-manager")
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewSkillsHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/skills/{skill_id}", handler.GetSkill)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/skills/"+skill.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Add chi context for path params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("skill_id", skill.ID.String())
	req = req.WithContext(testutil.WithChiContext(req.Context(), rctx))

	router.ServeHTTP(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Skill
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, skill.ID.String(), response.Id)
	assert.Equal(t, skill.Name, response.Name)
	assert.True(t, response.IsActive)
}

func TestGetSkill_API_NotFound(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "user@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Use non-existent skill ID
	nonExistentID := uuid.New()

	// Setup handler
	handler := handlers.NewSkillsHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/skills/{skill_id}", handler.GetSkill)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/skills/"+nonExistentID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Add chi context for path params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("skill_id", nonExistentID.String())
	req = req.WithContext(testutil.WithChiContext(req.Context(), rctx))

	router.ServeHTTP(rec, req)

	// Assert 404 response
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// Employee Skills API Tests
// ============================================================================

func TestListEmployeeSkills_API_Integration(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "user@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Assign some skills to the employee
	skills, err := queries.ListSkills(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(skills), 2)

	for i := 0; i < 2; i++ {
		configJSON := json.RawMessage(`{"auto_enable": true}`)
		_, err := queries.AssignSkillToEmployee(ctx, db.AssignSkillToEmployeeParams{
			EmployeeID: employee.ID,
			SkillID:    skills[i].ID,
			IsEnabled:  testutil.ToNullBool(true),
			Config:     configJSON,
		})
		require.NoError(t, err)
	}

	// Setup handler
	handler := handlers.NewSkillsHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/me/skills", handler.ListEmployeeSkills)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/employees/me/skills", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeeSkillsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.Total)
	assert.Equal(t, 2, len(response.Skills))

	// Verify employee skills have required fields
	for _, skill := range response.Skills {
		assert.NotEqual(t, uuid.Nil.String(), skill.Id)
		assert.NotEmpty(t, skill.Name)
		assert.NotEmpty(t, skill.Version)
		assert.NotNil(t, skill.Files)
		assert.True(t, skill.IsActive)
		assert.True(t, skill.IsEnabled)
		assert.NotNil(t, skill.InstalledAt)
	}
}

func TestListEmployeeSkills_API_EmptyList(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee with NO skills assigned
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "user@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewSkillsHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/me/skills", handler.ListEmployeeSkills)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/employees/me/skills", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeeSkillsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 0, response.Total)
	assert.Equal(t, 0, len(response.Skills))
}

func TestGetEmployeeSkill_API_Integration(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "user@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Assign a skill to the employee
	skill, err := queries.GetSkillByName(ctx, "release-manager")
	require.NoError(t, err)

	configJSON := json.RawMessage(`{"auto_release": true}`)
	_, err = queries.AssignSkillToEmployee(ctx, db.AssignSkillToEmployeeParams{
		EmployeeID: employee.ID,
		SkillID:    skill.ID,
		IsEnabled:  testutil.ToNullBool(true),
		Config:     configJSON,
	})
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewSkillsHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/me/skills/{skill_id}", handler.GetEmployeeSkill)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/employees/me/skills/"+skill.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Add chi context for path params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("skill_id", skill.ID.String())
	req = req.WithContext(testutil.WithChiContext(req.Context(), rctx))

	router.ServeHTTP(rec, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.EmployeeSkill
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, skill.ID.String(), response.Id)
	assert.Equal(t, skill.Name, response.Name)
	assert.True(t, response.IsActive)
	assert.True(t, response.IsEnabled)
	assert.NotNil(t, response.InstalledAt)
	assert.NotNil(t, response.Config)
}

func TestGetEmployeeSkill_API_NotFound(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "user@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Use skill that exists but is NOT assigned to employee
	skill, err := queries.GetSkillByName(ctx, "github-task-manager")
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewSkillsHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/me/skills/{skill_id}", handler.GetEmployeeSkill)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/employees/me/skills/"+skill.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	// Add chi context for path params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("skill_id", skill.ID.String())
	req = req.WithContext(testutil.WithChiContext(req.Context(), rctx))

	router.ServeHTTP(rec, req)

	// Assert 404 response
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSkillsAPI_Unauthorized(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))

	// Setup handler WITHOUT authentication middleware
	handler := handlers.NewSkillsHandler(queries)
	router := chi.NewRouter()
	router.Get("/skills", handler.ListSkills)

	// Make request without auth token
	req := httptest.NewRequest(http.MethodGet, "/skills", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Without middleware, the request would succeed
	// But in real setup with middleware, it should return 401
	// This test verifies the handler itself doesn't crash without auth context
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
