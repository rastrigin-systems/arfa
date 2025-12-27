---
sidebar_position: 1
---

# API Service

REST API server with WebSocket support for the Arfa platform.

## Structure

```
services/api/
├── cmd/server/          # Main entrypoint
├── internal/
│   ├── handlers/        # HTTP request handlers
│   ├── middleware/      # Auth, logging, RLS
│   ├── auth/            # JWT authentication
│   ├── database/        # Database layer
│   ├── service/         # Business logic
│   └── websocket/       # Real-time updates
└── tests/integration/   # Integration tests
```

## Request Flow

```
HTTP Request
    │
    ▼
┌─────────────────┐
│   Chi Router    │
└────────┬────────┘
         │
┌────────▼────────┐
│   Middleware    │
│ 1. Logging      │
│ 2. CORS         │
│ 3. JWT Auth     │
│ 4. RLS Context  │
└────────┬────────┘
         │
┌────────▼────────┐
│    Handler      │
│ 1. Parse input  │
│ 2. Validate     │
│ 3. Call sqlc    │
│ 4. Map types    │
│ 5. Return JSON  │
└─────────────────┘
```

## Key Endpoints

| Category | Endpoints |
|----------|-----------|
| Auth | `/auth/login`, `/auth/register`, `/auth/refresh` |
| Organizations | `/organizations`, `/teams`, `/employees` |
| Configuration | `/agents`, `/policies`, `/employee-agent-configs` |
| Sync | `/sync/employee/{id}`, WebSocket `/ws` |
| Webhooks | `/webhooks`, `/webhooks/{id}/test` |

## Dependencies

The API service depends on:
- `generated/api/` - OpenAPI types
- `generated/db/` - sqlc queries
- `pkg/types/` - Shared types

It does NOT depend on `services/cli` or `services/web`.
