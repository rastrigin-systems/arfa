# Testing Guide — Pivot Enterprise Platform

**Complete guide to testing in the Pivot platform**

---

## Table of Contents
1. [Quick Start](#quick-start)
2. [Testing Philosophy](#testing-philosophy)
3. [TDD Workflow](#tdd-workflow)
4. [Test Types](#test-types)
5. [Common Patterns](#common-patterns)
6. [Commands](#commands)
7. [Troubleshooting](#troubleshooting)

---

## Quick Start

### TL;DR - 5 Minutes

```bash
# 1. Install tools
make install-tools

# 2. Generate mocks
make generate-mocks

# 3. Run tests
make test

# 4. View coverage
make test-coverage
```

### Quick Test Example

```go
package handlers_test

import (
	"testing"
	"go.uber.org/mock/gomock"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

func TestGetEmployee_Success(t *testing.T) {
	// 1. Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 2. Create mock database
	mockDB := mocks.NewMockQuerier(ctrl)

	// 3. Set expectations
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), gomock.Any()).
		Return(db.Employee{Email: "test@example.com"}, nil)

	// 4. Test your handler
	handler := NewEmployeeHandler(mockDB)
	// ... test assertions
}
```

---

## Testing Philosophy

### What to Test

✅ **YES - Test These**:
- Handler logic and validation
- Service business logic
- Type mappers (DB ↔ API)
- SQL query behavior (integration tests)
- Error handling and edge cases
- OpenAPI contract compliance

❌ **NO - Don't Test These**:
- Generated code in `generated/`
- Database schema DDL
- Third-party libraries

### Testing Pyramid

```
        ┌─────────────┐
        │ Integration │  30% - Handler + Real DB
        │  (30 tests) │
        └─────────────┘
     ┌───────────────────┐
     │  Unit Tests       │  70% - Logic + Mocked DB
     │   (70 tests)      │
     └───────────────────┘
```

**Target Coverage**: 85% overall (excluding generated code)

---

## TDD Workflow

### RED → GREEN → REFACTOR

#### 1. Write Tests FIRST (RED Phase 🔴)

```bash
# Write failing test
vim internal/handlers/employees_test.go

# Run tests - they should FAIL
go test -v -short ./internal/handlers
```

**Expected**: Tests fail because handler doesn't exist yet. This is GOOD!

#### 2. Implement Code (GREEN Phase 🟢)

```bash
# Add SQL queries if needed
vim platform/database/sqlc/queries/employees.sql

# Regenerate DB code and mocks
make generate-db && make generate-mocks

# Implement handler (for API service)
vim services/api/internal/handlers/employees.go

# Run tests - they should PASS
go test -v -short ./internal/handlers
```

#### 3. Add Integration Tests (FULL STACK 🔄)

```bash
# Write integration test with REAL database
vim tests/integration/employees_integration_test.go

# Run integration test
go test -v -run TestEmployees_Integration ./tests/integration
```

#### 4. Refactor (CLEAN UP 🧹)

```bash
# Run all tests to ensure nothing breaks
go test -v ./...

# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## Test Types

### 1. Unit Tests (Fast - Use Mocks)

**Run with:** `go test -v -short ./internal/...`

**When to use**: Testing handler logic, validation, error paths

**Example**:
```go
func TestLogout_Success(t *testing.T) {
	mockDB.EXPECT().DeleteSession(gomock.Any(), tokenHash).Return(nil)
	handler.Logout(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}
```

**Coverage Target**: 90%+ for handlers

### 2. Integration Tests (Slow - Real DB)

**Run with:** `go test -v -run Integration ./tests/integration`

**When to use**: Testing full stack with database side effects

**Example**:
```go
func TestLogin_Integration_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup real database with testcontainers
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

**Coverage Target**: 80%+ of critical user flows

**What Integration Tests Catch** (That Unit Tests Miss):
- Missing database columns
- Type conversion issues
- Schema drift
- Multi-tenant isolation

---

## Common Patterns

### Pattern 1: Table-Driven Tests

```go
tests := []struct {
	name           string
	input          string
	setupMock      func(*mocks.MockQuerier)
	expectedStatus int
}{
	{name: "success", input: "valid", expectedStatus: 200, setupMock: ...},
	{name: "not found", input: "missing", expectedStatus: 404, setupMock: ...},
	{name: "unauthorized", input: "invalid", expectedStatus: 401, setupMock: ...},
}

for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockDB := mocks.NewMockQuerier(ctrl)
		tt.setupMock(mockDB)
		// ... test logic
	})
}
```

### Pattern 2: Multiple Expectations

```go
mockDB.EXPECT().
	GetEmployeeByEmail(ctx, email).
	Return(db.Employee{}, pgx.ErrNoRows) // First call

mockDB.EXPECT().
	CreateEmployee(ctx, gomock.Any()).
	Return(db.Employee{...}, nil)         // Second call
```

### Pattern 3: Any Matcher

```go
mockDB.EXPECT().
	GetEmployee(
		gomock.Any(),           // Any context
		gomock.Eq(employeeID),  // Exact UUID match
	).Return(...)
```

### Pattern 4: JWT Authentication Flow

```go
// Extract token from Authorization header
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

### Pattern 5: Type Conversion (DB ↔ API)

```go
// Convert database types to OpenAPI types
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

### Pattern 6: Test Fixture Creation

```go
// Use testutil helpers for consistent test data
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

## Commands

### Code Generation

```bash
make install-tools      # Install mockgen + other tools
make generate-mocks     # Generate mocks from db.Querier
make generate           # Generate all (ERD + API + DB + Mocks)
```

### Running Tests

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

### Makefile Shortcuts

```bash
make test               # Run all tests with coverage
make test-unit          # Fast unit tests only
make test-integration   # Integration tests (Docker required)
make test-coverage      # HTML coverage report
```

---

## Troubleshooting

### Mock Generation Version Mismatch

**Problem**: Using wrong mockgen version

```bash
# ❌ Wrong: Old mockgen (github.com/golang/mock)
go install github.com/golang/mock/mockgen@latest

# ✅ Correct: New mockgen (go.uber.org/mock)
go install go.uber.org/mock/mockgen@latest
```

**Always use**: `go.uber.org/mock` (matches import in code)

### SQL Query vs Schema Mismatch

**Problem**: Query references column that doesn't exist

```sql
-- Query references deleted_at
SELECT * FROM employees WHERE deleted_at IS NULL

-- But schema is missing the column!
```

**Solution**: Always run integration tests to catch this!

```bash
make db-reset           # Apply schema changes
go test -v -run Integration ./tests/integration
```

### Testcontainers Docker Issues

**Problem**: Integration tests hang or fail to start

```bash
# Check Docker is running
docker ps

# Check testcontainers can access Docker
docker run hello-world

# Ensure platform/database/schema.sql path is correct
schemaPath, err := filepath.Abs("../../platform/database/schema.sql")
```

### Coverage Not Calculating

**Problem**: Generated code skewing coverage

**Solution**: Exclude generated directories

```bash
go test -coverprofile=coverage.out ./...
# Manually exclude or use:
go test -coverprofile=coverage.out $(go list ./... | grep -v generated)
```

---

## Test Success Criteria

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

## Key Files

| File | Purpose |
|------|---------|
| `generated/db/querier.go` | Interface (source for mocks) |
| `generated/mocks/db_mock.go` | Generated mock (don't edit!) |
| `internal/handlers/*_test.go` | Unit tests (use mocks) |
| `tests/integration/*_test.go` | Integration tests (real DB) |
| `tests/testutil/db.go` | Testcontainers setup |
| `tests/testutil/fixtures.go` | Test data factories |

---

## Dependencies

```go
require (
	github.com/stretchr/testify v1.8.4
	go.uber.org/mock v0.4.0
	github.com/testcontainers/testcontainers-go v0.32.0
)
```

Run `go mod download` to install.

---

## Summary

✅ **All mocks are auto-generated** — no manual maintenance
✅ **Use unit tests (mocks) for speed** — test handler logic
✅ **Use integration tests (real DB) for correctness** — test SQL queries
✅ **Follow TDD workflow** — RED → GREEN → REFACTOR
✅ **Integration tests catch what unit tests miss** — schema drift, type mismatches

**Target**: 85% coverage excluding generated code.

**TDD provides:**
- ✅ Confidence in refactoring
- ✅ Living documentation
- ✅ Faster debugging (tests show exactly what broke)
- ✅ Better API design (think from caller's perspective)
