# Implementation Roadmap - Phase 2 API Development

**Last Updated**: 2025-10-28 (Updated Priority)
**Current Progress**: 9/9 core employee endpoints complete (100%) + **82 tests passing**

---

## 🎯 NEW PRIORITY: Agent Configuration APIs (Core Value Proposition)

**Strategic Decision**: Skip full org/team/role CRUD for now. Focus on the unique value proposition - **AI Agent Management**.

**Rationale**:
- We have enough employee infrastructure to support agent configuration
- This is what makes the platform unique and valuable
- Fastest path to MVP
- Org/team/role full CRUD can wait until needed

---

## ✅ COMPLETED: Core Employee CRUD (9/9 endpoints)

1. ✅ **POST /auth/login** - Full authentication (5 unit + 4 integration tests)
2. ✅ **POST /auth/logout** - Session invalidation (3 unit + 1 integration tests)
3. ✅ **GET /auth/me** - Current employee data (5 unit tests)
4. ✅ **JWT Middleware** - Centralized authentication (9 unit + 2 integration tests)
5. ✅ **GET /employees** - List employees (4 unit + 4 integration tests)
6. ✅ **GET /employees/{id}** - Get employee by ID (4 unit + 3 integration tests)
7. ✅ **POST /employees** - Create employee (6 unit + 3 integration tests)
8. ✅ **PATCH /employees/{id}** - Update employee (5 unit + 2 integration tests)
9. ✅ **DELETE /employees/{id}** - Soft delete employee (4 unit + 2 integration tests)

**Status**: Employee CRUD complete with **82/82 tests passing** (as of 2025-10-28)

---

## 📋 NEXT: Agent Configuration APIs (Core Value Proposition)

### Phase 1: Agent Catalog & Employee Agent Configuration

**Goal**: Enable employees to view available AI agents and assign them to their account.

#### Endpoint 1: GET /agents - List available AI agents

**Purpose**: Return catalog of available AI agents (Claude Code, Cursor, Windsurf, Continue, Copilot)

**Implementation Plan:**
- **Estimated Time**: 1 hour
- **Tests to Write**: 4 unit + 2 integration

**SQL Query Needed:**
```sql
-- name: ListAgents :many
SELECT id, name, provider, description, logo_url,
       is_active, supported_platforms, pricing_tier,
       created_at, updated_at
FROM agent_catalog
WHERE is_active = true
ORDER BY name ASC;
```

**Unit Tests (4):**
1. `TestListAgents_Success` - Returns all active agents
2. `TestListAgents_EmptyResult` - No agents available
3. `TestListAgents_OnlyActiveAgents` - Filters out inactive agents
4. `TestListAgents_OrderedByName` - Returns agents alphabetically

**Integration Tests (2):**
5. `TestListAgents_Integration_WithSeedData` - Real DB with catalog data
6. `TestListAgents_Integration_Authentication` - Requires valid JWT

**Response Format:**
```json
{
  "agents": [
    {
      "id": "uuid",
      "name": "Claude Code",
      "provider": "anthropic",
      "description": "AI-powered code assistant with deep codebase understanding",
      "logo_url": "https://...",
      "supported_platforms": ["macos", "linux", "windows"],
      "pricing_tier": "enterprise"
    }
  ],
  "total": 5
}
```

---

#### Endpoint 2: GET /employees/{employee_id}/agent-configs - List employee's assigned agents

**Purpose**: Return all agents assigned to a specific employee with their configurations

**Implementation Plan:**
- **Estimated Time**: 1.5 hours
- **Tests to Write**: 6 unit + 3 integration

**SQL Query Needed:**
```sql
-- name: ListEmployeeAgentConfigs :many
SELECT
    eac.id,
    eac.employee_id,
    eac.agent_id,
    eac.config,
    eac.is_enabled,
    eac.last_synced_at,
    eac.created_at,
    eac.updated_at,
    ac.name as agent_name,
    ac.provider as agent_provider,
    ac.logo_url as agent_logo
FROM employee_agent_configs eac
JOIN agent_catalog ac ON eac.agent_id = ac.id
WHERE eac.employee_id = $1
  AND eac.deleted_at IS NULL
ORDER BY eac.created_at DESC;
```

**Unit Tests (6):**
1. `TestListEmployeeAgentConfigs_Success` - Returns employee's agents
2. `TestListEmployeeAgentConfigs_WithConfig` - Includes config JSON
3. `TestListEmployeeAgentConfigs_EmptyResult` - No agents assigned
4. `TestListEmployeeAgentConfigs_OnlyEnabled` - Can filter by is_enabled
5. `TestListEmployeeAgentConfigs_WrongEmployee` - Cannot view other employee's agents (org isolation)
6. `TestListEmployeeAgentConfigs_InvalidEmployeeID` - Returns 400

