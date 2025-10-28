# Testing Strategy — Pivot Enterprise Platform

**Last Updated**: 2025-10-28
**Status**: Testing infrastructure setup guide

---

## Overview

This document describes the comprehensive testing strategy for the Pivot platform, including:

1. **Generated Mocks** (using `mockgen`)
2. **Database Layer Testing** (with testcontainers)
3. **Handler Testing** (unit + integration)
4. **OpenAPI Spec Validation**
5. **Schema Alignment Verification**

---

## Testing Philosophy

### What to Test

✅ **YES - Test These**:
1. **Business Logic** in `internal/service/`
2. **HTTP Handlers** in `internal/handlers/`
3. **Middleware** (auth, RLS, validation)
4. **Type Mappers** (generated → domain models)
5. **SQL Query Correctness** (integration tests)
6. **OpenAPI Contract Compliance**

❌ **NO - Don't Test These**:
1. Generated code in `generated/` (trust the generators)
2. Database schema (trust PostgreSQL + migrations)
3. Third-party libraries

### Testing Pyramid

```
        ┌─────────────┐
        │  E2E Tests  │  10% - Full HTTP → DB flow
        │   (10 tests)│
        └─────────────┘
       ┌───────────────┐
       │ Integration   │  30% - Handler + Real DB
       │  (50 tests)   │
       └───────────────┘
     ┌───────────────────┐
     │  Unit Tests       │  60% - Logic + Mocked DB
     │   (100 tests)     │
     └───────────────────┘
```

**Target Coverage**: 80% overall (excluding generated code)

---

## 1. Generated Mocks Setup

### Using `mockgen` for Database Interface

**Why mockgen?**
- ✅ Auto-generated from interface definitions
- ✅ Type-safe mock methods
- ✅ Easy to regenerate when interfaces change
- ✅ No manual mock maintenance

### Installation

```bash
go install go.uber.org/mock/mockgen@latest
```

### Mock Generation Configuration

Add to `Makefile`:

```makefile
# Mock generation
generate-mocks:
	@echo "🎭 Generating mocks..."
	@mkdir -p generated/mocks
	mockgen -source=generated/db/querier.go \
		-destination=generated/mocks/db_mock.go \
		-package=mocks \
		-mock_names=Querier=MockQuerier
	@echo "✅ Mocks generated at generated/mocks/"

# Update generate target
generate: generate-erd generate-api generate-db generate-mocks
```

### Mock Files Generated

```
generated/
├── api/
│   └── server.gen.go       # OpenAPI types + router
├── db/
│   ├── querier.go          # Interface (source for mocks)
│   ├── models.go           # DB models
│   └── *.sql.go            # Query implementations
└── mocks/
    └── db_mock.go          # 🎭 GENERATED MOCK
```

### Using Mocks in Tests

```go
package handlers_test

import (
	"testing"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"go.uber.org/mock/gomock"
)

func TestGetEmployee_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock
	mockDB := mocks.NewMockQuerier(ctrl)

	// Set expectations
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), gomock.Any()).
		Return(db.Employee{
			ID:    uuid.New(),
			Email: "test@example.com",
		}, nil)

	// Test handler with mock
	handler := NewEmployeeHandler(mockDB)
	// ... rest of test
}
```

---

## 2. Database Layer Testing

### Should You Test the DB Layer?

**Answer: Partial Yes**

✅ **Test Query Correctness** (integration tests):
```go
// Test that SQL queries return expected data
func TestListEmployees_WithFilters(t *testing.T) {
	db := setupTestDB(t) // Real PostgreSQL via testcontainers
	queries := db.New(dbConn)

	employees, err := queries.ListEmployees(ctx, db.ListEmployeesParams{
		OrgID:  orgID,
		Status: "active",
		Limit:  10,
	})

	assert.NoError(t, err)
	assert.Len(t, employees, 5) // Expect 5 active employees
}
```

