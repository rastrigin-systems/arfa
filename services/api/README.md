# Arfa API Server

Multi-tenant REST API for the Arfa platform.

## Overview

The API server provides:
- **Authentication & Sessions:** JWT-based authentication with session management
- **Multi-tenant Management:** Organizations, teams, and employee management
- **Agent Configuration:** AI agent (Claude Code, Cursor, etc.) configuration management
- **MCP Configuration:** Model Context Protocol server configuration
- **Approval Workflows:** Request/approval system for agent and MCP access
- **Activity Logging:** Comprehensive activity tracking and usage analytics
- **WebSocket Support:** Real-time updates for live configuration sync

## Quick Start

### Prerequisites
- Go 1.24+
- PostgreSQL 15+
- Docker (for integration tests)

### Setup

```bash
# From project root
make db-up          # Start PostgreSQL
make generate       # Generate code from schema and OpenAPI spec

# Build and run
cd services/api
make build          # Build server binary
../../bin/arfa-server
```

### Development

```bash
# Run tests
cd services/api
make test           # All tests
make test-unit      # Unit tests only
make test-integration  # Integration tests only

# Test coverage
make coverage

# Docker testing
make docker-build   # Build Docker image
make docker-test    # Verify Docker image structure
make docker-run     # Run container locally
```

## Project Structure

```
services/api/
├── cmd/
│   └── server/          # Main entrypoint
│       └── main.go
├── internal/            # Private API code
│   ├── handlers/        # HTTP request handlers (39 endpoints)
│   ├── middleware/      # HTTP middleware (auth, logging, etc.)
│   ├── auth/            # JWT authentication logic
│   ├── database/        # Database layer
│   ├── service/         # Business logic services
│   ├── websocket/       # WebSocket hub and handlers
│   └── pkg/             # Internal utilities
├── tests/
│   ├── integration/     # Integration tests (testcontainers)
│   └── testutil/        # Test fixtures and helpers
├── build/               # Deployment configurations
│   ├── Dockerfile.gcp   # GCP Cloud Run Dockerfile
│   └── cloudbuild.yaml  # GCP Cloud Build config
├── scripts/             # Service-specific scripts (currently empty)
├── docs/                # Service-specific docs (currently empty)
├── Makefile             # Service build commands
├── go.mod               # Go module dependencies
└── README.md            # This file
```

## API Documentation

**OpenAPI Specification:** `../../platform/api-spec/spec.yaml`

**Interactive Documentation:**
- Local: http://localhost:8080/api/docs
- Swagger UI with live testing capabilities

**Key Endpoints:**

**Authentication:**
- `POST /api/v1/auth/register` - Employee registration
- `POST /api/v1/auth/login` - Login and JWT generation
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `POST /api/v1/auth/logout` - Logout and session cleanup

**Organizations & Teams:**
- `GET/POST /api/v1/organizations` - Organization management
- `GET/POST /api/v1/teams` - Team management
- `GET/POST /api/v1/employees` - Employee management

**Agent Configuration:**
- `GET /api/v1/agents` - List available agents
- `GET/POST /api/v1/employee-agent-configs` - Employee agent configs
- `GET/POST /api/v1/team-agent-configs` - Team agent configs
- `GET/POST /api/v1/org-agent-configs` - Organization agent configs

**Configuration Sync:**
- `GET /api/v1/sync/employee/{id}` - Sync employee configs
- `WS /api/v1/ws` - WebSocket for real-time updates

See OpenAPI spec for complete endpoint documentation.

## Architecture

### Code Generation

The API uses automatic code generation:

**Source Files (Edit These):**
- `../../platform/api-spec/spec.yaml` - OpenAPI specification
- `../../platform/database/schema/schema.sql` - PostgreSQL schema
- `../../platform/database/sqlc/queries/*.sql` - SQL queries

**Generated Code (Never Edit):**
- `../../generated/api/` - API types and Chi server
- `../../generated/db/` - Type-safe database code (sqlc)
- `../../generated/mocks/` - Test mocks

After changing source files, regenerate code:
```bash
cd ../..          # Go to root
make generate     # Regenerate everything
```

### Multi-Tenancy

**All database queries are organization-scoped** using Row-Level Security (RLS) policies.

**Critical Rule:** Every query must include `org_id` filtering to prevent data leakage.

```go
// ✅ GOOD - Scoped to organization
employees, err := db.ListEmployees(ctx, orgID, status)

// ❌ BAD - Exposes all organizations!
employees, err := db.ListAllEmployees(ctx)
```