**Integration Tests (3):**
7. `TestListEmployeeAgentConfigs_Integration_MultipleAgents` - Real DB with multiple configs
8. `TestListEmployeeAgentConfigs_Integration_OrgIsolation` - Cannot access other org's data
9. `TestListEmployeeAgentConfigs_Integration_WithPolicies` - Includes policy restrictions

**Response Format:**
```json
{
  "agent_configs": [
    {
      "id": "uuid",
      "employee_id": "uuid",
      "agent_id": "uuid",
      "agent_name": "Claude Code",
      "agent_provider": "anthropic",
      "config": {
        "model": "claude-3-5-sonnet-20241022",
        "max_tokens": 8192,
        "temperature": 0.7
      },
      "is_enabled": true,
      "last_synced_at": "2025-10-28T10:00:00Z",
      "created_at": "2025-10-20T15:30:00Z"
    }
  ],
  "total": 3
}
```

---

#### Endpoint 3: POST /employees/{employee_id}/agent-configs - Assign agent to employee

**Purpose**: Create new agent configuration for an employee

**Implementation Plan:**
- **Estimated Time**: 2 hours
- **Tests to Write**: 8 unit + 4 integration

**SQL Query Needed:**
```sql
-- name: CreateEmployeeAgentConfig :one
INSERT INTO employee_agent_configs (
    employee_id,
    agent_id,
    config,
    is_enabled
) VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetAgentById :one
SELECT * FROM agent_catalog WHERE id = $1;

-- name: CheckEmployeeAgentExists :one
SELECT EXISTS(
    SELECT 1 FROM employee_agent_configs
    WHERE employee_id = $1 AND agent_id = $2 AND deleted_at IS NULL
);
```

**Request Body:**
```json
{
  "agent_id": "uuid",
  "config": {
    "model": "claude-3-5-sonnet-20241022",
    "max_tokens": 8192
  },
  "is_enabled": true
}
```

**Unit Tests (8):**
1. `TestCreateEmployeeAgentConfig_Success` - Creates config
2. `TestCreateEmployeeAgentConfig_WithDefaultConfig` - Uses default if not provided
3. `TestCreateEmployeeAgentConfig_InvalidAgentID` - Returns 400
4. `TestCreateEmployeeAgentConfig_AgentNotFound` - Returns 404
5. `TestCreateEmployeeAgentConfig_DuplicateAgent` - Returns 409 (already assigned)
6. `TestCreateEmployeeAgentConfig_InvalidConfig` - Returns 400 (malformed JSON)
7. `TestCreateEmployeeAgentConfig_WrongEmployee` - Cannot create for other employee (org isolation)
8. `TestCreateEmployeeAgentConfig_MissingAgentID` - Returns 422

**Integration Tests (4):**
9. `TestCreateEmployeeAgentConfig_Integration_Success` - Real DB creation
10. `TestCreateEmployeeAgentConfig_Integration_DuplicateCheck` - Uniqueness constraint
11. `TestCreateEmployeeAgentConfig_Integration_OrgIsolation` - Cross-org validation
12. `TestCreateEmployeeAgentConfig_Integration_WithPolicies` - Policy inheritance

---

#### Endpoint 4: PATCH /employees/{employee_id}/agent-configs/{config_id} - Update agent config

**Purpose**: Update existing agent configuration (model, settings, enabled status)

**Implementation Plan:**
- **Estimated Time**: 1.5 hours
- **Tests to Write**: 6 unit + 2 integration

**SQL Query Needed:**
```sql
-- name: UpdateEmployeeAgentConfig :one
UPDATE employee_agent_configs
SET
    config = COALESCE($2, config),
    is_enabled = COALESCE($3, is_enabled),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;

-- name: GetEmployeeAgentConfig :one
SELECT * FROM employee_agent_configs
WHERE id = $1 AND deleted_at IS NULL;
```

**Unit Tests (6):**
1. `TestUpdateEmployeeAgentConfig_Success` - Updates config
2. `TestUpdateEmployeeAgentConfig_PartialUpdate` - Updates only provided fields
3. `TestUpdateEmployeeAgentConfig_DisableAgent` - Sets is_enabled = false
4. `TestUpdateEmployeeAgentConfig_NotFound` - Returns 404
5. `TestUpdateEmployeeAgentConfig_WrongEmployee` - Cannot update other's config
6. `TestUpdateEmployeeAgentConfig_InvalidConfig` - Returns 400

**Integration Tests (2):**
7. `TestUpdateEmployeeAgentConfig_Integration_Success` - Real DB update
8. `TestUpdateEmployeeAgentConfig_Integration_OrgIsolation` - Cross-org check

---

