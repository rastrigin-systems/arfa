# Arfa API Server

Multi-tenant REST API for the Arfa platform.

## Quick Start

```bash
# From project root
make db-up              # Start PostgreSQL
make generate           # Generate code from schema and OpenAPI spec

# Build and run
cd services/api
make build              # Build server binary
../../bin/arfa-server   # Run server
```

## Project Structure

```
services/api/
├── cmd/server/         Main entrypoint
├── internal/
│   ├── handlers/       HTTP request handlers
│   ├── middleware/     HTTP middleware (auth, logging)
│   ├── auth/           JWT authentication
│   ├── database/       Database layer
│   ├── service/        Business logic
│   └── websocket/      WebSocket handlers
├── tests/
│   ├── integration/    Integration tests (testcontainers)
│   └── testutil/       Test fixtures
├── build/              Dockerfile, cloudbuild.yaml
└── Makefile            Service build commands
```

## Commands

```bash
make test               # All tests
make test-unit          # Unit tests only
make test-integration   # Integration tests (requires Docker)
make coverage           # Coverage report
make docker-build       # Build Docker image
make docker-run         # Run container locally
```

## Code Generation

Source files (edit these):
- `../../platform/api-spec/spec.yaml` - OpenAPI specification
- `../../platform/database/schema.sql` - PostgreSQL schema
- `../../platform/database/sqlc/queries/*.sql` - SQL queries

Generated files (never edit):
- `../../generated/api/` - API types and Chi server
- `../../generated/db/` - Type-safe database code (sqlc)
- `../../generated/mocks/` - Test mocks

After changing source files:
```bash
cd ../.. && make generate
```

## Multi-Tenancy

All database queries MUST include `org_id` filtering to prevent data leakage.

```go
// GOOD - Scoped to organization
employees, err := db.ListEmployees(ctx, orgID, status)

// BAD - Exposes all organizations!
employees, err := db.ListAllEmployees(ctx)
```

## API Documentation

- OpenAPI Spec: `../../platform/api-spec/spec.yaml`
- Interactive docs: http://localhost:8080/api/docs

Key endpoints:
- `POST /api/v1/auth/login` - Login and JWT generation
- `GET /api/v1/auth/me` - Get current user
- `GET /api/v1/employees` - List employees
- `WS /api/v1/ws` - WebSocket for real-time updates

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string (required)
- `PORT` - Server port (default: 8080)
- `JWT_SECRET` - JWT signing secret (required in production)

## Documentation

- [Architecture](../../docs/architecture/overview.md)
- [Database schema](../../docs/database/README.md)
- [Testing](../../docs/development/testing.md)
- [Contributing](../../docs/development/contributing.md)
