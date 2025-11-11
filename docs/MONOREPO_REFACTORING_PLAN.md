# Ubik Enterprise Monorepo Refactoring Plan

**Version:** 1.0
**Date:** 2025-11-11
**Status:** Proposal

---

## Executive Summary

This document proposes a comprehensive refactoring of the Ubik Enterprise monorepo structure to establish clear service boundaries, improve maintainability, and prepare for future growth. The current structure has good separation at the Go workspace level but suffers from:

1. **Mixed concerns** - Deployment configs at root, tests scattered across services
2. **Unclear boundaries** - Some internal/ code at root duplicates service-specific code
3. **Incomplete service encapsulation** - Services don't fully own their deployment artifacts
4. **Shared resources ambiguity** - Schema, OpenAPI, and SQL queries live in different locations

The proposed refactoring follows industry best practices from Google, Uber, and the golang-standards project-layout, adapted for our multi-language platform (Go + Next.js + Docker).

---

## Current State Analysis

### Current Directory Structure

```
ubik-enterprise/                      # Monorepo root
├── go.work                           # Go workspace (API, CLI, types, generated)
├── go.mod                            # Root module (legacy?)
├── Makefile                          # All build commands
├── docker-compose.yml                # Local dev environment
├── cloudbuild.yaml                   # GCP deployment (all services)
│
├── services/
│   ├── api/                          # API Server Module
│   │   ├── go.mod                    # Own dependencies (240 lines go.sum)
│   │   ├── cmd/server/               # Server entrypoint
│   │   ├── internal/                 # Service-specific code
│   │   │   ├── auth/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   ├── service/
│   │   │   └── websocket/
│   │   ├── pkg/                      # API-specific packages (auth, email)
│   │   ├── tests/                    # API tests
│   │   │   ├── integration/
│   │   │   └── testutil/
│   │   └── Dockerfile.gcp            # GCP deployment only
│   │
│   ├── cli/                          # CLI Client Module
│   │   ├── go.mod                    # Own dependencies (105 lines go.sum)
│   │   ├── cmd/ubik/                 # CLI entrypoint
│   │   ├── internal/                 # CLI-specific code
│   │   │   ├── commands/
│   │   │   └── logging/
│   │   ├── tests/                    # CLI tests
│   │   │   └── integration/
│   │   └── bin/                      # Build output
│   │
│   └── web/                          # Next.js Web UI
│       ├── package.json              # Node dependencies
│       ├── app/                      # Next.js 14 app router
│       ├── components/
│       ├── lib/
│       ├── tests/
│       │   └── e2e/                  # Playwright E2E tests
│       └── Dockerfile                # Web deployment
│
├── pkg/types/                        # Shared Go types module
│   └── go.mod
│
├── generated/                        # Auto-generated code (not committed)
│   ├── go.mod
│   ├── api/                          # From openapi/spec.yaml
│   ├── db/                           # From sqlc queries
│   └── mocks/                        # From mockgen
│
├── shared/                           # Cross-language shared resources
│   ├── schema/
│   │   ├── schema.sql                # Database source of truth
│   │   ├── migrations/
│   │   └── seeds/
│   ├── openapi/
│   │   ├── spec.yaml                 # API source of truth
│   │   └── oapi-codegen.yaml
│   └── docker/                       # Dockerized agents/MCPs
│       ├── agents/claude-code/
│       └── mcp/filesystem/
│
├── sqlc/                             # SQL queries for code generation
│   ├── sqlc.yaml                     # References ../shared/schema/
│   └── queries/
│
├── docs/                             # All documentation
│   ├── ERD.md                        # Auto-generated from DB
│   ├── TESTING.md
│   ├── wireframes/
│   └── user-stories/
│
├── cmd/                              # Root-level commands (LEGACY)
│   ├── cli/
│   └── server/
│
├── internal/                         # Root-level internal code (DUPLICATE)
│   ├── auth/
│   ├── cli/
│   ├── handlers/
│   ├── middleware/
│   └── service/
│
├── tests/                            # Root-level tests (DUPLICATE)
│   ├── integration/
│   └── testutil/
│
└── scripts/                          # Utility scripts
    ├── generate-erd-overview.py
    └── seed-claude-config.sh
```

### Issues Identified

#### 1. **Code Duplication**
- `cmd/` at root duplicates `services/api/cmd/` and `services/cli/cmd/`
- `internal/` at root duplicates `services/api/internal/`
- `tests/` at root duplicates `services/api/tests/` and `services/cli/tests/`
- Root `go.mod` may be legacy/unused

