package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
	mock_db "github.com/rastrigin-systems/ubik-enterprise/generated/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestRequireRole_AllowsMatchingRole tests that a user with a matching role is allowed
func TestRequireRole_AllowsMatchingRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_db.NewMockQuerier(ctrl)

	// Setup: admin user trying to access admin-only endpoint
	roleID := uuid.New()
	sessionData := &db.GetSessionWithEmployeeRow{
		EmployeeID: uuid.New(),
		RoleID:     roleID,
	}

	// Mock GetRole to return admin role
	mockDB.EXPECT().
		GetRole(gomock.Any(), roleID).
		Return(db.Role{
			ID:   roleID,
			Name: "admin",
		}, nil)

	// Create middleware
	middleware := RequireRole(mockDB, "admin")

	// Create test handler that should be called
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create request with session data in context
	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	ctx := context.WithValue(req.Context(), sessionDataKey, sessionData)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Execute
	middleware(testHandler).ServeHTTP(rr, req)

	// Assert: handler was called, request succeeded
	assert.True(t, handlerCalled, "Handler should have been called")
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestRequireRole_BlocksNonMatchingRole tests that a user without the required role is blocked
func TestRequireRole_BlocksNonMatchingRole(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_db.NewMockQuerier(ctrl)

	// Setup: developer user trying to access admin-only endpoint
	roleID := uuid.New()
	sessionData := &db.GetSessionWithEmployeeRow{
		EmployeeID: uuid.New(),
		RoleID:     roleID,
	}

	// Mock GetRole to return developer role
	mockDB.EXPECT().
		GetRole(gomock.Any(), roleID).
		Return(db.Role{
			ID:   roleID,
			Name: "developer",
		}, nil)

	// Create middleware requiring admin role
	middleware := RequireRole(mockDB, "admin")

	// Create test handler that should NOT be called
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create request with session data in context
	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	ctx := context.WithValue(req.Context(), sessionDataKey, sessionData)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Execute
	middleware(testHandler).ServeHTTP(rr, req)

	// Assert: handler was NOT called, 403 returned
	assert.False(t, handlerCalled, "Handler should NOT have been called")
	assert.Equal(t, http.StatusForbidden, rr.Code)

	// Check error message
	var errResp struct {
		Error string `json:"error"`
	}
	err := json.NewDecoder(rr.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, "Insufficient permissions", errResp.Error)
}

// TestRequireRole_AllowsMultipleRoles tests that any of the allowed roles can access
func TestRequireRole_AllowsMultipleRoles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_db.NewMockQuerier(ctrl)

	// Setup: manager user trying to access endpoint allowing admin OR manager
	roleID := uuid.New()
	sessionData := &db.GetSessionWithEmployeeRow{
		EmployeeID: uuid.New(),
		RoleID:     roleID,
	}

	// Mock GetRole to return manager role
	mockDB.EXPECT().
		GetRole(gomock.Any(), roleID).
		Return(db.Role{
			ID:   roleID,
			Name: "manager",
		}, nil)

	// Create middleware allowing admin OR manager
	middleware := RequireRole(mockDB, "admin", "manager")

	// Create test handler
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create request with session data in context
	req := httptest.NewRequest(http.MethodGet, "/employees", nil)
	ctx := context.WithValue(req.Context(), sessionDataKey, sessionData)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Execute
	middleware(testHandler).ServeHTTP(rr, req)

	// Assert: handler was called
	assert.True(t, handlerCalled, "Handler should have been called")
	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestRequireRole_NoSessionData tests that missing session data returns 401
func TestRequireRole_NoSessionData(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_db.NewMockQuerier(ctrl)

	// Create middleware
	middleware := RequireRole(mockDB, "admin")

	// Create test handler that should NOT be called
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	// Create request WITHOUT session data in context
	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	rr := httptest.NewRecorder()

	// Execute
	middleware(testHandler).ServeHTTP(rr, req)

	// Assert: 401 returned (not authenticated)
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

// TestRequireRole_RoleNotFound tests that missing role in DB returns 403
func TestRequireRole_RoleNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mock_db.NewMockQuerier(ctrl)

	// Setup: session with role_id that doesn't exist in DB
	roleID := uuid.New()
	sessionData := &db.GetSessionWithEmployeeRow{
		EmployeeID: uuid.New(),
		RoleID:     roleID,
	}

	// Mock GetRole to return error (not found)
	mockDB.EXPECT().
		GetRole(gomock.Any(), roleID).
		Return(db.Role{}, pgx.ErrNoRows)

	// Create middleware
	middleware := RequireRole(mockDB, "admin")

	// Create test handler that should NOT be called
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	// Create request with session data in context
	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	ctx := context.WithValue(req.Context(), sessionDataKey, sessionData)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Execute
	middleware(testHandler).ServeHTTP(rr, req)

	// Assert: 403 returned
	assert.False(t, handlerCalled)
	assert.Equal(t, http.StatusForbidden, rr.Code)
}

// TestGetRoleName_ReturnsRoleFromContext tests that we can get the role name from context
func TestGetRoleName_ReturnsRoleFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), roleNameKey, "admin")

	roleName, err := GetRoleName(ctx)

	assert.NoError(t, err)
	assert.Equal(t, "admin", roleName)
}

// TestGetRoleName_ReturnsErrorWhenMissing tests error when role not in context
func TestGetRoleName_ReturnsErrorWhenMissing(t *testing.T) {
	ctx := context.Background()

	roleName, err := GetRoleName(ctx)

	assert.Error(t, err)
	assert.Equal(t, "", roleName)
}
