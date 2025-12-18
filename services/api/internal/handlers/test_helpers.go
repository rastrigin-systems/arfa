package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/middleware"
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

// GetOrgID wraps middleware.GetOrgID for convenience
func GetOrgID(ctx context.Context) (uuid.UUID, error) {
	return middleware.GetOrgID(ctx)
}