#### Endpoint 5: DELETE /employees/{employee_id}/agent-configs/{config_id} - Remove agent assignment

**Purpose**: Soft delete agent configuration

**Implementation Plan:**
- **Estimated Time**: 45 minutes
- **Tests to Write**: 4 unit + 2 integration

**SQL Query Needed:**
```sql
-- name: SoftDeleteEmployeeAgentConfig :exec
UPDATE employee_agent_configs
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
```

**Unit Tests (4):**
1. `TestDeleteEmployeeAgentConfig_Success` - Soft deletes config
2. `TestDeleteEmployeeAgentConfig_NotFound` - Returns 404
3. `TestDeleteEmployeeAgentConfig_WrongEmployee` - Cannot delete other's config
4. `TestDeleteEmployeeAgentConfig_Idempotent` - Can delete twice

**Integration Tests (2):**
5. `TestDeleteEmployeeAgentConfig_Integration_SoftDelete` - Sets deleted_at
6. `TestDeleteEmployeeAgentConfig_Integration_CannotSync` - Deleted config not returned by list

---

### Phase 2: MCP Server Configuration (Lower Priority)

Similar structure for MCP server endpoints:
- GET /mcp-servers - List available MCP servers
- GET /employees/{employee_id}/mcp-configs - List employee's MCP configs
- POST /employees/{employee_id}/mcp-configs - Assign MCP server
- PATCH /employees/{employee_id}/mcp-configs/{config_id} - Update MCP config
- DELETE /employees/{employee_id}/mcp-configs/{config_id} - Remove MCP assignment

**Deferred until Phase 1 complete**

---

### Phase 3: Policy & Tool Management (Even Lower Priority)

- GET /policies - List available policies
- GET /tools - List available tools
- GET /agents/{agent_id}/policies - Get agent's policies
- GET /agents/{agent_id}/tools - Get agent's available tools

**Deferred until Phase 2 complete**

---

## 📊 Updated Summary Table

| Priority | Endpoint | Method | Tests | Est. Time | Status |
|----------|----------|--------|-------|-----------|--------|
| **COMPLETED** | **Authentication & Employees** | | | | |
| 0 | /auth/login | POST | 9 | - | ✅ Complete |
| 0 | /auth/logout | POST | 4 | - | ✅ Complete |
| 0 | /auth/me | GET | 5 | - | ✅ Complete |
| 0 | JWT Middleware | - | 11 | - | ✅ Complete |
| 1 | /employees | GET | 8 | - | ✅ Complete |
| 2 | /employees/{id} | GET | 7 | - | ✅ Complete |
| 3 | /employees | POST | 9 | - | ✅ Complete |
| 4 | /employees/{id} | PATCH | 7 | - | ✅ Complete |
| 5 | /employees/{id} | DELETE | 6 | - | ✅ Complete |
| **NEXT** | **Agent Configuration (Phase 1)** | | | | |
| 6 | /agents | GET | 6 | 1 hr | ⏳ **NEXT** |
| 7 | /employees/{id}/agent-configs | GET | 9 | 1.5 hrs | ⏸️ Pending |
| 8 | /employees/{id}/agent-configs | POST | 12 | 2 hrs | ⏸️ Pending |
| 9 | /employees/{id}/agent-configs/{cid} | PATCH | 8 | 1.5 hrs | ⏸️ Pending |
| 10 | /employees/{id}/agent-configs/{cid} | DELETE | 6 | 45 min | ⏸️ Pending |
| **LATER** | **MCP & Policies (Phase 2-3)** | | | | |
| 11+ | MCP endpoints | Various | TBD | ~6 hrs | ⏸️ Deferred |

**Completed**: 9 endpoints (84 tests passing)
**Next Phase**: Hierarchical agent config system
**Total Remaining**: ~20 hours for full MVP

---

## 🎯 NEW SCHEMA: Hierarchical Configuration System

### Schema Changes (2025-10-29)

**✅ COMPLETED:**
1. Renamed `agent_catalog` → `agents` (consistency with other tables)
2. Split `employee_agent_configs` into 3-level hierarchy:
   - `org_agent_configs` - Organization defaults (full config)
   - `team_agent_configs` - Team overrides (config_override)
   - `employee_agent_configs` - Employee overrides (config_override)
3. Added `system_prompts` table - Additive prompts across org/team/employee
4. Added `employee_policies` table - Completes policy hierarchy
5. All 84 tests passing ✅

### Configuration Resolution

**Hierarchy**: Org (base) → Team (override) → Employee (final override)

**Example**:
```javascript
// Org config (base)
{
  "model": "claude-3-5-sonnet-20241022",
  "temperature": 0.2,
  "max_tokens": 8192
}

// Team override
{
  "temperature": 0.5  // Override org
}

// Employee override
{
  "max_tokens": 16384  // Override org
}

// RESOLVED (merged)
{
  "model": "claude-3-5-sonnet-20241022",    // From org
  "temperature": 0.5,                        // From team
  "max_tokens": 16384                        // From employee
}
```