❌ **Don't Test CRUD Methods** (trust sqlc):
```go
// DON'T write tests like this:
func TestCreateEmployee_InsertsRow(t *testing.T) {
	// This just tests that sqlc works - waste of time
}
```

### Test Query Behavior, Not Generation

**Test These Scenarios**:
1. **Multi-tenant isolation** - Queries respect org_id filtering
2. **Soft delete** - Deleted records don't appear in listings
3. **Join correctness** - Related data is properly assembled
4. **Pagination** - LIMIT/OFFSET work as expected
5. **Unique constraints** - Duplicate emails are rejected

### Testcontainers Setup

```go
package testutil

import (
	"context"
	"testing"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func SetupTestDB(t *testing.T) *db.Queries {
	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("pivot_test"),
		postgres.WithUsername("pivot"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts("../schema.sql"),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		pgContainer.Terminate(ctx)
	})

	// Connect to test DB
	connStr, _ := pgContainer.ConnectionString(ctx)
	conn, _ := pgx.Connect(ctx, connStr)

	return db.New(conn)
}
```

---

## 3. Handler Testing Strategy

### Unit Tests (with Mocks)

**Test Structure**:
```go
func TestEmployeeHandler_GetEmployee(t *testing.T) {
	tests := []struct {
		name           string
		employeeID     string
		setupMock      func(*mocks.MockQuerier)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:       "success",
			employeeID: "valid-uuid",
			setupMock: func(m *mocks.MockQuerier) {
				m.EXPECT().
					GetEmployee(gomock.Any(), gomock.Any()).
					Return(db.Employee{...}, nil)
			},
			expectedStatus: 200,
		},
		{
			name:       "not found",
			employeeID: "missing-uuid",
			setupMock: func(m *mocks.MockQuerier) {
				m.EXPECT().
					GetEmployee(gomock.Any(), gomock.Any()).
					Return(db.Employee{}, pgx.ErrNoRows)
			},
			expectedStatus: 404,
		},
		{
			name:       "invalid uuid",
			employeeID: "not-a-uuid",
			setupMock:  func(m *mocks.MockQuerier) {}, // No DB call
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockQuerier(ctrl)
			tt.setupMock(mockDB)

			handler := handlers.NewEmployeeHandler(mockDB)

			req := httptest.NewRequest("GET", "/api/v1/employees/"+tt.employeeID, nil)
			rec := httptest.NewRecorder()

			handler.GetEmployee(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
```

### Integration Tests (with Real DB)

**Test Structure**:
```go
func TestEmployeeAPI_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup real PostgreSQL
	db := testutil.SetupTestDB(t)

	// Seed test data
	org := testutil.CreateTestOrg(t, db)
	employee := testutil.CreateTestEmployee(t, db, org.ID)

	// Create real HTTP server
	router := chi.NewRouter()
	api.HandlerFromMux(handlers.NewServer(db), router)
	server := httptest.NewServer(router)
	defer server.Close()

	// Test full HTTP flow
	resp, err := http.Get(server.URL + "/api/v1/employees/" + employee.ID.String())
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var result api.EmployeeResponse
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, employee.Email, result.Email)
}
```

---

## 4. OpenAPI Spec Validation

### Ensuring Spec Matches Implementation

**Strategy**: Use OpenAPI validation middleware

