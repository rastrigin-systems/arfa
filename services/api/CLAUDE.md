# API Server Development Guide

**You are working on the Ubik API Server** - the REST API backend for the Ubik Enterprise platform.

---

## Quick Context

**What is this service?**
Multi-tenant REST API providing authentication, organization management, agent configuration, and WebSocket support for real-time config sync.

**Key capabilities:**
- JWT-based authentication with session management
- Multi-tenant organization, team, employee management
- AI agent configuration management
- MCP server configuration
- Approval workflows
- Activity logging and usage analytics
- WebSocket for real-time updates

---

## Essential Commands

```bash
# From services/api/ directory
make build              # Build server binary to ../../bin/ubik-server
make test               # Run all tests with coverage
make test-unit          # Unit tests only (fast)
make test-integration   # Integration tests (requires Docker)
make coverage           # View coverage report

# From repository root
make db-up              # Start PostgreSQL
make db-reset           # Reset database (⚠️ deletes data)
make generate           # Regenerate all code (after schema/API changes)
make generate-api       # Regenerate API code only
make generate-db        # Regenerate database code only

# Docker testing
make docker-build       # Build Docker image
make docker-test        # Verify image contents
make docker-run         # Run container locally
```

---

## Code Generation

**CRITICAL: NEVER edit generated files!**

### Source Files (Edit These)
- `../../platform/api-spec/spec.yaml` - OpenAPI specification (API contract)
- `../../platform/database/schema.sql` - PostgreSQL schema
- `../../platform/database/sqlc/queries/*.sql` - SQL queries

### Generated Files (Never Edit)
- `../../generated/api/` - API types, Chi server stubs
- `../../generated/db/` - Type-safe database code (sqlc)
- `../../generated/mocks/` - Test mocks

### Workflow
```bash
# 1. Edit source files
vim ../../platform/api-spec/spec.yaml
vim ../../platform/database/sqlc/queries/employees.sql

# 2. Regenerate from repository root
cd ../.. && make generate

# 3. Implement handlers using generated code
cd services/api
vim internal/handlers/employees.go
```

**IMPORTANT:** CI/CD regenerates code automatically. If docs are stale, CI FAILS.

---

## Architecture

### Request Flow
```
HTTP Request → Chi Router → Middleware → Handler → Service → Database → Response
                              ↓
                       (auth, logging, cors)
```

### Directory Structure
```
services/api/
├── cmd/server/main.go      # Entry point, route wiring
├── internal/
│   ├── handlers/           # HTTP request handlers (39 endpoints)
│   ├── middleware/         # Auth, logging, CORS, org context
│   ├── auth/               # JWT token generation/validation
│   ├── database/           # Database connection, migrations
│   ├── service/            # Business logic layer
│   └── websocket/          # WebSocket hub and handlers
└── tests/
    ├── integration/        # Full API flow tests (testcontainers)
    └── testutil/           # Test fixtures, helpers
```

### Layer Responsibilities

**Handlers** (`internal/handlers/`):
- Parse HTTP request (params, body)
- Call service layer
- Return HTTP response
- Handle HTTP-specific errors

**Service** (`internal/service/`):
- Business logic
- Transaction management
- Database queries via generated code
- Domain validations

**Database** (`../../generated/db/`):
- Type-safe SQL queries (sqlc generated)
- MUST be org-scoped for multi-tenancy

---

## Multi-Tenancy

**CRITICAL: ALL queries MUST be organization-scoped**

```go
// ✅ GOOD - Org-scoped query
employees, err := queries.ListEmployees(ctx, db.ListEmployeesParams{
    OrgID:  orgID,
    Status: "active",
})

// ❌ BAD - Exposes all organizations!
employees, err := queries.ListAllEmployees(ctx)
```

**Enforcement:**
1. Middleware extracts `org_id` from JWT
2. Adds to request context
3. Handlers retrieve from context
4. Pass to all database queries

**Row-Level Security (RLS):**
PostgreSQL RLS policies provide safety net, but queries MUST explicitly include `org_id`.