#### 2. **Incomplete Service Encapsulation**
- Deployment configs scattered:
  - `cloudbuild.yaml` at root (should be per-service)
  - `docker-compose.yml` at root (orchestration OK, but references services)
  - `services/api/Dockerfile.gcp` only (no local dev Dockerfile)
  - `services/web/Dockerfile` only
- Tests not fully co-located with services
- Service-specific scripts not co-located

#### 3. **Shared Resources Organization**
- `shared/` is good but mixed concerns:
  - `shared/docker/` is for agent/MCP containers (product feature)
  - `shared/schema/` is infrastructure (DB source of truth)
  - `shared/openapi/` is infrastructure (API source of truth)
- `sqlc/` separate from `shared/` but depends on it

#### 4. **Build System**
- Single root `Makefile` knows about all services
- No per-service build/test isolation
- Hard to extract a service into separate repo

---

## Proposed Target Structure

### Design Principles

Following industry best practices from Google, Uber, golang-standards, and modern monorepo patterns:

1. **Service Self-Containment** - Each service folder contains ALL service-specific code, configs, tests, and deployment artifacts
2. **Clear Ownership** - Every folder has clear OWNERS (via CODEOWNERS file)
3. **Extractability** - Any service can be moved to separate repo with minimal changes
4. **Shared Code Explicitness** - Shared code in dedicated locations with clear visibility
5. **Build Isolation** - Each service can build/test independently
6. **Documentation Co-location** - Service-specific docs with service, platform docs at root

### Target Directory Structure