```go
package middleware

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func OpenAPIValidator(specPath string) func(http.Handler) http.Handler {
	loader := openapi3.NewLoader()
	doc, _ := loader.LoadFromFile(specPath)
	router, _ := gorillamux.NewRouter(doc)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Validate request against spec
			route, pathParams, _ := router.FindRoute(r)
			requestValidationInput := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
			}

			if err := openapi3filter.ValidateRequest(ctx, requestValidationInput); err != nil {
				http.Error(w, "Invalid request", 400)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

**Add to Integration Tests**:
```go
func TestAPI_OpenAPICompliance(t *testing.T) {
	// Load OpenAPI spec
	spec := loadOpenAPISpec(t, "openapi/spec.yaml")

	// Create router with validator
	router := chi.NewRouter()
	router.Use(middleware.OpenAPIValidator("openapi/spec.yaml"))
	api.HandlerFromMux(handlers.NewServer(db), router)

	// Test all endpoints match spec
	testCases := []struct {
		method string
		path   string
		body   interface{}
	}{
		{"POST", "/api/v1/auth/login", api.LoginRequest{...}},
		{"GET", "/api/v1/employees", nil},
		// ... all endpoints
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := createRequest(tc.method, tc.path, tc.body)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			// If this passes, request/response match OpenAPI spec
			assert.NotEqual(t, 400, rec.Code, "OpenAPI validation failed")
		})
	}
}
```

---

## 5. Schema Alignment Verification

### Problem: DB Schema ≠ API DTOs

**Example Mismatch**:
```go
// Database model (generated/db/models.go)
type Employee struct {
	ID           uuid.UUID
	OrgID        uuid.UUID
	Email        string
	PasswordHash string  // ⚠️ Should NEVER be in API response
	Status       string
	CreatedAt    time.Time
}

// API response (generated/api/server.gen.go)
type EmployeeResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	// ✅ No password_hash field
}
```

### How to Ensure Alignment

#### 1. Type Mapping Layer

**Create explicit mappers**:
```go
package mapper

import (
	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

// DBEmployeeToAPI converts DB model to API response
func DBEmployeeToAPI(emp db.Employee) api.EmployeeResponse {
	return api.EmployeeResponse{
		ID:        emp.ID.String(),
		Email:     emp.Email,
		Status:    emp.Status,
		CreatedAt: emp.CreatedAt,
		// ⚠️ Explicitly exclude PasswordHash
	}
}

// APIEmployeeRequestToDB converts API request to DB params
func APIEmployeeRequestToDB(req api.CreateEmployeeRequest, orgID uuid.UUID) db.CreateEmployeeParams {
	return db.CreateEmployeeParams{
		OrgID:        orgID,
		Email:        req.Email,
		PasswordHash: hashPassword(req.Password),
		RoleID:       parseUUID(req.RoleID),
		TeamID:       parseOptionalUUID(req.TeamID),
	}
}
```

#### 2. Mapper Tests

**Test that mappings are correct**:
```go
func TestMapper_DBEmployeeToAPI(t *testing.T) {
	dbEmp := db.Employee{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "secret", // Should NOT appear in API
		Status:       "active",
	}

	apiEmp := mapper.DBEmployeeToAPI(dbEmp)

	assert.Equal(t, dbEmp.Email, apiEmp.Email)
	assert.Equal(t, dbEmp.Status, apiEmp.Status)

	// Verify password is excluded
	json, _ := json.Marshal(apiEmp)
	assert.NotContains(t, string(json), "secret")
	assert.NotContains(t, string(json), "password_hash")
}
```

#### 3. Automated Drift Detection

**Static Analysis Tool** (future):
```bash
# Compare OpenAPI spec fields vs DB columns
make check-drift

# Output:
# ⚠️  Field mismatch in Employee:
#   - DB has 'password_hash' but API does not
#   - API has 'full_name' but DB does not
#   - Type mismatch: 'created_at' (DB: timestamptz, API: string)
```

---

## 6. Test Organization

### Directory Structure

```
pivot/
├── internal/
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── auth_test.go          # Unit tests (mocked DB)
│   │   ├── employees.go
│   │   └── employees_test.go
│   ├── service/
│   │   ├── employee_service.go
│   │   └── employee_service_test.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── auth_test.go
│   └── mapper/
│       ├── employee.go
│       └── employee_test.go      # Mapper tests
│
├── tests/
│   ├── integration/
│   │   ├── setup_test.go         # Test helpers
│   │   ├── auth_integration_test.go
│   │   └── employee_integration_test.go
│   ├── e2e/
│   │   └── full_flow_test.go
│   └── testutil/
│       ├── db.go                 # testcontainers setup
│       ├── fixtures.go           # Test data factories
│       └── assertions.go         # Custom matchers
│
└── generated/
    ├── mocks/
    │   └── db_mock.go            # 🎭 Generated mock
    ├── api/
    └── db/
```

### Test Naming Convention

```go
// Unit tests
func TestHandlerName_ScenarioName(t *testing.T)
// Example: TestGetEmployee_NotFound

// Integration tests
func TestAPI_FeatureName_Integration(t *testing.T)
// Example: TestAPI_EmployeeCRUD_Integration

// E2E tests
func TestE2E_UserStoryName(t *testing.T)
// Example: TestE2E_AdminManagesEmployees
```

---

## 7. Test Fixtures and Helpers

### Factory Functions

```go
package testutil

import "github.com/google/uuid"

// CreateTestOrg creates an organization for testing
func CreateTestOrg(t *testing.T, db *db.Queries) db.Organization {
	org, err := db.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name:  "Test Corp",
		Slug:  "test-corp-" + uuid.NewString(),
		Email: "admin@testcorp.com",
	})
	require.NoError(t, err)
	return org
}