### Testing Strategy

**Test-Driven Development (TDD) is mandatory:**
1. Write failing tests first
2. Implement minimal code to pass tests
3. Refactor with tests passing

**Test Types:**
- **Unit Tests:** In `internal/*/` alongside source code, test individual functions
- **Integration Tests:** In `tests/integration/`, test full API flows with real database (testcontainers)

**Coverage Target:** 85% overall

**Running Tests:**
```bash
make test-unit         # Fast unit tests
make test-integration  # Full integration tests with Docker
make test              # All tests
make coverage          # View coverage report
```

See [../../docs/TESTING.md](../../docs/TESTING.md) for complete testing guide.

## Deployment

### GCP Cloud Run

```bash
# Deploy to GCP
cd services/api
make deploy

# This will:
# 1. Build Docker image with Cloud Build
# 2. Push to Artifact Registry
# 3. Deploy to Cloud Run
```

**Environment Variables:**
- `DATABASE_URL` - PostgreSQL connection string (required)
- `PORT` - Server port (default: 8080)
- `JWT_SECRET` - JWT signing secret (required in production)

### Docker

```bash
# Build image
cd services/api
make docker-build

# Test locally
make docker-test

# Run locally
make docker-run

# Or with custom config
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://user:pass@host:5432/dbname" \
  arfa-api:latest
```

## Development Workflow

### Before Starting Work

1. Check GitHub issues for assigned tasks
2. Create feature branch: `git checkout -b feature/123-description`
3. Start database: `make db-up` (from root)
4. Ensure code is generated: `make generate` (from root)

### During Development

1. **Write tests first** (TDD)
2. Implement minimal code to pass tests
3. Refactor with tests passing
4. Run tests: `make test`
5. Verify coverage: `make coverage`

### After Implementation

1. Verify all tests pass: `make test`
2. Commit changes with descriptive message
3. Push branch: `git push -u origin feature/123-description`
4. Create Pull Request with issue number in title
5. Wait for CI checks to pass

See [../../docs/DEV_WORKFLOW.md](../../docs/DEV_WORKFLOW.md) for complete workflow guide.

## Contributing

### Code Standards

- **Multi-Tenancy:** All queries must be org-scoped
- **Error Handling:** Use proper error wrapping (`fmt.Errorf("...: %w", err)`)
- **Testing:** TDD with 85%+ coverage
- **Code Generation:** Never edit generated files
- **Documentation:** Update OpenAPI spec for endpoint changes

### Common Tasks

**Add New Endpoint:**
1. Update `../../platform/api-spec/spec.yaml`
2. Regenerate API code: `make generate-api` (from root)
3. Write handler tests in `internal/handlers/*_test.go`
4. Implement handler in `internal/handlers/*.go`
5. Wire route in `cmd/server/main.go`
6. Run tests: `make test`

**Add Database Query:**
1. Write query in `../../platform/database/sqlc/queries/*.sql`
2. Regenerate DB code: `make generate-db` (from root)
3. Use generated code in handlers

**Schema Change:**
1. Update `../../platform/database/schema/schema.sql`
2. Reset database: `make db-reset` (from root)
3. Regenerate all: `make generate` (from root)

## Troubleshooting

### Tests Failing

```bash
# Ensure database is running
make db-up

# Regenerate code
cd ../.. && make generate

# Rebuild binary
cd services/api && make build

# Clear test cache
go clean -testcache
```

### Docker Build Fails

```bash
# Verify files are copied correctly
make docker-test

# Check Dockerfile paths
cat build/Dockerfile.gcp

# Rebuild from scratch
docker system prune -a
make docker-build
```

### Import Errors

```bash
# Sync workspace
cd ../.. && go work sync

# Tidy modules
cd services/api && go mod tidy

# Regenerate code
cd ../.. && make generate
```

## Resources

**Documentation:**
- [Complete System Docs](../../CLAUDE.md)
- [Database ERD](../../docs/ERD.md)
- [Testing Guide](../../docs/TESTING.md)
- [Development Workflow](../../docs/DEV_WORKFLOW.md)
- [Quick Reference](../../docs/QUICK_REFERENCE.md)

**Tools:**
- OpenAPI Spec: http://localhost:8080/api/docs
- Adminer (DB UI): http://localhost:8081
- Database: `postgres://arfa:arfa_dev_password@localhost:5432/arfa`

## License

MIT License - Arfa Platform
