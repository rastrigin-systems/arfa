---
sidebar_position: 3
---

# ADR-002: Transparent HTTPS Proxy

**Status:** Accepted
**Date:** 2025-10

## Context

AI coding assistants use HTTPS to communicate with LLM APIs. Organizations need:

1. **Visibility**: See tool calls and file accesses
2. **Control**: Block dangerous operations
3. **Audit**: Complete activity log for compliance

## Decision

Implement a **transparent MITM HTTPS proxy** in the CLI:

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
│ AI Client   │────▶│ Arfa Proxy      │────▶│ LLM APIs    │
└─────────────┘     └─────────────────┘     └─────────────┘
```

**Implementation:**
1. Generate local CA certificate on first run
2. Configure clients via `HTTPS_PROXY` and `NODE_EXTRA_CA_CERTS`
3. Dynamically generate certificates for each domain
4. Extract tool calls from SSE streams in real-time

## Consequences

### Positive

- **Full visibility**: Complete request/response access
- **Real-time blocking**: Stop tool calls before execution
- **Client agnostic**: Works with any client that respects proxy settings

### Negative

- **Certificate management**: Users must trust CA certificate
- **Security responsibility**: Proxy has access to all traffic
- **Performance overhead**: ~5-10ms latency added

## Alternatives Considered

1. **Client SDK/Plugin** - Rejected: Requires vendor cooperation
2. **Browser Extension** - Rejected: Most tools are CLI, not browser
3. **Network Firewall** - Rejected: Can't decrypt HTTPS anyway