// CreateTestEmployee creates an employee for testing
func CreateTestEmployee(t *testing.T, db *db.Queries, orgID uuid.UUID) db.Employee {
	emp, err := db.CreateEmployee(ctx, db.CreateEmployeeParams{
		OrgID:        orgID,
		Email:        "test-" + uuid.NewString() + "@example.com",
		PasswordHash: "hashed",
		Status:       "active",
	})
	require.NoError(t, err)
	return emp
}

// CreateAuthToken creates a valid JWT token
func CreateAuthToken(t *testing.T, employeeID uuid.UUID) string {
	// ... JWT generation
}
```

---

## 8. Running Tests

### Makefile Targets

```makefile
# Run all tests
test:
	@echo "🧪 Running all tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run unit tests only (fast)
test-unit:
	@echo "⚡ Running unit tests..."
	go test -v -short ./internal/...

# Run integration tests (slower)
test-integration:
	@echo "🔄 Running integration tests..."
	go test -v -run Integration ./tests/integration/...

# Run E2E tests (slowest)
test-e2e:
	@echo "🌐 Running E2E tests..."
	go test -v ./tests/e2e/...

# Coverage report
test-coverage:
	@echo "📊 Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage: coverage.html"

# Watch mode (requires gotestsum)
test-watch:
	gotestsum --watch -- -short ./...
```

### CI/CD Pipeline

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Install dependencies
        run: go mod download

      - name: Generate mocks
        run: make generate-mocks

      - name: Run unit tests
        run: make test-unit

      - name: Run integration tests
        run: make test-integration

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

---

## 9. Example Test Suite

### Complete Handler Test Example

```go
package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/internal/handlers"
)

