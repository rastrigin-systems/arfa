# Real-Time Policy Updates Design

## Overview

This document describes the architecture for delivering policy updates to proxies in real-time, eliminating the need for proxy restarts when policies change.

## Requirements

### Functional Requirements

| ID | Requirement | Details |
|----|-------------|---------|
| FR1 | Real-time delivery | Policy updates delivered to affected proxies in < 1 second |
| FR2 | Scoped delivery | Proxies only receive policies applicable to their authenticated employee |
| FR3 | Initial sync | Proxy receives all applicable policies on connect before processing requests |
| FR4 | Fail-closed | Block all requests after 5 minutes of disconnection from API |
| FR5 | Immediate revocation | When employee is deactivated, their proxy immediately blocks all requests |

### Non-Functional Requirements

| ID | Requirement | Details |
|----|-------------|---------|
| NFR1 | Scalability | Support 1000+ concurrent proxy connections per organization |
| NFR2 | Availability | In-memory policies only; no disk cache |
| NFR3 | Security | JWT authentication, TLS encryption, tenant isolation |
| NFR4 | Performance | Minimal CPU/memory overhead on proxy |

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────────┐
│                         API Server                               │
│                                                                  │
│  ┌─────────────────┐         ┌─────────────────────────────────┐│
│  │  Policy CRUD    │         │  WebSocket Hub                  ││
│  │  Handlers       │         │  - Connection registry          ││
│  │                 │         │  - Indexed by org/team/employee ││
│  └────────┬────────┘         └──────────────┬──────────────────┘│
│           │                                  │                   │
│           ▼                                  │                   │
│  ┌─────────────────┐                         │                   │
│  │  PostgreSQL     │                         │                   │
│  │  NOTIFY         │─────────────────────────┘                   │
│  │  'policy_change'│      (Hub listens, broadcasts to proxies)  │
│  └─────────────────┘                                             │
└─────────────────────────────────────────────────────────────────┘
                                      │
        ┌─────────────────────────────┼─────────────────────────────┐
        ▼                             ▼                             ▼
   ┌─────────┐                   ┌─────────┐                   ┌─────────┐
   │ Proxy 1 │                   │ Proxy 2 │                   │ Proxy N │
   │ (emp A) │                   │ (emp B) │                   │ (emp C) │
   └─────────┘                   └─────────┘                   └─────────┘
```

### Components

#### 1. WebSocket Hub (API Server)

Manages all proxy connections and broadcasts policy updates.

```go
type PolicyHub struct {
    // All connections indexed by connection ID
    connections map[string]*PolicyConn

    // Connections indexed by scope for efficient broadcast
    byEmployee  map[uuid.UUID][]*PolicyConn  // employee_id -> connections
    byTeam      map[uuid.UUID][]*PolicyConn  // team_id -> connections
    byOrg       map[uuid.UUID][]*PolicyConn  // org_id -> connections

    mu sync.RWMutex
}

type PolicyConn struct {
    ID          string
    Conn        *websocket.Conn
    EmployeeID  uuid.UUID
    OrgID       uuid.UUID
    TeamID      *uuid.UUID  // nullable
    ConnectedAt time.Time
}
```

#### 2. PostgreSQL NOTIFY/LISTEN

Used for decoupling policy CRUD handlers from the WebSocket hub.

**Channel:** `policy_change`

**Payload format:**
```json
{
  "action": "create|update|delete|revoke",
  "policy": { /* policy object for create/update/delete */ },
  "employee_id": "uuid",  // for revoke action
  "org_id": "uuid",
  "team_id": "uuid"       // nullable
}
```

#### 3. Proxy Policy Client (CLI)

Manages WebSocket connection and in-memory policy storage.

```go
type PolicyClient struct {
    conn        *websocket.Conn
    policies    []Policy
    state       ProxyState
    lastContact time.Time
    mu          sync.RWMutex
}

type ProxyState string

const (
    StateConnecting   ProxyState = "connecting"   // Waiting for initial sync
    StateReady        ProxyState = "ready"        // Normal operation
    StateDisconnected ProxyState = "disconnected" // Lost connection
    StateRevoked      ProxyState = "revoked"      // Access revoked
)
```

## Flows

### 1. Proxy Startup & Initial Sync

```
Proxy                          API Server                    Database
  │                                │                            │
  │──── WS Connect + JWT ─────────▶│                            │
  │                                │─── Validate JWT ──────────▶│
  │                                │◀── Employee + Team info ───│
  │                                │                            │
  │                                │─── Get policies for ──────▶│
  │                                │    employee/team/org       │
  │                                │◀── Policies ───────────────│
  │                                │                            │
  │◀─── Init message ──────────────│                            │
  │     {type: "init",             │                            │
  │      policies: [...]}          │                            │
  │                                │                            │
  │     [State: ready]             │                            │
  │     Start processing requests  │                            │
