# Development Guide — Pivot Enterprise Platform

**Essential development workflow and best practices**

---

## Table of Contents
1. [Development Workflow](#development-workflow)
2. [Code Generation](#code-generation)
3. [TDD Strategy](#tdd-strategy)
4. [Common Gotchas](#common-gotchas)
5. [Best Practices](#best-practices)

---

## Development Workflow

### Making Changes

```bash
# 1. Update database schema
vim shared/schema/schema.sql

# 2. Apply to database
make db-reset

# 3. Update OpenAPI spec (if API changes)
vim openapi/spec.yaml

# 4. Update SQL queries (if needed)
vim sqlc/queries/employees.sql

# 5. Regenerate all code
make generate

# 6. Implement handlers
vim internal/handlers/employees.go

# 7. Run tests
make test

# 8. Build and test locally
go run cmd/server/main.go
```

### Code Generation Pipeline

```
shared/schema/schema.sql → PostgreSQL → tbls → schema.json, README.md, public.*.md, schema.svg
                        ↓         ↓
                       sqlc      Python script → ERD.md (user-friendly)
                        ↓
                  generated/db/*.go

openapi/spec.yaml → oapi-codegen → generated/api/server.gen.go

Your code (internal/) → Uses generated types
```

**ERD Documentation:**
- `make generate-erd` creates **both** README.md (tbls) and ERD.md (custom script)
- README.md = Technical reference with table index
- ERD.md = User-friendly overview with categories

---

## Code Generation

### Manual Code Generation (No Git Hooks)

**As of v0.3.0, code generation is NO LONGER automatic on commit.**

**Two sources of truth:**
1. **shared/schema/schema.sql** - Database structure
2. **openapi/spec.yaml** - API contract

These are maintained separately because:
- DB tables ≠ API DTOs (different concerns)
- DB can have more tables than API exposes
- API can aggregate/transform DB data

### Generation Commands

```bash
# Backend (Go) - Generate everything
make generate

# Backend (Go) - Generate specific parts
make generate-erd        # Database documentation
make generate-api        # API types + router
make generate-db         # Database code
make generate-mocks      # Test mocks

# Frontend (Web) - Generate TypeScript types
cd services/web
npm run generate:api     # Generate from OpenAPI spec
```

### When to Regenerate

**Run `make generate` after:**
- Changing `shared/schema/schema.sql` (then run `make db-reset`)
- Changing `openapi/spec.yaml`
- Changing SQL queries in `sqlc/queries/`
- Pulling changes that modify any of the above

**Important Notes:**
- ⚠️ Never edit generated files - they are completely overwritten!
  - `generated/` directory (Go code)
  - `services/web/lib/api/schema.ts` (TypeScript types)
  - `docs/` directory (ERD documentation)
- ✅ Generated code is NOT committed to git (`generated/`, `services/web/lib/api/schema.ts`)
- ✅ Generated docs ARE committed to git (`docs/` - for GitHub visibility)
- ✅ CI/CD automatically regenerates everything and fails if docs are stale
- ✅ This ensures consistency between local and CI environments

### Committing Schema Changes

**Complete workflow:**
```bash
# 1. Edit schema
vim shared/schema/schema.sql

# 2. Reset database and regenerate everything
make db-reset
make generate  # Generates code + docs

# 3. Commit source AND docs
git add shared/schema/schema.sql docs/
git commit -m "feat: Add notifications table"
git push
```

**CI will verify:**
- ✅ Go code regenerates correctly
- ✅ ERD docs are up to date (fails if you forgot `make generate-erd`)
- ✅ TypeScript types regenerate correctly

### Why No Git Hooks?

**Benefits of manual generation:**
- ✅ Faster commits (no 5-10 second hook delay)
- ✅ Simpler setup (no tool installation required immediately)
- ✅ Cleaner git history (no generated code in commits)
- ✅ Fewer merge conflicts (generated code doesn't conflict)
- ✅ Easier PR reviews (only source changes visible)
- ✅ CI reliability (code always freshly generated)

---

## TDD Strategy

### When to Use TDD (Write Tests First)

✅ **Use TDD for:**
- Complex business logic
- Security-critical features (auth, permissions)
- Edge cases and error handling
- Bug fixes (write failing test, then fix)

### When to Use Implementation-First

✅ **Use Implementation-First for:**
- Simple CRUD operations
- When contract is already defined (OpenAPI spec)
- Type conversion/mappers
- Basic validation

### Hybrid Approach (Recommended)

**For new endpoints:**
1. **Implement handler** (30 min) - Follow OpenAPI contract
2. **Write unit tests immediately** (30 min) - Validate behavior
3. **Add integration test** (20 min) - Test full stack
4. **Refactor if needed** (10 min) - With test safety net

**Example:**
```go
// 1. Implement (following OpenAPI spec)
func (h *Handler) GetEmployee(w http.ResponseWriter, r *http.Request) {
    // Basic implementation
}

// 2. Write unit tests
func TestGetEmployee_Success(t *testing.T) { ... }
func TestGetEmployee_NotFound(t *testing.T) { ... }
func TestGetEmployee_Unauthorized(t *testing.T) { ... }

// 3. Integration test
func TestGetEmployee_Integration(t *testing.T) { ... }

// 4. Refactor with confidence
```

---

## Common Gotchas

### 1. Mock Generation Version Mismatch

```bash
# ❌ Wrong: Old mockgen
go install github.com/golang/mock/mockgen@latest

# ✅ Correct: New mockgen
go install go.uber.org/mock/mockgen@latest
```

Always use `go.uber.org/mock` (matches import in code).

### 2. Forgetting to Regenerate After Schema Changes

**Problem**: SQL query fails because column doesn't exist

**Solution**:
```bash
# After changing shared/schema/schema.sql
make db-reset              # Apply to database
make generate-db           # Regenerate DB code
make generate-mocks        # Regenerate mocks
go test ./...              # Run tests to verify
```

### 3. Schema Drift

**Problem**: Database schema doesn't match OpenAPI spec

**Solution**: Use consistent naming and validation
```go
// OpenAPI spec
type Employee struct {
    Email string `json:"email"`
}

// Database query
-- name: GetEmployee :one
SELECT email FROM employees WHERE id = $1;

// Mapper layer bridges differences
func DBEmployeeToAPI(db db.Employee) api.Employee { ... }
```

### 4. Multi-Tenant Data Leakage

**Problem**: Forgetting org_id filtering

**Solution**: Always scope queries by organization
```go
// ❌ BAD - Returns all employees across all orgs!
employees, err := db.ListAllEmployees(ctx)

// ✅ GOOD - Scoped to organization
employees, err := db.ListEmployees(ctx, orgID)
```

Use Row-Level Security (RLS) as additional safety net.

### 5. Testing Against Real Database

**Problem**: Tests fail locally but shared/schema/schema.sql is correct

**Solution**: Check testcontainers path
```go
schemaPath, err := filepath.Abs("../../shared/schema/schema.sql")
if err != nil {
    t.Fatal(err)
}
```

---

## Best Practices

### 1. Always Test Immediately

**Don't do this:**
```
❌ Implement 5 handlers → Write tests later
```

**Do this:**
```
✅ Implement 1 handler → Write tests → Next handler
```

### 2. Use Type-Safe SQL

**sqlc generates type-safe Go code from SQL:**

```sql
-- sqlc/queries/employees.sql
-- name: GetEmployee :one
SELECT * FROM employees WHERE id = $1 AND org_id = $2;

-- name: ListEmployees :many
SELECT * FROM employees
WHERE org_id = $1
  AND ($2::varchar IS NULL OR status = $2)
  AND ($3::uuid IS NULL OR team_id = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;
```

**Generated code is type-safe:**
```go
// Auto-generated
func (q *Queries) GetEmployee(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (Employee, error)
func (q *Queries) ListEmployees(ctx context.Context, arg ListEmployeesParams) ([]Employee, error)
```

### 3. Separate DB and API Types

**Don't expose database types directly:**

```go
// ❌ BAD - Leaks password_hash
func GetEmployee(w http.ResponseWriter, r *http.Request) {
    emp, _ := db.GetEmployee(ctx, id)
    json.NewEncoder(w).Encode(emp) // Includes password_hash!
}

// ✅ GOOD - Use mapper layer
func GetEmployee(w http.ResponseWriter, r *http.Request) {
    emp, _ := db.GetEmployee(ctx, id)
    apiEmp := mapper.DBEmployeeToAPI(emp) // Excludes password_hash
    json.NewEncoder(w).Encode(apiEmp)
}
```

### 4. Integration Tests Validate Schema

**Unit tests with mocks don't catch:**
- Missing database columns
- Type mismatches
- SQL syntax errors
- Constraint violations

**Integration tests catch all of these!**

### 5. Use Table-Driven Tests

```go
tests := []struct {
    name           string
    input          RequestBody
    setupMock      func(*mocks.MockQuerier)
    expectedStatus int
    expectedBody   string
}{
    {name: "success", ...},
    {name: "validation error", ...},
    {name: "not found", ...},
    {name: "unauthorized", ...},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test implementation
    })
}
```

---

## Daily Workflow

### Starting a New Feature

1. **Check IMPLEMENTATION_ROADMAP.md** - What to build next
2. **Update OpenAPI spec** - Define contract
3. **Add SQL queries** - Define data access
4. **Run `make generate`** - Generate types
5. **Implement handler** - Follow generated types
6. **Write tests** - Unit + integration
7. **Run `make test`** - Verify all pass
8. **Commit** - With descriptive message

### Pre-Commit Checklist

```bash
# 1. All code generated
make generate

# 2. All tests pass
go test ./...

# 3. No race conditions
go test -race ./...

# 4. Coverage acceptable
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1

# 5. Code formatted
go fmt ./...

# 6. Build succeeds
go build ./...
```

---

## Project Structure

```
ubik-enterprise/
├── shared/schema/schema.sql                 # Database schema (source of truth)
├── openapi/spec.yaml          # API contract (source of truth)
│
├── generated/                 # ⚠️ AUTO-GENERATED (never edit!)
│   ├── api/                   # From OpenAPI
│   ├── db/                    # From SQL
│   └── mocks/                 # From interfaces
│
├── internal/                  # Your code goes here
│   ├── handlers/              # HTTP handlers
│   ├── middleware/            # Auth, RLS, logging
│   ├── mapper/                # Type conversion
│   └── validation/            # Custom validators
│
├── tests/
│   ├── integration/           # Full stack tests
│   └── testutil/              # Test helpers
│
└── cmd/
    └── server/                # Main application
```

---

## Summary

✅ **Use hybrid TDD approach** - Implementation-first for simple, TDD for complex
✅ **Always regenerate after changes** - `make generate`
✅ **Write tests immediately** - Don't batch them
✅ **Integration tests catch schema issues** - Essential safety net
✅ **Follow OpenAPI contract** - Generated types guide implementation
✅ **Multi-tenant everything** - Always scope by org_id

**Target**: 85% test coverage, all tests passing, zero race conditions

---

**See also:**
- [docs/TESTING.md](./TESTING.md) - Complete testing guide
- [IMPLEMENTATION_ROADMAP.md](../IMPLEMENTATION_ROADMAP.md) - Next endpoints to build
- [docs/ERD.md](./ERD.md) - Database schema reference