```
ubik-enterprise/                      # Monorepo root
│
├── .github/                          # GitHub workflows
│   ├── workflows/
│   │   ├── api-ci.yml               # API-specific CI
│   │   ├── cli-ci.yml               # CLI-specific CI
│   │   ├── web-ci.yml               # Web-specific CI
│   │   └── monorepo-ci.yml          # Cross-cutting checks
│   └── CODEOWNERS                    # Ownership mapping
│
├── go.work                           # Go workspace definition
├── Makefile                          # Orchestration only (delegates to services)
├── docker-compose.yml                # Local dev environment (orchestration)
├── README.md                         # Quick overview
├── CLAUDE.md                         # Complete documentation hub
│
├── services/                         # 🎯 All Services
│   │
│   ├── api/                          # ✅ API Server Service (Go)
│   │   ├── README.md                 # Service-specific docs
│   │   ├── go.mod                    # Service dependencies
│   │   ├── Makefile                  # Service-specific build commands
│   │   │
│   │   ├── cmd/                      # Service entrypoints
│   │   │   └── server/
│   │   │       └── main.go
│   │   │
│   │   ├── internal/                 # Private service code (Go compiler enforced)
│   │   │   ├── app/                  # Application layer
│   │   │   │   └── server.go         # HTTP server setup
│   │   │   ├── handlers/             # HTTP handlers
│   │   │   │   ├── auth.go
│   │   │   │   ├── employees.go
│   │   │   │   └── ...
│   │   │   ├── middleware/           # HTTP middleware
│   │   │   ├── service/              # Business logic layer
│   │   │   ├── auth/                 # Auth utilities
│   │   │   └── websocket/            # WebSocket handlers
│   │   │
│   │   ├── pkg/                      # Public packages (could be extracted)
│   │   │   ├── auth/                 # Auth utilities (if reusable)
│   │   │   └── email/                # Email utilities
│   │   │
│   │   ├── tests/                    # All API tests
│   │   │   ├── unit/                 # Unit tests (near code also OK)
│   │   │   ├── integration/          # Integration tests
│   │   │   └── testutil/             # Test utilities
│   │   │
│   │   ├── build/                    # Build artifacts & configs
│   │   │   ├── Dockerfile            # Local/dev build
│   │   │   ├── Dockerfile.gcp        # GCP production build
│   │   │   └── cloudbuild.yaml       # GCP deployment config
│   │   │
│   │   ├── scripts/                  # Service-specific scripts
│   │   │   └── seed-data.sh
│   │   │
│   │   └── docs/                     # Service-specific documentation
│   │       ├── API.md                # API design decisions
│   │       └── DEPLOYMENT.md         # Deployment guide
│   │
│   ├── cli/                          # ✅ CLI Client Service (Go)
│   │   ├── README.md
│   │   ├── go.mod
│   │   ├── Makefile
│   │   │
│   │   ├── cmd/
│   │   │   └── ubik/
│   │   │       └── main.go
│   │   │
│   │   ├── internal/                 # Private CLI code
│   │   │   ├── app/                  # CLI app setup
│   │   │   ├── commands/             # Cobra commands
│   │   │   │   ├── login.go
│   │   │   │   ├── sync.go
│   │   │   │   └── agents.go
│   │   │   ├── docker/               # Docker integration
│   │   │   ├── config/               # Config management
│   │   │   └── logging/              # Logging utilities
│   │   │
│   │   ├── tests/
│   │   │   ├── integration/
│   │   │   └── testutil/
│   │   │
│   │   ├── build/                    # Build artifacts
│   │   │   ├── Dockerfile            # For CLI container builds (if needed)
│   │   │   └── install.sh            # Installation script
│   │   │
│   │   └── docs/
│   │       ├── CLI_USAGE.md
│   │       └── CLI_ARCHITECTURE.md
│   │
│   ├── web/                          # ✅ Web UI Service (Next.js)
│   │   ├── README.md
│   │   ├── package.json
│   │   ├── Makefile                  # Optional (npm scripts may suffice)
│   │   │
│   │   ├── app/                      # Next.js 14 app router
│   │   ├── components/
│   │   ├── lib/
│   │   │
│   │   ├── tests/
│   │   │   ├── unit/
│   │   │   ├── integration/
│   │   │   └── e2e/                  # Playwright tests
│   │   │
│   │   ├── build/                    # Build artifacts
│   │   │   ├── Dockerfile
│   │   │   └── cloudbuild.yaml
│   │   │
│   │   └── docs/
│   │       └── WEB_ARCHITECTURE.md
│   │
│   └── worker/                       # 🔮 Future: Background Workers (Go)
│       ├── README.md
│       ├── go.mod
│       ├── cmd/
│       ├── internal/
│       ├── tests/
│       ├── build/
│       └── docs/
│
├── pkg/                              # 📦 Shared Go Packages (Public)
│   │
│   ├── types/                        # Shared domain types
│   │   ├── go.mod
│   │   ├── employee.go
│   │   ├── organization.go
│   │   └── ...
│   │
│   ├── errors/                       # Common error types (future)
│   │   └── go.mod
│   │
│   └── clients/                      # API clients (future)
│       ├── go.mod
│       └── api/
│           └── client.go
│
├── internal/                         # 🔒 Shared Internal Code (Private)
│   │
│   ├── db/                           # Database utilities (if shared)
│   │   └── connection.go
│   │
│   └── testutil/                     # Shared test utilities
│       └── postgres.go               # TestContainer setup
│
├── platform/                         # 🏗️ Platform Infrastructure
│   │
│   ├── database/                     # Database source of truth
│   │   ├── README.md
│   │   ├── schema.sql                # Complete schema
│   │   ├── migrations/               # Migration SQL files
│   │   │   └── 001_skills_and_mcp.sql
│   │   ├── seeds/                    # Seed data
│   │   │   └── 002_claude_config.sql
│   │   └── sqlc/                     # SQL query definitions
│   │       ├── sqlc.yaml             # Points to ../schema.sql
│   │       └── queries/
│   │           ├── employees.sql
│   │           ├── organizations.sql
│   │           └── ...
│   │
│   ├── api-spec/                     # API contract (OpenAPI)
│   │   ├── README.md
│   │   ├── spec.yaml                 # OpenAPI 3.0.3 spec
│   │   └── oapi-codegen.yaml         # Code generation config
│   │
│   └── docker-images/                # Dockerized agents/MCPs (product features)
│       ├── README.md
│       ├── agents/
│       │   └── claude-code/
│       │       ├── Dockerfile
│       │       ├── entrypoint.sh
│       │       └── README.md
│       └── mcp-servers/
│           ├── filesystem/
│           └── git/
│
├── generated/                        # ⚠️ AUTO-GENERATED (not committed)
│   ├── go.mod
│   ├── api/                          # From platform/api-spec/spec.yaml
│   │   └── server.gen.go
│   ├── db/                           # From platform/database/sqlc/queries/
│   │   ├── models.go
│   │   ├── employees.sql.go
│   │   └── ...
│   └── mocks/                        # From mockgen
│       └── db_mock.go
│
├── docs/                             # 📚 Platform Documentation
│   ├── ERD.md                        # Auto-generated DB diagram
│   ├── README.md                     # Auto-generated table index (tbls)
│   ├── public.*.md                   # Auto-generated table docs (27 files)
│   ├── schema.svg                    # Visual ERD
│   ├── schema.json                   # Machine-readable schema
│   │
│   ├── architecture/                 # Architecture docs
│   │   ├── DECISIONS.md              # ADRs (Architecture Decision Records)
│   │   └── MONOREPO.md               # This document
│   │
│   ├── guides/                       # How-to guides
│   │   ├── QUICKSTART.md
│   │   ├── DEVELOPMENT.md
│   │   ├── TESTING.md
│   │   ├── DEBUGGING.md
│   │   └── DEV_WORKFLOW.md
│   │
│   ├── product/                      # Product documentation
│   │   ├── user-stories/
│   │   │   ├── epic-1-authentication/
│   │   │   └── epic-2-dashboard/
│   │   └── wireframes/
│   │       ├── epic-1-authentication/
│   │       └── epic-2-dashboard/
│   │
│   └── operations/                   # Operational docs
│       ├── DEPLOYMENT.md
│       ├── MONITORING.md
│       └── RUNBOOKS.md
│
├── scripts/                          # 🛠️ Build & Utility Scripts
│   ├── generate-erd-overview.py      # ERD generation
│   ├── install-hooks.sh              # Git hooks
│   └── check-drift.js                # Drift detection
│
└── tools/                            # 🔧 Development Tools
    ├── go.mod                        # Tools dependencies
    └── tools.go                      # Tool imports
```

