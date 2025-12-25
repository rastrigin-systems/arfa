package handlers

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/services/api/internal/middleware"
)

// SetOrgIDInContext is a test helper to set org_id in context
// This simulates what the JWT middleware does in production
func SetOrgIDInContext(ctx context.Context, orgID uuid.UUID) context.Context {
	return middleware.SetOrgIDForTest(ctx, orgID)
}

// SetEmployeeIDInContext is a test helper to set employee_id in context
func SetEmployeeIDInContext(ctx context.Context, employeeID uuid.UUID) context.Context {
	return middleware.SetEmployeeIDForTest(ctx, employeeID)
}

// SetSessionDataInContext is a test helper to set session_data in context
func SetSessionDataInContext(ctx context.Context, sessionData *db.GetSessionWithEmployeeRow) context.Context {
	return middleware.SetSessionDataForTest(ctx, sessionData)
}

// GetOrgID wraps middleware.GetOrgID for convenience
func GetOrgID(ctx context.Context) (uuid.UUID, error) {
	return middleware.GetOrgID(ctx)
}

// WithChiContext is a test helper to add chi route context to a context
func WithChiContext(ctx context.Context, chiCtx *chi.Context) context.Context {
	return context.WithValue(ctx, chi.RouteCtxKey, chiCtx)
}