**System Prompts** (concatenated):
```
Org:      "Follow company coding standards. Use TypeScript."
Team:     "Focus on React components."
Employee: "I prefer functional programming style."

RESULT: All three concatenated with priority order
```

### Access Model

**Team Members Inherit Team's Agents Automatically**
- If Team A has ChatGPT configured → All Team A members get ChatGPT
- If Team B has Claude → All Team B members get Claude
- Employees can add overrides but don't need individual assignment

**Example Scenario**:
- Org enables ChatGPT and Claude (org_agent_configs)
- Team A gets only ChatGPT (team_agent_configs)
- Team B gets only Claude (team_agent_configs)
- Team C gets both (team_agent_configs for both)
- Alice (Team B) wants higher temperature (employee_agent_configs override)

---

## 📋 IMPLEMENTATION PLAN: Hierarchical Config System

### Phase 1: Config Resolution Service (Core Logic)

**Goal**: Build the service that resolves org → team → employee configs

**Tasks**:
1. Create `internal/service/config_resolver.go`
2. Implement config merge logic (deep merge JSONB)
3. Implement system prompt concatenation
4. Implement policy resolution (most restrictive wins)
5. Write comprehensive tests

**Estimated Time**: 3-4 hours

**Key Functions**:
```go
type ConfigResolver struct {
    db db.Querier
}

// Resolve full config for an employee + agent
func (r *ConfigResolver) ResolveAgentConfig(ctx context.Context, employeeID, agentID uuid.UUID) (*ResolvedConfig, error)

// Resolve all agents for an employee (for CLI sync)
func (r *ConfigResolver) ResolveEmployeeAgents(ctx context.Context, employeeID uuid.UUID) ([]ResolvedAgentConfig, error)

type ResolvedConfig struct {
    AgentID      uuid.UUID
    Config       map[string]interface{}  // Merged config
    SystemPrompt string                   // Concatenated prompts
    Policies     []Policy                 // Resolved policies
    IsEnabled    bool                     // All levels must be enabled
}
```

---

### Phase 2: Org-Level Management (Admin Operations)

**Goal**: Allow admins to manage org-level agent configurations

#### Endpoint 1: POST /organizations/current/agent-configs
Create org-level agent configuration (makes agent available to org)

**Tests**: 6 unit + 2 integration
**Time**: 1.5 hours

#### Endpoint 2: GET /organizations/current/agent-configs
List all agents configured at org level

**Tests**: 4 unit + 2 integration
**Time**: 1 hour

#### Endpoint 3: PATCH /organizations/current/agent-configs/{config_id}
Update org-level config (affects all teams/employees)

**Tests**: 5 unit + 2 integration
**Time**: 1 hour

#### Endpoint 4: DELETE /organizations/current/agent-configs/{config_id}
Remove agent from org (cascades to teams/employees)

**Tests**: 4 unit + 2 integration
**Time**: 45 min

**Phase 2 Total**: ~4.5 hours

---

### Phase 3: Employee Resolved View (CLI Sync)

**Goal**: Employees can fetch their fully resolved configs for CLI sync

#### Endpoint: GET /employees/{employee_id}/agent-configs/resolved
Returns fully resolved configs (org + team + employee merged)

**This is THE MOST IMPORTANT endpoint** - CLI needs this to sync

**Tests**: 8 unit + 4 integration
**Time**: 2 hours

**Response Format**:
```json
{
  "configs": [
    {
      "agent": {
        "id": "uuid",
        "name": "Claude Code",
        "type": "claude-code"
      },
      "config": {
        "model": "claude-3-5-sonnet-20241022",
        "temperature": 0.5,
        "max_tokens": 16384
      },
      "system_prompt": "Org prompt\n\nTeam prompt\n\nEmployee prompt",
      "policies": [...],
      "source": {
        "org_config_id": "uuid",
        "team_config_id": "uuid",
        "employee_config_id": null
      }
    }
  ]
}
```

---

### Phase 4: Team-Level Management (Optional for MVP)

Can be deferred - most orgs will use org-level + employee overrides

#### Endpoints (if needed):
- POST /teams/{team_id}/agent-configs
- GET /teams/{team_id}/agent-configs
- PATCH /teams/{team_id}/agent-configs/{config_id}
- DELETE /teams/{team_id}/agent-configs/{config_id}

**Phase 4 Total**: ~5 hours (DEFERRED)

---

### Phase 5: Employee-Level Overrides

**Goal**: Employees can request custom overrides (pending approval)

