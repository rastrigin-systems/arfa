package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/internal/handlers"
	"github.com/sergeirastrigin/ubik-enterprise/tests/testutil"
)

// TestLogin_Integration_Success tests the complete login flow with a REAL database
//
// Integration Test Lesson:
// - Uses real PostgreSQL via testcontainers
// - Tests full HTTP → Handler → Database → Response flow
// - Verifies side effects (session creation, last login update)
// - Slower but more comprehensive than unit tests
func TestLogin_Integration_Success(t *testing.T) {
	// Skip in short mode (for quick local testing)
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// === ARRANGE === Setup real database and test data

	// Step 1: Create real PostgreSQL database
	// Integration Test Lesson: This is the KEY difference from unit tests!
	// We're using a REAL database, not a mock.
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))

	ctx := testutil.GetContext(t)

	// Step 2: Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Step 3: Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "Admin")

	// Step 4: Create test employee with known password
	password := "SecurePass123!"
	passwordHash, err := auth.HashPassword(password)
	require.NoError(t, err)

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org.ID,
		RoleID:       role.ID,
		Email:        "alice@acme.com",
		FullName:     "Alice Smith",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	// Debug: Verify employee was created
	t.Logf("Created employee: ID=%s, Email=%s", employee.ID, employee.Email)

	// Debug: Verify we can retrieve the employee
	retrievedEmp, err := queries.GetEmployeeByEmail(ctx, "alice@acme.com")
	require.NoError(t, err, "Should be able to retrieve employee by email")
	t.Logf("Retrieved employee: ID=%s, Email=%s, PasswordHash length=%d",
		retrievedEmp.ID, retrievedEmp.Email, len(retrievedEmp.PasswordHash))

	// Step 5: Create HTTP router with real handler
	// Integration Test Lesson: We're testing the REAL handler, not a mock
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/login", authHandler.Login)

	// === ACT === Make real HTTP request

	loginRequest := api.LoginRequest{
		Email:    openapi_types.Email("alice@acme.com"),
		Password: password,
	}
	bodyBytes, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// === ASSERT === Verify response and database state

	// Debug: Print response if not 200
	if rec.Code != http.StatusOK {
		t.Logf("Response Code: %d", rec.Code)
		t.Logf("Response Body: %s", rec.Body.String())
	}

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, rec.Code, "Should return 200 OK")

	var response api.LoginResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify token is returned
	assert.NotEmpty(t, response.Token, "Should return JWT token")
	assert.Greater(t, len(response.Token), 50, "Token should be reasonably long")

	// Verify employee data
	require.NotNil(t, response.Employee.Id)
	assert.Equal(t, employee.ID.String(), response.Employee.Id.String())
	assert.Equal(t, "alice@acme.com", string(response.Employee.Email))
	assert.Equal(t, "Alice Smith", response.Employee.FullName)

	// === INTEGRATION TEST UNIQUE PART ===
	// Verify DATABASE side effects (this is what unit tests can't do!)

	// Verify 1: Session was created in database
	tokenHash := auth.HashToken(response.Token)
	session, err := queries.GetSession(ctx, tokenHash)
	require.NoError(t, err, "Session should exist in database")
	assert.Equal(t, employee.ID, session.EmployeeID, "Session should belong to employee")
	assert.True(t, session.ExpiresAt.Time.After(time.Now()), "Session should not be expired")

	// Verify 2: Last login timestamp was updated
	updatedEmployee, err := queries.GetEmployee(ctx, employee.ID)
	require.NoError(t, err)
	require.True(t, updatedEmployee.LastLoginAt.Valid, "Last login should be set")
	assert.WithinDuration(t, time.Now(), updatedEmployee.LastLoginAt.Time, 5*time.Second,
		"Last login should be recent")

	// Verify 3: Token can be verified and contains correct claims
	claims, err := auth.VerifyJWT(response.Token)
	require.NoError(t, err, "Token should be valid")
	assert.Equal(t, employee.ID.String(), claims.EmployeeID)
	assert.Equal(t, org.ID.String(), claims.OrgID)

	// Integration Test Lesson: We just verified:
	// ✅ HTTP request/response work
	// ✅ Handler logic works
	// ✅ Database queries work
	// ✅ Sessions are created
	// ✅ Timestamps are updated
	// ✅ JWT tokens are valid
	//
	// This gives us CONFIDENCE the whole system works together!
}