```

### 2. Policy Create/Update/Delete

```
Admin UI                 API Server                    Proxies
  │                          │                            │
  │── POST /policies ───────▶│                            │
  │                          │─── Insert to DB            │
  │                          │─── NOTIFY policy_change    │
  │                          │                            │
  │◀── 201 Created ──────────│                            │
  │                          │                            │
  │                          │    [Hub receives NOTIFY]   │
  │                          │    [Determine scope]       │
  │                          │    [Find affected conns]   │
  │                          │                            │
  │                          │─── WS: policy update ─────▶│ (affected only)
  │                          │                            │
  │                          │                            │ [Update in-memory]
```

### 3. Employee Revocation

```
Admin UI                 API Server                    Proxy
  │                          │                            │
  │── DELETE /employees/X ──▶│                            │
  │   or PATCH status=       │                            │
  │   inactive               │                            │
  │                          │─── Update DB               │
  │                          │─── NOTIFY policy_change    │
  │                          │    {action: "revoke",      │
  │                          │     employee_id: X}        │
  │◀── 200 OK ───────────────│                            │
  │                          │                            │
  │                          │    [Hub receives NOTIFY]   │
  │                          │    [Find employee's conn]  │
  │                          │                            │
  │                          │─── WS: revoke message ────▶│
  │                          │─── Close WebSocket ───────▶│
  │                          │                            │
  │                          │                            │ [State: revoked]
  │                          │                            │ [Block ALL requests]
```

### 4. Disconnect & Reconnect

```
Proxy                          API Server
  │                                │
  │     [Connection lost]          │
  │     [State: disconnected]      │
  │     [Start 5-min timer]        │
  │                                │
  │     ... continue with          │
  │     cached policies ...        │
  │                                │
  │──── WS Reconnect + JWT ───────▶│
  │                                │
  │◀─── Init message ──────────────│
  │                                │
  │     [State: ready]             │
  │     [Cancel timer]             │
```

### 5. Disconnect Timeout (Fail-Closed)

```
Proxy
  │
  │     [Connection lost]
  │     [State: disconnected]
  │     [Start 5-min timer]
  │
  │     ... 5 minutes pass ...
  │     ... reconnect attempts fail ...
  │
  │     [Timer expires]
  │     [State: disconnected, but past grace period]
  │     [Block ALL requests]
  │
  │     Request comes in:
  │     → "Connection to policy server lost. All requests blocked."
```

## Message Protocol

### Server → Proxy Messages

```typescript
// Initial sync - sent on connect
{
  "type": "init",
  "policies": [
    {
      "id": "uuid",
      "tool_name": "Bash",
      "action": "deny",
      "reason": "Shell commands blocked",
      "conditions": {"patterns": ["rm -rf"]},
      "scope": "organization"
    }
  ],
  "version": 12345
}

// Policy created or updated
{
  "type": "upsert",
  "policy": {
    "id": "uuid",
    "tool_name": "Write",
    "action": "audit",
    ...
  }
}

// Policy deleted
{
  "type": "delete",
  "policy_id": "uuid"
}

// Access revoked
{
  "type": "revoke",
  "reason": "Employee account deactivated"
}

// Heartbeat (keep-alive)
{
  "type": "ping"
}
```

### Proxy → Server Messages

```typescript
// Heartbeat response
{
  "type": "pong"
}
```

## Proxy State Machine

```
                    ┌──────────────┐
                    │  connecting  │
        ┌───────────┤  (block all) │
        │           └──────┬───────┘
        │                  │ init received
        │                  ▼
        │           ┌──────────────┐
        │      ┌───▶│    ready     │◀───┐
        │      │    │ (enforce)    │    │
        │      │    └──────┬───────┘    │
        │      │           │            │
        │   reconnect   disconnect   reconnect
        │   success        │         success
        │      │           ▼            │
        │      │    ┌──────────────┐    │
        │      └────│ disconnected │────┘
        │           │ (cached, 5m) │
        │           └──────┬───────┘
        │                  │ 5 min timeout
        │                  │ OR revoke message
        │                  ▼
        │           ┌──────────────┐
        └──────────▶│   revoked    │
         revoke msg │ (block all)  │
                    └──────────────┘
```

## Database Changes

### New: policy_notifications trigger

```sql
-- Trigger function to notify on policy changes
CREATE OR REPLACE FUNCTION notify_policy_change()
RETURNS TRIGGER AS $$
DECLARE
  payload JSON;
BEGIN
  IF TG_OP = 'DELETE' THEN
    payload := json_build_object(
      'action', 'delete',
      'policy_id', OLD.id,
      'org_id', OLD.org_id,
      'team_id', OLD.team_id,
      'employee_id', OLD.employee_id
    );
  ELSE
    payload := json_build_object(
      'action', LOWER(TG_OP),
      'policy', row_to_json(NEW),
      'org_id', NEW.org_id,
      'team_id', NEW.team_id,
      'employee_id', NEW.employee_id
    );
  END IF;

  PERFORM pg_notify('policy_change', payload::text);
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach trigger to tool_policies table
CREATE TRIGGER policy_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON tool_policies
FOR EACH ROW EXECUTE FUNCTION notify_policy_change();
```

### Employee deactivation trigger

```sql
-- Trigger to notify when employee is deactivated
CREATE OR REPLACE FUNCTION notify_employee_revoke()
RETURNS TRIGGER AS $$
BEGIN
  IF NEW.status != 'active' AND OLD.status = 'active' THEN
    PERFORM pg_notify('policy_change', json_build_object(
      'action', 'revoke',
      'employee_id', NEW.id,
      'org_id', NEW.org_id
    )::text);
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER employee_revoke_trigger
AFTER UPDATE ON employees
FOR EACH ROW EXECUTE FUNCTION notify_employee_revoke();
```

## API Changes

### New WebSocket Endpoint

```
GET /ws/policies
Authorization: Bearer <jwt>
Upgrade: websocket
```

**Authentication:** JWT token validated on connect. Connection rejected if:
- Token invalid/expired
- Employee not found
- Employee status != active

### Modified Policy Handlers

Policy CRUD handlers don't need changes - the PostgreSQL trigger handles notification automatically.

## CLI Changes

### New: PolicyClient

Replace file-based `PolicyHandler` with WebSocket-based `PolicyClient`:

```go
// services/cli/internal/control/policy_client.go

type PolicyClient struct {
    apiURL      string
    token       string
    conn        *websocket.Conn
    policies    []api.ToolPolicy
    state       ProxyState
    lastContact time.Time
    mu          sync.RWMutex

    // Callbacks
    onStateChange func(ProxyState)
}

func (c *PolicyClient) Connect(ctx context.Context) error
func (c *PolicyClient) IsBlocked(toolName string, input map[string]any) (string, bool)
func (c *PolicyClient) ShouldBlockAll() bool
func (c *PolicyClient) Close() error
```

### Proxy Integration

```go
// services/cli/internal/control/service.go

func NewService(config ServiceConfig) (*Service, error) {
    // Create policy client instead of policy handler
    policyClient := NewPolicyClient(config.APIURL, config.Token)

    // Connect with retry
    go policyClient.ConnectWithRetry(ctx)

    // ...
}
```

## Security Considerations

1. **Authentication**: Every WebSocket connection authenticated with JWT
2. **Authorization**: Policies filtered by employee's org/team membership
3. **Tenant isolation**: Connection registry indexed by org prevents cross-tenant access
4. **TLS**: All WebSocket connections over wss://
5. **Revocation**: Immediate disconnect on employee deactivation

## Monitoring & Observability

### Metrics to Track

- `policy_ws_connections_total` - Total active WebSocket connections
- `policy_ws_connections_by_org` - Connections per organization
- `policy_broadcast_latency_ms` - Time from NOTIFY to delivery
- `policy_broadcast_count` - Number of policies broadcast
- `proxy_state` - Current state per proxy (for debugging)

### Logging

- Connection established/closed (with employee_id, org_id)
- Policy broadcast (action, scope, affected count)
- Revocation events
- Reconnection attempts

## Implementation Plan

### Phase 1: API Server (WebSocket Hub)
1. Create `PolicyHub` with connection registry
2. Add `/ws/policies` endpoint
3. Implement PostgreSQL LISTEN
4. Add broadcast logic

### Phase 2: Database Triggers
1. Add `notify_policy_change()` trigger
2. Add `notify_employee_revoke()` trigger
3. Test notifications

### Phase 3: CLI Policy Client
1. Create `PolicyClient` with WebSocket connection
2. Implement state machine
3. Add reconnection logic
4. Integrate with proxy

### Phase 4: Testing & Hardening
1. Unit tests for all components
2. Integration tests for full flow
3. Load testing (1000 connections)
4. Failure mode testing

## Future Considerations

- **Multi-instance API**: Use Redis pub/sub if PostgreSQL NOTIFY doesn't scale
- **Policy versioning**: Track policy version to detect missed updates
- **Batch updates**: Debounce rapid policy changes
- **Compression**: Compress large policy payloads
