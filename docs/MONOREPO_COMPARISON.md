# Monorepo Structure Comparison

**Visual comparison of current vs. proposed structure**

---

## Current Structure (As-Is)

```
ubik-enterprise/                      ❌ Mixed concerns, duplicates
│
├── cmd/                              ❌ DUPLICATE of services/*/cmd/
│   ├── cli/
│   └── server/
│
├── internal/                         ❌ DUPLICATE of services/api/internal/
│   ├── auth/
│   ├── handlers/
│   └── middleware/
│
├── tests/                            ❌ DUPLICATE of services/api/tests/
│   ├── integration/
│   └── testutil/
│
├── shared/                           ⚠️ Mixed concerns
│   ├── schema/                       ✅ Good - DB source of truth
│   ├── openapi/                      ✅ Good - API source of truth
│   └── docker/                       ⚠️ Product feature, not "shared"
│
├── sqlc/                             ⚠️ Separated from schema
│   ├── sqlc.yaml
│   └── queries/
│
├── services/
│   ├── api/                          ⚠️ Missing deployment artifacts at service level
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── tests/
│   │   └── Dockerfile.gcp            ⚠️ Only GCP, no local Dockerfile
│   │
│   ├── cli/                          ⚠️ Missing build artifacts
│   │   ├── cmd/
│   │   └── internal/
│   │
│   └── web/                          ⚠️ Missing build directory
│       └── Dockerfile
│
├── pkg/types/                        ✅ Good - shared types
│
├── generated/                        ✅ Good - auto-generated
│
├── docs/                             ⚠️ Flat structure
│   ├── ERD.md
│   ├── TESTING.md
│   ├── user-stories/
│   └── wireframes/
│
├── cloudbuild.yaml                   ❌ Root-level deployment (should be per-service)
└── Makefile                          ⚠️ Knows about all services (tight coupling)
```

**Problems:**
- 🔴 Duplicate `cmd/`, `internal/`, `tests/` at root AND in services
- 🔴 Deployment configs scattered (root, services)
- 🔴 Unclear service boundaries (what belongs where?)
- 🔴 Hard to extract a service to separate repo
- 🔴 Tests spread across root and services
- 🔴 No clear ownership (CODEOWNERS would be complex)

---

## Proposed Structure (To-Be)

```
ubik-enterprise/                      ✅ Clear boundaries, no duplicates
│
├── services/                         ✅ Complete service encapsulation
│   │
│   ├── api/                          ✅ SELF-CONTAINED
│   │   ├── cmd/                      ✅ Service entrypoint
│   │   ├── internal/                 ✅ Private code
│   │   ├── pkg/                      ✅ Public packages (if any)
│   │   ├── tests/                    ✅ All tests co-located
│   │   │   ├── unit/
│   │   │   └── integration/
│   │   ├── build/                    ✅ ALL deployment artifacts
│   │   │   ├── Dockerfile            ✅ Local dev
│   │   │   ├── Dockerfile.gcp        ✅ Production
│   │   │   └── cloudbuild.yaml       ✅ Service-specific CI/CD
│   │   ├── scripts/                  ✅ Service-specific scripts
│   │   ├── docs/                     ✅ Service documentation
│   │   ├── Makefile                  ✅ Service build commands
│   │   └── README.md                 ✅ Service overview
│   │
│   ├── cli/                          ✅ SELF-CONTAINED
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── tests/
│   │   ├── build/
│   │   ├── scripts/
│   │   ├── docs/
│   │   ├── Makefile
│   │   └── README.md
│   │
│   └── web/                          ✅ SELF-CONTAINED
│       ├── app/
│       ├── components/
│       ├── tests/
│       ├── build/
│       ├── docs/
│       └── README.md
│
├── platform/                         ✅ Infrastructure source-of-truth
│   ├── database/                     ✅ Complete DB package
│   │   ├── schema.sql
│   │   ├── migrations/
│   │   ├── seeds/
│   │   └── sqlc/                     ✅ Co-located with schema
│   │       ├── sqlc.yaml
│   │       └── queries/
│   │
│   ├── api-spec/                     ✅ API contract
│   │   ├── spec.yaml
│   │   └── oapi-codegen.yaml
│   │
│   └── docker-images/                ✅ Dockerized agents/MCPs
│       ├── agents/
│       └── mcp-servers/
│
├── pkg/                              ✅ Shared PUBLIC packages
│   ├── types/
│   ├── errors/                       (future)
│   └── clients/                      (future)
│
├── internal/                         ✅ Shared PRIVATE code
│   ├── db/                           (if truly shared)
│   └── testutil/                     ✅ Shared test utilities
│
├── generated/                        ✅ Auto-generated (not committed)
│   ├── api/
│   ├── db/
│   └── mocks/
│
├── docs/                             ✅ Organized platform docs
│   ├── architecture/                 ✅ ADRs, designs
│   ├── guides/                       ✅ How-to guides
│   ├── product/                      ✅ User stories, wireframes
│   └── operations/                   ✅ Runbooks, deployments
│
├── scripts/                          ✅ Platform-wide scripts only
├── Makefile                          ✅ Orchestration (delegates to services)
└── .github/
    ├── workflows/
    │   ├── api-ci.yml                ✅ Per-service CI
    │   ├── cli-ci.yml
    │   └── web-ci.yml
    └── CODEOWNERS                    ✅ Clear ownership
```

