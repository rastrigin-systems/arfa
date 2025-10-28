# Next Steps — Phase 2 Implementation Kickoff

**Current Status**: Phase 1 Complete (Database + Code Generation + Documentation)
**Next Phase**: Phase 2 - Core API Implementation
**Estimated Time**: 1-2 weeks
**Start Date**: Now

---

## 🎯 Immediate Actions (Today)

### 1. Setup Test Infrastructure (2 hours)

```bash
# Navigate to project
cd pivot

# Install mock generation tool
make install-tools

# Create test helpers directory
mkdir -p tests/testutil tests/integration

# Generate mocks
make generate-mocks
```

**Files to create**:

#### `tests/testutil/db.go`
```go
package testutil

import (
	"context"
	"testing"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func SetupTestDB(t *testing.T) *pgx.Conn {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("pivot_test"),
		postgres.WithUsername("pivot"),
		postgres.WithPassword("test"),
		postgres.WithInitScripts("../../schema.sql"),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		pgContainer.Terminate(ctx)
	})

	connStr, _ := pgContainer.ConnectionString(ctx)
	conn, _ := pgx.Connect(ctx, connStr)

	return conn
}
```

#### `tests/testutil/fixtures.go`
```go
package testutil

import (
	"testing"
	"github.com/google/uuid"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

func CreateTestOrg(t *testing.T, queries *db.Queries) db.Organization {
	org, err := queries.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name:  "Test Corp",
		Slug:  "test-" + uuid.NewString(),
		Email: "admin@test.com",
	})
	require.NoError(t, err)
	return org
}

func CreateTestEmployee(t *testing.T, queries *db.Queries, orgID uuid.UUID, email string) db.Employee {
	// TODO: Implement
}
```

### 2. Verify Setup Works (30 min)

```bash
# Run example test
go test ./internal/handlers/example_test.go -v

# Should see:
# ✅ Mock generation working
# ✅ Test infrastructure ready
```

### 3. Update Dependencies (10 min)

```bash
# Add testcontainers
go get github.com/testcontainers/testcontainers-go/modules/postgres

# Add JWT library
go get github.com/golang-jwt/jwt/v5

# Add bcrypt
go get golang.org/x/crypto/bcrypt

# Download all dependencies
go mod download
```

---

## 📅 Week 1 Schedule

### Monday: Test Setup + JWT Helpers

**Morning (4 hours)**:
- [ ] Create `tests/testutil/db.go`
- [ ] Create `tests/testutil/fixtures.go`
- [ ] Run `make generate-mocks`
- [ ] Verify test setup with dummy test

**Afternoon (4 hours)**:
- [ ] Create `internal/auth/jwt.go`
- [ ] TDD: `TestGenerateJWT`
- [ ] TDD: `TestVerifyJWT`
- [ ] TDD: `TestHashPassword`
- [ ] TDD: `TestVerifyPassword`

**End of Day**:
- [ ] All JWT helper tests passing
- [ ] Coverage >90% for auth package

### Tuesday: Login Handler

**Morning (4 hours)**:
- [ ] Create `internal/handlers/auth.go`
- [ ] Write test: `TestLogin_Success`
- [ ] Implement `Login()` happy path
- [ ] Write test: `TestLogin_InvalidPassword`
- [ ] Write test: `TestLogin_UserNotFound`
- [ ] Implement error handling

**Afternoon (4 hours)**:
- [ ] Write test: `TestLogin_InvalidJSON`
- [ ] Write test: `TestLogin_MissingFields`
- [ ] Integration test: `tests/integration/auth_integration_test.go`
- [ ] Verify with real PostgreSQL

**End of Day**:
- [ ] Login endpoint fully working
- [ ] 8+ unit tests passing
- [ ] 2+ integration tests passing

### Wednesday: JWT Middleware + Logout

**Morning (4 hours)**:
- [ ] Create `internal/middleware/auth.go`
- [ ] TDD: `TestAuthMiddleware_ValidToken`
- [ ] TDD: `TestAuthMiddleware_ExpiredToken`
- [ ] TDD: `TestAuthMiddleware_MissingToken`
- [ ] Implement `AuthMiddleware()`