#### Endpoint: POST /employees/{employee_id}/agent-configs
Create employee-level override (may require approval)

**Tests**: 7 unit + 3 integration
**Time**: 2 hours

---

### Phase 6: System Prompts API

**Goal**: Manage hierarchical system prompts

#### Endpoints:
- GET /organizations/current/system-prompts
- POST /organizations/current/system-prompts
- GET /teams/{team_id}/system-prompts (if team management enabled)
- POST /employees/{employee_id}/system-prompts

**Phase 6 Total**: ~4 hours (DEFERRED to post-MVP)

---

## 🎯 MVP RECOMMENDATION (Next 2 Weeks)

### Week 1: Core + Org Management
1. ✅ Config Resolution Service (3-4 hours) - **START HERE**
2. ✅ Org-level CRUD (4.5 hours)
3. ✅ Employee resolved view (2 hours)

**Total**: ~10 hours
**Deliverable**: Admins can configure agents at org level, employees can sync to CLI

### Week 2: Employee Overrides + Polish
1. ✅ Employee override requests (2 hours)
2. ✅ Integration tests for full flow (2 hours)
3. ✅ Documentation (1 hour)

**Total**: ~5 hours
**Deliverable**: Complete MVP with hierarchical config system

### Post-MVP (Later)
- Team-level management
- System prompts API
- Approval workflows
- Policy resolution UI

---

## ✅ COMPLETED: JWT Middleware

### Why This is Critical

**Code Duplication Problem:**
```go
// This pattern appears in Logout and GetMe, will appear in 7+ more endpoints
authHeader := r.Header.Get("Authorization")
if authHeader == "" {
    writeError(w, http.StatusUnauthorized, "Missing authorization header")
    return
}

const bearerPrefix = "Bearer "
token := authHeader[len(bearerPrefix):]

claims, err := auth.VerifyJWT(token)
if err != nil {
    writeError(w, http.StatusUnauthorized, "Invalid token")
    return
}

tokenHash := auth.HashToken(token)
sessionData, err := h.db.GetSessionWithEmployee(ctx, tokenHash)
// ... 20+ lines of duplicate code
```

**After Middleware:**
```go
// Handler becomes this simple:
func (h *EmployeesHandler) List(w http.ResponseWriter, r *http.Request) {
    // Authentication already done by middleware!
    employeeID, _ := middleware.GetEmployeeID(r.Context())
    orgID, _ := middleware.GetOrgID(r.Context())

    // Just business logic here
    employees, err := h.db.ListEmployees(ctx, orgID, filters)
    // ...
}
```

### Implementation Plan

**Estimated Time**: 1-2 hours
**Tests to Write**: 8 (6 unit + 2 integration)

#### Step 1: Write Tests (RED Phase 🔴)

```bash
# Create middleware test file
vim internal/middleware/auth_test.go
```

**Unit Tests (6):**
1. `TestAuthMiddleware_ValidToken` - Sets context and calls next handler
2. `TestAuthMiddleware_InvalidToken` - Returns 401
3. `TestAuthMiddleware_ExpiredToken` - Returns 401
4. `TestAuthMiddleware_MissingToken` - Returns 401
5. `TestAuthMiddleware_SessionNotFound` - Returns 401 (logged out)
6. `TestAuthMiddleware_MalformedHeader` - Returns 401

**Integration Tests (2):**
7. `TestAuthMiddleware_Integration_ProtectedRoute` - Full stack test
8. `TestAuthMiddleware_Integration_ChainedHandlers` - Multiple middleware

#### Step 2: Implement Middleware (GREEN Phase 🟢)

```bash
# Create middleware implementation
vim internal/middleware/auth.go
```

**Required Functions:**
```go
// JWTAuth middleware - extracts and verifies JWT, adds to context
func JWTAuth(queries db.Querier) func(http.Handler) http.Handler

// Context helpers
func GetEmployeeID(ctx context.Context) (uuid.UUID, error)
func GetOrgID(ctx context.Context) (uuid.UUID, error)
func GetSessionData(ctx context.Context) (*db.GetSessionWithEmployeeRow, error)
```

#### Step 3: Update Existing Handlers (REFACTOR Phase 🧹)

```bash
# Refactor Logout to use middleware
vim internal/handlers/auth.go

# Refactor GetMe to use middleware
# Remove duplicate auth code, use context helpers instead
```

### Files to Create/Modify

**New Files:**
- `internal/middleware/auth.go` - Middleware implementation
- `internal/middleware/auth_test.go` - Unit tests
- `tests/integration/middleware_integration_test.go` - Integration tests

**Modified Files:**
- `internal/handlers/auth.go` - Refactor Logout and GetMe to use middleware

### ✅ COMPLETION SUMMARY (2025-10-28)

