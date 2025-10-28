# Pivot Project Testing Infrastructure Analysis

## Executive Summary

The Pivot project is in **Phase 1 (Foundation)** with a complete PostgreSQL schema, code generation pipeline, and type-safe database layer. However, **NO testing infrastructure currently exists** - no test files, helpers, mocks, or fixtures. The project is ready for comprehensive testing implementation in Phase 2.

**Status**: 🟢 Code generation pipeline operational | 🔴 Testing infrastructure not yet implemented

---

## 1. Current Testing Infrastructure

### Test Files
- **Count**: 0 test files (`.go` files)
- **Test Functions**: 0
- **Test Fixtures**: None
- **Test Helpers**: None
- **Mocks**: None
- **Test Data**: None

### Test Configuration
- **Makefile Target**: `make test` exists but will fail (no tests to run)
  ```bash
  test:
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out | tail -1
  ```
- **Coverage Tracking**: Infrastructure present but unused
- **Test Coverage Tools**: None configured

### Current Test State
```
❌ No unit tests
❌ No integration tests
❌ No handler tests
❌ No database layer tests
❌ No middleware tests
❌ No service/business logic tests
❌ No validation tests
❌ No test fixtures or factories
❌ No mock implementations
❌ No test containers (testcontainers-go not in go.mod)
```

---

## 2. Generated Code Structure

### A. API Code Generation (OpenAPI → Go)

**Tool**: `oapi-codegen/v2`  
**Input**: `/Users/sergeirastrigin/Projects/ubik/pivot/openapi/spec.yaml`  
**Output**: `/Users/sergeirastrigin/Projects/ubik/pivot/generated/api/server.gen.go` (698 lines)

#### Generated Components

1. **Type Models** (~140 lines)
   - `LoginRequest`, `LoginResponse`
   - `Employee`, `CreateEmployeeRequest`, `UpdateEmployeeRequest`
   - `Organization`, `Team`, `Role`
   - `EmployeeList`, `Error`, `PaginationMeta`
   - All with proper JSON marshaling tags

2. **ServerInterface** (11 methods)
   ```go
   type ServerInterface interface {
       Login(w http.ResponseWriter, r *http.Request)
       Logout(w http.ResponseWriter, r *http.Request)
       GetCurrentEmployee(w http.ResponseWriter, r *http.Request)
       ListEmployees(w http.ResponseWriter, r *http.Request, params ListEmployeesParams)
       CreateEmployee(w http.ResponseWriter, r *http.Request)
       DeleteEmployee(w http.ResponseWriter, r *http.Request, employeeId EmployeeId)
       GetEmployee(w http.ResponseWriter, r *http.Request, employeeId EmployeeId)
       UpdateEmployee(w http.ResponseWriter, r *http.Request, employeeId EmployeeId)
       GetCurrentOrganization(w http.ResponseWriter, r *http.Request)
       ListRoles(w http.ResponseWriter, r *http.Request)
   }
   ```

3. **Unimplemented Server** - Base implementation returning 501 (Not Implemented)

4. **ServerInterfaceWrapper** - Middleware injection and parameter binding

5. **Handler Registration**
   - `Handler(si ServerInterface) http.Handler`
   - `HandlerFromMux(si ServerInterface, r chi.Router) http.Handler`
   - `HandlerFromMuxWithBaseURL(...)`
   - `HandlerWithOptions(...)`

6. **Parameter Types**
   - `ListEmployeesParams` - Query parameter struct with optional page, per_page, status

#### Configuration
**File**: `openapi/oapi-codegen.yaml`
```yaml
package: api
generate:
  models: true
  chi-server: true
  embedded-spec: true
output: ../generated/api/server.gen.go
compatibility:
  always-prefix-enum-values: true
  apply-chi-middleware-first-to-last: true
```

### B. Database Code Generation (SQL → Go)

**Tool**: `sqlc/v1.30.0`  
**Input**: 3 SQL query files  
**Output**: `/Users/sergeirastrigin/Projects/ubik/pivot/generated/db/` (6 Go files, ~30KB)

#### Generated Files

