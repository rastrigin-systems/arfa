# Monorepo Migration Plan

**Date:** 2025-10-29
**Version:** 0.2.0 → Monorepo Structure
**Status:** 🚧 In Progress

## Overview

Migrating from single Go module to Go workspace-based monorepo.

## Why Migrate?

1. **Cleaner Dependencies** - CLI won't carry server dependencies (DB drivers, API handlers)
2. **Independent Versioning** - API and CLI can have separate versions
3. **Better Modularity** - Clear boundaries between components
4. **Smaller Binaries** - CLI binary won't include unused server code
5. **Web UI Ready** - Structure ready for Next.js frontend
6. **Team Scaling** - Clear ownership boundaries

## Target Structure

```
ubik-enterprise/                  # Monorepo root
├── go.work                       # Go workspace file
├── Makefile                      # Root orchestration
├── docker-compose.yml
│
├── services/
│   ├── api/                      # API Server Module
│   │   ├── go.mod
│   │   ├── cmd/server/main.go
│   │   ├── internal/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   ├── auth/
│   │   │   └── service/
│   │   └── tests/integration/
│   │
│   └── cli/                      # CLI Client Module
│       ├── go.mod
│       ├── cmd/ubik/main.go
│       ├── internal/
│       │   ├── client/
│       │   ├── config/
│       │   ├── docker/
│       │   └── commands/
│       └── tests/
│
├── pkg/                          # Shared Go Code
│   ├── go.mod
│   └── types/                    # Shared types/models
│       ├── org.go
│       ├── employee.go
│       └── agent.go
│
├── shared/                       # Cross-language shared
│   ├── openapi/spec.yaml
│   ├── schema/schema.sql
│   └── docker/
│
├── generated/                    # Generated code (root level)
│   ├── api/
│   └── db/
│
└── docs/
```

## Migration Steps

### Phase 1: Backup & Setup
- [x] Create backup branch
- [ ] Create go.work file
- [ ] Create services/, pkg/, shared/ directories

### Phase 2: Move API Server
- [ ] Create services/api/go.mod
- [ ] Move cmd/server/ → services/api/cmd/server/
- [ ] Move relevant internal/ → services/api/internal/
- [ ] Move tests/integration/ → services/api/tests/
- [ ] Update import paths in API code

### Phase 3: Move CLI Client
- [ ] Create services/cli/go.mod
- [ ] Move cmd/cli/ → services/cli/cmd/ubik/
- [ ] Move internal/cli/ → services/cli/internal/
- [ ] Update import paths in CLI code

### Phase 4: Create Shared Package
- [ ] Create pkg/types/go.mod
- [ ] Extract shared types to pkg/types/
- [ ] Update imports in API and CLI

### Phase 5: Organize Shared Resources
- [ ] Move openapi/ → shared/openapi/
- [ ] Move schema.sql → shared/schema/
- [ ] Move docker/ → shared/docker/
- [ ] Keep generated/ at root (or move to shared/)

### Phase 6: Update Build System
- [ ] Update Makefile for workspace
- [ ] Update docker-compose.yml paths
- [ ] Update sqlc.yaml paths
- [ ] Update oapi-codegen paths

### Phase 7: Documentation
- [ ] Update CLAUDE.md
- [ ] Update README.md
- [ ] Update docs/DEVELOPMENT.md
- [ ] Create docs/MONOREPO.md guide

### Phase 8: Testing & CI/CD
- [ ] Test API server build
- [ ] Test CLI build
- [ ] Run all tests
- [ ] Update GitHub workflows
- [ ] Test Docker builds

### Phase 9: Cleanup
- [ ] Remove old cmd/ directory
- [ ] Remove old internal/ directory
- [ ] Clean up root go.mod dependencies
- [ ] Archive old structure docs

## Key Changes

### Go Workspace (go.work)

```go
go 1.24.5

use (
    ./services/api
    ./services/cli
    ./pkg/types
)
```

### API Server Module (services/api/go.mod)

```go
module github.com/sergeirastrigin/ubik-enterprise/services/api

go 1.24.5

require (
    github.com/sergeirastrigin/ubik-enterprise/pkg/types v0.0.0
    github.com/go-chi/chi/v5 v5.0.11
    github.com/jackc/pgx/v5 v5.5.3
    // ... API-specific dependencies
)

replace github.com/sergeirastrigin/ubik-enterprise/pkg/types => ../../pkg/types
```

### CLI Module (services/cli/go.mod)

```go
module github.com/sergeirastrigin/ubik-enterprise/services/cli

go 1.24.5

require (
    github.com/sergeirastrigin/ubik-enterprise/pkg/types v0.0.0
    github.com/spf13/cobra v1.10.1
    github.com/docker/docker v28.5.1+incompatible
    // ... CLI-specific dependencies
)

replace github.com/sergeirastrigin/ubik-enterprise/pkg/types => ../../pkg/types
```

### Shared Types (pkg/types/go.mod)

```go
module github.com/sergeirastrigin/ubik-enterprise/pkg/types

go 1.24.5

require (
    github.com/google/uuid v1.6.0
)
```

## Import Path Changes

### Before
```go
import (
    "github.com/sergeirastrigin/ubik-enterprise/internal/auth"
    "github.com/sergeirastrigin/ubik-enterprise/generated/api"
)
```

### After (API)
```go
import (
    "github.com/sergeirastrigin/ubik-enterprise/services/api/internal/auth"
    "github.com/sergeirastrigin/ubik-enterprise/generated/api"
    "github.com/sergeirastrigin/ubik-enterprise/pkg/types"
)
```

### After (CLI)
```go
import (
    "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/config"
    "github.com/sergeirastrigin/ubik-enterprise/pkg/types"
)
```

## Makefile Changes

### Before
```makefile
build:
	go build -o bin/server cmd/server/main.go
	go build -o bin/ubik cmd/cli/main.go
```

### After
```makefile
build:
	cd services/api && go build -o ../../bin/server cmd/server/main.go
	cd services/cli && go build -o ../../bin/ubik cmd/ubik/main.go

test:
	go work sync
	cd services/api && go test ./...
	cd services/cli && go test ./...
	cd pkg/types && go test ./...
```

## Rollback Plan

If migration fails:
```bash
git checkout main
git branch -D feature/monorepo-migration
```

All changes are in a feature branch and can be discarded.

## Success Criteria

- [ ] `make build` produces working binaries
- [ ] All tests pass (API, CLI, shared)
- [ ] Docker builds work
- [ ] CI/CD workflows pass
- [ ] Documentation updated
- [ ] CLI binary size reduced (~30-50% smaller)

## Estimated Timeline

- **Total:** 2-4 hours
- **Phase 1-3:** 1 hour (structure + moves)
- **Phase 4-6:** 1 hour (shared code + build)
- **Phase 7-9:** 1-2 hours (docs + testing + cleanup)

## Notes

- Keep generated/ at root for now (shared by both services)
- Use Go workspace replace directives for local development
- Eventually publish pkg/types as versioned module
- Web UI will be added as services/web/ later
