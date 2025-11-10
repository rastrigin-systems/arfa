package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//
// POST /auth/forgot-password Tests
//

func TestForgotPassword_Integration_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup database
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "Member")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "alice@acme.com",
		FullName: "Alice Smith",
	})

	// Setup HTTP router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/forgot-password", authHandler.ForgotPassword)

	// Prepare request
	reqBody := api.ForgotPasswordRequest{
		Email: openapi_types.Email("alice@acme.com"),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Make request
	router.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ForgotPasswordResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Generic message (doesn't reveal if email exists)
	assert.Contains(t, response.Message, "If an account exists")

	// Verify password reset token count increased
	count, err := queries.CountRecentPasswordResetRequests(ctx, employee.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	t.Logf("✅ Password reset token created for employee %s", employee.Email)
}

func TestForgotPassword_Integration_NonExistentEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup database
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))

	// Setup HTTP router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/forgot-password", authHandler.ForgotPassword)

	// Prepare request with non-existent email
	reqBody := api.ForgotPasswordRequest{
		Email: openapi_types.Email("nonexistent@example.com"),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Make request
	router.ServeHTTP(rec, req)

	// Verify response (should still return success to prevent email enumeration)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ForgotPasswordResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Same generic message (security - don't reveal email doesn't exist)
	assert.Contains(t, response.Message, "If an account exists")

	t.Logf("✅ Generic success message returned for non-existent email")
}

func TestForgotPassword_Integration_RateLimited(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup database
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "Member")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "ratelimit@acme.com",
		FullName: "Rate Limited",
	})

	// Create 3 password reset tokens (reaching rate limit)
	for i := 0; i < 3; i++ {
		token := fmt.Sprintf("test-token-%d", i)
		_, err := queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
			EmployeeID: employee.ID,
			Token:      token,
			ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
		})
		require.NoError(t, err)
	}

	// Setup HTTP router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/forgot-password", authHandler.ForgotPassword)

	// Prepare request
	reqBody := api.ForgotPasswordRequest{
		Email: openapi_types.Email("ratelimit@acme.com"),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Make request (should be rate limited)
	router.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	var errResp api.Error
	err := json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error, "Too many")

	t.Logf("✅ Rate limit enforced after 3 requests")
}

//
// GET /auth/verify-reset-token Tests
//

func TestVerifyResetToken_Integration_ValidToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup database
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "Member")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "verify@acme.com",
		FullName: "Verify Token",
	})

	// Create valid password reset token
	token := "valid-test-token-123"
	_, err := queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		EmployeeID: employee.ID,
		Token:      token,
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
	})
	require.NoError(t, err)

	// Setup HTTP router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Get("/auth/verify-reset-token", authHandler.VerifyResetToken)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/auth/verify-reset-token?token="+token, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.VerifyResetTokenResponse
	err = json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.True(t, response.Valid)

	t.Logf("✅ Valid token verified successfully")
}

func TestVerifyResetToken_Integration_ExpiredToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup database
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "Member")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "expired@acme.com",
		FullName: "Expired Token",
	})

	// Create expired password reset token
	token := "expired-test-token-123"
	_, err := queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		EmployeeID: employee.ID,
		Token:      token,
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(-1 * time.Hour), Valid: true}, // Expired
	})
	require.NoError(t, err)

	// Setup HTTP router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Get("/auth/verify-reset-token", authHandler.VerifyResetToken)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/auth/verify-reset-token?token="+token, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp api.Error
	err = json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error, "expired")

	t.Logf("✅ Expired token rejected")
}

func TestVerifyResetToken_Integration_UsedToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup database
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "Member")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "used@acme.com",
		FullName: "Used Token",
	})

	// Create password reset token and mark as used
	token := "used-test-token-123"
	_, err := queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		EmployeeID: employee.ID,
		Token:      token,
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
	})
	require.NoError(t, err)

	// Mark token as used
	err = queries.MarkPasswordResetTokenUsed(ctx, token)
	require.NoError(t, err)

	// Setup HTTP router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Get("/auth/verify-reset-token", authHandler.VerifyResetToken)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/auth/verify-reset-token?token="+token, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp api.Error
	err = json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error, "used")

	t.Logf("✅ Already used token rejected")
}

func TestVerifyResetToken_Integration_InvalidToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup database
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))

	// Setup HTTP router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Get("/auth/verify-reset-token", authHandler.VerifyResetToken)

	// Make request with invalid token
	req := httptest.NewRequest(http.MethodGet, "/auth/verify-reset-token?token=invalid-token", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp api.Error
	err := json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error, "Invalid")

	t.Logf("✅ Invalid token rejected")
}

//
// POST /auth/reset-password Tests
//

func TestResetPassword_Integration_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup database
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "Member")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "reset@acme.com",
		FullName: "Reset Password",
	})

	// Create valid password reset token
	token := "reset-test-token-123"
	_, err := queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		EmployeeID: employee.ID,
		Token:      token,
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
	})
	require.NoError(t, err)

	// Setup HTTP router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/reset-password", authHandler.ResetPassword)

	// Prepare request
	newPassword := "NewSecurePassword123!"
	reqBody := api.ResetPasswordRequest{
		Token:       token,
		NewPassword: newPassword,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Make request
	router.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ResetPasswordResponse
	err = json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.Contains(t, response.Message, "successful")

	// Verify password was updated in database
	updatedEmployee, err := queries.GetEmployee(ctx, employee.ID)
	require.NoError(t, err)
	assert.True(t, auth.VerifyPassword(newPassword, updatedEmployee.PasswordHash))

	// Verify token was marked as used
	_, err = queries.GetPasswordResetToken(ctx, token)
	assert.Error(t, err) // Should not find token (it's marked as used)

	t.Logf("✅ Password reset successful and token marked as used")
}

func TestResetPassword_Integration_WeakPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup database
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "Member")

	// Create test employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "weak@acme.com",
		FullName: "Weak Password",
	})

	// Create valid password reset token
	token := "weak-password-token"
	_, err := queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		EmployeeID: employee.ID,
		Token:      token,
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
	})
	require.NoError(t, err)

	// Setup HTTP router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/reset-password", authHandler.ResetPassword)

	// Prepare request with weak password (less than 8 characters)
	reqBody := api.ResetPasswordRequest{
		Token:       token,
		NewPassword: "weak", // Too short
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Make request
	router.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp api.Error
	err = json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error, "8 characters")

	t.Logf("✅ Weak password rejected")
}

func TestResetPassword_Integration_InvalidToken(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup database
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))

	// Setup HTTP router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/reset-password", authHandler.ResetPassword)

	// Prepare request with invalid token
	reqBody := api.ResetPasswordRequest{
		Token:       "invalid-token",
		NewPassword: "ValidPassword123!",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/reset-password", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// Make request
	router.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp api.Error
	err := json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error, "Invalid")

	t.Logf("✅ Invalid token rejected")
}
