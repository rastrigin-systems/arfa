# Development Approach — TDD vs Implementation-First

**Last Updated**: 2025-10-28
**Current Phase**: Phase 2 - Core API Implementation

---

## TL;DR Recommendation

**Use Hybrid Approach**: Implementation-first with immediate test coverage

**Why?**
- ✅ You have stable foundations (schema, OpenAPI spec, generated code)
- ✅ Requirements are clear (OpenAPI spec = contract)
- ✅ Faster initial progress on MVP features
- ✅ Tests validate implementation matches spec
- ⚠️ Pure TDD slower when contract already defined

**When to use strict TDD**:
- Complex business logic (approval workflows, policy enforcement)
- Edge cases and error handling
- Security-critical features (auth, RLS)

---

## Current Project State

### ✅ What You Have (Phase 1 Complete)

```
Database Schema          OpenAPI Spec           Generated Code
     ↓                        ↓                       ↓
20 tables + 3 views     10 endpoints          Types + Interfaces
RLS policies            Request/Response      Database queries
Seed data               Error schemas         Router setup
```

**Key Insight**: Your **contracts are already defined** (OpenAPI spec + DB schema).

### ❌ What You Don't Have (Phase 2 Needed)

```
internal/
├── handlers/              # ❌ Empty - needs implementation
├── service/               # ❌ Empty - needs business logic
├── middleware/            # ❌ Empty - needs auth/RLS
├── mapper/                # ❌ Empty - needs DB ↔ API conversion
└── validation/            # ❌ Empty - needs input validation

cmd/
└── server/main.go         # ❌ Doesn't exist yet

tests/
├── integration/           # ❌ No tests yet
└── testutil/              # ❌ No test helpers yet
```

---

## Recommended Approach: "Contract-First Development"

### Strategy

Since you have **OpenAPI spec** (contract), use it to drive implementation:

```
OpenAPI Spec (Contract)
    ↓
1. Generate API types          ✅ Done (make generate-api)
    ↓
2. Implement handlers          ← You are here
    ↓
3. Write tests immediately     ← Validate against spec
    ↓
4. Test validates contract     ← OpenAPI validator middleware
```

### Why This Works

1. **Spec = Requirements** - OpenAPI spec defines exact behavior
2. **Generated Types** - Eliminates type mismatches
3. **Fast Feedback** - Tests validate spec compliance
4. **Refactor Safely** - Tests catch regressions

---

## Phase 2 Implementation Plan

### Week 1: Authentication (3-4 days)

#### Day 1-2: Core Auth Implementation

**Step 1: Write handler signature (5 min)**
```go
// internal/handlers/auth.go
type AuthHandler struct {
	db db.Querier
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// TODO
}
```

**Step 2: Write test FIRST (TDD for business logic)** (20 min)
```go
// internal/handlers/auth_test.go
func TestLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	mockDB.EXPECT().
		GetEmployeeByEmail(gomock.Any(), "user@example.com").
		Return(db.Employee{
			PasswordHash: hashPassword("correct"),
		}, nil)

	mockDB.EXPECT().
		CreateSession(gomock.Any(), gomock.Any()).
		Return(db.Session{Token: "jwt-token"}, nil)

	handler := NewAuthHandler(mockDB)

	body := `{"email":"user@example.com","password":"correct"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	assert.Equal(t, 200, rec.Code)
	var resp api.LoginResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NotEmpty(t, resp.Token)
}
```

**Step 3: Implement to make test pass** (30 min)
```go
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req api.LoginRequest
	json.NewDecoder(r.Body).Decode(&req)

	// Get employee
	emp, err := h.db.GetEmployeeByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", 401)
		return
	}

	// Verify password
	if !verifyPassword(req.Password, emp.PasswordHash) {
		http.Error(w, "Invalid credentials", 401)
		return
	}

	// Create session
	token := generateJWT(emp.ID)
	session, _ := h.db.CreateSession(r.Context(), db.CreateSessionParams{
		EmployeeID: emp.ID,
		TokenHash:  hashToken(token),
	})

	// Return token
	json.NewEncoder(w).Encode(api.LoginResponse{Token: token})
}
```

**Step 4: Add error cases (TDD)** (30 min)
```go
func TestLogin_InvalidPassword(t *testing.T) { ... }
func TestLogin_UserNotFound(t *testing.T) { ... }
func TestLogin_InvalidJSON(t *testing.T) { ... }
```

**Step 5: Add integration test** (20 min)
```go
// tests/integration/auth_integration_test.go
func TestAPI_Login_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)
	router := setupRouter(db)

	// Create test user
	testutil.CreateTestEmployee(t, db, "user@example.com", "password")

	// Test login
	body := `{"email":"user@example.com","password":"password"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, 200, rec.Code)
}
```

