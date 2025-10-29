package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/middleware"
)

// TDD Lesson: Testing middleware - verify it adds auth data to context and calls next handler
func TestAuthMiddleware_ValidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Create valid token
	employeeID := uuid.New()
	orgID := uuid.New()
	roleID := uuid.New()
	token, _ := auth.GenerateJWT(employeeID, orgID, 24*time.Hour)
	tokenHash := auth.HashToken(token)

	// Expect database to return session data
	mockDB.EXPECT().
		GetSessionWithEmployee(gomock.Any(), tokenHash).
		Return(db.GetSessionWithEmployeeRow{
			ID:                uuid.New(),
			EmployeeID:        employeeID,
			TokenHash:         tokenHash,
			ExpiresAt:         pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
			CreatedAt:         pgtype.Timestamp{Time: time.Now(), Valid: true},
			OrgID:             orgID,
			TeamID:            pgtype.UUID{},
			RoleID:            roleID,
			Email:             "test@example.com",
			FullName:          "Test User",
			Status:            "active",
			Preferences:       []byte("{}"),
			LastLoginAt:       pgtype.Timestamp{},
			EmployeeCreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
			EmployeeUpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil)

	// Create a test handler that verifies context was set
	var capturedEmployeeID uuid.UUID
	var capturedOrgID uuid.UUID
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract from context to verify middleware set it
		empID, err := middleware.GetEmployeeID(r.Context())
		require.NoError(t, err, "Should have employee_id in context")
		capturedEmployeeID = empID

		orgIDVal, err := middleware.GetOrgID(r.Context())
		require.NoError(t, err, "Should have org_id in context")
		capturedOrgID = orgIDVal

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap handler with middleware
	authMiddleware := middleware.JWTAuth(mockDB)
	wrappedHandler := authMiddleware(testHandler)

	// Make request with valid token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Verify handler was called (status 200)
	assert.Equal(t, http.StatusOK, rec.Code, "Next handler should be called")

	// Verify context was populated correctly
	assert.Equal(t, employeeID, capturedEmployeeID, "Employee ID should be in context")
	assert.Equal(t, orgID, capturedOrgID, "Org ID should be in context")
}

// TDD Lesson: Test invalid token
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// No database calls expected (token validation fails first)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called with invalid token")
	})

	authMiddleware := middleware.JWTAuth(mockDB)
	wrappedHandler := authMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Should return 401 and NOT call next handler
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TDD Lesson: Test expired token
func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Create expired token
	employeeID := uuid.New()
	orgID := uuid.New()
	token, _ := auth.GenerateJWT(employeeID, orgID, 1*time.Nanosecond)
	time.Sleep(2 * time.Millisecond) // Ensure expired

	// No database calls expected (token verification fails)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called with expired token")
	})

	authMiddleware := middleware.JWTAuth(mockDB)
	wrappedHandler := authMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TDD Lesson: Test missing token
func TestAuthMiddleware_MissingToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called without token")
	})

	authMiddleware := middleware.JWTAuth(mockDB)
	wrappedHandler := authMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No Authorization header
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Should return 401
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TDD Lesson: Test session not found (logged out user)
func TestAuthMiddleware_SessionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	employeeID := uuid.New()
	orgID := uuid.New()
	token, _ := auth.GenerateJWT(employeeID, orgID, 24*time.Hour)
	tokenHash := auth.HashToken(token)

	// Expect database to return "not found" error
	mockDB.EXPECT().
		GetSessionWithEmployee(gomock.Any(), tokenHash).
		Return(db.GetSessionWithEmployeeRow{}, assert.AnError)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called when session not found")
	})

	authMiddleware := middleware.JWTAuth(mockDB)
	wrappedHandler := authMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	// Should return 401 (session doesn't exist - user logged out)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TDD Lesson: Test malformed Authorization header
func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called with malformed header")
	})

	authMiddleware := middleware.JWTAuth(mockDB)
	wrappedHandler := authMiddleware(testHandler)

	// Test cases for malformed headers
	testCases := []struct {
		name   string
		header string
	}{
		{"Missing Bearer prefix", "just-a-token"},
		{"Wrong prefix", "Basic dXNlcjpwYXNz"},
		{"Empty after Bearer", "Bearer "},
		{"Just Bearer", "Bearer"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", tc.header)
			rec := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusUnauthorized, rec.Code, "Should return 401 for: "+tc.name)
		})
	}
}
