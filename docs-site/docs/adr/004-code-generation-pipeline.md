---
sidebar_position: 5
---

# ADR-004: Code Generation Pipeline

**Status:** Accepted
**Date:** 2025-10

## Context

Arfa needs consistency between database schema, SQL queries, API contract, and Go/TypeScript code. Manual synchronization leads to type mismatches and runtime errors.

## Decision

Implement a **code generation pipeline** with two sources of truth:

```
Sources of Truth
├── platform/database/schema.sql    → Database structure
└── platform/api-spec/spec.yaml     → API contract

Generated Code (Never Edit)
├── generated/db/                   → sqlc
├── generated/api/                  → oapi-codegen
└── services/web/lib/api/schema.ts  → openapi-typescript
```

**Workflow:**
```bash
make db-reset    # Apply schema
make generate    # Regenerate all code
make test        # Verify everything works
```

## Consequences

### Positive

- **Type safety**: Compile-time errors instead of runtime
- **Consistency**: Single source of truth per domain
- **Documentation**: API spec doubles as documentation

### Negative

- **Regeneration overhead**: Need `make generate` after changes
- **Tool dependencies**: Must install sqlc, oapi-codegen
- **CI complexity**: Need to verify generated code is fresh

## Alternatives Considered

1. **ORM (GORM)** - Rejected: Less control over SQL
2. **Manual types** - Rejected: Easy to get out of sync
3. **GraphQL** - Rejected: Overkill for our use case
