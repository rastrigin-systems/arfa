# Authorization Guide

This document describes how role-based authorization works in the Arfa API.

## Overview

The API uses **role-based access control (RBAC)** to protect endpoints. Every employee has a role, and each role defines what actions they can perform.

## Roles

Four default roles are seeded in the database:

| Role | Description | Typical Use |
|------|-------------|-------------|
| `admin` | Full access to all endpoints | Organization owners, IT admins |
| `manager` | Can manage employees and teams | Team leads, department heads |
| `developer` | Standard employee access | Engineers, designers |
| `viewer` | Read-only access | Auditors, observers |

## Endpoint Classification

### CLI Endpoints (Any Authenticated Employee)

These endpoints are accessible to **any authenticated employee** regardless of role. They are self-service endpoints for the CLI tool.

```
GET  /api/v1/auth/me                           # Get own profile
GET  /api/v1/employees/me/agent-configs/resolved  # Get own agent configs
PUT  /api/v1/employees/me/claude-token         # Set own Claude token
GET  /api/v1/employees/me/claude-token/status  # Check own token status
```

### Admin Endpoints (Admin Only)

These endpoints require the `admin` role. They manage high-privilege organization settings.

```
GET    /api/v1/roles           # List all roles
POST   /api/v1/roles           # Create new role
GET    /api/v1/roles/{id}      # Get role details
PATCH  /api/v1/roles/{id}      # Update role
DELETE /api/v1/roles/{id}      # Delete role
```

### Manager Endpoints (Admin or Manager)

These endpoints require either `admin` or `manager` role. They manage employees and teams.

```
# Employees
GET    /api/v1/employees              # List employees
POST   /api/v1/employees              # Create employee
GET    /api/v1/employees/{id}         # Get employee details
PATCH  /api/v1/employees/{id}         # Update employee
DELETE /api/v1/employees/{id}         # Delete employee

# Teams
GET    /api/v1/teams                  # List teams
POST   /api/v1/teams                  # Create team
GET    /api/v1/teams/{id}             # Get team details
PATCH  /api/v1/teams/{id}             # Update team
DELETE /api/v1/teams/{id}             # Delete team
```

## How It Works

### Middleware Chain

```
Request → JWTAuth → RequireRole → Handler
              ↓           ↓
         Extracts    Checks role
         employee    against
         & role_id   allowed list
```

1. **JWTAuth middleware** validates the JWT token and loads session data (including `role_id`) into the request context.

2. **RequireRole middleware** queries the role from the database and checks if the role name matches any allowed role.

3. **Handler** executes if authorized, otherwise 403 is returned.

### Authorization Flow

```go
// Example: Admin-only endpoint
r.Group(func(r chi.Router) {
    r.Use(authmiddleware.RequireRole(queries, "admin"))
    r.Get("/roles", rolesHandler.ListRoles)
})

// Example: Admin or Manager endpoint
r.Group(func(r chi.Router) {
    r.Use(authmiddleware.RequireRole(queries, "admin", "manager"))
    r.Get("/employees", employeesHandler.ListEmployees)
})
```

## Error Responses

### 401 Unauthorized

Returned when:
- No JWT token provided
- JWT token is invalid or expired
- Session not found in database

```json
{
  "error": "Unauthorized"
}
```

### 403 Forbidden

Returned when:
- User's role doesn't match required roles
- User's role not found in database

```json
{
  "error": "Insufficient permissions"
}
```

## Audit Logging

All authorization failures are logged for security auditing:

```
AUTHZ_FAIL: Employee <id> (role: developer) attempted GET /roles, requires: [admin]
```

These logs can be used to:
- Detect unauthorized access attempts
- Monitor for suspicious activity
- Audit compliance

## Adding Authorization to New Endpoints

### Step 1: Determine Access Level

- **Any authenticated**: Use only `JWTAuth` middleware
- **Admin only**: Add `RequireRole(queries, "admin")`
- **Manager or Admin**: Add `RequireRole(queries, "admin", "manager")`

### Step 2: Add Middleware

```go
// In cmd/server/main.go

// Admin only
r.Group(func(r chi.Router) {
    r.Use(authmiddleware.RequireRole(queries, "admin"))
    r.Post("/new-admin-endpoint", handler.NewAdminEndpoint)
})

// Manager + Admin
r.Group(func(r chi.Router) {
    r.Use(authmiddleware.RequireRole(queries, "admin", "manager"))
    r.Post("/new-manager-endpoint", handler.NewManagerEndpoint)
})
```

### Step 3: Test Authorization

Write tests that verify:
1. Authorized roles can access the endpoint
2. Unauthorized roles receive 403
3. Unauthenticated requests receive 401

## Future Improvements

### Phase 2: Permission-Based Checks

Fine-grained authorization using permission strings:

```go
// Future: RequirePermission middleware
r.Use(authmiddleware.RequirePermission(queries, "employees:create"))
```

Permissions stored in role's `permissions` JSONB column:
```json
["employees:read", "employees:create", "teams:manage"]
```

### Phase 3: Contextual Authorization

Complex authorization rules like:
- "Edit own profile" (any employee)
- "Manage own team members" (team leads)
- Resource ownership checks

## Related Documentation

- [Database Schema](../database/schema-reference.md) - roles and employees tables
- [API Spec](../../platform/api-spec/spec.yaml) - endpoint definitions
- [Testing Guide](../development/testing.md) - writing authorization tests
