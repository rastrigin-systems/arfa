package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rastrigin-systems/arfa/services/api/internal/auth"
	"github.com/rastrigin-systems/arfa/services/api/internal/middleware"
	"github.com/rastrigin-systems/arfa/services/api/tests/testutil"
)

// TestAuthMiddleware_Integration_ProtectedRoute tests middleware with real database
//
// Integration Test Lesson: This verifies the full auth stack:
// - Real PostgreSQL database
// - Real session lookup
// - Context propagation to handler
func TestAuthMiddleware_Integration_ProtectedRoute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conn, queries := testutil.SetupTestDB(t)
	defer func() { _ = conn.Close(testutil.GetContext(t)) }()
	ctx := testutil.GetContext(t)

	// Create test data
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "User")

	password := "TestPassword123"
	passwordHash, _ := auth.HashPassword(password)

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org.ID,
		RoleID:       role.ID,
		Email:        "middleware@example.com",
		FullName:     "Middleware Test User",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	// Generate a valid token and create session
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)

	// Create session in database
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err, "Should create session")

	// Create a protected endpoint
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract auth data from context (set by middleware)
		employeeID, err := middleware.GetEmployeeID(r.Context())
		require.NoError(t, err, "Should have employee_id in context")

		orgID, err := middleware.GetOrgID(r.Context())
		require.NoError(t, err, "Should have org_id in context")

		sessionData, err := middleware.GetSessionData(r.Context())
		require.NoError(t, err, "Should have session data in context")

		// Verify it's the correct employee
		assert.Equal(t, employee.ID, employeeID)
		assert.Equal(t, org.ID, orgID)
		assert.Equal(t, "middleware@example.com", sessionData.Email)
		assert.Equal(t, "active", sessionData.Status)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Protected resource accessed"))
	})

	// Setup router with middleware
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/protected", protectedHandler)

	// Test 1: Valid token should access protected route
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code, "Should access protected route with valid token")
	assert.Contains(t, rec.Body.String(), "Protected resource accessed")

	// Test 2: No token should be rejected
	reqNoToken := httptest.NewRequest(http.MethodGet, "/protected", nil)
	recNoToken := httptest.NewRecorder()

	router.ServeHTTP(recNoToken, reqNoToken)

	assert.Equal(t, http.StatusUnauthorized, recNoToken.Code, "Should reject request without token")

	// Test 3: Invalid token should be rejected
	reqInvalid := httptest.NewRequest(http.MethodGet, "/protected", nil)
	reqInvalid.Header.Set("Authorization", "Bearer invalid-token")
	recInvalid := httptest.NewRecorder()

	router.ServeHTTP(recInvalid, reqInvalid)

	assert.Equal(t, http.StatusUnauthorized, recInvalid.Code, "Should reject invalid token")

	// Integration Test Lesson: We verified:
	// ✅ Middleware fetches session from real database
	// ✅ Context is populated with auth data
	// ✅ Protected handler receives correct employee/org IDs
	// ✅ Invalid tokens are rejected before reaching handler
}

// TestAuthMiddleware_Integration_AfterLogout tests middleware rejects logged-out users
//
// Integration Test Lesson: Tests the complete logout → middleware flow
func TestAuthMiddleware_Integration_AfterLogout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conn, queries := testutil.SetupTestDB(t)
	defer func() { _ = conn.Close(testutil.GetContext(t)) }()
	ctx := testutil.GetContext(t)

	// Create test data
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "User")

	password := "TestPassword123"
	passwordHash, _ := auth.HashPassword(password)

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org.ID,
		RoleID:       role.ID,
		Email:        "logout@example.com",
		FullName:     "Logout Test User",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	// Generate token and create session
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)

	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err, "Should create session")

	// Create protected endpoint
	protectedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Success"))
	})

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/protected", protectedHandler)

	// Step 1: Access protected route WITH session - should work
	req1 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	rec1 := httptest.NewRecorder()

	router.ServeHTTP(rec1, req1)

	assert.Equal(t, http.StatusOK, rec1.Code, "Should access with valid session")

	// Step 2: Delete session (simulate logout)
	err = queries.DeleteSession(ctx, tokenHash)
	require.NoError(t, err, "Should delete session")

	// Step 3: Try to access protected route again - should FAIL
	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	rec2 := httptest.NewRecorder()

	router.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusUnauthorized, rec2.Code, "Should reject after session deleted")

	// Integration Test Lesson: We verified:
	// ✅ Middleware checks session exists in database
	// ✅ Deleted sessions cause authentication to fail
	// ✅ JWT token alone is not enough - session must exist
	// ✅ Logout effectively invalidates all future requests
}