**Tests Implemented:**
- ✅ 9 unit tests (6 core + 3 subtests for malformed headers)
- ✅ 2 integration tests (protected route + after logout)
- ✅ **Total: 11 new tests, all passing**

**Files Created:**
- `internal/middleware/auth.go` (127 lines)
- `internal/middleware/auth_test.go` (247 lines)
- `tests/integration/middleware_integration_test.go` (169 lines)

**Helper Functions Added:**
- `GetEmployeeID(ctx)` - Extract employee ID from context
- `GetOrgID(ctx)` - Extract org ID from context
- `GetSessionData(ctx)` - Extract full session data from context

**Test Results:**
```
✅ 43/43 total tests passing
   - 14 auth utility tests
   - 13 handler tests (Login, Logout, GetMe)
   - 9 middleware unit tests
   - 2 middleware integration tests
   - 6 auth integration tests
```

**Key Achievements:**
1. ✅ Centralized authentication - eliminates 20+ lines of duplicate code per handler
2. ✅ Context-based auth data - employee_id and org_id available to all handlers
3. ✅ Database session validation - ensures logged-out users can't access protected routes
4. ✅ Comprehensive error handling - all edge cases covered (missing token, expired, malformed, etc.)
5. ✅ Production-ready - tested with real PostgreSQL, ready for use in all endpoints

**Next Step:** Use middleware in employee endpoints (Priority 2)

### Success Criteria (ALL MET ✅)

- ✅ All 11 middleware tests passing
- ✅ No code duplication
- ✅ Coverage > 95% for middleware
- ✅ Integration tests verify full stack
- ✅ Context helpers working correctly

### Benefits

1. **DRY Principle** - Auth logic written once, used everywhere
2. **Security** - Single point of enforcement, easier to audit
3. **Performance** - Can add session caching later
4. **Testability** - Test auth once, not in every handler
5. **Unblocks Everything** - All employee endpoints need this

---

## 📋 PRIORITY 2: Employee List Endpoint

**GET /employees** - List employees with filtering and pagination

### Why This is Next

- Simpler than POST (no request body validation)
- Tests org isolation and multi-tenancy
- Establishes patterns for other list endpoints
- Lower risk, good learning opportunity

### Implementation Plan

**Estimated Time**: 1 hour
**Tests to Write**: 6 (4 unit + 2 integration)

#### SQL Queries Needed

```sql
-- name: ListEmployees :many
SELECT * FROM employees
WHERE org_id = $1
  AND deleted_at IS NULL
  AND ($2::VARCHAR IS NULL OR status = $2)
  AND ($3::UUID IS NULL OR team_id = $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountEmployees :one
SELECT COUNT(*) FROM employees
WHERE org_id = $1
  AND deleted_at IS NULL
  AND ($2::VARCHAR IS NULL OR status = $2)
  AND ($3::UUID IS NULL OR team_id = $3);
```

#### Unit Tests (4)

1. `TestListEmployees_Success` - Returns multiple employees
2. `TestListEmployees_WithFilters` - Filters by status and team
3. `TestListEmployees_Pagination` - Limit and offset work
4. `TestListEmployees_EmptyResult` - No employees found

#### Integration Tests (2)

5. `TestListEmployees_Integration_MultipleEmployees` - Real DB with data
6. `TestListEmployees_Integration_OrgIsolation` - Cannot see other org's employees

### Handler Implementation

```go
// internal/handlers/employees.go
type EmployeesHandler struct {
    db db.Querier
}

func NewEmployeesHandler(database db.Querier) *EmployeesHandler {
    return &EmployeesHandler{db: database}
}

func (h *EmployeesHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
    // Get org_id from middleware context
    orgID, _ := middleware.GetOrgID(r.Context())

    // Parse query params (status, team_id, limit, offset)
    // Call db.ListEmployees()
    // Convert to API types
    // Return JSON
}
```

### Success Criteria

- ✅ 6 tests passing
- ✅ Pagination works correctly
- ✅ Filtering by status and team works
- ✅ Org isolation verified (integration test)
- ✅ Coverage > 85%

---

## 📋 PRIORITY 3: Employee Get By ID

**GET /employees/{id}** - Fetch single employee

### Implementation Plan

**Estimated Time**: 30 minutes
**Tests to Write**: 3 (2 unit + 1 integration)

#### Unit Tests (2)

1. `TestGetEmployee_Success` - Returns employee data
2. `TestGetEmployee_NotFound` - Returns 404

#### Integration Tests (1)

3. `TestGetEmployee_Integration_OrgIsolation` - Cannot fetch employee from different org

### Handler Implementation

```go
func (h *EmployeesHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
    employeeID := chi.URLParam(r, "employee_id")
    orgID, _ := middleware.GetOrgID(r.Context())

    // Fetch employee
    // Verify org_id matches
    // Return JSON
}
```