**Afternoon (4 hours)**:
- [ ] Create `Logout()` handler
- [ ] Write tests for logout
- [ ] Create `GetMe()` handler
- [ ] Integration tests
- [ ] Test full auth flow end-to-end

**End of Day**:
- [ ] Auth system complete
- [ ] Can login → get token → access protected endpoints → logout

### Thursday: Employee Read Operations

**Morning (4 hours)**:
- [ ] Create `internal/handlers/employees.go`
- [ ] Implement `ListEmployees()`
- [ ] Write unit tests (3+ scenarios)
- [ ] Implement query parameter parsing

**Afternoon (4 hours)**:
- [ ] Implement `GetEmployee()`
- [ ] Write unit tests (success, not found, invalid ID)
- [ ] Integration tests with real DB
- [ ] Test with auth middleware

**End of Day**:
- [ ] Can list and get employees
- [ ] Auth protected
- [ ] Tests passing

### Friday: Employee Write Operations

**Morning (4 hours)**:
- [ ] Implement `CreateEmployee()`
- [ ] Write unit tests (success, duplicate email, validation)
- [ ] Implement `UpdateEmployee()`
- [ ] Write unit tests

**Afternoon (4 hours)**:
- [ ] Implement `DeleteEmployee()` (soft delete)
- [ ] Write unit tests
- [ ] Integration tests for all CRUD
- [ ] Run full test suite
- [ ] Check coverage (`make test-coverage`)

**End of Day**:
- [ ] Full Employee CRUD working
- [ ] Auth protected
- [ ] >80% coverage
- [ ] All tests passing

---

## 🎯 Week 1 Goals (Must Have)

By end of Week 1, you should have:

### Functional Features
- [x] ✅ Authentication system (login, logout, JWT middleware)
- [x] ✅ Employee CRUD endpoints (list, get, create, update, delete)
- [x] ✅ Multi-tenant isolation (org_id filtering)
- [x] ✅ Soft delete for employees

### Testing
- [x] ✅ 20+ unit tests passing
- [x] ✅ 10+ integration tests passing
- [x] ✅ >80% code coverage
- [x] ✅ All tests run in CI/CD

### Infrastructure
- [x] ✅ Test helpers (testcontainers setup)
- [x] ✅ Mock generation working
- [x] ✅ Test fixtures for common data
- [x] ✅ Integration test suite

### Documentation
- [x] ✅ API endpoints documented
- [x] ✅ Test strategy documented
- [x] ✅ Development approach documented

---

## 📝 Week 2 Preview

### Organization Management API (2-3 days)
- [ ] Get organization details
- [ ] List teams
- [ ] List roles
- [ ] Team CRUD (create, update, delete)

### Mapper Layer (1 day)
- [ ] DB ↔ API type converters
- [ ] Test security (password exclusion)
- [ ] Test nullable fields

### OpenAPI Validation (1 day)
- [ ] Add validation middleware
- [ ] Verify all endpoints match spec
- [ ] Fix any drift issues

### Polish & Deploy (1 day)
- [ ] Error handling consistency
- [ ] Logging middleware
- [ ] Health check endpoint
- [ ] Docker build
- [ ] Deploy to staging

---

## 🚀 Getting Started Right Now

### Option 1: Start with Test Setup (Recommended)

```bash
# 1. Create test helpers
mkdir -p tests/testutil
vim tests/testutil/db.go          # Copy from DEVELOPMENT_APPROACH.md

# 2. Generate mocks
make generate-mocks

# 3. Verify it works
go test ./internal/handlers/example_test.go -v
```

### Option 2: Start with JWT Helpers (If impatient)

```bash
# 1. Create auth package
mkdir -p internal/auth
vim internal/auth/jwt.go

# 2. Write first test
vim internal/auth/jwt_test.go

# 3. Run TDD cycle
go test ./internal/auth/... -v
```

### Option 3: Start with Handler Skeleton (Quick win)

```bash
# 1. Create handlers
mkdir -p internal/handlers
vim internal/handlers/auth.go

# 2. Wire up to router
vim cmd/server/main.go

# 3. Start server
go run cmd/server/main.go
```