**Benefits:**
- ✅ No duplicates - single location for each concern
- ✅ Clear service boundaries - each service is self-contained
- ✅ Easy extraction - `cd services/api/` is complete service
- ✅ Clear ownership - CODEOWNERS maps cleanly
- ✅ Independent CI - only affected services run
- ✅ Discoverable - know where to find things

---

## Side-by-Side: API Service

### Current (Partial Containment)

```
services/api/
├── cmd/server/                       ✅ Entrypoint
├── internal/                         ✅ Code
│   ├── auth/
│   ├── handlers/
│   └── middleware/
├── tests/                            ✅ Tests
│   ├── integration/
│   └── testutil/
├── Dockerfile.gcp                    ⚠️ Only GCP version
└── go.mod                            ✅ Dependencies

❌ Missing:
- Local Dockerfile
- Service-specific cloudbuild.yaml
- Service-specific scripts
- Service documentation
- Build directory
- Makefile
```

### Proposed (Full Containment)

```
services/api/
├── README.md                         ✅ Service overview
├── Makefile                          ✅ Build commands
├── go.mod                            ✅ Dependencies
│
├── cmd/
│   └── server/
│       └── main.go                   ✅ Entrypoint
│
├── internal/                         ✅ Private service code
│   ├── app/
│   ├── handlers/
│   ├── middleware/
│   ├── service/
│   └── websocket/
│
├── pkg/                              ✅ Public packages (if reusable)
│   ├── auth/
│   └── email/
│
├── tests/                            ✅ All tests co-located
│   ├── unit/
│   ├── integration/
│   └── testutil/
│
├── build/                            ✅ ALL deployment artifacts
│   ├── Dockerfile                    ✅ Local development
│   ├── Dockerfile.gcp                ✅ Production build
│   └── cloudbuild.yaml               ✅ Service-specific CI/CD
│
├── scripts/                          ✅ Service-specific scripts
│   └── seed-data.sh
│
└── docs/                             ✅ Service documentation
    ├── API.md                        ✅ Design decisions
    └── DEPLOYMENT.md                 ✅ How to deploy

✅ Complete service package - can extract to separate repo
```

---

## Side-by-Side: Database Infrastructure

### Current (Separated)

```
shared/schema/
├── schema.sql                        ✅ Source of truth
├── migrations/
└── seeds/

sqlc/                                 ❌ Separated from schema
├── sqlc.yaml                         ⚠️ References ../shared/schema/
└── queries/
    ├── employees.sql
    └── ...
```

**Problem:** SQL queries and schema are logically one unit but physically separated.

### Proposed (Co-located)

```
platform/database/
├── README.md                         ✅ Database documentation
├── schema.sql                        ✅ Source of truth
├── migrations/                       ✅ Migration files
├── seeds/                            ✅ Seed data
└── sqlc/                             ✅ Co-located with schema
    ├── sqlc.yaml                     ✅ References ../schema.sql
    └── queries/                      ✅ SQL queries
        ├── employees.sql
        ├── organizations.sql
        └── ...
```

**Benefits:**
- ✅ Complete database package in one place
- ✅ Clear ownership - "platform team owns platform/database/"
- ✅ Easier to understand - everything DB-related in one directory

---

## Side-by-Side: Documentation

### Current (Flat)

```
docs/
├── ERD.md
├── README.md
├── TESTING.md
├── DEVELOPMENT.md
├── DEBUGGING.md
├── QUICKSTART.md
├── user-stories/
│   ├── epic-1-authentication/
│   └── epic-2-dashboard/
└── wireframes/
    ├── epic-1-authentication/
    └── epic-2-dashboard/
```

