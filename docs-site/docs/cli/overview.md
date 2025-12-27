---
sidebar_position: 1
---

# CLI Service

Transparent HTTPS proxy for intercepting AI agent traffic.

## Structure

```
services/cli/
├── cmd/arfa/           # Main CLI entrypoint
├── internal/
│   ├── control/        # Control service (proxy + handlers)
│   ├── commands/       # CLI command implementations
│   ├── auth.go         # JWT token management
│   ├── sync.go         # Configuration sync
│   └── logging/        # Activity logging
└── tests/
    ├── integration/
    └── e2e/
```

## Core Commands

| Command | Description |
|---------|-------------|
| `arfa login` | Authenticate with platform |
| `arfa logout` | Clear session |
| `arfa proxy start` | Start HTTPS proxy |
| `arfa proxy stop` | Stop proxy |
| `arfa sync` | Sync agent configurations |
| `arfa logs view` | View activity logs |

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  CLI Process                                                │
│  ┌─────────────┐    ┌───────────────────────────────────┐   │
│  │   Control   │───▶│        Handler Pipeline           │   │
│  │   Service   │    │ Policy → Logger → PII → Analytics │   │
│  └─────────────┘    └───────────────────────────────────┘   │
│         │                         │                         │
│  ┌──────▼──────┐          ┌───────▼───────┐                 │
│  │ MITM Proxy  │◀────────▶│ State Manager │                 │
│  └──────┬──────┘          └───────────────┘                 │
└─────────┼───────────────────────────────────────────────────┘
          │
     ┌────▼────┐
     │ LLM API │
     └─────────┘
```

## Design Constraints

### Self-Contained Binary

The CLI intentionally does NOT depend on:
- `generated/db` (database code)
- `generated/api` (server code)
- PostgreSQL drivers

This keeps the binary small (~10MB).

### Async Operations

Non-blocking design for:
- Log upload (queued, batched)
- Config sync (background refresh)
- Policy updates (WebSocket)