1. **models.go** (10.7 KB, ~230 lines)
   - 20 struct types mapping all database tables
   - 3 view types (materialized views)
   - Type-safe model definitions with proper JSON tags
   
   **Table Models** (20):
   - `ActivityLog`, `AgentCatalog`, `AgentPolicy`, `AgentRequest`, `AgentTool`
   - `Approval`, `Employee`, `EmployeeAgentConfig`, `EmployeeMcpConfig`
   - `McpCatalog`, `McpCategory`, `Organization`, `Policy`, `Role`
   - `Session`, `Subscription`, `Team`, `TeamPolicy`, `Tool`, `UsageRecord`
   
   **View Models** (3):
   - `VEmployeeAgent` - Employee agents with catalog details
   - `VEmployeeMcp` - Employee MCPs with catalog details
   - `VPendingApproval` - Approval requests with requester context

2. **querier.go** (2.1 KB, ~46 lines)
   - `Querier` interface with all database operations
   - Example methods:
     ```go
     type Querier interface {
         GetEmployee(ctx context.Context, id uuid.UUID) (Employee, error)
         ListEmployees(ctx context.Context, arg ListEmployeesParams) ([]Employee, error)
         CreateEmployee(ctx context.Context, arg CreateEmployeeParams) (Employee, error)
         UpdateEmployee(ctx context.Context, arg UpdateEmployeeParams) (Employee, error)
         CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error)
         GetSessionWithEmployee(ctx context.Context, tokenHash string) (GetSessionWithEmployeeRow, error)
         // ... 20+ methods total
     }
     ```

3. **db.go** (564 bytes)
   - Database connection wrapper
   - Implements `Querier` interface

4. **employees.sql.go** (8.3 KB)
   - 10 database operation implementations
   - Handles employee CRUD and queries
   - Type-safe parameter structs

5. **auth.sql.go** (4.1 KB)
   - 5 authentication/session operations
   - Session creation, lookup, deletion
   - Employee with role queries

6. **organizations.sql.go** (4.9 KB)
   - Organization and team management
   - Role queries
   - Team CRUD operations

#### sqlc Configuration
**File**: `sqlc/sqlc.yaml`
```yaml
version: "2"
sql:
  - schema: "../schema.sql"
    queries: "./queries"
    engine: "postgresql"
    gen:
      go:
        package: "db"
        out: "../generated/db"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true
        emit_pointers_for_null_types: true
        json_tags_case_style: "snake"
```

---

## 3. SQL Queries Organization

### Query Files Structure

**Location**: `/Users/sergeirastrigin/Projects/ubik/pivot/sqlc/queries/`

#### A. employees.sql (10 queries)
```
✓ GetEmployee :one                  - Fetch by ID, excludes deleted
✓ GetEmployeeByEmail :one          - Fetch by email
✓ ListEmployees :many              - Paginated list with org_id filtering
✓ CountEmployees :one              - Count employees in org
✓ CreateEmployee :one              - Insert with RETURNING
✓ UpdateEmployee :one              - Partial update with COALESCE
✓ UpdateEmployeeLastLogin :exec    - Update last login timestamp
✓ SoftDeleteEmployee :exec         - Mark deleted_at = NOW()
✓ GetEmployeesByTeam :many         - Team member listing
✓ GetEmployeeWithRole :one         - Employee + role details JOIN
```

**Pattern**: All queries are org-scoped (multi-tenant safe) and soft-delete aware

#### B. auth.sql (6 queries)
```
✓ CreateSession :one               - Insert JWT session with expiry
✓ GetSession :one                  - Lookup by token_hash
✓ DeleteSession :exec              - Invalidate single session
✓ DeleteExpiredSessions :exec      - Cleanup old sessions
✓ DeleteEmployeeSessions :exec     - Logout all sessions for employee
✓ GetSessionWithEmployee :one       - Session + employee details JOIN
```

**Pattern**: Session security focused - requires valid token_hash AND active employee status

#### C. organizations.sql (8 queries)
```
✓ GetOrganization :one             - Fetch org by ID
✓ GetOrganizationBySlug :one       - Fetch org by slug
✓ ListTeams :many                  - List teams in org
✓ GetTeam :one                     - Fetch team (org-scoped)
✓ CreateTeam :one                  - Insert team
✓ UpdateTeam :one                  - Update team details
✓ DeleteTeam :exec                 - Hard delete team
✓ ListRoles :many                  - List all roles
✓ GetRole :one                     - Fetch role by ID
```

### Query Coverage Summary