### Success Criteria

- ✅ 3 tests passing
- ✅ Org isolation enforced
- ✅ 404 for non-existent employees

---

## 📋 PRIORITY 4: Employee Create

**POST /employees** - Create new employee

### Implementation Plan

**Estimated Time**: 1.5 hours
**Tests to Write**: 9 (6 unit + 3 integration)

#### Unit Tests (6)

1. `TestCreateEmployee_Success` - Creates employee
2. `TestCreateEmployee_InvalidEmail` - Returns 400
3. `TestCreateEmployee_WeakPassword` - Returns 400
4. `TestCreateEmployee_MissingFields` - Returns 400
5. `TestCreateEmployee_InvalidRoleID` - Returns 400
6. `TestCreateEmployee_InvalidTeamID` - Returns 400

#### Integration Tests (3)

7. `TestCreateEmployee_Integration_Success` - Real DB creation
8. `TestCreateEmployee_Integration_DuplicateEmail` - Email uniqueness
9. `TestCreateEmployee_Integration_OrgIsolation` - Email unique per org

### Validation Requirements

```go
// Email validation
- Valid email format
- Unique within organization
- Max 255 characters

// Password validation
- Minimum 8 characters
- At least one uppercase, lowercase, number, special char
- Not in common passwords list (optional)

// Role/Team validation
- Role ID must exist
- Team ID must exist and belong to org
- Team ID is optional
```

### Success Criteria

- ✅ 9 tests passing
- ✅ All validation rules enforced
- ✅ Password hashed with bcrypt
- ✅ Email uniqueness verified (integration)

---

## 📋 PRIORITY 5: Employee Update

**PATCH /employees/{id}** - Update employee (partial updates)

### Implementation Plan

**Estimated Time**: 1 hour
**Tests to Write**: 7 (5 unit + 2 integration)

#### Unit Tests (5)

1. `TestUpdateEmployee_Success_FullUpdate` - Updates all fields
2. `TestUpdateEmployee_Success_PartialUpdate` - Updates some fields
3. `TestUpdateEmployee_NotFound` - Returns 404
4. `TestUpdateEmployee_InvalidData` - Returns 400
5. `TestUpdateEmployee_CannotChangeOrgID` - Security check

#### Integration Tests (2)

6. `TestUpdateEmployee_Integration_Success` - Real DB update
7. `TestUpdateEmployee_Integration_CannotUpdateDeleted` - Soft delete check

### SQL Query

```sql
-- name: UpdateEmployee :one
UPDATE employees
SET
    team_id = COALESCE($2, team_id),
    role_id = COALESCE($3, role_id),
    full_name = COALESCE($4, full_name),
    status = COALESCE($5, status),
    preferences = COALESCE($6, preferences),
    updated_at = NOW()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING *;
```

### Success Criteria

- ✅ 7 tests passing
- ✅ Partial updates work (COALESCE)
- ✅ Cannot update deleted employees
- ✅ Cannot change org_id (security)

---

## 📋 PRIORITY 6: Employee Delete

**DELETE /employees/{id}** - Soft delete employee

### Implementation Plan

**Estimated Time**: 45 minutes
**Tests to Write**: 4 (2 unit + 2 integration)

#### Unit Tests (2)

1. `TestDeleteEmployee_Success` - Soft deletes employee
2. `TestDeleteEmployee_NotFound` - Returns 404

#### Integration Tests (2)

3. `TestDeleteEmployee_Integration_SoftDelete` - Sets deleted_at timestamp
4. `TestDeleteEmployee_Integration_CannotLogin` - Deleted employee cannot login

### SQL Query

```sql
-- name: SoftDeleteEmployee :exec
UPDATE employees
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;
```

### Success Criteria

- ✅ 4 tests passing
- ✅ Soft delete (sets deleted_at)
- ✅ Idempotent (can delete twice)
- ✅ Deleted employees cannot login (integration)

---

## 📋 PRIORITY 7: Organization Endpoints (Lower Priority)

### GET /organizations/current

**Estimated Time**: 20 minutes
**Tests**: 2

Returns current organization details based on org_id from JWT.

### GET /roles

**Estimated Time**: 20 minutes
**Tests**: 2

Returns list of available roles (for dropdowns in UI).

---

## 📊 Summary Table