// TestLogin_Integration_InvalidPassword tests login with wrong password
//
// Integration Test Lesson: Even error cases should be tested with real DB
// to ensure database queries don't have unexpected side effects
func TestLogin_Integration_InvalidPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test data
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "User")

	correctPassword := "CorrectPassword123"
	passwordHash, _ := auth.HashPassword(correctPassword)

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org.ID,
		RoleID:       role.ID,
		Email:        "bob@acme.com",
		FullName:     "Bob Jones",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	// Setup router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/login", authHandler.Login)

	// Try to login with WRONG password
	loginRequest := api.LoginRequest{
		Email:    openapi_types.Email("bob@acme.com"),
		Password: "WrongPassword",
	}
	bodyBytes, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Verify error response
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// === VERIFY NO SIDE EFFECTS ===
	// Integration Test Lesson: Failed login should NOT create session or update last_login

	// Verify 1: No session was created (checking with non-existent hash)
	_, sessionErr := queries.GetSessionWithEmployee(ctx, "nonexistent-hash")
	assert.Error(t, sessionErr, "Session should not exist")

	// Verify 2: Last login was NOT updated
	unchangedEmployee, err := queries.GetEmployee(ctx, employee.ID)
	require.NoError(t, err)
	assert.False(t, unchangedEmployee.LastLoginAt.Valid, "Last login should still be null")
}

// TestLogin_Integration_SuspendedUser tests that suspended users cannot login
func TestLogin_Integration_SuspendedUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "User")

	password := "Password123"
	passwordHash, _ := auth.HashPassword(password)

	// Create SUSPENDED employee
	_ = testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org.ID,
		RoleID:       role.ID,
		Email:        "suspended@acme.com",
		FullName:     "Suspended User",
		PasswordHash: passwordHash,
		Status:       "suspended", // Key difference: not active!
	})

	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/login", authHandler.Login)

	loginRequest := api.LoginRequest{
		Email:    openapi_types.Email("suspended@acme.com"),
		Password: password,
	}
	bodyBytes, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var errorResponse api.Error
	json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.Contains(t, errorResponse.Error, "suspended")
}

// TestLogin_Integration_MultipleEmployees tests that login works correctly
// when multiple employees exist (tests query filtering)
//
// Integration Test Lesson: Real databases can have complex scenarios
// that are hard to mock - test them with real data!
func TestLogin_Integration_MultipleEmployees(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create TWO organizations
	org1 := testutil.CreateTestOrg(t, queries, ctx)
	org2 := testutil.CreateTestOrg(t, queries, ctx)

	role := testutil.CreateTestRole(t, queries, ctx, "User")

	password := "Password123"
	passwordHash, _ := auth.HashPassword(password)

	// Create employee in org1 with email alice@acme.com
	emp1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org1.ID,
		RoleID:       role.ID,
		Email:        "alice@company1.com",
		FullName:     "Alice from Org 1",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	// Create employee in org2 with different email
	emp2 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org2.ID,
		RoleID:       role.ID,
		Email:        "bob@company2.com",
		FullName:     "Bob from Org 2",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/login", authHandler.Login)

	// Login as alice@company1.com
	loginRequest := api.LoginRequest{
		Email:    openapi_types.Email("alice@company1.com"),
		Password: password,
	}
	bodyBytes, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.LoginResponse
	json.Unmarshal(rec.Body.Bytes(), &response)

	// Verify we got the CORRECT employee (from org1, not org2)
	require.NotNil(t, response.Employee.Id)
	assert.Equal(t, emp1.ID.String(), response.Employee.Id.String())
	assert.Equal(t, org1.ID.String(), response.Employee.OrgId.String())
	assert.NotEqual(t, emp2.ID.String(), response.Employee.Id.String(), "Should not get wrong employee")

	// Verify JWT contains correct org_id
	claims, err := auth.VerifyJWT(response.Token)
	require.NoError(t, err)
	assert.Equal(t, org1.ID.String(), claims.OrgID, "Token should have org1's ID")
}

// TestLogout_Integration_Success tests the logout flow with real database
func TestLogout_Integration_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test data
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "User")

	password := "Password123"
	passwordHash, _ := auth.HashPassword(password)

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org.ID,
		RoleID:       role.ID,
		Email:        "logout@acme.com",
		FullName:     "Logout User",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	// Step 1: Login to get a token
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/login", authHandler.Login)
	router.Post("/auth/logout", authHandler.Logout)

	loginRequest := api.LoginRequest{
		Email:    openapi_types.Email("logout@acme.com"),
		Password: password,
	}
	bodyBytes, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var loginResponse api.LoginResponse
	json.Unmarshal(rec.Body.Bytes(), &loginResponse)
	token := loginResponse.Token
	tokenHash := auth.HashToken(token)

	// Verify session exists
	session, err := queries.GetSession(ctx, tokenHash)
	require.NoError(t, err, "Session should exist after login")
	assert.Equal(t, employee.ID, session.EmployeeID)

	// Step 2: Logout
	logoutReq := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+token)
	logoutRec := httptest.NewRecorder()

	router.ServeHTTP(logoutRec, logoutReq)

	// Verify logout succeeded
	assert.Equal(t, http.StatusOK, logoutRec.Code)

	var logoutResponse map[string]string
	json.Unmarshal(logoutRec.Body.Bytes(), &logoutResponse)
	assert.Equal(t, "Logged out successfully", logoutResponse["message"])

	// Verify session was deleted from database
	_, err = queries.GetSession(ctx, tokenHash)
	assert.Error(t, err, "Session should not exist after logout")

	// Integration Test Lesson: We verified:
	// ✅ Login creates session
	// ✅ Logout deletes session
	// ✅ Session no longer retrievable after logout
}