**Total Time: ~2 hours per endpoint**

#### Day 3: JWT Middleware + Auth Tests

**TDD Approach**:
1. Write test for JWT extraction → Implement
2. Write test for token validation → Implement
3. Write test for expired tokens → Implement
4. Write test for missing tokens → Implement

#### Day 4: Auth Completion

- Integration tests with real DB
- OpenAPI validation
- Error handling polish
- Documentation

---

### Week 1-2: Employee Management (3-4 days)

**Pattern: Implementation-First with Immediate Tests**

For each endpoint (`GET /employees`, `POST /employees`, etc.):

#### 1. Implement Happy Path (30 min)
```go
func (h *EmployeeHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	employees, _ := h.db.ListEmployees(r.Context(), ...)
	apiEmployees := mapper.DBEmployeesToAPI(employees)
	json.NewEncoder(w).Encode(apiEmployees)
}
```

#### 2. Write Unit Tests (30 min)
```go
func TestListEmployees_Success(t *testing.T) { ... }
func TestListEmployees_EmptyResult(t *testing.T) { ... }
func TestListEmployees_DatabaseError(t *testing.T) { ... }
```

#### 3. Add Error Handling (20 min)
```go
employees, err := h.db.ListEmployees(r.Context(), ...)
if err != nil {
	http.Error(w, "Internal server error", 500)
	return
}
```

#### 4. Integration Test (20 min)
```go
func TestAPI_ListEmployees_Integration(t *testing.T) {
	// Test with real DB
}
```

**Total: ~2 hours per endpoint × 5 endpoints = 10 hours**

---

### Week 2: Organization API (2 days)

Follow same pattern:
1. Implement handler (30 min)
2. Unit tests (30 min)
3. Integration test (20 min)
4. Polish error handling (20 min)

---

## TDD vs Implementation-First: Decision Matrix

| Scenario | Approach | Reason |
|----------|----------|--------|
| **CRUD endpoints** | Implementation-first + immediate tests | Spec defines behavior, straightforward logic |
| **Business logic** (approval workflows) | **Pure TDD** | Complex rules, edge cases unknown |
| **Security features** (auth, RLS) | **Pure TDD** | Critical correctness, test all cases |
| **Mappers** (DB ↔ API) | Implementation-first + immediate tests | Simple transformations |
| **Middleware** | **Pure TDD** | Cross-cutting concerns, many edge cases |
| **Validation** | **Pure TDD** | Many edge cases, security implications |

---

## Recommended Next Steps (Prioritized)

### 🎯 Immediate Next Steps (Week 1)

#### 1. Setup Test Infrastructure (2 hours)

```bash
# Create test helpers
mkdir -p tests/testutil
vim tests/testutil/db.go          # testcontainers setup
vim tests/testutil/fixtures.go    # Test data factories
vim tests/testutil/assertions.go  # Custom assertions

# Generate mocks
make generate-mocks

# Verify test setup works
go test ./tests/testutil/...
```

**Files to create**:
- `tests/testutil/db.go` - PostgreSQL testcontainers setup
- `tests/testutil/fixtures.go` - Factory functions for test data
- `tests/testutil/server.go` - HTTP server setup helper

#### 2. Implement Authentication (Days 1-3)

**Order**:
1. JWT helper functions (TDD)
   - `generateJWT()` - test first
   - `verifyJWT()` - test first
   - `hashPassword()` - test first

2. Login handler (hybrid)
   - Write test for success case
   - Implement handler
   - Write tests for error cases
   - Implement error handling

3. JWT middleware (TDD)
   - Write test for valid token
   - Write test for expired token
   - Write test for missing token
   - Implement middleware
   - All tests should pass

4. Integration tests (implementation-first)
   - Full login flow with real DB
   - Token refresh flow

#### 3. Employee Management API (Days 4-6)

**For each endpoint** (GET, POST, PATCH, DELETE):

```go
// 1. Handler stub (5 min)
func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	// TODO
}

// 2. Unit test (15 min)
func TestGetEmployee_Success(t *testing.T) { ... }

// 3. Implement (15 min)
// 4. Error tests (15 min)
// 5. Integration test (15 min)
```