---

## Design Rationale

### 1. Service Self-Containment

**Decision:** Each service contains ALL service-specific resources.

**Rationale:**
- **Google/Uber Pattern:** Each service is a folder with complete ownership
- **Extractability:** Can move service to separate repo with `git subtree` or manual copy
- **Clear Boundaries:** No ambiguity about what belongs to which service
- **Independent Evolution:** Services can adopt different patterns without affecting others

**Example:** The API service now owns:
- `services/api/build/Dockerfile` - Local dev build
- `services/api/build/Dockerfile.gcp` - Production build
- `services/api/build/cloudbuild.yaml` - GCP deployment
- `services/api/scripts/` - Service-specific scripts
- `services/api/docs/` - Service documentation
- `services/api/Makefile` - Service build commands

### 2. Shared Code Strategy

**Decision:** Three-tier sharing strategy:
1. `pkg/` - Public shared packages (could be extracted to separate modules)
2. `internal/` - Private shared code (monorepo-only, Go compiler enforced)
3. `platform/` - Infrastructure source-of-truth files (DB schema, API spec)

**Rationale:**
- **golang-standards Pattern:** Clear separation of public (`pkg/`) and private (`internal/`)
- **Go Compiler Enforcement:** `internal/` cannot be imported by external projects
- **Explicitness:** Developers know exactly where to put shared code
- **Minimal Coupling:** Services import from `pkg/types`, not from each other

**Trade-offs:**
- ✅ Clear visibility of dependencies
- ✅ Go workspace makes local development seamless
- ❌ Slightly more complex than "shared libs everywhere"
- ✅ Prevents accidental coupling between services

### 3. Platform Infrastructure

**Decision:** New `platform/` directory for source-of-truth infrastructure.

**Rationale:**
- **Semantic Clarity:** "Platform" clearly means "infrastructure everyone depends on"
- **Discoverability:** All infrastructure in one place
- **Separation of Concerns:**
  - `platform/database/` - DB schema, migrations, SQL queries
  - `platform/api-spec/` - OpenAPI specification
  - `platform/docker-images/` - Dockerized agents/MCPs (product feature)

**Why not keep `shared/`?**
- "Shared" is too generic (shared code? shared configs? shared data?)
- "Platform" is more descriptive (infrastructure that platforms are built on)
- Industry precedent: Many monorepos use `infra/`, `platform/`, or `foundation/`

**Alternative Considered:** Keep `shared/` but rename subdirectories
- Rejected because top-level name matters for discoverability

### 4. Build Isolation

**Decision:** Each service has its own `Makefile` (or equivalent), root `Makefile` delegates.

**Rationale:**
- **Service Independence:** Can build/test a service without understanding the whole monorepo
- **Parallel Development:** Teams can work on different services without build conflicts
- **Gradual Migration:** Can move one service to separate repo without rewriting build scripts

**Root Makefile Pattern:**
```makefile
# Orchestration only
test-all:
    @cd services/api && make test
    @cd services/cli && make test
    @cd services/web && npm test

build-all:
    @cd services/api && make build
    @cd services/cli && make build
    @cd services/web && npm run build
```

### 5. Documentation Organization