| Category | Count | Type |
|----------|-------|------|
| SELECT (one) | 9 | :one |
| SELECT (many) | 4 | :many |
| SELECT (count) | 1 | :one |
| INSERT | 3 | :one |
| UPDATE | 5 | :one or :exec |
| DELETE | 5 | :exec |
| **Total** | **27** | |

---

## 4. OpenAPI Endpoints Defined

**Location**: `/Users/sergeirastrigin/Projects/ubik/pivot/openapi/spec.yaml` (668 lines)

### Endpoint Summary

#### Authentication (3 endpoints)
```
POST   /auth/login              - Login with email/password → JWT token
POST   /auth/logout             - Invalidate session
GET    /auth/me                 - Get current employee context
```

#### Employees (4 endpoints)
```
GET    /employees               - List employees (paginated, filterable)
POST   /employees               - Create new employee
GET    /employees/{employee_id} - Get single employee
PATCH  /employees/{employee_id} - Update employee
DELETE /employees/{employee_id} - Soft delete employee
```

#### Organizations (3 endpoints)
```
GET    /organizations/current   - Get authenticated employee's org
GET    /roles                   - List available roles
GET    /roles/{id}              - Get role details
```

### Request/Response Schemas

**Login Request**
```json
{
  "email": "alice@acme.com",
  "password": "SecurePass123!"
}
```

**Create Employee Request**
```json
{
  "email": "string",
  "full_name": "string",
  "role_id": "uuid",
  "team_id": "uuid (optional)",
  "preferences": {} (optional)
}
```

**Update Employee Request**
```json
{
  "full_name": "string (optional)",
  "team_id": "uuid (optional)",
  "role_id": "uuid (optional)",
  "status": "active|suspended|inactive (optional)",
  "preferences": {} (optional)
}
```

**Error Response**
```json
{
  "error": "validation_error",
  "message": "Email is required",
  "details": {} (optional)
}
```

### Security
- **Default**: Bearer token (JWT) required on all endpoints except `/auth/login`
- **Status Codes**:
  - 200/201: Success
  - 204: No content (successful delete)
  - 401: Unauthorized
  - 403: Forbidden (insufficient permissions)
  - 404: Not found
  - 422: Validation error

---

## 5. Project Structure

### Directory Layout

```
pivot/
├── generated/                    # ⚠️ Auto-generated (DO NOT EDIT)
│   ├── api/
│   │   └── server.gen.go         # 698 lines - Router, types, interfaces
│   └── db/
│       ├── models.go             # 230 lines - All table/view structs
│       ├── querier.go            # 46 lines - Database interface
│       ├── db.go                 # Connection wrapper
│       ├── employees.sql.go      # 10 operations
│       ├── auth.sql.go           # 6 operations
│       └── organizations.sql.go  # 8 operations
│
├── internal/                     # User code (implement here)
│   ├── handlers/                 # HTTP request handlers
│   ├── service/                  # Business logic layer
│   ├── middleware/               # Auth, logging, RLS
│   ├── mapper/                   # Type conversions (API ↔ DB)
│   └── validation/               # Input validation
│
├── cmd/                          # Entry points
│   ├── server/                   # API server
│   └── cli/                      # Employee CLI (future)
│
├── sqlc/
│   ├── sqlc.yaml                 # sqlc configuration
│   └── queries/
│       ├── employees.sql         # 10 queries
│       ├── auth.sql              # 6 queries
│       └── organizations.sql     # 8 queries
│
├── openapi/
│   ├── spec.yaml                 # OpenAPI 3.0.3 spec
│   └── oapi-codegen.yaml         # oapi-codegen config
│
├── docs/                         # Documentation
│   ├── ERD.md
│   ├── INDEX.md
│   └── *.md (auto-generated)
│
├── schema.sql                    # PostgreSQL schema (20 tables + 3 views)
├── docker-compose.yml            # Local PostgreSQL + Adminer
├── go.mod                        # Empty (1.24.5)
├── Makefile                      # 20+ commands
└── .gitignore                    # Excludes generated/, bin/, coverage.*
```

### Code Statistics

| Component | Lines | Status |
|-----------|-------|--------|
| Generated API Code | 698 | ✅ Complete |
| Generated DB Code | ~1000 | ✅ Complete |
| User Implementation | 0 | 🔴 Not started |
| **Total Generated** | **1787** | |
| **Tests** | **0** | 🔴 |

