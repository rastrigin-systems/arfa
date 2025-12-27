---
sidebar_position: 7
---

# ADR-006: Async Log Upload

**Status:** Accepted
**Date:** 2025-11

## Context

The CLI captures activity logs and uploads them to the platform. Requirements:

1. **Non-blocking**: Upload must not slow AI operations
2. **Reliable**: Logs must not be lost on network issues
3. **Efficient**: Minimize API calls through batching

## Decision

Implement **async log upload with local queue fallback**:

```
Event → Buffer (100 items) → Batch Upload
                                  │
                          ┌───────▼───────┐
                          │   Success?    │
                          └───────┬───────┘
                            yes │   │ no
                                │   ▼
                           Done │ Retry (5x)
                                │   │
                                │   ▼
                                │ Disk Queue
                                │ ~/.arfa/log_queue/
```

**Configuration:**
- Buffer: 100 entries
- Batch interval: 5 seconds
- Max retries: 5 (exponential backoff)
- Disk retry: Every 10 seconds

## Consequences

### Positive

- **Zero latency impact**: Logging never blocks agent
- **Reliability**: Disk queue survives process restart
- **Efficiency**: Batching reduces API calls ~95%

### Negative

- **Eventual consistency**: Logs may appear with delay
- **Disk usage**: Failed logs accumulate on disk

## Alternatives Considered

1. **Synchronous upload** - Rejected: Adds latency
2. **Fire and forget** - Rejected: Loses logs on failure
3. **Local database** - Rejected: Overkill for this use case