---

## 📚 Key Documentation

**Before You Start, Read**:
1. [DEVELOPMENT_APPROACH.md](./docs/DEVELOPMENT_APPROACH.md) - TDD strategy
2. [TESTING_QUICKSTART.md](./docs/TESTING_QUICKSTART.md) - Test setup
3. [docs/ERD.md](./docs/ERD.md) - Database schema
4. [openapi/spec.yaml](./openapi/spec.yaml) - API contract

**While Coding, Reference**:
- [TESTING_STRATEGY.md](./docs/TESTING_STRATEGY.md) - Test patterns
- [generated/db/querier.go](./generated/db/querier.go) - Available DB queries
- [generated/api/server.gen.go](./generated/api/server.gen.go) - API types

---

## 🎯 Success Metrics

### Daily Check
```bash
make test               # Should pass
make test-coverage      # Should be >80%
go build ./cmd/server   # Should compile
```

### Weekly Check
- [ ] All planned features working
- [ ] All tests passing
- [ ] Coverage >80%
- [ ] API matches OpenAPI spec
- [ ] Can run locally: `make dev`
- [ ] Can run in Docker: `docker-compose up`

---

## 💡 Pro Tips

### 1. Run Tests Frequently
```bash
# Fast feedback loop
make test-unit

# After each feature
make test-integration

# Before commit
make test
```

### 2. Use Watch Mode
```bash
# Auto-run tests on file changes
gotestsum --watch -- -short ./...
```

### 3. Check Coverage Often
```bash
make test-coverage
# Open coverage.html to see gaps
```

### 4. Commit Often
```bash
# After each working feature
git add .
git commit -m "feat: implement login endpoint with tests"
```

### 5. Keep Tickets Small
- [ ] ✅ "Implement login" - Good (2-3 hours)
- [ ] ❌ "Implement all auth" - Too big (2 days)

---

## 🆘 If You Get Stuck

### Problem: Tests Won't Run
```bash
# Check dependencies
go mod download

# Regenerate mocks
make generate-mocks

# Check test file syntax
go test -c ./internal/handlers/...
```

### Problem: Mock Expectations Failing
```go
// Add .AnyTimes() if call count doesn't matter
mockDB.EXPECT().GetEmployee(...).Return(...).AnyTimes()

// Use gomock.Any() for arguments you don't care about
mockDB.EXPECT().GetEmployee(gomock.Any(), gomock.Any())
```

### Problem: Integration Tests Failing
```bash
# Ensure Docker is running
docker ps

# Reset test database
make db-reset

# Check PostgreSQL logs
docker-compose logs postgres
```

### Problem: OpenAPI Types Don't Match
```bash
# Regenerate API types
make generate-api

# Check OpenAPI spec syntax
cat openapi/spec.yaml | yq .
```

---

## ✅ Checklist Before Starting

### Environment Setup
- [ ] Go 1.24+ installed
- [ ] Docker Desktop running
- [ ] PostgreSQL running (`make db-up`)
- [ ] Tools installed (`make install-tools`)
- [ ] Mocks generated (`make generate-mocks`)

### Documentation Read
- [ ] Read DEVELOPMENT_APPROACH.md
- [ ] Read TESTING_QUICKSTART.md
- [ ] Reviewed OpenAPI spec
- [ ] Reviewed database schema

### Ready to Code
- [ ] Test infrastructure created
- [ ] Dependencies installed
- [ ] First test written
- [ ] First test passing

---

## 🎉 You're Ready!

**Recommended first task**: Create test helpers

```bash
cd pivot
mkdir -p tests/testutil
vim tests/testutil/db.go
# Copy testcontainers setup from DEVELOPMENT_APPROACH.md
```

**Estimated time to first working endpoint**: 4-6 hours

**Estimated time to complete Week 1 goals**: 32-40 hours (1 week full-time)

Good luck! 🚀

---

**See [DEVELOPMENT_APPROACH.md](./docs/DEVELOPMENT_APPROACH.md) for detailed daily schedule.**