**See [../../docs/DATABASE.md](../../docs/DATABASE.md#multi-tenancy) for RLS details.**

---

## Testing Strategy

**CRITICAL: ALWAYS follow strict TDD (Test-Driven Development)**

### TDD Workflow (Mandatory)
1. ✅ Write failing test FIRST
2. ✅ Implement minimal code to pass test
3. ✅ Refactor with tests passing
4. ❌ NEVER write implementation before tests

### Test Types

**Unit Tests** (`internal/*_test.go`):
- Test individual functions, handlers
- Mock external dependencies (database, HTTP calls)
- Fast execution (<1s)
- Target: 85%+ coverage

**Integration Tests** (`tests/integration/`):
- Test full API flows with real database
- Use testcontainers for PostgreSQL
- Test multi-tenant isolation
- Test transaction rollbacks
- Slower execution (~10-30s)

### Running Tests

```bash
# Fast feedback loop
make test-unit          # ~1-2 seconds

# Full test suite
make test               # ~30-60 seconds

# Integration tests only
make test-integration   # ~20-40 seconds

# Coverage report
make coverage           # Opens HTML report
```

### Test Patterns

**Handler tests with mocks:**
```go
func TestListEmployees(t *testing.T) {
    // Setup
    ctrl := gomock.NewController(t)
    mockDB := mock_db.NewMockQuerier(ctrl)

    // Expect
    mockDB.EXPECT().
        ListEmployees(gomock.Any(), gomock.Any()).
        Return([]db.Employee{{...}}, nil)

    // Execute
    handler := NewEmployeeHandler(mockDB)
    resp := handler.ListEmployees(req)

    // Assert
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

**Integration tests with testcontainers:**
```go
func TestEmployeeIntegration(t *testing.T) {
    // Setup real database
    db := testutil.SetupTestDB(t)
    defer db.Close()

    // Seed test data
    testutil.SeedOrganization(t, db, orgID)

    // Test API flow
    resp := testutil.CallAPI(t, "POST", "/api/v1/employees", body)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    // Verify database state
    employees := testutil.QueryEmployees(t, db, orgID)
    assert.Len(t, employees, 1)
}
```

**See [../../docs/TESTING.md](../../docs/TESTING.md) for complete testing guide.**

---

## Common Tasks

### Adding New Endpoint

1. **Update OpenAPI spec:**
   ```yaml
   # ../../platform/api-spec/spec.yaml
   paths:
     /api/v1/employees/{id}:
       get:
         summary: Get employee by ID
         parameters: [...]
         responses: [...]
   ```

2. **Regenerate API code:**
   ```bash
   cd ../.. && make generate-api
   ```

3. **Write handler tests:**
   ```go
   // internal/handlers/employees_test.go
   func TestGetEmployee(t *testing.T) {
       // Write failing test FIRST
   }
   ```

4. **Implement handler:**
   ```go
   // internal/handlers/employees.go
   func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
       // Implement to pass tests
   }
   ```

5. **Wire route:**
   ```go
   // cmd/server/main.go
   r.Get("/api/v1/employees/{id}", employeeHandler.GetEmployee)
   ```

6. **Run tests:**
   ```bash
   make test
   ```

### Adding Database Query

1. **Write SQL query:**
   ```sql
   -- ../../platform/database/sqlc/queries/employees.sql
   -- name: GetEmployeeByID :one
   SELECT * FROM employees
   WHERE id = $1 AND org_id = $2;
   ```

2. **Regenerate database code:**
   ```bash
   cd ../.. && make generate-db
   ```

3. **Use in service layer:**
   ```go
   employee, err := queries.GetEmployeeByID(ctx, db.GetEmployeeByIDParams{
       ID:    employeeID,
       OrgID: orgID,
   })
   ```

### Schema Change

1. **Update schema:**
   ```bash
   vim ../../platform/database/schema.sql
   ```

2. **Reset database (⚠️ deletes data):**
   ```bash
   cd ../.. && make db-reset
   ```

3. **Regenerate all code:**
   ```bash
   make generate
   ```

4. **Update affected queries and handlers**

---

## Common Pitfalls

### 1. Missing Org Scoping
```go
// ❌ BAD - No org_id check
employees, err := queries.ListEmployees(ctx)

