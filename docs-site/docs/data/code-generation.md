---
sidebar_position: 2
---

# Code Generation

Arfa uses extensive code generation to ensure type safety and consistency.

## Generation Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    Sources of Truth                         │
│ platform/database/schema.sql    platform/api-spec/spec.yaml │
└─────────────────────┬────────────────────────┬──────────────┘
                      │                        │
          ┌───────────┴───────────┐  ┌─────────┴─────────┐
          │         sqlc          │  │   oapi-codegen    │
          └───────────┬───────────┘  └─────────┬─────────┘
                      │                        │
          ┌───────────▼───────────┐  ┌─────────▼─────────┐
          │   generated/db/       │  │  generated/api/   │
          └───────────────────────┘  └───────────────────┘
                                             │
                                     ┌───────▼───────┐
                                     │openapi-typescript│
                                     └───────┬───────┘
                                     ┌───────▼───────────────┐
                                     │services/web/lib/api/  │
                                     │  schema.ts            │
                                     └───────────────────────┘
```

## Tools

| Tool | Input | Output |
|------|-------|--------|
| sqlc | SQL schema + queries | Go database code |
| oapi-codegen | OpenAPI spec | Go types + Chi router |
| openapi-typescript | OpenAPI spec | TypeScript types |
| mockgen | Go interfaces | Test mocks |

## Commands

```bash
# Generate everything
make generate

# Individual generators
make generate-db      # sqlc
make generate-api     # oapi-codegen
make generate-mocks   # mockgen
```

## Workflow

After changing source files:

```bash
# 1. Edit source
vim platform/database/schema.sql

# 2. Reset database (if schema changed)
make db-reset

# 3. Regenerate all code
make generate

# 4. Build and test
make build && make test
```

## Critical Rules

### Never Edit Generated Files

Generated files are completely overwritten:
- `generated/` directory
- `services/web/lib/api/schema.ts`

### Generated Code Not Committed

The `generated/` directory is in `.gitignore`. CI regenerates from source.

## Type Mapping

| PostgreSQL | Go (sqlc) | TypeScript |
|------------|-----------|------------|
| UUID | uuid.UUID | string |
| TIMESTAMPTZ | time.Time | string (ISO8601) |
| JSONB | json.RawMessage | `Record<string, unknown>` |
| VARCHAR | string | string |
| BOOLEAN | bool | boolean |