// TestFullAuthFlow_Integration tests the complete authentication lifecycle:
// Login → GetMe → Logout → GetMe (should fail)
//
// Integration Test Lesson: This tests the ENTIRE auth system working together
func TestFullAuthFlow_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Setup test data
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "User")

	password := "SecurePassword123"
	passwordHash, _ := auth.HashPassword(password)

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org.ID,
		RoleID:       role.ID,
		Email:        "fullflow@acme.com",
		FullName:     "Full Flow User",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	// Setup router with all endpoints
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/login", authHandler.Login)
	router.Get("/auth/me", authHandler.GetMe)
	router.Post("/auth/logout", authHandler.Logout)

	// ========================================================================
	// Step 1: Login
	// ========================================================================

	loginRequest := api.LoginRequest{
		Email:    openapi_types.Email("fullflow@acme.com"),
		Password: password,
	}
	bodyBytes, _ := json.Marshal(loginRequest)

	loginReq := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	loginReq.Header.Set("Content-Type", "application/json")
	loginRec := httptest.NewRecorder()

	router.ServeHTTP(loginRec, loginReq)

	assert.Equal(t, http.StatusOK, loginRec.Code, "Login should succeed")

	var loginResponse api.LoginResponse
	json.Unmarshal(loginRec.Body.Bytes(), &loginResponse)
	token := loginResponse.Token

	assert.NotEmpty(t, token, "Should receive token")
	assert.Equal(t, employee.ID.String(), loginResponse.Employee.Id.String())

	// ========================================================================
	// Step 2: GetMe (should work with valid token)
	// ========================================================================

	getMeReq1 := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	getMeReq1.Header.Set("Authorization", "Bearer "+token)
	getMeRec1 := httptest.NewRecorder()

	router.ServeHTTP(getMeRec1, getMeReq1)

	assert.Equal(t, http.StatusOK, getMeRec1.Code, "GetMe should succeed with valid token")

	var meResponse1 api.Employee
	json.Unmarshal(getMeRec1.Body.Bytes(), &meResponse1)

	assert.Equal(t, employee.ID.String(), meResponse1.Id.String())
	assert.Equal(t, "fullflow@acme.com", string(meResponse1.Email))
	assert.Equal(t, "Full Flow User", meResponse1.FullName)
	assert.Equal(t, "active", string(meResponse1.Status))

	// ========================================================================
	// Step 3: Logout
	// ========================================================================

	logoutReq := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	logoutReq.Header.Set("Authorization", "Bearer "+token)
	logoutRec := httptest.NewRecorder()

	router.ServeHTTP(logoutRec, logoutReq)

	assert.Equal(t, http.StatusOK, logoutRec.Code, "Logout should succeed")

	var logoutResponse map[string]string
	json.Unmarshal(logoutRec.Body.Bytes(), &logoutResponse)
	assert.Equal(t, "Logged out successfully", logoutResponse["message"])

	// ========================================================================
	// Step 4: GetMe again (should FAIL - session is gone)
	// ========================================================================

	getMeReq2 := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	getMeReq2.Header.Set("Authorization", "Bearer "+token)
	getMeRec2 := httptest.NewRecorder()

	router.ServeHTTP(getMeRec2, getMeReq2)

	assert.Equal(t, http.StatusUnauthorized, getMeRec2.Code, "GetMe should fail after logout")

	var errorResponse api.Error
	json.Unmarshal(getMeRec2.Body.Bytes(), &errorResponse)
	assert.Contains(t, errorResponse.Error, "Session not found")

	// Integration Test Lesson: Complete auth lifecycle verified!
	// ✅ Login → creates session and returns token
	// ✅ GetMe → validates token and returns employee data
	// ✅ Logout → invalidates session
	// ✅ GetMe (after logout) → correctly rejects invalidated token
	//
	// This is the GOLD STANDARD for integration testing:
	// We tested the entire system working together with a real database!
}
