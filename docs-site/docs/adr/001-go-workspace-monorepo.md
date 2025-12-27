---
sidebar_position: 2
---

# ADR-001: Go Workspace Monorepo

**Status:** Accepted
**Date:** 2025-10

## Context

Arfa consists of multiple services (API, CLI, Web) with different deployment targets:

- **API Service**: REST server deployed to Cloud Run, needs database drivers
- **CLI Service**: Binary distributed to end users, needs to be small (~10MB)
- **Web Service**: Next.js app, separate deployment

Initially, a single Go module caused bloated CLI binary and dependency conflicts.

## Decision

Use a **Go workspace monorepo** with self-contained service modules:

```
arfa/
├── go.work              # Go workspace config
├── services/
│   ├── api/go.mod       # Independent module
│   └── cli/go.mod       # Independent module
├── pkg/types/go.mod     # Shared types (minimal)
└── generated/go.mod     # Generated code
```

## Consequences

### Positive

- **Smaller CLI binary**: ~60% reduction
- **Clear boundaries**: Compile-time enforcement
- **Independent versioning**: API and CLI version separately
- **Faster builds**: Only rebuild affected service

### Negative

- **More complex setup**: Requires `go work sync`
- **Duplicate utilities**: Some code duplicated (intentional)
- **Learning curve**: Developers need to understand workspaces

## Alternatives Considered

1. **Single Module** - Rejected: Binary size issues
2. **Separate Repositories** - Rejected: Too much overhead
3. **Modules without Workspace** - Rejected: Harder to develop locally
