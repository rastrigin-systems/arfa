# pkg/types

Shared types and models used across the ubik-enterprise platform.

## Purpose

This package contains common domain types that are shared between:
- API Server (`services/api`)
- CLI Client (`services/cli`)
- Future services (web UI, etc.)

## Usage

```go
import "github.com/sergeirastrigin/ubik-enterprise/pkg/types"

func example() {
    agent := types.Agent{
        ID:          uuid.New(),
        Name:        "Claude Code",
        Description: "AI coding assistant",
        Version:     "1.0.0",
        IsActive:    true,
    }
}
```

## Design Principles

1. **Minimal Dependencies** - Only core types (uuid, time)
2. **API Agnostic** - No HTTP/transport concerns
3. **Shared Domain** - Only types needed by multiple services
4. **Stable Interface** - Changes here affect all services

## What Belongs Here

✅ Domain models (Agent, Organization, Employee)
✅ Common enums/constants
✅ Shared value objects
✅ Business logic types

❌ HTTP request/response DTOs (stay in services/api)
❌ Database models (stay in services/api)
❌ Transport-specific types
❌ Service-specific business logic