**Total per endpoint**: ~1 hour
**5 endpoints**: ~5 hours

#### 4. Mapper Layer (Day 7)

**Use TDD for security-critical mappings**:

```go
// Test FIRST
func TestMapper_ExcludesPasswordHash(t *testing.T) {
	dbEmp := db.Employee{PasswordHash: "secret"}
	apiEmp := mapper.DBEmployeeToAPI(dbEmp)
	json, _ := json.Marshal(apiEmp)
	assert.NotContains(t, string(json), "secret")
	assert.NotContains(t, string(json), "password")
}

// Then implement
func DBEmployeeToAPI(emp db.Employee) api.EmployeeResponse {
	return api.EmployeeResponse{
		ID:     emp.ID.String(),
		Email:  emp.Email,
		// Exclude PasswordHash
	}
}
```

#### 5. Integration with OpenAPI Validation (Day 7)

```go
// Add validator middleware
router.Use(middleware.OpenAPIValidator("openapi/spec.yaml"))

// Run all integration tests
make test-integration

// Any spec violation = test failure
```

---

## Week 1 Detailed Schedule

### Monday: Test Infrastructure + Auth Setup

**Morning (4 hours)**:
- [ ] Create `tests/testutil/db.go` with testcontainers
- [ ] Create `tests/testutil/fixtures.go` with factory functions
- [ ] Create `tests/testutil/server.go` with HTTP helpers
- [ ] Run `make generate-mocks`
- [ ] Write example test to verify setup

**Afternoon (4 hours)**:
- [ ] Implement JWT helpers (TDD)
  - [ ] Test: `TestGenerateJWT_ValidClaims`
  - [ ] Implement: `generateJWT()`
  - [ ] Test: `TestVerifyJWT_ValidToken`
  - [ ] Implement: `verifyJWT()`
  - [ ] Test: `TestHashPassword_BCrypt`
  - [ ] Implement: `hashPassword()`, `verifyPassword()`

### Tuesday: Login Handler

**Morning (4 hours)**:
- [ ] Test: `TestLogin_Success` (write first)
- [ ] Implement: `AuthHandler.Login()` happy path
- [ ] Test: `TestLogin_InvalidPassword`
- [ ] Implement: password verification
- [ ] Test: `TestLogin_UserNotFound`
- [ ] Implement: error handling
- [ ] Test: `TestLogin_InvalidJSON`
- [ ] Implement: request validation

**Afternoon (4 hours)**:
- [ ] Integration test: `TestAPI_Login_Integration`
- [ ] Test: `TestLogin_CreatesSession`
- [ ] Implement: session creation
- [ ] Test: `TestLogin_UpdatesLastLogin`
- [ ] Implement: last login update

### Wednesday: JWT Middleware + Logout

**Morning (4 hours)**:
- [ ] Test: `TestAuthMiddleware_ValidToken` (TDD)
- [ ] Test: `TestAuthMiddleware_ExpiredToken`
- [ ] Test: `TestAuthMiddleware_MissingToken`
- [ ] Test: `TestAuthMiddleware_MalformedToken`
- [ ] Implement: `AuthMiddleware()`
- [ ] All tests pass

**Afternoon (4 hours)**:
- [ ] Test: `TestLogout_Success`
- [ ] Implement: `AuthHandler.Logout()`
- [ ] Test: `TestLogout_InvalidToken`
- [ ] Integration test: `TestAPI_Logout_Integration`
- [ ] Test: `TestGetMe_Success`
- [ ] Implement: `AuthHandler.GetMe()`

### Thursday: Employee Management - Read Operations

**Morning (4 hours)**:
- [ ] Test: `TestListEmployees_Success`
- [ ] Implement: `EmployeeHandler.ListEmployees()`
- [ ] Test: `TestListEmployees_WithFilters`
- [ ] Implement: query parameter parsing
- [ ] Test: `TestListEmployees_Pagination`
- [ ] Implement: pagination logic

**Afternoon (4 hours)**:
- [ ] Test: `TestGetEmployee_Success`
- [ ] Implement: `EmployeeHandler.GetEmployee()`
- [ ] Test: `TestGetEmployee_NotFound`
- [ ] Implement: error handling
- [ ] Integration test: `TestAPI_GetEmployee_Integration`

### Friday: Employee Management - Write Operations

