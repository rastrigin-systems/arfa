---
sidebar_position: 8
---

# ADR-007: Webhook Event Forwarding

**Status:** Accepted
**Date:** 2025-12

## Context

Organizations using Arfa have existing security infrastructure:
- SIEM systems (Splunk, Elastic, Datadog)
- Alerting platforms (PagerDuty, Slack)
- Compliance archives (S3, data warehouses)

They need to integrate Arfa data into these systems.

## Decision

Implement **real-time webhook event forwarding**:

```
Activity Log → Event Forwarder → Webhook Destination
                 (background)
```

**Features:**
- Authentication: Bearer token, custom headers, basic auth
- Event filtering: Subscribe to specific event types
- Signature verification: HMAC-SHA256
- Retry logic: Exponential backoff, max 3 attempts

**Payload:**
```json
{
  "id": "37e7c9c3-e286-4e27-8abc-4999fe50a6a5",
  "event_type": "tool_call",
  "timestamp": "2025-12-23T16:02:19.890Z",
  "org_id": "e5d10009-0988-44b6-b313-67ffbbbb1ef8",
  "payload": {
    "tool_name": "Bash",
    "blocked": false
  }
}
```

## Consequences

### Positive

- **Integration**: Works with existing SIEM infrastructure
- **Real-time**: Events delivered within seconds
- **Flexible**: Multiple destinations per organization
- **Secure**: Signature verification prevents tampering

### Negative

- **Complexity**: Need to manage delivery state
- **Latency**: Processing delay for forwarding

## Alternatives Considered

1. **Polling API** - Rejected: Higher latency
2. **Message queue export** - Rejected: Requires additional infrastructure
3. **Direct database access** - Rejected: Security risks