**Problem:** All docs at same level - hard to distinguish platform vs. product vs. service docs.

### Proposed (Organized)

```
docs/
├── architecture/                     ✅ Architecture decisions
│   ├── DECISIONS.md                  ✅ ADRs
│   └── MONOREPO.md                   ✅ This document
│
├── guides/                           ✅ How-to guides
│   ├── QUICKSTART.md
│   ├── DEVELOPMENT.md
│   ├── TESTING.md
│   └── DEBUGGING.md
│
├── product/                          ✅ Product documentation
│   ├── user-stories/
│   │   ├── epic-1-authentication/
│   │   └── epic-2-dashboard/
│   └── wireframes/
│       ├── epic-1-authentication/
│       └── epic-2-dashboard/
│
├── operations/                       ✅ Operational docs
│   ├── DEPLOYMENT.md
│   └── RUNBOOKS.md
│
├── ERD.md                            ✅ Auto-generated (stays at root)
└── README.md                         ✅ Auto-generated table index

services/api/docs/                    ✅ Service-specific docs
├── API.md
└── DEPLOYMENT.md

services/cli/docs/
└── CLI_ARCHITECTURE.md

services/web/docs/
└── WEB_ARCHITECTURE.md
```

**Benefits:**
- ✅ Clear categorization - know where to look
- ✅ Service docs with service code
- ✅ Platform docs at top level
- ✅ Product docs separated from technical docs

---

## Migration Impact Summary

### What Changes

| Area | Before | After | Impact |
|------|--------|-------|--------|
| **Service Structure** | Partial containment | Full containment | 🟡 Medium - file moves |
| **Deployment Configs** | Root + services | Per-service | 🟡 Medium - config updates |
| **Tests** | Root + services | Per-service only | 🟢 Low - just moves |
| **Shared Code** | `shared/` | `platform/` + `pkg/` + `internal/` | 🟡 Medium - renames |
| **Documentation** | Flat | Organized | 🟢 Low - just moves |
| **Build System** | Root Makefile | Root + per-service | 🟡 Medium - new Makefiles |
| **CI/CD** | Single workflow | Per-service workflows | 🔴 High - workflow changes |

### What Stays the Same

- ✅ Go workspace still manages dependencies
- ✅ `generated/` at root (single source of truth)
- ✅ `pkg/types` still shared across services
- ✅ Code generation pipeline unchanged
- ✅ Docker Compose orchestration unchanged
- ✅ Testing patterns unchanged

---

## Key Improvements Visualized

### Ownership Clarity

**Before:**
```
Who owns cmd/server?
Who owns internal/handlers?
Who owns tests/integration?
```
❌ Ambiguous - could be root-level or service-level

**After:**
```
services/api/cmd/server/     → API team
services/api/internal/       → API team
services/api/tests/          → API team
platform/database/           → Platform team
pkg/types/                   → Platform team
```
✅ Crystal clear ownership via CODEOWNERS

---

### Build Independence

**Before:**
```bash
# Must build from root
cd /path/to/ubik-enterprise
make build-server

# Must understand full monorepo
```
❌ Coupled to monorepo structure

**After:**
```bash
# Can build from service directory
cd /path/to/ubik-enterprise/services/api
make build

# Service knows how to build itself
```
✅ Service independence

---

### Service Extraction

**Before:**
```
To extract API service to separate repo:
1. Copy services/api/
2. Copy cmd/server/ (but not cmd/cli/)
3. Copy internal/ (but only API parts)
4. Copy tests/ (but only API tests)
5. Figure out which scripts are needed
6. Recreate Dockerfile
7. Recreate cloudbuild.yaml
```
❌ Complex, error-prone

**After:**
```
To extract API service to separate repo:
1. cp -r services/api/ ../ubik-api/
2. Add references to platform/ and pkg/types/
   (or extract as Go modules)
```
✅ Simple, clean

---

## Conclusion

The proposed refactoring transforms the monorepo from:

**"Good enough for 2-3 services"**

to:

**"Best practice for 50+ services"**

With clear boundaries, full service containment, and scalable patterns that follow industry standards from Google, Uber, and the Go community.

---

**See:** [MONOREPO_REFACTORING_PLAN.md](./MONOREPO_REFACTORING_PLAN.md) for complete migration details.
