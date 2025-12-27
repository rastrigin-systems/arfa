---
sidebar_position: 3
---

# Multi-Tenancy

Arfa uses Row-Level Security (RLS) for strict data isolation between organizations.

## Implementation Layers

### Layer 1: Application Filtering

Every SQL query must include `org_id`:

```go
// ✅ CORRECT
employees, err := db.ListEmployees(ctx, orgID, status)

// ❌ WRONG - data leakage
employees, err := db.ListAllEmployees(ctx)
```

### Layer 2: sqlc Enforcement

sqlc generates functions requiring org_id:

```go
func (q *Queries) ListEmployees(ctx context.Context, orgID uuid.UUID) ([]Employee, error)
```

### Layer 3: PostgreSQL RLS

Database-level enforcement:

```sql
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;

CREATE POLICY employees_org_isolation ON employees
    FOR ALL
    USING (org_id = current_setting('app.current_org_id')::uuid);
```

### Layer 4: Middleware Context

Set org context per request:

```go
func RLSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        orgID := extractOrgID(r)
        db.ExecContext(ctx, "SET LOCAL app.current_org_id = $1", orgID)
        next.ServeHTTP(w, r)
    })
}
```

## Security Audit

When reviewing code, check:

1. ✅ Every SQL query includes `org_id` parameter
2. ✅ No direct table scans without WHERE clause
3. ✅ Foreign keys reference correct organization
4. ✅ RLS policies enabled on all tenant tables
