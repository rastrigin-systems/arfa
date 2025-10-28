# Testing Quickstart — Pivot Enterprise

**5-Minute Guide to Testing Setup**

---

## TL;DR

```bash
# 1. Install mock generation tool
make install-tools

# 2. Generate mocks from database interface
make generate-mocks

# 3. Run tests
make test

# 4. View coverage
make test-coverage
```

---

## What Gets Generated?

### Automatic Mock Generation

```
generated/
├── db/
│   └── querier.go          # ← Source interface (27 methods)
└── mocks/
    └── db_mock.go          # ← Generated mock (auto-created)
```

**Key Point**: Run `make generate-mocks` after any database query changes.

---

## Quick Test Example

```go
package handlers_test

import (
	"testing"
	"go.uber.org/mock/gomock"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

func TestGetEmployee(t *testing.T) {
	// 1. Create mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 2. Create mock database
	mockDB := mocks.NewMockQuerier(ctrl)

	// 3. Set expectations
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), gomock.Any()).
		Return(db.Employee{Email: "test@example.com"}, nil)

	// 4. Test your handler with mockDB
	// handler := NewEmployeeHandler(mockDB)
	// result := handler.GetEmployee(...)
	// assert.Equal(t, "test@example.com", result.Email)
}
```

---

## Test Types

### 1. Unit Tests (Fast - Use Mocks)

```bash
make test-unit
```

**What to test**:
- Handler logic
- Service business logic
- Input validation
- Error handling

**Mock the database** — no real PostgreSQL needed.

### 2. Integration Tests (Slower - Real DB)

```bash
make test-integration
```

**What to test**:
- SQL query correctness
- Full HTTP → Handler → DB → Response flow
- Multi-tenant isolation
- Join logic

**Use testcontainers** — spins up real PostgreSQL per test.

### 3. Full Coverage Report

```bash
make test-coverage
# Opens coverage.html in browser
```

---

## Testing Checklist

### ✅ DO Test These

- [ ] Handler input validation
- [ ] Handler error responses (400, 404, 500)
- [ ] Service business logic
- [ ] Type mappers (DB → API)
- [ ] SQL query behavior (integration tests)
- [ ] OpenAPI contract compliance

### ❌ DON'T Test These

- [ ] ~~Generated code in `generated/`~~
- [ ] ~~Database schema DDL~~
- [ ] ~~sqlc CRUD methods~~

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

---

## OpenAPI Validation

### Ensuring Spec Matches Implementation

**Add OpenAPI validator middleware** (future):

```go
router.Use(middleware.OpenAPIValidator("openapi/spec.yaml"))
```

This validates:
- Request structure matches spec
- Response structure matches spec
- Status codes are defined
- Content-Type headers correct

---

## Schema Alignment

### Problem: DB Models ≠ API DTOs

**Solution: Explicit mapper layer**

```go
// mapper/employee.go
func DBEmployeeToAPI(emp db.Employee) api.EmployeeResponse {
	return api.EmployeeResponse{
		ID:     emp.ID.String(),
		Email:  emp.Email,
		Status: emp.Status,
		// ⚠️ Exclude password_hash from API
	}
}
```

**Test the mapper**:

```go
func TestMapper_ExcludesPasswordHash(t *testing.T) {
	dbEmp := db.Employee{
		Email:        "test@example.com",
		PasswordHash: "secret",
	}
	apiEmp := mapper.DBEmployeeToAPI(dbEmp)
	json, _ := json.Marshal(apiEmp)
	assert.NotContains(t, string(json), "secret")
}
```

---

## Updated Makefile Targets

```bash
make install-tools      # Install mockgen + other tools
make generate-mocks     # Generate mocks from db.Querier
make generate           # Generate all (ERD + API + DB + Mocks)

make test               # Run all tests with coverage
make test-unit          # Fast unit tests only
make test-integration   # Integration tests (Docker required)
make test-coverage      # HTML coverage report
```

---

## Next Steps

1. **Read**: [docs/TESTING_STRATEGY.md](./TESTING_STRATEGY.md) for complete guide
2. **Run**: `make generate-mocks` to create mock files
3. **Write**: First test in `internal/handlers/auth_test.go`
4. **Verify**: `make test` passes
5. **Check**: `make test-coverage` shows >80%

---

## Key Files

| File | Purpose |
|------|---------|
| `generated/db/querier.go` | Interface (source for mocks) |
| `generated/mocks/db_mock.go` | Generated mock (don't edit!) |
| `internal/handlers/*_test.go` | Unit tests (use mocks) |
| `tests/integration/*_test.go` | Integration tests (real DB) |
| `go.mod` | Includes `go.uber.org/mock` |

---

## Dependencies Added

```go
require (
	github.com/stretchr/testify v1.8.4
	go.uber.org/mock v0.4.0
)
```

Run `go mod download` to install.

---

## Summary

✅ **All mocks are auto-generated** — no manual maintenance
✅ **Use unit tests (mocks) for speed** — test handler logic
✅ **Use integration tests (real DB) for correctness** — test SQL queries
✅ **Create mapper layer** — bridge DB ↔ API types
✅ **Test mappers** — verify schema alignment

**Target**: 80% coverage excluding generated code.
