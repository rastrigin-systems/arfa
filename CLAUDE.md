# Ubik Enterprise — AI Agent Management Platform

## System Purpose

Multi-tenant SaaS platform for companies to centrally manage AI agent (Claude Code, Cursor, Windsurf, etc.) and MCP server configurations for their employees.

**Core Value**: Centralized control, policy enforcement, and visibility into AI agent usage across an organization.

**Current Status**: 🟢 **Phase 2 - Authentication Complete** - Full auth system with 33 passing tests (as of 2025-10-28)

**⭐ NEXT STEPS**: See **[IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md)** for priority order of next endpoints to implement

---

## Quick Start

```bash
# Start database
cd pivot
make db-up

# Install tools (one-time)
make install-tools

# Generate all code
make generate

# View documentation
open docs/ERD.md
open docs/INDEX.md
```

**See [QUICKSTART.md](./docs/setup/QUICKSTART.md) for complete setup instructions.**

---

## What This Platform Does

### For Companies
- ✅ Manage employees, teams, and roles
- ✅ Control which AI agents employees can use (Claude Code, Cursor, Windsurf, Continue, Copilot)
- ✅ Configure MCP servers and access per employee
- ✅ Set usage policies (path restrictions, rate limits, cost limits)
- ✅ Approve/reject employee requests for new agents or MCPs
- ✅ Track usage, costs, and activity across the organization
- ✅ Enforce compliance and security policies

### For Employees
- ✅ Sync agent configurations to local machines via CLI
- ✅ Request access to new agents or MCP servers
- ✅ View their assigned agents and policies
- ✅ Use AI agents with centrally-managed configurations

---

## Architecture Overview

```
PostgreSQL Schema (DB source of truth)
    ↓
    ├─→ tbls → ERD documentation (auto-generated)
    └─→ sqlc → Type-safe Go database code

OpenAPI Spec (API source of truth)
    ↓
    └─→ oapi-codegen → Go API types + Chi server

Employee CLI Client (future)
    ↓
    └─→ Syncs configs from central server
```

**Hybrid Approach**: Database schema and API spec maintained separately, both generate code automatically.

---

## 📚 Documentation Map

### ⭐ MOST IMPORTANT - START HERE
- **[IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md)** ⭐⭐⭐ - **PRIORITY ORDER for next endpoints** with detailed implementation plans, TDD workflow, and success criteria

### 🚀 Getting Started
- **[QUICKSTART.md](./docs/setup/QUICKSTART.md)** - 5-minute setup guide
- **[README.md](./README.md)** - Project overview and common commands
- **[INIT_COMPLETE.md](./docs/setup/INIT_COMPLETE.md)** - Phase 1 completion summary
- **[SETUP_COMPLETE.md](./docs/setup/SETUP_COMPLETE.md)** - What's done and what's next
- **[DOCUMENTATION_COMPLETE.md](./docs/setup/DOCUMENTATION_COMPLETE.md)** - Documentation overview

### 📋 Planning & Architecture
- **[MIGRATION_PLAN.md](./docs/planning/MIGRATION_PLAN.md)** - Complete 10-week roadmap
- **[DATABASE_SCHEMA.md](./docs/planning/DATABASE_SCHEMA.md)** - Original ERD with table definitions

### 📊 Database Documentation
- **[docs/ERD.md](./docs/ERD.md)** ⭐ - Mermaid ERD diagram (start here!)
- **[docs/INDEX.md](./docs/INDEX.md)** - Complete documentation index
- **[docs/README.md](./docs/README.md)** - Auto-generated table index
- **[docs/schema.json](./docs/schema.json)** - Machine-readable schema

### 🧪 Testing & Development
- **[docs/DEVELOPMENT_APPROACH.md](./docs/DEVELOPMENT_APPROACH.md)** ⭐ - TDD vs Implementation-First strategy + Week 1 schedule
- **[docs/TESTING_QUICKSTART.md](./docs/TESTING_QUICKSTART.md)** ⭐ - 5-minute testing setup guide
- **[docs/TESTING_STRATEGY.md](./docs/TESTING_STRATEGY.md)** - Complete testing guide with mock generation
- **[docs/TESTING_ANALYSIS.md](./docs/TESTING_ANALYSIS.md)** - Codebase testing analysis

