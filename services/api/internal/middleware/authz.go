package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/rastrigin-systems/arfa/generated/db"
)

// Context key for role name
const roleNameKey contextKey = "role_name"

// ErrNoRoleName is returned when role_name is not in context
var ErrNoRoleName = errors.New("role_name not found in context")

// RequireRole middleware restricts access to users with specific roles.
//
// This middleware:
// 1. Extracts session data from context (set by JWTAuth middleware)
// 2. Queries the employee's role from database using role_id
// 3. Checks if role name matches any in allowedRoles
// 4. Returns 403 Forbidden if unauthorized
// 5. Adds role_name to context for downstream handlers
// 6. Logs all authorization failures for audit
//
// Usage:
//
//	r.With(RequireRole(db, "admin")).Post("/roles", handler.CreateRole)
//	r.With(RequireRole(db, "admin", "manager")).Get("/employees", handler.ListEmployees)
func RequireRole(queries db.Querier, allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Step 1: Get session data from context (set by JWTAuth middleware)
			sessionData, err := GetSessionData(ctx)
			if err != nil {
				// No session data means JWTAuth middleware didn't run or failed
				// This is a programming error (middleware order) or auth bypass attempt
				log.Printf("AUTHZ_FAIL: No session data in context for %s %s", r.Method, r.URL.Path)
				writeError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			// Step 2: Query role from database using role_id from session
			role, err := queries.GetRole(ctx, sessionData.RoleID)
			if err != nil {
				// Role not found in database
				log.Printf("AUTHZ_FAIL: Role %s not found for employee %s, attempted %s %s",
					sessionData.RoleID, sessionData.EmployeeID, r.Method, r.URL.Path)
				writeError(w, http.StatusForbidden, "Role not found")
				return
			}

			// Step 3: Check if employee's role is in allowed list
			for _, allowedRole := range allowedRoles {
				if role.Name == allowedRole {
					// Authorized - add role name to context and proceed
					ctx = context.WithValue(ctx, roleNameKey, role.Name)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// Step 4: Not authorized - log and return 403
			log.Printf("AUTHZ_FAIL: Employee %s (role: %s) attempted %s %s, requires: %v",
				sessionData.EmployeeID, role.Name, r.Method, r.URL.Path, allowedRoles)
			writeError(w, http.StatusForbidden, "Insufficient permissions")
		})
	}
}

// GetRoleName extracts the role name from the request context.
// This is set by RequireRole middleware after successful authorization.
//
// Usage in handlers:
//
//	roleName, err := middleware.GetRoleName(r.Context())
//	if roleName == "admin" {
//	    // admin-specific logic
//	}
func GetRoleName(ctx context.Context) (string, error) {
	roleName, ok := ctx.Value(roleNameKey).(string)
	if !ok {
		return "", ErrNoRoleName
	}
	return roleName, nil
}

// SetRoleNameForTest sets role_name in context (for testing only)
func SetRoleNameForTest(ctx context.Context, roleName string) context.Context {
	return context.WithValue(ctx, roleNameKey, roleName)
}