**Decision:** Three-tier documentation:
1. Service-specific docs in `services/*/docs/`
2. Platform-wide docs in `docs/`
3. Quick reference at root (`README.md`, `CLAUDE.md`)

**Rationale:**
- **Context Relevance:** API design decisions stay with API service
- **Discoverability:** Platform decisions visible at top level
- **Auto-generation:** `docs/ERD.md` and `docs/public.*.md` auto-generated from DB

### 6. Generated Code Location

**Decision:** Keep `generated/` at root, NOT per-service.

**Rationale:**
- **Single Source of Truth:** DB schema and API spec are platform-level
- **Avoid Duplication:** Multiple services need same generated code (API types, DB models)
- **Go Workspace:** Services import via `github.com/sergeirastrigin/ubik-enterprise/generated`
- **Clear Boundary:** `generated/` is never committed, always regenerated

**Why not per-service?**
- Would require duplicating code generation
- Would create sync issues (API and CLI both need API types)

### 7. SQL Queries with Database

**Decision:** Move `sqlc/` into `platform/database/sqlc/`.

**Rationale:**
- **Co-location:** SQL queries logically belong with database schema
- **Single Concern:** `platform/database/` is the complete database package
- **Discoverability:** Developers look in one place for all DB-related files

**Current:**
```
├── shared/schema/schema.sql
└── sqlc/
    ├── sqlc.yaml (references ../shared/schema/)
    └── queries/
```

**Proposed:**
```
└── platform/database/
    ├── schema.sql
    ├── migrations/
    └── sqlc/
        ├── sqlc.yaml (references ../schema.sql)
        └── queries/
```

---

## Comparison with Industry Best Practices

### Google's Monorepo

**What we're adopting:**
- ✅ Clear service ownership (each folder = one team)
- ✅ Shared libraries in dedicated locations
- ✅ Build system that supports service isolation