### 🔧 Configuration Files
- **[openapi/spec.yaml](./openapi/spec.yaml)** - OpenAPI 3.0 specification
- **[sqlc/sqlc.yaml](./sqlc/sqlc.yaml)** - sqlc configuration
- **[sqlc/queries/*.sql](./sqlc/queries/)** - Type-safe SQL queries
- **[docker-compose.yml](./docker-compose.yml)** - Local development environment
- **[Makefile](./Makefile)** - Automation commands

---

## Project Structure

```
pivot/
├── CLAUDE.md                  # This file - main documentation hub
├── README.md                  # Quick reference
├── schema.sql                 # PostgreSQL schema (20 tables + 3 views)
├── Makefile                   # Automation commands
├── docker-compose.yml         # Local environment
├── go.mod                     # Go dependencies
│
├── docs/                      # All documentation
│   ├── planning/              # Planning documents
│   │   ├── MIGRATION_PLAN.md  # 10-week roadmap
│   │   └── DATABASE_SCHEMA.md # Original ERD documentation
│   ├── setup/                 # Setup guides
│   │   ├── QUICKSTART.md      # 5-minute setup guide
│   │   ├── INIT_COMPLETE.md   # Phase 1 completion
│   │   ├── SETUP_COMPLETE.md  # What's done/next
│   │   └── DOCUMENTATION_COMPLETE.md # Docs overview
│   ├── ERD.md                 # ⭐ Mermaid ERD (start here!)
│   ├── INDEX.md               # Documentation index
│   ├── README.md              # Auto-generated table index
│   ├── schema.svg             # Full schema diagram
│   ├── schema.json            # Machine-readable schema
│   └── public.*.md            # Table docs (24 files)
│
├── openapi/
│   ├── spec.yaml              # OpenAPI 3.0.3 spec (API source of truth)
│   └── oapi-codegen.yaml      # Generator config
│
├── sqlc/
│   ├── sqlc.yaml              # Generator config
│   └── queries/
│       ├── employees.sql      # Employee CRUD
│       ├── auth.sql           # Sessions
│       └── organizations.sql  # Org/team/roles
│
├── generated/                 # ⚠️ Auto-generated (don't edit!)
│   ├── api/                   # From OpenAPI spec
│   │   └── server.gen.go
│   └── db/                    # From SQL queries
│       ├── models.go
│       ├── employees.sql.go
│       └── ...
│
├── internal/                  # Your code goes here
│   ├── handlers/              # HTTP handlers
│   ├── service/               # Business logic
│   ├── middleware/            # Auth, RLS, logging
│   ├── mapper/                # Type conversion
│   └── validation/            # Custom validators
│
├── cmd/
│   ├── server/                # API server
│   └── cli/                   # Employee CLI (future)
│
└── scripts/                   # Utility scripts
```

---

## Database Schema Summary

### 20 Tables + 3 Views

**Core Organization** (5 tables)
- `organizations` - Top-level tenants
- `subscriptions` - Billing and budgets
- `teams` - Group employees
- `roles` - Permission definitions
- `employees` - User accounts

**Agent Management** (7 tables)
- `agent_catalog` - Available AI agents
- `tools` - Tool registry (fs, git, http, shell, docker)
- `policies` - Usage policies and restrictions
- `agent_tools` - Agent ↔ Tool mapping
- `agent_policies` - Agent ↔ Policy mapping
- `team_policies` - Team-specific overrides
- `employee_agent_configs` - Per-employee agent instances

**MCP Configuration** (3 tables)
- `mcp_categories` - Organize MCP servers
- `mcp_catalog` - Available MCP servers
- `employee_mcp_configs` - Per-employee MCP instances

**Authentication** (1 table)
- `sessions` - JWT session tracking

**Approvals** (2 tables)
- `agent_requests` - Employee requests
- `approvals` - Manager approval workflow

**Analytics** (2 tables)
- `activity_logs` - Audit trail
- `usage_records` - Cost and resource tracking

**Views** (3)
- `v_employee_agents` - Employee agents with details
- `v_employee_mcps` - Employee MCPs with details
- `v_pending_approvals` - Approval queue

**See [docs/ERD.md](./docs/ERD.md) for complete visual schema.**

---

## Key Features

### ✅ Phase 1: Foundation (COMPLETE)
1. **Database Schema** - PostgreSQL with 20 tables, RLS, seed data
2. **Code Generation** - oapi-codegen, sqlc, tbls all configured
3. **OpenAPI Spec** - Auth + Employee endpoints defined
4. **SQL Queries** - Type-safe queries for employees, auth, orgs
5. **Documentation** - 50+ docs including Mermaid ERD
6. **Automation** - Makefile with 20+ commands
7. **Local Environment** - Docker Compose with PostgreSQL + Adminer

### 📋 Phase 2: Core API (Next - Week of 2025-11-04)
1. **Authentication & Authorization** - JWT, sessions, RLS middleware
2. **Employee Management API** - CRUD endpoints
3. **Organization API** - Org, team, role management
4. **Integration Tests** - Full test coverage

### 📋 Phase 3-8: See [MIGRATION_PLAN.md](./docs/planning/MIGRATION_PLAN.md)

---

## Technology Stack

- **Language**: Go 1.24+
- **Database**: PostgreSQL 15+ (multi-tenant with RLS) - **20 tables + 3 views**
- **API Specification**: OpenAPI 3.0.3
- **Code Generation**: oapi-codegen, sqlc, tbls
- **HTTP Router**: Chi
- **Web UI**: Next.js 14 (future)
- **Testing**: testcontainers-go
- **Deployment**: Docker, Docker Compose

---

## Common Commands

```bash
# Database
make db-up              # Start PostgreSQL
make db-down            # Stop PostgreSQL
make db-reset           # Reset database (⚠️ deletes data)

# Code Generation
make generate           # Generate everything
make generate-erd       # Generate ERD docs only
make generate-api       # Generate API code only
make generate-db        # Generate DB code only

# Development
make dev                # Start dev server (once implemented)
make test               # Run tests
make build              # Build binaries
make clean              # Clean generated files

# Help
make help               # Show all commands
```

---

## Quick Reference

### Database Access
```bash
# PostgreSQL connection
postgres://pivot:pivot_dev_password@localhost:5432/pivot

# Adminer web UI
open http://localhost:8080

# psql CLI
docker exec pivot-postgres psql -U pivot -d pivot
```

### API Endpoints (OpenAPI)
```bash
# View spec
cat openapi/spec.yaml

# Once server is running:
# POST /api/v1/auth/login
# GET  /api/v1/auth/me
# GET  /api/v1/employees
# POST /api/v1/employees
# GET  /api/v1/employees/{id}
# PATCH /api/v1/employees/{id}
# DELETE /api/v1/employees/{id}
```

### Generated Code
```bash
# API types
cat generated/api/server.gen.go

# Database queries
cat generated/db/employees.sql.go

# Database models
cat generated/db/models.go
```

---

## Development Workflow

### Making Changes

```bash
# 1. Update database schema
vim schema.sql

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
schema.sql → PostgreSQL → tbls → docs/ERD.md, docs/*.md
                        ↓
                       sqlc → generated/db/*.go

openapi/spec.yaml → oapi-codegen → generated/api/server.gen.go

Your code (internal/) → Uses generated types
```

---

## Critical Usage Notes

### ✅ Hybrid Architecture

**Two Sources of Truth**:
1. **schema.sql** - Database structure
2. **openapi/spec.yaml** - API contract

These are maintained separately because:
- DB tables ≠ API DTOs (different concerns)
- DB can have more tables than API exposes
- API can aggregate/transform DB data

**Keep in sync manually** or use drift detection script (future).

### ✅ Generated Code

**Never edit files in `generated/`**:
- They are completely regenerated on `make generate`
- Add `.gitignore` entry to exclude them from commits
- Treat as read-only artifacts

### ✅ Multi-Tenancy

**All queries must be org-scoped**:
```go
// GOOD
db.ListEmployees(ctx, org_id, status)

// BAD
db.ListAllEmployees(ctx) // Exposes all orgs!
```

Use Row-Level Security (RLS) policies as safety net.

---

## 🧪 TDD Development Best Practices

### ✅ TDD Workflow (RED → GREEN → REFACTOR)

**CRITICAL: Always follow this sequence when implementing new features**

#### 1. Write Tests FIRST (RED Phase 🔴)

```bash
# Step 1: Write unit tests in *_test.go
vim internal/handlers/auth_test.go

# Step 2: Run tests - they should FAIL
go test -v -short ./internal/handlers

# Expected: Tests fail because handler doesn't exist yet
# This is GOOD! Confirms tests are working.
```

**Example Test Structure:**
```go
func TestLogin_Success(t *testing.T) {
    // === ARRANGE === Setup test data and mocks
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    mockDB := mocks.NewMockQuerier(ctrl)

    // Define expectations
    mockDB.EXPECT().GetEmployeeByEmail(gomock.Any(), email).Return(...)

    // === ACT === Call the handler
    handler := handlers.NewAuthHandler(mockDB)
    handler.Login(rec, req)

    // === ASSERT === Verify the results
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

#### 2. Implement Code (GREEN Phase 🟢)

```bash
# Step 1: Add SQL queries if needed
vim sqlc/queries/auth.sql

# Step 2: Regenerate DB code and mocks
make generate-db && make generate-mocks

# Step 3: Implement handler
vim internal/handlers/auth.go

# Step 4: Run tests - they should PASS
go test -v -short ./internal/handlers

# Expected: All tests pass ✅
```

#### 3. Add Integration Tests (FULL STACK 🔄)

```bash
# Step 1: Write integration test with REAL database
vim tests/integration/auth_integration_test.go

# Step 2: Run integration test
go test -v -run TestLogin_Integration ./tests/integration

# Expected: Test passes with real PostgreSQL ✅
```

**Integration Test Structure:**
```go
func TestLogin_Integration_Success(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Setup real database
    conn, queries := testutil.SetupTestDB(t)
    defer conn.Close(testutil.GetContext(t))

    // Create test data in REAL database
    org := testutil.CreateTestOrg(t, queries, ctx)
    employee := testutil.CreateTestEmployee(t, queries, ctx, params)

    // Make HTTP request to handler
    router := chi.NewRouter()
    authHandler := handlers.NewAuthHandler(queries)
    router.Post("/auth/login", authHandler.Login)

    router.ServeHTTP(rec, req)

    // Verify DATABASE side effects
    session, err := queries.GetSession(ctx, tokenHash)
    require.NoError(t, err, "Session should exist in database")
}
```

#### 4. Refactor (CLEAN UP 🧹)

```bash
# Step 1: Run all tests to ensure nothing breaks
go test -v ./...

# Step 2: Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Step 3: Refactor if needed, re-run tests
```

---

### ✅ Code Generation Workflow

**CRITICAL: Always regenerate code after schema/query changes**

```bash
# After changing schema.sql
make db-reset           # Reset database with new schema
make generate-db        # Regenerate sqlc code
make generate-mocks     # Regenerate mocks from new interfaces
make generate-erd       # Update ERD documentation

# Or regenerate everything at once
make generate           # Runs all generators
```

**Common Mistakes:**
- ❌ Forgetting to regenerate mocks after SQL query changes
- ❌ Editing generated code directly (always regenerate!)
- ❌ Not running `make db-reset` after schema changes

---

### ✅ Testing Patterns

#### Unit Tests (Fast, Mocked)

**When to use:** Testing handler logic, validation, error paths

```go
// Good: Tests ONE handler function with mocked database
func TestLogout_Success(t *testing.T) {
    mockDB.EXPECT().DeleteSession(gomock.Any(), tokenHash).Return(nil)
    handler.Logout(rec, req)
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

**Run with:** `go test -v -short ./internal/...`

**Coverage Target:** 90%+ for handlers, 95%+ for auth utilities

#### Integration Tests (Slow, Real DB)

**When to use:** Testing full stack with database side effects

```go
// Good: Tests ENTIRE flow with real PostgreSQL
func TestLogin_Integration_Success(t *testing.T) {
    // Real database, real handler, real HTTP request
    // Verify session created in database
    // Verify last_login updated
    // Verify JWT contains correct claims
}
```

**Run with:** `go test -v -run Integration ./tests/integration`

**Coverage Target:** 80%+ of critical user flows

#### What Integration Tests Catch (That Unit Tests Miss)

**Real Examples from Our Development:**
1. ✅ **Missing deleted_at column** - SQL query referenced non-existent column
2. ✅ **Type conversion issues** - OpenAPI types vs database types
3. ✅ **Schema drift** - SQL queries out of sync with schema
4. ✅ **Multi-tenant isolation** - Verified org_id filtering works

---

### ✅ Common Patterns

#### 1. JWT Authentication Flow

```go
// Pattern: Extract token from Authorization header
authHeader := r.Header.Get("Authorization")
if authHeader == "" {
    writeError(w, http.StatusUnauthorized, "Missing authorization header")
    return
}

const bearerPrefix = "Bearer "
if len(authHeader) < len(bearerPrefix) {
    writeError(w, http.StatusUnauthorized, "Invalid format")
    return
}
token := authHeader[len(bearerPrefix):]

// Verify JWT
claims, err := auth.VerifyJWT(token)
if err != nil {
    writeError(w, http.StatusUnauthorized, "Invalid token")
    return
}

// Hash for database lookup
tokenHash := auth.HashToken(token)
```

**Note:** This pattern appears in Logout, GetMe, and will appear in ALL protected endpoints → Extract to middleware!

#### 2. Type Conversion (DB ↔ API)

```go
// Pattern: Convert database types to OpenAPI types
empID := openapi_types.UUID(emp.ID)
orgID := openapi_types.UUID(emp.OrgID)
email := openapi_types.Email(emp.Email)

employee := api.Employee{
    Id:    &empID,
    OrgId: orgID,
    Email: email,
}

// Handle nullable fields
if emp.TeamID.Valid {
    teamID := openapi_types.UUID(emp.TeamID.Bytes)
    employee.TeamId = &teamID
}
```

#### 3. Test Fixture Creation

```go
// Pattern: Use testutil helpers for consistent test data
org := testutil.CreateTestOrg(t, queries, ctx)
role := testutil.CreateTestRole(t, queries, ctx, "Admin")
employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
    OrgID:        org.ID,
    RoleID:       role.ID,
    Email:        "test@example.com",
    PasswordHash: passwordHash,
    Status:       "active",
})
```

---

### ✅ Testing Commands

```bash
# Run all tests
go test -v ./...

# Run only unit tests (fast)
go test -v -short ./internal/...

# Run only integration tests (slow)
go test -v -run Integration ./tests/integration

# Run specific test
go test -v -run TestLogin_Success ./internal/handlers

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detection
go test -v -race ./...
```

---

### ✅ Critical Gotchas

#### 1. Mock Generation Version Mismatch

**Problem:** Using wrong mockgen version

```bash
# ❌ Wrong: Old mockgen (github.com/golang/mock)
go install github.com/golang/mock/mockgen@latest

# ✅ Correct: New mockgen (go.uber.org/mock)
go install go.uber.org/mock/mockgen@latest
```

**Always use:** `go.uber.org/mock` (matches import in code)

#### 2. SQL Query vs Schema Mismatch

**Problem:** Query references column that doesn't exist

```sql
-- Query references deleted_at
SELECT * FROM employees WHERE deleted_at IS NULL

-- But schema is missing the column!
-- ❌ Integration test will fail: "column deleted_at does not exist"
```

**Solution:** Always run integration tests to catch this!

```bash
make db-reset           # Apply schema changes
go test -v -run Integration ./tests/integration
```

#### 3. Testcontainers Docker Issues

**Problem:** Integration tests hang or fail to start

```bash
# Check Docker is running
docker ps

# Check testcontainers can access Docker
docker run hello-world

# Ensure schema.sql path is correct
schemaPath, err := filepath.Abs("../../schema.sql")
```

---

### ✅ File Organization

```
pivot/
├── internal/
│   ├── auth/
│   │   ├── jwt.go           # JWT utilities
│   │   └── jwt_test.go      # Unit tests (14 tests)
│   └── handlers/
│       ├── auth.go          # HTTP handlers
│       └── auth_test.go     # Unit tests (13 tests)
│
├── tests/
│   ├── integration/
│   │   └── auth_integration_test.go  # Integration tests (6 tests)
│   └── testutil/
│       ├── db.go            # Testcontainers setup
│       └── fixtures.go      # Test data factories
│
├── sqlc/queries/
│   ├── auth.sql             # Auth SQL queries
│   ├── employees.sql        # Employee SQL queries
│   └── organizations.sql    # Org SQL queries
│
└── generated/               # ⚠️ NEVER EDIT!
    ├── api/                 # From OpenAPI spec
    ├── db/                  # From sqlc
    └── mocks/               # From mockgen
```

---

### ✅ Test Success Criteria

**Before merging any PR:**

1. ✅ All tests passing: `go test ./...`
2. ✅ Coverage > 85%: `go test -coverprofile=coverage.out ./...`
3. ✅ No race conditions: `go test -race ./...`
4. ✅ Integration tests pass: `go test -run Integration ./tests/integration`
5. ✅ Code generated: `make generate` (no diffs)

**Coverage Targets:**
- Auth utilities: 95%+
- HTTP handlers: 90%+
- Integration: 80%+ of user flows

---

### ✅ Example: Complete TDD Cycle (Logout Handler)

**Step 1: Write Tests (RED 🔴)**
```bash
vim internal/handlers/auth_test.go
# Add TestLogout_Success, TestLogout_InvalidToken, TestLogout_MissingToken

go test -v -short ./internal/handlers -run TestLogout
# ❌ FAIL: handler.Logout undefined
```

**Step 2: Add SQL Query**
```bash
vim sqlc/queries/auth.sql
# Add: -- name: DeleteSession :exec

make generate-db && make generate-mocks
```

**Step 3: Implement Handler (GREEN 🟢)**
```bash
vim internal/handlers/auth.go
# Implement Logout() handler

go test -v -short ./internal/handlers -run TestLogout
# ✅ PASS: TestLogout_Success (0.00s)
# ✅ PASS: TestLogout_InvalidToken (0.00s)
# ✅ PASS: TestLogout_MissingToken (0.00s)
```

**Step 4: Add Integration Test**
```bash
vim tests/integration/auth_integration_test.go
# Add TestLogout_Integration_Success

go test -v -run TestLogout_Integration ./tests/integration
# ✅ PASS: TestLogout_Integration_Success (2.13s)
```

**Result:** 4 tests, 100% coverage, handler complete! 🎉

---

### ✅ Key Learnings

1. **Integration tests are ESSENTIAL** - They catch bugs unit tests cannot
2. **Write tests BEFORE code** - Clarifies requirements, prevents overengineering
3. **Use testcontainers** - Real PostgreSQL in Docker, isolated per test
4. **Generate mocks automatically** - No manual maintenance, always in sync
5. **Test the full stack** - HTTP → Handler → Database → Response
6. **Document as you go** - Add comments linking code to test requirements

**TDD provides:**
- ✅ Confidence in refactoring
- ✅ Living documentation
- ✅ Faster debugging (tests show exactly what broke)
- ✅ Better API design (think from caller's perspective)

---

## Success Metrics

### Phase 1 (Complete)
- ✅ 20 tables + 3 views in PostgreSQL
- ✅ 50+ documentation files generated
- ✅ OpenAPI spec with 10+ endpoints
- ✅ 15+ type-safe SQL queries
- ✅ Code generation working end-to-end
- ✅ Complete automation via Makefile

### Phase 2 (In Progress - As of 2025-10-28)
- ✅ **Authentication working** (JWT + sessions) - **COMPLETE!**
  - ✅ Login endpoint (5 unit + 4 integration tests)
  - ✅ Logout endpoint (3 unit + 1 integration tests)
  - ✅ GetMe endpoint (5 unit tests)
  - ✅ Full auth lifecycle test (login → getMe → logout)
  - ✅ 33/33 tests passing
  - ✅ ~85% code coverage
- [ ] JWT Middleware (planned)
- [ ] Employee CRUD endpoints (0/5 complete)
- [ ] Integration tests with >80% coverage (auth at 100%, need employee tests)
- [ ] API response time <100ms (p95) - not yet measured

---

## 📖 Documentation Standards

### When to Update Docs

**Always update when**:
- Adding new tables → Regenerate ERD: `make generate-erd`
- Adding API endpoints → Update `openapi/spec.yaml`
- Adding SQL queries → Add to `sqlc/queries/*.sql`
- Changing architecture → Update `docs/planning/MIGRATION_PLAN.md`

**How to update**:
```bash
# Database docs (auto-generated)
make generate-erd

# Manual docs (update manually)
vim MIGRATION_PLAN.md
vim docs/ERD.md  # If schema structure changes significantly
```

---

## Roadmap

### Phase 1: ✅ COMPLETED (Foundation)
- Database schema, code generation, documentation

### Phase 2: In Progress (Core API)
- Authentication, employee management, org management

### Phase 3-8: Planned
- Agent/MCP configuration APIs
- Approval workflows
- Analytics endpoints
- Employee CLI client
- Admin web UI
- Production deployment

**See [MIGRATION_PLAN.md](./docs/planning/MIGRATION_PLAN.md) for complete 10-week plan.**

---

## 🎯 Next Actions

### For New Developers

1. Read [QUICKSTART.md](./docs/setup/QUICKSTART.md)
2. Review [docs/ERD.md](./docs/ERD.md)
3. Explore generated code in `generated/`
4. Start implementing handlers in `internal/handlers/`

### For This Week (Phase 2)

1. Create `internal/middleware/auth.go`
2. Implement `internal/handlers/auth.go`
3. Set up JWT token generation
4. Add integration tests
5. Create `cmd/server/main.go`

---

## 📞 Important Links

### Planning
- [MIGRATION_PLAN.md](./docs/planning/MIGRATION_PLAN.md) - Complete roadmap
- [DATABASE_SCHEMA.md](./docs/planning/DATABASE_SCHEMA.md) - Original ERD
- [../CLAUDE.md](../CLAUDE.md) - Parent project (original Ubik)

### Code
- [openapi/spec.yaml](./openapi/spec.yaml) - API specification
- [sqlc/queries/](./sqlc/queries/) - SQL queries
- [generated/](./generated/) - Auto-generated code

### Documentation
- [docs/ERD.md](./docs/ERD.md) - Visual schema
- [docs/INDEX.md](./docs/INDEX.md) - Complete docs index
- [QUICKSTART.md](./docs/setup/QUICKSTART.md) - Setup guide

---

## Status Summary

**Last Updated**: 2025-10-28
**Version**: 0.2.0 (Phase 2 - Authentication Complete)
**Status**: 🟢 Authentication System Production-Ready

**Phase 1 Achievements**:
- ✅ Complete database schema (20 tables + 3 views)
- ✅ Code generation pipeline working
- ✅ 50+ documentation files
- ✅ OpenAPI spec for auth + employees
- ✅ Type-safe SQL queries
- ✅ Local development environment
- ✅ Comprehensive Mermaid ERD

**Phase 2 Achievements** (UPDATED - 2025-10-28):
- ✅ **Complete authentication system** with TDD
- ✅ **JWT Middleware** - Centralized auth for all endpoints ⭐ **NEW!**
- ✅ **43/43 tests passing** (36 unit + 7 integration) - **+10 tests!**
- ✅ **~88% code coverage** across auth, handlers, and middleware
- ✅ **3/10 API endpoints + middleware implemented**:
  - POST /auth/login
  - POST /auth/logout
  - GET /auth/me
  - JWT Middleware (context-based auth)
- ✅ **Real PostgreSQL integration tests** with testcontainers
- ✅ **JWT token lifecycle** fully tested and working
- ✅ **TDD best practices** documented in CLAUDE.md
- ✅ **Code duplication eliminated** - auth logic centralized

**Test Breakdown**:
- 14 JWT helper tests (auth utilities)
- 13 handler unit tests (Login, Logout, GetMe)
- 9 middleware unit tests ⭐ **NEW!**
- 2 middleware integration tests ⭐ **NEW!**
- 6 auth integration tests (full stack with PostgreSQL)
- 100% of auth user flows covered

**Next Milestone**: Employee CRUD endpoints (GET, POST, PATCH, DELETE)

---

**For detailed information, see the [Documentation Map](#-documentation-map) above.**
- save .md files inside /docs. Update cloud.md if .md files were updated or created