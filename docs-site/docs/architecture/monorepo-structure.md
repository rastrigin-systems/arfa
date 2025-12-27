---
sidebar_position: 1
---

# Monorepo Structure

Arfa uses a **Go workspace monorepo** with self-contained services and shared platform resources.

## Directory Layout

```
arfa/
├── services/                     # Self-contained services
│   ├── api/                      # REST API (Go)
│   ├── cli/                      # Proxy CLI (Go)
│   └── web/                      # Admin UI (Next.js)
│
├── platform/                     # Shared resources (source of truth)
│   ├── api-spec/                 # OpenAPI 3.0.3 specification
│   └── database/                 # PostgreSQL schema, queries
│
├── pkg/types/                    # Shared Go types
│
├── generated/                    # Auto-generated code (not committed)
│   ├── api/                      # From OpenAPI spec
│   ├── db/                       # From SQL schema
│   └── mocks/                    # From interfaces
│
├── docs/                         # Documentation
├── go.work                       # Go workspace config
└── Makefile                      # Build automation
```

## Design Principles

### Service Independence

Each service is a complete Go/Node module with its own:
- `go.mod` or `package.json`
- `internal/` package for private code
- `tests/` directory
- `build/` deployment configs

**Rule:** Services never import from other services' `internal/` packages.

### Clear Boundaries

| Service | Dependencies | No Dependencies |
|---------|-------------|-----------------|
| API | `generated/api`, `generated/db`, `pkg/types` | `services/cli`, `services/web` |
| CLI | `pkg/types` | `generated/`, database packages |
| Web | `lib/api/schema.ts` | Go code |

### Binary Size Benefits

| Service | Binary Size | Why |
|---------|-------------|-----|
| CLI | ~10MB | No DB drivers, HTTP handlers |
| API | ~25MB | Full stack |

The CLI is distributed to end users, so keeping it small matters.