**Morning (4 hours)**:
- [ ] Test: `TestCreateEmployee_Success`
- [ ] Implement: `EmployeeHandler.CreateEmployee()`
- [ ] Test: `TestCreateEmployee_DuplicateEmail`
- [ ] Implement: duplicate detection
- [ ] Test: `TestCreateEmployee_InvalidData`
- [ ] Implement: validation

**Afternoon (4 hours)**:
- [ ] Test: `TestUpdateEmployee_Success`
- [ ] Implement: `EmployeeHandler.UpdateEmployee()`
- [ ] Test: `TestDeleteEmployee_Success` (soft delete)
- [ ] Implement: `EmployeeHandler.DeleteEmployee()`
- [ ] Integration tests for all endpoints
- [ ] Run `make test-coverage` - verify >80%

---

## Testing Workflow

### Daily Cycle

```bash
# Morning: Start fresh
make db-reset
make generate

# During development
make test-unit           # Fast feedback loop
make test-integration    # After each feature

# Before commit
make test                # Full test suite
make test-coverage       # Check coverage
git add . && git commit
```

### Pre-Commit Checklist

- [ ] All tests pass (`make test`)
- [ ] Coverage >80% for new code (`make test-coverage`)
- [ ] No `.only` or `.skip` in tests
- [ ] Integration tests pass
- [ ] OpenAPI validation passes
- [ ] No `TODO` comments in production code

---

## Key Principles

### 1. **Test Business Logic with TDD**

```go
// Complex approval logic → TDD
func TestApprovalWorkflow_RequiresTwoManagers(t *testing.T) {
	// Write test first, define behavior
}
```

### 2. **Test CRUD with Implementation-First**

```go
// Simple CRUD → Implement then test
func (h *Handler) GetEmployee(...) { /* implement */ }
func TestGetEmployee_Success(t *testing.T) { /* validate */ }
```

### 3. **Always Write Tests Immediately**

```go
// ❌ BAD: Implement many endpoints, test later
func GetEmployee() { ... }
func ListEmployees() { ... }
func CreateEmployee() { ... }
// TODO: Write tests

// ✅ GOOD: Implement → Test → Next
func GetEmployee() { ... }
func TestGetEmployee() { ... }  // Write now!
func ListEmployees() { ... }
func TestListEmployees() { ... }  // Write now!
```

### 4. **Integration Tests Catch Gaps**

```go
// Unit tests with mocks might miss issues
// Integration tests with real DB catch them
func TestAPI_CreateEmployee_Integration(t *testing.T) {
	db := testutil.SetupTestDB(t)  // Real PostgreSQL
	// Catches: SQL errors, constraint violations, RLS issues
}
```

---

## Measuring Progress

### Daily Metrics

```bash
make test-coverage
# Target: Total coverage: 82.5%

# Breakdown:
# - Handlers: 90%+
# - Services: 85%+
# - Mappers: 95%+
# - Middleware: 80%+
```

### Weekly Milestone

**Week 1 Goal**:
- [ ] Authentication working (login, logout, JWT)
- [ ] Employee CRUD operational
- [ ] 15+ handler tests written
- [ ] 10+ integration tests passing
- [ ] >80% coverage

---

## Summary: Best Next Step

### 🎯 Recommended Path

**Start with**: Hybrid approach (implementation-first + immediate tests)

**Step 1 (Today)**: Setup test infrastructure
```bash
# 1. Create test helpers (2 hours)
mkdir -p tests/testutil
vim tests/testutil/db.go
vim tests/testutil/fixtures.go

# 2. Generate mocks
make generate-mocks

# 3. Write first test
vim internal/handlers/auth_test.go
```

**Step 2 (Days 1-3)**: Authentication (TDD for security)
- JWT helpers (TDD - test first)
- Login handler (hybrid - implement + test immediately)
- Middleware (TDD - test first)

**Step 3 (Days 4-7)**: Employee API (implementation-first + immediate tests)
- Implement endpoint → Write tests → Next endpoint

**Why This Works**:
- ✅ Faster progress than pure TDD
- ✅ High test coverage from day 1
- ✅ OpenAPI spec provides contract
- ✅ TDD where it matters (security, complex logic)

---

## Final Recommendation

**Use TDD selectively, test immediately always**

```
Complex/Security    →  Pure TDD (test → implement → refactor)
Simple CRUD        →  Implement → Test → Next
Everything         →  Test coverage required before moving on
```

**Success = 80% coverage + All endpoints working + OpenAPI validated**

Start tomorrow with test infrastructure setup! 🚀