func TestEmployeeHandler_CreateEmployee(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    api.CreateEmployeeRequest
		setupMock      func(*mocks.MockQuerier)
		expectedStatus int
		validateBody   func(t *testing.T, body []byte)
	}{
		{
			name: "success - creates employee with valid data",
			requestBody: api.CreateEmployeeRequest{
				Email:  "newuser@example.com",
				RoleID: uuid.NewString(),
			},
			setupMock: func(m *mocks.MockQuerier) {
				m.EXPECT().
					CreateEmployee(gomock.Any(), gomock.Any()).
					Return(db.Employee{
						ID:     uuid.New(),
						Email:  "newuser@example.com",
						Status: "active",
					}, nil)
			},
			expectedStatus: 201,
			validateBody: func(t *testing.T, body []byte) {
				var resp api.EmployeeResponse
				json.Unmarshal(body, &resp)
				assert.Equal(t, "newuser@example.com", resp.Email)
			},
		},
		{
			name: "error - duplicate email",
			requestBody: api.CreateEmployeeRequest{
				Email: "existing@example.com",
			},
			setupMock: func(m *mocks.MockQuerier) {
				m.EXPECT().
					CreateEmployee(gomock.Any(), gomock.Any()).
					Return(db.Employee{}, &pgconn.PgError{
						Code: "23505", // Unique violation
					})
			},
			expectedStatus: 409,
			validateBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "already exists")
			},
		},
		{
			name: "error - invalid email format",
			requestBody: api.CreateEmployeeRequest{
				Email: "not-an-email",
			},
			setupMock:      func(m *mocks.MockQuerier) {}, // No DB call
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockQuerier(ctrl)
			tt.setupMock(mockDB)

			handler := handlers.NewEmployeeHandler(mockDB)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/v1/employees", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.CreateEmployee(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.validateBody != nil {
				tt.validateBody(t, rec.Body.Bytes())
			}
		})
	}
}
```

---

## 10. Coverage Goals

### Target Coverage by Component

| Component            | Target | Priority |
|---------------------|--------|----------|
| Handlers            | 90%    | P0       |
| Services            | 85%    | P0       |
| Middleware          | 80%    | P1       |
| Mappers             | 95%    | P0       |
| Validation          | 90%    | P0       |
| Integration Tests   | 70%    | P1       |

### Coverage Exclusions

```go
// coverage:ignore - exclude from coverage
func (h *Handler) debugEndpoint(w http.ResponseWriter, r *http.Request) {
	// Debug code not covered by tests
}
```

---

## 11. Testing Checklist

### Before Committing

- [ ] All unit tests pass (`make test-unit`)
- [ ] Integration tests pass (`make test-integration`)
- [ ] Coverage >80% for new code
- [ ] No `.only` or `.skip` in tests
- [ ] Mocks regenerated (`make generate-mocks`)
- [ ] OpenAPI validation passes

### Before Deploying

- [ ] E2E tests pass (`make test-e2e`)
- [ ] Load tests completed (future)
- [ ] Security tests passed (future)
- [ ] Performance benchmarks acceptable

---

## Summary: Answering Your Questions

### 1. How to Test Database Models?

**Answer**: Test query **behavior**, not CRUD operations.

```go
✅ DO: Test multi-tenant isolation
✅ DO: Test join correctness
✅ DO: Test edge cases (soft delete, pagination)
❌ DON'T: Test basic CRUD (trust sqlc)
```

### 2. How to Test OpenAPI Spec with Generated Handlers?

**Answer**: Use OpenAPI validation middleware + integration tests.

```go
✅ DO: Validate requests/responses match spec
✅ DO: Test all endpoints in integration tests
✅ DO: Use OpenAPI validator middleware
```

### 3. Should I Test the DB Layer?

**Answer**: Partial yes — test query correctness with testcontainers.

```go
✅ DO: Integration tests with real PostgreSQL
✅ DO: Test SQL query logic (joins, filters)
❌ DON'T: Test sqlc-generated CRUD methods
```

### 4. How Do I Understand Models and Spec Match?

**Answer**: Create explicit mapper layer + mapper tests.

```go
✅ DO: Write DBEmployeeToAPI() mappers
✅ DO: Test mappers verify field mappings
✅ DO: Test sensitive fields excluded (passwords)
❌ DON'T: Expose DB models directly in API
```

---

## Next Steps

1. **Run**: `make generate-mocks` to create mock files
2. **Create**: `tests/testutil/db.go` with testcontainers setup
3. **Write**: First handler test in `internal/handlers/employees_test.go`
4. **Add**: OpenAPI validator middleware
5. **Create**: Mapper tests in `internal/mapper/employee_test.go`
6. **Run**: `make test` and aim for 80% coverage

---

**Key Insight**: You have **two sources of truth** (DB schema + OpenAPI spec). Use:
- **Mappers** to bridge the gap
- **Integration tests** to verify alignment
- **Generated mocks** to avoid testing infrastructure code

This approach ensures schema and spec stay aligned while maintaining fast, reliable tests.
