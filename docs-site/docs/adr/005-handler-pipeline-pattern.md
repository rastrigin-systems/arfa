---
sidebar_position: 6
---

# ADR-005: Handler Pipeline Pattern

**Status:** Accepted
**Date:** 2025-11

## Context

The CLI proxy needs to process LLM API traffic with multiple concerns:

- Policy enforcement (block dangerous tools)
- Logging (capture activity)
- PII detection (identify sensitive data)
- Analytics (track usage)

These concerns may be enabled/disabled independently.

## Decision

Implement a **handler pipeline pattern**:

```go
type Handler interface {
    Name() string
    Priority() int
    HandleRequest(ctx HandlerContext, req *http.Request) Result
    HandleResponse(ctx HandlerContext, res *http.Response) Result
}
```

**Pipeline flow:**

```
Request → [Policy] → [Logger] → [PII] → [Analytics] → LLM
                                                        │
Response ← [Analytics] ← [PII] ← [Logger] ← [Policy] ←─┘
```

| Handler | Priority | Purpose |
|---------|----------|---------|
| PolicyHandler | 100 | Block/allow decisions |
| LoggerHandler | 50 | Capture tool calls |
| PIIHandler | 40 | Detect sensitive data |
| AnalyticsHandler | 10 | Track metrics |

## Consequences

### Positive

- **Extensibility**: Add handlers without modifying existing code
- **Testability**: Each handler tested independently
- **Configurability**: Enable/disable handlers via config

### Negative

- **Complexity**: More abstractions to understand
- **Overhead**: Pipeline invocation has small cost

## Alternatives Considered

1. **Single handler** - Rejected: Hard to test, no modularity
2. **Event-based** - Rejected: Harder to control ordering