---

## 6. Gaps in Testing Coverage

### Critical Gaps

#### 1. No Unit Tests
- No handler tests
- No service layer tests
- No validation tests
- No mapper tests
- No middleware tests

#### 2. No Integration Tests
- No database integration tests
- No API endpoint-to-database tests
- No transaction tests
- No concurrent operation tests

#### 3. No Test Infrastructure
- No test database setup
- No fixtures/factories
- No mock implementations
- No test helpers
- No test containers (testcontainers-go)

#### 4. No Fixtures/Test Data
- No seed data for tests
- No test payloads
- No example requests/responses

#### 5. No Database Testing
- No pgx mock
- No transaction rollback tests
- No constraint violation tests
- No multi-tenant isolation tests

#### 6. Missing Dependencies
Go modules not yet initialized - needs:
- `testify/assert` or `testing/fstest` for assertions
- `testify/mock` or similar for mocking
- `testcontainers-go` for integration tests
- `golang.org/x/oauth2` for JWT testing
- `pgx` for database testing

---

## 7. Recommended Testing Approach

### Phase 2 (Weeks 1-2): Test Foundation

#### 1. Set Up Test Infrastructure

**Create test helpers** (`internal/testing/helpers.go`):
```go
type TestDB struct {
    conn *pgx.Conn
    tx   pgx.Tx
}

func SetupTestDB(t *testing.T) *TestDB {
    // Initialize from environment
}

func (db *TestDB) Cleanup() {
    // Rollback transaction
}

type TestClient struct {
    router http.Handler
}

func NewTestClient(handlers api.ServerInterface) *TestClient {
    // Create test HTTP client
}
```

**Create fixtures** (`internal/testing/fixtures/`):
```
fixtures/
├── employees.go     # Employee factory functions
├── sessions.go      # Session factory functions
├── organizations.go # Org/team/role factories
└── payloads.go      # Request/response examples
```

**Create mocks** (`internal/testing/mocks/`):
```
mocks/
├── querier.go       # Mock database.Querier
├── service.go       # Mock business logic
└── middleware.go    # Mock authentication
```

#### 2. Test Database Strategy

**Option A: Testcontainers** (Recommended)
```go
import "github.com/testcontainers/testcontainers-go"

func setupPostgresContainer() (*Container, *pgx.Conn) {
    // Spin up isolated PostgreSQL container per test
    // Run migrations
    // Return connection for cleanup
}
```

**Option B: Shared Test Database**
```bash
# In docker-compose.yml
services:
  test-postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: pivot_test
```

#### 3. Test Structure Pattern

```go
package handlers_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/sergeirastrigin/ubik-enterprise/internal/handlers"
    "github.com/sergeirastrigin/ubik-enterprise/generated/api"
)

func TestLoginHandler_Success(t *testing.T) {
    // Arrange: Create test data
    mockDB := setupMockDB()
    handler := handlers.NewAuthHandler(mockDB)
    
    // Act: Call handler
    req := createTestRequest("POST", "/auth/login", loginPayload)
    w := executeRequest(handler, req)
    
    // Assert: Verify response
    assert.Equal(t, http.StatusOK, w.Code)
    var resp api.LoginResponse
    json.Unmarshal(w.Body.Bytes(), &resp)
    assert.NotEmpty(t, resp.Token)
}

func TestLoginHandler_InvalidCredentials(t *testing.T) {
    // Test error cases
}

func TestLoginHandler_ValidationError(t *testing.T) {
    // Test input validation
}
```

#### 4. Coverage Targets

| Component | Target | Priority |
|-----------|--------|----------|
| Handlers | 80% | High |
| Services | 85% | High |
| Middleware | 75% | Medium |
| Validation | 90% | High |
| Mappers | 70% | Low |
| **Overall** | **80%** | |

### Phase 2 Breakdown

**Week 1: Infrastructure**
- [ ] Add test dependencies to go.mod
- [ ] Create test helpers and factories
- [ ] Set up test database (testcontainers or shared)
- [ ] Create mock implementations
- [ ] Document testing standards

**Week 2: Handler Tests**
- [ ] Auth handler tests (3 endpoints × 5 cases = 15 tests)
- [ ] Employee handler tests (5 endpoints × 4 cases = 20 tests)
- [ ] Organization handler tests (3 endpoints × 3 cases = 9 tests)
- [ ] Error handling tests (validation, auth, 404s)