// ✅ GOOD - Org-scoped
employees, err := queries.ListEmployees(ctx, db.ListEmployeesParams{
    OrgID: orgID,
})
```

### 2. Editing Generated Files
```bash
# ❌ NEVER edit these:
generated/api/server.gen.go
generated/db/queries.sql.go

# ✅ Edit source files instead:
platform/api-spec/spec.yaml
platform/database/sqlc/queries/*.sql
```

### 3. Not Running Tests
```bash
# ✅ ALWAYS run tests before committing
make test
```

### 4. Stale Generated Code
```bash
# ✅ After pulling changes
cd ../.. && make generate
cd services/api && make build
```

### 5. Docker Issues
```bash
# ✅ Test Docker builds locally
cd ../.. && docker build -f services/api/Dockerfile.gcp -t test .
docker run --rm test ls -la /app/platform/api-spec/
```

**See [../../docs/DEBUGGING.md](../../docs/DEBUGGING.md) for debugging strategies.**

---

## Debugging

**Golden Rule: Check the data, not just the code**

### Quick Debug Checklist
1. ✅ Add request/response logging
2. ✅ Check database state (foreign keys, seed data)
3. ✅ Verify org-scoping in queries
4. ✅ Check for stale binaries
5. ✅ Rebuild: `make clean && make build`

### Logging
```go
// Add temporary debug logging
log.Printf("Request params: %+v", params)
log.Printf("Database result: %+v", result)
log.Printf("Org ID from context: %s", orgID)
```

### Database Inspection
```bash
# Connect to database
make db-connect  # From root

# Or directly
psql -U ubik -h localhost -d ubik

# Check data
SELECT * FROM employees WHERE org_id = 'org-uuid';
```

---

## Docker & Deployment

### Local Docker Testing

**MANDATORY before deploying:**
```bash
# 1. Build image (from root)
docker build -f services/api/Dockerfile.gcp -t ubik-api-test .

# 2. Verify files in image
docker run --rm ubik-api-test ls -la /app/
docker run --rm ubik-api-test ls -la /app/platform/api-spec/

# 3. Test container
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://ubik:ubik_dev_password@host.docker.internal:5432/ubik?sslmode=disable" \
  ubik-api-test

# 4. Verify endpoints
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/docs/
```

### GCP Deployment

```bash
# Deploy to Cloud Run (from root)
gcloud builds submit --config=cloudbuild-api.yaml

# This will:
# 1. Build Docker image with Cloud Build
# 2. Push to Artifact Registry
# 3. Deploy to Cloud Run
```

**See [../../docs/DOCKER_TESTING_CHECKLIST.md](../../docs/DOCKER_TESTING_CHECKLIST.md) for complete guide.**

---

## Related Documentation

**Root Documentation:**
- [../../CLAUDE.md](../../CLAUDE.md) - Monorepo overview, critical rules
- [../../docs/QUICKSTART.md](../../docs/QUICKSTART.md) - First-time setup
- [../../docs/QUICK_REFERENCE.md](../../docs/QUICK_REFERENCE.md) - Command reference

**Development:**
- [../../docs/DEVELOPMENT.md](../../docs/DEVELOPMENT.md) - Development workflow
- [../../docs/DEV_WORKFLOW.md](../../docs/DEV_WORKFLOW.md) - PR workflow (mandatory)
- [../../docs/TESTING.md](../../docs/TESTING.md) - Complete testing guide
- [../../docs/DEBUGGING.md](../../docs/DEBUGGING.md) - Debugging strategies

**Database:**
- [../../docs/DATABASE.md](../../docs/DATABASE.md) - Database operations
- [../../docs/ERD.md](../../docs/ERD.md) - Visual schema

**Other Services:**
- [../cli/CLAUDE.md](../cli/CLAUDE.md) - CLI client development
- [../web/CLAUDE.md](../web/CLAUDE.md) - Web UI development

---

**Quick Links:**
- API Docs (local): http://localhost:8080/api/docs
- Adminer (DB UI): http://localhost:8081
- Database: `postgres://ubik:ubik_dev_password@localhost:5432/ubik`
