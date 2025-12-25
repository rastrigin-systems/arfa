package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/services/api/internal/auth"
)

// Context keys for storing auth data
type contextKey string

const (
	employeeIDKey  contextKey = "employee_id"
	orgIDKey       contextKey = "org_id"
	sessionDataKey contextKey = "session_data"
)

var (
	// ErrNoEmployeeID is returned when employee_id is not in context
	ErrNoEmployeeID = errors.New("employee_id not found in context")
	// ErrNoOrgID is returned when org_id is not in context
	ErrNoOrgID = errors.New("org_id not found in context")
	// ErrNoSessionData is returned when session_data is not in context
	ErrNoSessionData = errors.New("session_data not found in context")
)

// JWTAuth middleware extracts and verifies JWT token, adds auth data to context
//
// TDD Lesson: This middleware eliminates duplicate auth code from every handler
//
// Implementation Steps (derived from tests):
// 1. Extract token from Authorization header
// 2. Verify JWT token is valid and not expired
// 3. Hash the token
// 4. Fetch session and employee from database
// 5. Add employee_id, org_id, and session data to context
// 6. Call next handler with enriched context
func JWTAuth(queries db.Querier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Step 1: Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "Missing authorization header")
				return
			}

			// Extract "Bearer <token>" format
			const bearerPrefix = "Bearer "
			if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
				writeError(w, http.StatusUnauthorized, "Invalid authorization header format")
				return
			}
			token := authHeader[len(bearerPrefix):]

			// Handle edge case: empty token after "Bearer "
			if token == "" {
				writeError(w, http.StatusUnauthorized, "Empty token")
				return
			}

			// Step 2: Verify JWT token (checks signature and expiration)
			_, err := auth.VerifyJWT(token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			// Step 3: Hash the token for database lookup
			tokenHash := auth.HashToken(token)

			// Step 4: Fetch session and employee data
			// This also validates:
			// - Session exists and is not expired
			// - Employee exists and is not deleted
			// - Employee status is active
			sessionData, err := queries.GetSessionWithEmployee(ctx, tokenHash)
			if err != nil {
				// Session not found (user logged out or session expired in DB)
				writeError(w, http.StatusUnauthorized, "Session not found")
				return
			}

			// Step 5: Add auth data to context
			ctx = context.WithValue(ctx, employeeIDKey, sessionData.EmployeeID)
			ctx = context.WithValue(ctx, orgIDKey, sessionData.OrgID)
			ctx = context.WithValue(ctx, sessionDataKey, &sessionData)

			// Step 6: Call next handler with enriched context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetEmployeeID extracts the employee ID from the request context
//
// TDD Lesson: Helper function for handlers to easily access authenticated employee
func GetEmployeeID(ctx context.Context) (uuid.UUID, error) {
	employeeID, ok := ctx.Value(employeeIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, ErrNoEmployeeID
	}
	return employeeID, nil
}

// GetOrgID extracts the organization ID from the request context
//
// TDD Lesson: Helper function for multi-tenant queries
func GetOrgID(ctx context.Context) (uuid.UUID, error) {
	orgID, ok := ctx.Value(orgIDKey).(uuid.UUID)
	if !ok {
		return uuid.UUID{}, ErrNoOrgID
	}
	return orgID, nil
}

// GetSessionData extracts the full session data from the request context
//
// TDD Lesson: Helper function if handlers need more than just IDs
func GetSessionData(ctx context.Context) (*db.GetSessionWithEmployeeRow, error) {
	sessionData, ok := ctx.Value(sessionDataKey).(*db.GetSessionWithEmployeeRow)
	if !ok {
		return nil, ErrNoSessionData
	}
	return sessionData, nil
}

// WithTestAuth creates a context with auth data for testing
// This is only for unit tests - in production, use JWTAuth middleware
func WithTestAuth(ctx context.Context, employeeID, orgID uuid.UUID) context.Context {
	ctx = context.WithValue(ctx, employeeIDKey, employeeID)
	ctx = context.WithValue(ctx, orgIDKey, orgID)
	return ctx
}

// writeError writes a JSON error response
//
// TDD Lesson: Centralized error handling for middleware
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(api.Error{
		Error: message,
	})
}

// Test helpers for setting context values in tests
// These should only be used in test code to simulate middleware behavior

// SetOrgIDForTest sets org_id in context (for testing only)
func SetOrgIDForTest(ctx context.Context, orgID uuid.UUID) context.Context {
	return context.WithValue(ctx, orgIDKey, orgID)
}

// SetEmployeeIDForTest sets employee_id in context (for testing only)
func SetEmployeeIDForTest(ctx context.Context, employeeID uuid.UUID) context.Context {
	return context.WithValue(ctx, employeeIDKey, employeeID)
}

// SetSessionDataForTest sets session_data in context (for testing only)
func SetSessionDataForTest(ctx context.Context, sessionData *db.GetSessionWithEmployeeRow) context.Context {
	return context.WithValue(ctx, sessionDataKey, sessionData)
}