**Week 3: Service/Business Logic**
- [ ] Employee service tests
- [ ] Auth service tests (password hashing, JWT)
- [ ] Role-based access control tests
- [ ] Integration tests (handler → service → DB)

---

## 8. Code Generation Dependencies

### Current Tools

| Tool | Version | Status |
|------|---------|--------|
| oapi-codegen | latest (v2.5.0+) | ✅ Installed |
| sqlc | latest (v1.30.0+) | ✅ Installed |
| tbls | latest | ✅ Installed (docs) |
| Go | 1.24.5+ | ✅ Required |
| PostgreSQL | 15+ | ✅ Docker image |

### Install Command
```bash
make install-tools
```

### Generation Flow

```
schema.sql ──┬──> PostgreSQL
             └──> tbls ──> docs/ERD.md

openapi/spec.yaml ──> oapi-codegen ──> generated/api/server.gen.go

schema.sql + queries/ ──> sqlc ──> generated/db/*.go

generated/ + internal/ ──> your code implements api.ServerInterface
```

---

## 9. Database Schema Summary

### 20 Tables

**Organization Tier** (5):
- `organizations` - Top-level tenants
- `subscriptions` - Billing and budgets
- `teams` - Groups of employees
- `roles` - Permission definitions
- `employees` - User accounts

**Agent Configuration** (7):
- `agent_catalog` - Available AI agents
- `tools` - Tool registry
- `policies` - Usage policies
- `agent_tools` - Agent ↔ Tool mapping
- `agent_policies` - Agent ↔ Policy mapping
- `team_policies` - Team-specific overrides
- `employee_agent_configs` - Per-employee agents

**MCP Configuration** (3):
- `mcp_categories` - Organization
- `mcp_catalog` - Available servers
- `employee_mcp_configs` - Per-employee MCPs

**Authentication** (1):
- `sessions` - JWT session tracking

**Approvals** (2):
- `agent_requests` - Employee requests
- `approvals` - Approval workflow

**Analytics** (2):
- `activity_logs` - Audit trail
- `usage_records` - Cost tracking

**Views** (3):
- `v_employee_agents` - Joined agent details
- `v_employee_mcps` - Joined MCP details
- `v_pending_approvals` - Approval queue

---

## 10. Next Steps for Testing

### Immediate Actions (This Week)

1. **Create test organization**
   - [ ] `internal/testing/helpers.go` - DB setup, cleanup
   - [ ] `internal/testing/fixtures/` - Factory functions
   - [ ] `internal/testing/mocks/` - Database/service mocks

2. **Update go.mod**
   ```bash
   go get -u \
     github.com/testcontainers/testcontainers-go \
     github.com/jackc/pgx/v5 \
     github.com/stretchr/testify \
     golang.org/x/crypto \
     github.com/golang-jwt/jwt/v5
   ```

3. **Create first test file**
   - [ ] `internal/handlers/auth_test.go` - Login/logout tests
   - [ ] Verify test runner works: `make test`

### Documentation

- [ ] Create `docs/TESTING.md` - Testing guide
- [ ] Add test patterns to architecture docs
- [ ] Document mock strategy

---

## Summary

### What's Ready ✅
- 698 lines of generated API code (types, interfaces, router)
- 1000+ lines of generated DB code (models, queries, interfaces)
- 27 type-safe SQL queries across 3 domains
- 10 fully specified API endpoints
- 20 database tables with 3 views
- Makefile with test targets

### What's Missing 🔴
- **0 test files**
- **0 test functions**
- **0 fixtures/factories**
- **0 mocks**
- **0 database integration tests**
- **0 endpoint tests**

### Recommended First Step

Create `/Users/sergeirastrigin/Projects/ubik/pivot/internal/testing/helpers.go` with:
1. Database setup/teardown
2. HTTP test client wrapper
3. Request/response builders
4. Mock factories

This unblocks writing the first batch of handler tests in Phase 2.

---

**Analysis Date**: 2025-10-28  
**Project Status**: Phase 1 (Foundation) Complete → Phase 2 (API Implementation) Ready  
**Recommendation**: Implement comprehensive test suite before shipping Phase 2 endpoints
