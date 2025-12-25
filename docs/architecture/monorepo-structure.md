# Monorepo Structure Documentation

**Arfa Platform - Architecture Guide**

---

## Table of Contents

1. [Overview](#overview)
2. [Directory Organization](#directory-organization)
3. [Service Boundaries](#service-boundaries)
4. [Shared Code Strategy](#shared-code-strategy)
5. [Code Generation Pipeline](#code-generation-pipeline)
6. [When to Add Code Where](#when-to-add-code-where)
7. [Benefits of This Structure](#benefits-of-this-structure)
8. [Migration History](#migration-history)

---

## Overview

Arfa uses a **Go workspace monorepo** architecture with **self-contained services** and **shared platform resources**.

### Key Principles

- **Service Independence**: Each service is a complete Go module with its own dependencies
- **Clear Boundaries**: Services never import from each other's `internal/` packages
- **Shared Platform**: Database schema, API spec, and Docker images are shared resources
- **Generated Code**: Type-safe code generated from source of truth files
- **Minimal Coupling**: Only `pkg/types` is shared between services

### Architecture Diagram

```
arfa/                       # Monorepo Root
│
├── services/                          # Self-Contained Services
│   ├── api/                           # API Server (Go)
│   │   ├── cmd/server/                # Main entrypoint
│   │   ├── internal/                  # Private implementation
│   │   ├── tests/                     # Service tests
│   │   ├── build/                     # Deployment configs
│   │   └── go.mod                     # Independent module
│   │
│   ├── cli/                           # CLI Client (Go)
│   │   ├── cmd/arfa/                  # Main entrypoint
│   │   ├── internal/                  # Private implementation
│   │   ├── tests/                     # Service tests
│   │   └── go.mod                     # Independent module
│   │
│   └── web/                           # Web UI (Next.js)
│       ├── app/                       # Next.js App Router
│       ├── components/                # React components
│       ├── lib/                       # Utilities and API client
│       ├── tests/                     # Service tests
│       └── package.json               # Independent dependencies
│
├── platform/                          # Shared Platform Resources
│   ├── api-spec/                      # OpenAPI 3.0.3 spec
│   │   └── spec.yaml                  # API contract (source of truth)
│   │
│   ├── database/                      # PostgreSQL resources
│   │   ├── schema.sql                 # Database schema (source of truth)
│   │   ├── sqlc/                      # SQL queries
│   │   │   ├── queries/               # sqlc query files
│   │   │   └── sqlc.yaml              # sqlc configuration
│   │   └── migrations/                # Database migrations (future)
│   │
│   └── docker-images/                 # Docker image definitions
│       ├── agents/                    # AI agent images
│       └── mcp/                       # MCP server images
│
├── pkg/types/                         # Shared Go Types
│   └── types.go                       # Common data structures
│
├── generated/                         # Generated Code (NOT committed)
│   ├── api/                           # From OpenAPI spec
│   ├── db/                            # From database schema
│   └── mocks/                         # From interfaces
│
├── docs/                              # Documentation
│   ├── architecture/                  # Architecture docs (this file)
│   ├── schema-reference.md                         # Database schema (auto-generated)
│   ├── README.md                      # Table index (auto-generated)
│   └── public.*.md                    # Per-table docs (auto-generated)
│
├── scripts/                           # Cross-cutting utility scripts
│   ├── erd/                           # ERD generation
│   └── release/                       # Release automation
│
├── go.work                            # Go workspace definition
├── Makefile                           # Root-level commands
└── docker-compose.yml                 # Local development environment
```

---

## Directory Organization

### Services (`services/`)

Self-contained applications with:
- Own `go.mod` (Go) or `package.json` (Node.js)
- Own `internal/` package (private implementation)
- Own `tests/` directory
- Own `build/` configs (Dockerfiles, cloud build)
- Own `README.md` documentation

**Key Rule**: Services NEVER import from other services' `internal/` packages.

### Platform (`platform/`)

Shared resources that define the system:
- **api-spec/**: OpenAPI 3.0.3 specification (API contract)
- **database/**: PostgreSQL schema, sqlc queries, migrations
- **docker-images/**: Reusable Docker images for agents and MCP servers

**Key Rule**: Platform resources are sources of truth - they generate code, not consume it.

### Shared Code (`pkg/`)

Minimal shared Go packages:
- **pkg/types/**: Common data structures shared across services
- **Note**: Only simple, stable types - no business logic

**Key Rule**: Keep `pkg/` minimal - services should be self-contained.

### Generated Code (`generated/`)

Auto-generated, never manually edited:
- **generated/api/**: Go types and server from OpenAPI spec
- **generated/db/**: Type-safe database code from sqlc
- **generated/mocks/**: Test mocks from interfaces

**Key Rule**: NEVER edit `generated/` - regenerate from source of truth.

### Documentation (`docs/`)

Comprehensive documentation:
- **architecture/**: System design docs (this file)
- **schema-reference.md**: User-friendly database schema (auto-generated)
- **README.md**: Technical database reference (auto-generated)
- **public.*.md**: Per-table documentation (auto-generated)
- **TESTING.md, DEVELOPMENT.md, etc.**: Development guides

**Key Rule**: Keep docs up-to-date with code changes.

---

## Service Boundaries

### API Server (`services/api/`)

**Responsibility**: REST API for the platform

**Dependencies**:
- ✅ `generated/api` - API types and server
- ✅ `generated/db` - Database layer
- ✅ `pkg/types` - Shared types
- ❌ NO dependency on `services/cli` or `services/web`

**Key Files**:
- `cmd/server/main.go` - Entrypoint
- `internal/handlers/` - HTTP request handlers
- `internal/middleware/` - HTTP middleware
- `internal/auth/` - JWT authentication
- `internal/database/` - Database layer
- `internal/service/` - Business logic

**Tests**: `tests/integration/` + `internal/*_test.go`

### CLI Client (`services/cli/`)

**Responsibility**: Command-line tool for syncing configurations

**Dependencies**:
- ✅ `pkg/types` - Shared types (minimal)
- ❌ NO dependency on `generated/` (keeps binary small)
- ❌ NO dependency on database packages

**Key Files**:
- `cmd/arfa/main.go` - Entrypoint
- `internal/commands/` - CLI commands
- `internal/auth.go` - Authentication
- `internal/sync.go` - Configuration sync
- `internal/docker.go` - Docker integration

**Tests**: `tests/integration/` + `tests/e2e/` + `internal/*_test.go`

### Web UI (`services/web/`)

**Responsibility**: Next.js admin panel

**Dependencies**:
- ✅ `lib/api/schema.ts` - API types from OpenAPI spec (auto-generated)
- ❌ NO dependency on Go code

**Key Files**:
- `app/` - Next.js App Router pages
- `components/` - React components
- `lib/api/` - API client and types
- `lib/auth/` - Authentication utilities

**Tests**: `tests/unit/` + `tests/e2e/`

---

## Shared Code Strategy

### What Goes in `pkg/types`

**Only simple, stable types**:
```go
// ✅ GOOD - Simple data structures
type AgentType string
type PolicyType string
type Status string

// ❌ BAD - Business logic
func (a *Agent) Validate() error { ... }

// ❌ BAD - Database code
func GetAgent(ctx context.Context, id uuid.UUID) (*Agent, error) { ... }
```

**Rule of Thumb**: If it changes frequently or has dependencies, it belongs in a service's `internal/` package.

### What Goes in `platform/`

**Sources of truth**:
- **platform/api-spec/spec.yaml**: API contract
- **platform/database/schema.sql**: Database schema
- **platform/database/sqlc/queries/**: SQL queries

**NOT business logic** - only data definitions.

### What Goes in `services/*/internal/`

**Everything else**:
- HTTP handlers
- Business logic
- Service implementations
- Database access code
- Authentication logic
- Tests

**Key Rule**: If only one service needs it, it goes in that service's `internal/`.

---

## Code Generation Pipeline

### Database Schema → Go Code

```
platform/database/schema.sql
    ↓
PostgreSQL (via make db-reset)
    ↓
    ├─→ tbls → docs/README.md, docs/public.*.md, schema.json
    ├─→ Python script → docs/database/schema-reference.md (user-friendly)
    └─→ sqlc → generated/db/*.go (type-safe queries)
```

**Commands**:
```bash
# Update schema
vim platform/database/schema.sql
make db-reset                    # Apply to database
make generate-erd                # Update docs
make generate-db                 # Update Go code
```

### API Spec → Go/TypeScript Code

```
platform/api-spec/spec.yaml
    ↓
    ├─→ oapi-codegen → generated/api/server.gen.go (Go types + server)
    └─→ openapi-typescript → services/web/lib/api/schema.ts (TypeScript types)
```

**Commands**:
```bash
# Update API spec
vim platform/api-spec/spec.yaml
make generate-api                # Update Go code
cd services/web && npm run generate:api  # Update TypeScript
```

### SQL Queries → Go Code

```
platform/database/sqlc/queries/*.sql
    ↓
sqlc → generated/db/queries.sql.go (type-safe Go functions)
```

**Commands**:
```bash
# Update queries
vim platform/database/sqlc/queries/employees.sql
make generate-db
```

---

## When to Add Code Where

### Decision Tree

**1. Is it a complete application?**
- Yes → New service in `services/`
- No → Continue

**2. Is it a source of truth (schema, API contract)?**
- Yes → `platform/`
- No → Continue

**3. Is it a simple type shared by multiple services?**
- Yes → `pkg/types/`
- No → Continue

**4. Is it auto-generated?**
- Yes → `generated/` (via code generation)
- No → Continue

**5. Is it documentation?**
- Yes → `docs/`
- No → Continue

**6. Is it a cross-cutting script?**
- Yes → `scripts/`
- No → Service-specific `internal/`

### Examples

**Adding a new API endpoint**:
1. Update `platform/api-spec/spec.yaml` (API contract)
2. Add SQL query to `platform/database/sqlc/queries/`
3. Run `make generate`
4. Implement handler in `services/api/internal/handlers/`
5. Write tests in `services/api/internal/handlers/*_test.go`

**Adding a new CLI command**:
1. Add command to `services/cli/internal/commands/`
2. Update `services/cli/cmd/arfa/main.go`
3. Write tests in `services/cli/tests/integration/`

**Adding a new UI page**:
1. Ensure API endpoint exists (see above)
2. Add page to `services/web/app/`
3. Add components to `services/web/components/`
4. Write tests in `services/web/tests/e2e/`

**Adding a shared type**:
1. Add to `pkg/types/types.go` (only if truly needed by multiple services)
2. Update all services that use it
3. Run tests in each service

---

## Benefits of This Structure

### Service Independence

- **Independent Versioning**: API v0.5 + CLI v1.0 possible
- **Parallel Development**: Teams can work on different services simultaneously
- **Clear Ownership**: Each service has clear boundaries and responsibilities

### Smaller Binaries

- **CLI**: ~60% smaller (no database drivers, HTTP handlers)
- **API**: No CLI dependencies
- **Web**: No Go dependencies

### Better Modularity

- **Clear Boundaries**: Services can't accidentally import from each other
- **Easier Testing**: Each service tested independently
- **Simpler Deployment**: Each service deployed independently

### Type Safety

- **Generated Code**: Database queries, API types all type-safe
- **Single Source of Truth**: Schema and API spec generate consistent code
- **Compile-Time Checks**: Type mismatches caught at compile time

### Developer Experience

- **Fast Iteration**: Only regenerate what changed
- **Clear Documentation**: Each service has own README
- **Easy Onboarding**: New developers start with one service
- **Consistent Patterns**: All services follow same structure

---

## Migration History

### Pre-Monorepo (v0.1.0 - v0.2.0)

**Structure**:
```
arfa/
├── cmd/server/           # API server
├── internal/             # Shared code (handlers, auth, etc.)
├── tests/                # All tests
├── shared/schema/        # Database schema
├── openapi/spec.yaml     # API spec
└── sqlc/                 # SQL queries
```

**Problems**:
- CLI binary included all API dependencies (database drivers, HTTP handlers)
- No clear service boundaries
- Difficult to version independently
- Large binary sizes

### Monorepo v1 (v0.3.0+)

**Structure**: See [Directory Organization](#directory-organization) above

**Benefits**:
- Self-contained services
- Clear dependency boundaries
- Smaller binaries
- Independent versioning
- Better modularity

**Migration**: See [docs/MONOREPO_MIGRATION.md](../MONOREPO_MIGRATION.md) for detailed migration process.

---

## Related Documentation

- **[CLAUDE.md](../../CLAUDE.md)** - Complete system documentation
- **[DEVELOPMENT.md](../DEVELOPMENT.md)** - Development workflow
- **[TESTING.md](../TESTING.md)** - Testing strategy
- **[services/api/README.md](../../services/api/README.md)** - API service docs
- **[services/cli/README.md](../../services/cli/README.md)** - CLI service docs
- **[services/web/README.md](../../services/web/README.md)** - Web UI docs
- **[docs/MONOREPO_MIGRATION.md](../MONOREPO_MIGRATION.md)** - Migration guide

---

**Last Updated**: 2025-11-13
**Version**: v0.3.0+
**Maintained By**: Tech Lead