| Priority | Endpoint | Method | Tests | Est. Time | Status |
|----------|----------|--------|-------|-----------|--------|
| 0 | /auth/login | POST | 9 | - | ✅ Complete |
| 0 | /auth/logout | POST | 4 | - | ✅ Complete |
| 0 | /auth/me | GET | 5 | - | ✅ Complete |
| 0 | **JWT Middleware** | **-** | **11** | **-** | ✅ **Complete** |
| **1** | **/employees** | **GET** | **6** | **1 hr** | **⏳ NEXT** |
| 2 | /employees/{id} | GET | 3 | 30 min | ⏸️ Pending |
| 3 | /employees | POST | 9 | 1.5 hrs | ⏸️ Pending |
| 4 | /employees/{id} | PATCH | 7 | 1 hr | ⏸️ Pending |
| 5 | /employees/{id} | DELETE | 4 | 45 min | ⏸️ Pending |
| 6 | /organizations/current | GET | 2 | 20 min | ⏸️ Pending |
| 6 | /roles | GET | 2 | 20 min | ⏸️ Pending |

**Completed**: 3 endpoints + middleware (43 tests passing)
**Remaining**: 7 endpoints (~5-7 hours to complete all)

---

## 🎯 Completion Milestones

### Milestone 1: Auth Complete ✅ (2025-10-28)
- All authentication endpoints working
- 33/33 tests passing
- ~85% coverage
- Login, Logout, GetMe fully functional

### Milestone 2: Middleware Complete ✅ (2025-10-28)
- JWT middleware tested and working
- Context-based auth data propagation
- Code duplication eliminated
- **43/43 tests passing**
- Ready for all protected endpoints

### Milestone 3: Employee CRUD Complete (Next - ~5-7 hours)
- All 5 employee endpoints working
- ~70+ tests passing
- Full CRUD operations tested
- Multi-tenancy verified

### Milestone 4: API v1 Complete (Final)
- All 10 endpoints operational
- ~75+ tests passing
- Ready for frontend integration

---

## 🔄 TDD Workflow for Each Endpoint

**For every endpoint, follow this exact sequence:**

```bash
# 1. Write unit tests FIRST (RED 🔴)
vim internal/handlers/{handler}_test.go
go test -v -short ./internal/handlers -run Test{Endpoint}
# Expected: FAIL

# 2. Add SQL queries if needed
vim sqlc/queries/{resource}.sql
make generate-db && make generate-mocks

# 3. Implement handler (GREEN 🟢)
vim internal/handlers/{handler}.go
go test -v -short ./internal/handlers -run Test{Endpoint}
# Expected: PASS

# 4. Write integration tests
vim tests/integration/{handler}_integration_test.go
go test -v -run Test{Endpoint}_Integration ./tests/integration
# Expected: PASS

# 5. Run full test suite
go test -v ./...
# Expected: All tests pass
```

---

## 📈 Coverage Goals

**Current Coverage**: ~85%

**Target Coverage**:
- Auth utilities: 95%+ ✅ (88.2% - close!)
- HTTP handlers: 90%+ ✅ (85%+ achieved)
- Middleware: 95%+ (target after Priority 1)
- Integration: 80%+ of user flows ✅
- **Overall target**: 90%+

---

## 🚨 Critical Reminders

### Before Starting Each Endpoint

1. ✅ **Run all existing tests** - Ensure nothing broken
2. ✅ **Update OpenAPI spec** - If endpoint signature changes
3. ✅ **Add SQL queries first** - Before handler implementation
4. ✅ **Regenerate code** - `make generate-db && make generate-mocks`
5. ✅ **Write tests FIRST** - TDD is mandatory

### After Completing Each Endpoint

1. ✅ **All tests passing** - `go test ./...`
2. ✅ **Check coverage** - `go test -coverprofile=coverage.out ./...`
3. ✅ **Integration tests pass** - With real PostgreSQL
4. ✅ **Update COVERAGE_ANALYSIS.md** - Document new tests
5. ✅ **Update this roadmap** - Mark as complete

---

## 📚 Reference Documents

- **[CLAUDE.md](./CLAUDE.md)** - Main project documentation + TDD guide
- **[COVERAGE_ANALYSIS.md](./COVERAGE_ANALYSIS.md)** - Current test coverage status
- **[docs/TESTING_STRATEGY.md](./docs/TESTING_STRATEGY.md)** - Complete testing guide
- **[docs/DEVELOPMENT_APPROACH.md](./docs/DEVELOPMENT_APPROACH.md)** - TDD vs implementation-first

---

## ✅ Success Criteria for "Done"

An endpoint is considered **DONE** when:

1. ✅ All unit tests passing (>90% coverage)
2. ✅ Integration tests passing (real DB)
3. ✅ OpenAPI spec updated
4. ✅ Handler documented with comments
5. ✅ No race conditions (`go test -race`)
6. ✅ Org isolation verified (multi-tenancy)
7. ✅ Error cases tested (400, 401, 404, etc.)
8. ✅ Code reviewed (or self-reviewed)

---

**Last Updated**: 2025-10-28
**Next Action**: Implement JWT Middleware (Priority 1)