**What we're NOT adopting:**
- ❌ Bazel (too heavyweight for our scale)
- ❌ Single language (we have Go + Next.js + Docker)
- ❌ Ultra-large scale tooling (we're ~5-10 services, not 10,000)

### Uber's Monorepo

**What we're adopting:**
- ✅ Service-per-folder with OWNERS
- ✅ Clear dependency graph via Go workspace
- ✅ Independent service deployment

**What we're NOT adopting:**
- ❌ Custom deployment orchestration (we use GCP Cloud Build)
- ❌ Cross-service dependency analysis (Go workspace handles this)

### golang-standards/project-layout

**What we're adopting:**
- ✅ `/cmd` for entrypoints (per service)
- ✅ `/internal` for private code (per service + shared)
- ✅ `/pkg` for public packages (shared across services)
- ✅ Service-specific structure within each service

**What we're adapting:**
- 🔄 `/api` becomes `platform/api-spec/` (OpenAPI spec, not Go code)
- 🔄 `/build` becomes `services/*/build/` (per-service deployments)
- 🔄 Added `platform/` for infrastructure source-of-truth

---

## Migration Strategy

### Phase 1: Preparation (No Breaking Changes)

**Goal:** Set up new structure without breaking existing build/test.

**Tasks:**
1. Create new directories:
   ```bash
   mkdir -p platform/database platform/api-spec platform/docker-images
   mkdir -p services/{api,cli,web}/{build,docs,scripts}
   mkdir -p docs/{architecture,guides,product,operations}
   mkdir -p internal/testutil
   ```

2. Move files (with git mv to preserve history):
   ```bash
   # Database
   git mv shared/schema platform/database/schema.sql
   git mv shared/schema/migrations platform/database/migrations
   git mv shared/schema/seeds platform/database/seeds
   git mv sqlc platform/database/sqlc

   # API spec
   git mv shared/openapi platform/api-spec

   # Docker images
   git mv shared/docker platform/docker-images

   # Documentation
   git mv docs/user-stories docs/product/user-stories
   git mv docs/wireframes docs/product/wireframes
   ```

3. Update references in config files:
   - `platform/database/sqlc/sqlc.yaml` - schema path
   - `platform/api-spec/oapi-codegen.yaml` - spec path
   - `Makefile` - update paths

4. Update imports (automated):
   ```bash
   find . -name "*.go" -exec sed -i '' 's|shared/schema/|platform/database/|g' {} \;
   ```

**Validation:** Run `make test` and ensure all tests pass.

**Duration:** 1-2 days

---

### Phase 2: Service Consolidation (API Service)

**Goal:** Move API service to full self-containment pattern.

**Tasks:**

1. Move deployment configs:
   ```bash
   mkdir -p services/api/build
   git mv services/api/Dockerfile.gcp services/api/build/
   # Create new services/api/build/Dockerfile for local dev
   # Create new services/api/build/cloudbuild.yaml (API-specific)
   ```

2. Create service-specific Makefile:
   ```bash
   # services/api/Makefile
   .PHONY: build test clean

   build:
       CGO_ENABLED=0 go build -o bin/server cmd/server/main.go

   test:
       go test -v ./...

   test-integration:
       go test -v ./tests/integration/...

   clean:
       rm -rf bin/
   ```

3. Move service-specific scripts:
   ```bash
   # Any API-specific scripts from root scripts/ to services/api/scripts/
   ```

4. Create service documentation:
   ```bash
   # services/api/docs/API.md - Design decisions
   # services/api/docs/DEPLOYMENT.md - How to deploy
   ```

5. Remove duplicates:
   ```bash
   # Compare cmd/ at root with services/api/cmd/ - remove root version
   # Compare internal/ at root with services/api/internal/ - remove root version
   # Compare tests/ at root with services/api/tests/ - remove root version
   ```

6. Update root Makefile to delegate:
   ```makefile
   build-api:
       @cd services/api && make build

   test-api:
       @cd services/api && make test
   ```

7. Update cloudbuild.yaml at root:
   ```yaml
   # Reference services/api/build/cloudbuild.yaml
   # Or keep orchestration at root but use per-service Dockerfiles
   ```

**Validation:**
- `cd services/api && make build` - succeeds
- `cd services/api && make test` - all tests pass
- Docker build from `services/api/build/Dockerfile.gcp` - succeeds

**Duration:** 2-3 days

---

### Phase 3: Service Consolidation (CLI Service)

**Goal:** Apply same pattern to CLI service.

**Tasks:** (Same pattern as Phase 2, for CLI)

1. Create `services/cli/build/` with Dockerfile (if needed)
2. Create `services/cli/Makefile`
3. Move CLI scripts to `services/cli/scripts/`
4. Create `services/cli/docs/`
5. Update root Makefile to delegate

**Validation:**
- `cd services/cli && make build` - succeeds
- `cd services/cli && make test` - all tests pass

**Duration:** 1-2 days

---

### Phase 4: Service Consolidation (Web Service)

**Goal:** Apply same pattern to Web service.

**Tasks:**

1. Create `services/web/build/cloudbuild.yaml`
2. Create `services/web/Makefile` (optional, npm scripts may suffice)
3. Create `services/web/docs/`
4. Update root Makefile

**Validation:**
- `cd services/web && npm test` - succeeds
- Docker build succeeds

**Duration:** 1 day

---

### Phase 5: Root Cleanup

**Goal:** Remove legacy/duplicate files from root.

**Tasks:**

1. Remove duplicates:
   ```bash
   # ONLY if confirmed they're duplicates from Phase 2
   rm -rf cmd/
   rm -rf internal/
   rm -rf tests/
   ```

2. Remove root `go.mod` (if unused):
   ```bash
   # Check if anything imports from root module
   # If not, delete go.mod
   ```

3. Update root README.md:
   ```markdown
   # See services/api/README.md for API documentation
   # See services/cli/README.md for CLI documentation
   # See services/web/README.md for Web UI documentation
   ```

4. Create CODEOWNERS:
   ```
   /services/api/ @backend-team
   /services/cli/ @cli-team
   /services/web/ @frontend-team
   /platform/ @platform-team
   /docs/ @tech-writers
   ```

**Validation:**
- `make test-all` - all tests pass
- `make build-all` - all services build
- No broken imports

**Duration:** 1 day

---

### Phase 6: CI/CD Updates

**Goal:** Update GitHub Actions to use new structure.

**Tasks:**

1. Split CI workflows:
   ```yaml
   # .github/workflows/api-ci.yml
   name: API CI
   on:
     push:
       paths:
         - 'services/api/**'
         - 'platform/**'
         - 'pkg/**'
   jobs:
     test:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4
         - name: Test API
           run: cd services/api && make test
   ```

2. Create per-service CD workflows:
   ```yaml
   # .github/workflows/api-deploy.yml
   # Triggered on API changes only
   ```

3. Update deployment configs:
   ```yaml
   # cloudbuild.yaml at root orchestrates all services
   # OR move to per-service cloudbuild.yaml files
   ```

**Validation:**
- Push to PR - only affected services run CI
- Deploy - only changed services redeploy

**Duration:** 2-3 days

---

### Phase 7: Documentation Updates

**Goal:** Update all documentation to reflect new structure.

**Tasks:**

1. Update CLAUDE.md:
   - Replace old structure diagrams
   - Update file paths
   - Update commands

2. Move existing docs to new locations:
   ```bash
   git mv docs/TESTING.md docs/guides/TESTING.md
   git mv docs/DEVELOPMENT.md docs/guides/DEVELOPMENT.md
   git mv docs/DEBUGGING.md docs/guides/DEBUGGING.md
   ```

3. Create new documentation:
   - `docs/architecture/MONOREPO.md` (this document)
   - `docs/architecture/DECISIONS.md` (ADRs)
   - `services/api/docs/API.md`
   - `services/cli/docs/CLI_ARCHITECTURE.md`
   - `services/web/docs/WEB_ARCHITECTURE.md`

4. Update all absolute paths in docs:
   ```bash
   # Find all references to old paths and update
   grep -r "shared/schema" docs/ | # update references
   ```

**Validation:**
- All docs render correctly
- No broken links
- Commands in docs work

**Duration:** 2-3 days

---

## Migration Timeline

**Total Duration:** 2-3 weeks (assuming 1 developer full-time)

```
Week 1:
  Mon-Tue:  Phase 1 (Preparation)
  Wed-Thu:  Phase 2 (API Service)
  Fri:      Phase 3 start (CLI Service)

Week 2:
  Mon:      Phase 3 complete (CLI Service)
  Tue:      Phase 4 (Web Service)
  Wed:      Phase 5 (Root Cleanup)
  Thu-Fri:  Phase 6 (CI/CD)

Week 3:
  Mon-Wed:  Phase 7 (Documentation)
  Thu-Fri:  Buffer for issues, validation
```

**Risk Mitigation:**
- Create feature branch for entire migration
- Keep main branch stable throughout
- Each phase is independently testable
- Can pause between phases if needed

---

## Trade-offs Analysis

### Advantages ✅

1. **Clear Ownership**
   - Every file has clear owner (via directory structure)
   - Easy to enforce via CODEOWNERS

2. **Service Independence**
   - Each service can be extracted to separate repo
   - Teams can work independently

3. **Better Discoverability**
   - New developers know where to find things
   - No duplicate/ambiguous locations

4. **Scalability**
   - Easy to add new services (copy pattern)
   - Can move services to separate repos as needed

5. **Build Isolation**
   - Test/build one service without others
   - Faster CI (only affected services)

6. **Documentation Co-location**
   - Service docs with service code
   - Platform docs at top level

### Disadvantages ❌

1. **More Directories**
   - Deeper nesting: `services/api/build/Dockerfile.gcp`
   - Can be mitigated with good tooling/scripts

2. **Migration Effort**
   - 2-3 weeks of work
   - Mitigated by phased approach

3. **Learning Curve**
   - Team needs to learn new structure
   - Mitigated by clear documentation

4. **Potential Overhead**
   - More Makefiles, more config files
   - Mitigated by shared patterns/templates

### Alternatives Considered

#### Alternative 1: Keep Current Structure

**Pros:**
- No migration effort
- Team already familiar

**Cons:**
- Unclear boundaries remain
- Scales poorly (already have duplicates)
- Hard to extract services

**Verdict:** Rejected. Current problems will worsen as we add more services.

---

#### Alternative 2: Separate Repos (Polyrepo)

**Pros:**
- Ultimate service independence
- Clear ownership

**Cons:**
- Harder to share code (`pkg/types` becomes separate module)
- Cross-service changes require multiple PRs
- Version coordination hell
- Duplicated CI/CD configs

**Verdict:** Rejected for now. May revisit when we have 20+ services.

---

#### Alternative 3: Bazel-based Monorepo

**Pros:**
- Google-scale build performance
- Precise dependency tracking
- Hermetic builds

**Cons:**
- Massive learning curve
- Overkill for ~5 services
- Requires rewriting all build logic

**Verdict:** Rejected. Too heavyweight for our scale.

---

#### Alternative 4: Keep `shared/` instead of `platform/`

**Pros:**
- Less renaming
- Familiar name

**Cons:**
- "Shared" is too generic
- Doesn't convey "infrastructure" clearly

**Verdict:** Rejected. `platform/` is more descriptive.

---

## Success Criteria

### Must Have (P0)

- ✅ All tests pass after migration
- ✅ All services build independently (`cd services/api && make build`)
- ✅ CI/CD works with new structure
- ✅ No broken imports or references
- ✅ Documentation updated

### Should Have (P1)

- ✅ Per-service CI triggers only for affected services
- ✅ CODEOWNERS file maps ownership
- ✅ Service-specific docs created
- ✅ Root Makefile delegates to services

### Nice to Have (P2)

- ✅ Template service for future services
- ✅ Automated tooling for creating new services
- ✅ Architecture Decision Records (ADRs) documented

---

## Post-Migration Benefits

### Immediate Benefits

1. **Clarity:** No more confusion about where files belong
2. **Independence:** Can build/test services in isolation
3. **Efficiency:** CI only runs for affected services

### Medium-Term Benefits

1. **Team Scalability:** Multiple teams can own different services
2. **Onboarding:** New developers understand structure faster
3. **Extractability:** Can split monorepo if needed

### Long-Term Benefits

1. **Microservices Ready:** Easy to deploy services independently
2. **Technology Flexibility:** Services can adopt different patterns
3. **Codebase Longevity:** Structure supports growth to 50+ services

---

## Open Questions

1. **Root `go.mod`:** Is it still needed? What imports from it?
   - **Action:** Audit imports, decide to keep or remove

2. **Docker Compose:** Should it reference per-service Dockerfiles?
   - **Proposal:** Keep at root for orchestration, reference `services/*/build/Dockerfile`

3. **Per-service CI:** One workflow per service, or smart monorepo CI?
   - **Proposal:** Start with per-service, optimize later if needed

4. **Generated code:** Should each service have its own `generated/` dir?
   - **Decision:** No, keep at root (single source of truth)

5. **Shared test utilities:** In `internal/testutil/` or `pkg/testutil/`?
   - **Decision:** `internal/testutil/` (test code is always internal)

---

## Appendices

### Appendix A: File Move Commands

Complete script for Phase 1:

```bash
#!/bin/bash
set -e

# Create new directories
mkdir -p platform/database platform/api-spec platform/docker-images
mkdir -p services/{api,cli,web}/{build,docs,scripts}
mkdir -p docs/{architecture,guides,product,operations}
mkdir -p internal/testutil

# Move platform infrastructure
git mv shared/schema/schema.sql platform/database/
git mv shared/schema/migrations platform/database/
git mv shared/schema/seeds platform/database/
git mv sqlc platform/database/

git mv shared/openapi platform/api-spec
git mv shared/docker platform/docker-images

# Move documentation
git mv docs/user-stories docs/product/
git mv docs/wireframes docs/product/

# Update config files
sed -i '' 's|../shared/schema/|../|g' platform/database/sqlc/sqlc.yaml
sed -i '' 's|shared/openapi/|platform/api-spec/|g' Makefile

echo "✅ Phase 1 complete. Run 'make test' to validate."
```

### Appendix B: Service Template

Template for creating new services:

```
services/new-service/
├── README.md                 # Service overview
├── go.mod                    # Dependencies
├── Makefile                  # Build commands
├── cmd/
│   └── main.go
├── internal/
│   └── app/
├── tests/
│   ├── unit/
│   └── integration/
├── build/
│   ├── Dockerfile
│   └── cloudbuild.yaml
├── scripts/
└── docs/
    └── ARCHITECTURE.md
```

### Appendix C: References

**Industry Examples:**
- [Google's Monorepo](https://abseil.io/resources/swe-book/html/ch16.html)
- [Uber's Monorepo Strategy](https://www.uber.com/blog/continuous-deployment/)
- [Golang Project Layout](https://github.com/golang-standards/project-layout)

**Blog Posts:**
- [Building a Monorepo in Golang](https://earthly.dev/blog/golang-monorepo/)
- [Shared Go Packages in a Monorepo](https://passage.1password.com/post/shared-go-packages-in-a-monorepo)

---

## Summary

This refactoring proposal transforms Ubik Enterprise from a "good enough" monorepo to a **best-practice monorepo** that:

1. ✅ Follows industry patterns from Google, Uber, and golang-standards
2. ✅ Establishes clear service boundaries and ownership
3. ✅ Enables service independence and extractability
4. ✅ Improves discoverability and onboarding
5. ✅ Scales to 50+ services without structural changes
6. ✅ Can be migrated in 2-3 weeks with low risk

**Next Steps:**

1. Review this proposal with team
2. Get approval from stakeholders
3. Create GitHub issue with phases as tasks
4. Execute Phase 1 (preparation)
5. Iterate through phases with validation at each step

---

**Document Version:** 1.0
**Last Updated:** 2025-11-11
**Author:** Tech Lead (Claude)
**Status:** Awaiting Review
