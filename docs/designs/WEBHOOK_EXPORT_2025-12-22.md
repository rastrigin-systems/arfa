# Webhook Export - Technical Design

**Created**: 2025-12-22
**Status**: Draft
**Priority**: P0 - Next feature to implement

---

## Overview

Enable real-time forwarding of activity logs to customer-configured webhook endpoints (Kibana, Splunk, Slack, etc.) while retaining all data in Ubik's database.

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ activity_   │────▶│ Event Forwarder │────▶│ Customer SIEM   │
│ logs table  │     │ (background)    │     │ (webhook URL)   │
└─────────────┘     └─────────────────┘     └─────────────────┘
      │
      └── Data stays in our DB (source of truth)
```

---

## Database Schema

### New Tables

```sql
-- Webhook destination configuration
CREATE TABLE webhook_destinations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Destination config
    name VARCHAR(100) NOT NULL,           -- "production-siem", "slack-alerts"
    url TEXT NOT NULL,                    -- https://hooks.slack.com/...

    -- Authentication
    auth_type VARCHAR(50) DEFAULT 'none', -- 'none', 'bearer', 'header', 'basic'
    auth_config JSONB DEFAULT '{}',       -- {"token": "xxx"} or {"header": "X-Api-Key", "value": "xxx"}

    -- Event filtering
    event_types TEXT[] DEFAULT '{}',      -- Empty = all events, or ['tool_call', 'policy_violation']
    event_filter JSONB DEFAULT '{}',      -- {"blocked": true} = only blocked events

    -- Delivery settings
    enabled BOOLEAN DEFAULT true,
    batch_size INT DEFAULT 1,             -- 1 = real-time, >1 = batched
    timeout_ms INT DEFAULT 5000,          -- Request timeout
    retry_max INT DEFAULT 3,              -- Max retry attempts
    retry_backoff_ms INT DEFAULT 1000,    -- Initial backoff (doubles each retry)

    -- Metadata
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    created_by VARCHAR(255),              -- Admin who created

    -- Secret for signature verification
    signing_secret VARCHAR(255),          -- For X-Ubik-Signature header

    UNIQUE(org_id, name)
);

-- Track delivery status for each log entry
CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    destination_id UUID NOT NULL REFERENCES webhook_destinations(id) ON DELETE CASCADE,
    log_id UUID NOT NULL REFERENCES activity_logs(id) ON DELETE CASCADE,

    -- Delivery status
    status VARCHAR(50) NOT NULL,          -- 'pending', 'delivered', 'failed', 'dead'
    attempts INT DEFAULT 0,
    last_attempt_at TIMESTAMP,
    next_retry_at TIMESTAMP,

    -- Response info
    response_status INT,                  -- HTTP status code
    response_body TEXT,                   -- First 1000 chars of response
    error_message TEXT,                   -- Error if failed

    -- Timing
    created_at TIMESTAMP DEFAULT NOW(),
    delivered_at TIMESTAMP,

    UNIQUE(destination_id, log_id)
);

-- Indexes for efficient querying
CREATE INDEX idx_webhook_destinations_org ON webhook_destinations(org_id);
CREATE INDEX idx_webhook_destinations_enabled ON webhook_destinations(enabled) WHERE enabled = true;

CREATE INDEX idx_webhook_deliveries_pending ON webhook_deliveries(status, next_retry_at)
    WHERE status IN ('pending', 'failed');
CREATE INDEX idx_webhook_deliveries_destination ON webhook_deliveries(destination_id);
```

### Migration

```sql
-- migrations/YYYYMMDD_add_webhook_destinations.sql

BEGIN;

CREATE TABLE webhook_destinations (...);
CREATE TABLE webhook_deliveries (...);

-- Add RLS policies
ALTER TABLE webhook_destinations ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_deliveries ENABLE ROW LEVEL SECURITY;

CREATE POLICY webhook_destinations_org_isolation ON webhook_destinations
    USING (org_id = current_setting('app.current_org_id')::uuid);

COMMIT;
```

---

## API Design

### Endpoints

```yaml
# Add to platform/api-spec/spec.yaml

/admin/webhooks:
  get:
    summary: List webhook destinations
    operationId: listWebhookDestinations
    tags: [admin, webhooks]
    responses:
      200:
        content:
          application/json:
            schema:
              type: object
              properties:
                destinations:
                  type: array
                  items:
                    $ref: '#/components/schemas/WebhookDestination'

  post:
    summary: Create webhook destination
    operationId: createWebhookDestination
    tags: [admin, webhooks]
    requestBody:
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/CreateWebhookRequest'
    responses:
      201:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/WebhookDestination'

/admin/webhooks/{id}:
  get:
    summary: Get webhook destination
    operationId: getWebhookDestination

  patch:
    summary: Update webhook destination
    operationId: updateWebhookDestination

  delete:
    summary: Delete webhook destination
    operationId: deleteWebhookDestination

/admin/webhooks/{id}/test:
  post:
    summary: Send test event to webhook
    operationId: testWebhookDestination
    responses:
      200:
        content:
          application/json:
            schema:
              type: object
              properties:
                success: boolean
                response_status: integer
                response_time_ms: integer
                error: string

/admin/webhooks/{id}/deliveries:
  get:
    summary: List delivery attempts for destination
    operationId: listWebhookDeliveries
    parameters:
      - name: status
        in: query
        schema:
          type: string
          enum: [pending, delivered, failed, dead]
```

### Request/Response Schemas

```yaml
components:
  schemas:
    WebhookDestination:
      type: object
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        url:
          type: string
        auth_type:
          type: string
          enum: [none, bearer, header, basic]
        event_types:
          type: array
          items:
            type: string
        enabled:
          type: boolean
        batch_size:
          type: integer
        created_at:
          type: string
          format: date-time
        # Note: auth_config secrets are never returned

    CreateWebhookRequest:
      type: object
      required: [name, url]
      properties:
        name:
          type: string
          maxLength: 100
        url:
          type: string
          format: uri
        auth_type:
          type: string
          enum: [none, bearer, header, basic]
          default: none
        auth_config:
          type: object
          description: |
            For bearer: {"token": "xxx"}
            For header: {"header": "X-Api-Key", "value": "xxx"}
            For basic: {"username": "xxx", "password": "xxx"}
        event_types:
          type: array
          items:
            type: string
            enum: [tool_call, api_request, api_response, policy_violation]
        event_filter:
          type: object
          description: 'Filter events, e.g., {"blocked": true}'
        batch_size:
          type: integer
          default: 1
          minimum: 1
          maximum: 100
```

---

## Event Forwarder Service

### Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    API SERVER                                   │
│                                                                 │
│  ┌─────────────────┐    ┌─────────────────────────────────────┐ │
│  │ Log Handler     │───▶│ activity_logs table                 │ │
│  │ (existing)      │    └───────────────┬─────────────────────┘ │
│  └─────────────────┘                    │                       │
│                                         │ NOTIFY                │
│                                         ▼                       │
│  ┌─────────────────────────────────────────────────────────────┐│
│  │                  Event Forwarder                            ││
│  │                                                             ││
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐  ││
│  │  │ Listener    │  │ Dispatcher  │  │ Delivery Workers    │  ││
│  │  │             │  │             │  │                     │  ││
│  │  │ • PG NOTIFY │─▶│ • Match     │─▶│ • HTTP POST         │  ││
│  │  │ • Or poll   │  │   filters   │  │ • Retry logic       │  ││
│  │  │             │  │ • Queue     │  │ • Status update     │  ││
│  │  └─────────────┘  └─────────────┘  └─────────────────────┘  ││
│  │                                                             ││
│  └─────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

### Go Implementation

```go
// services/api/internal/webhooks/forwarder.go

package webhooks

import (
    "context"
    "time"
)

type ForwarderConfig struct {
    PollInterval    time.Duration // How often to check for new logs
    WorkerCount     int           // Concurrent delivery workers
    BatchSize       int           // Max logs to process per poll
    MaxRetries      int           // Max retries before marking dead
    RetryBackoff    time.Duration // Initial retry backoff
}

type Forwarder struct {
    config     ForwarderConfig
    db         *sql.DB
    httpClient *http.Client
    logger     *slog.Logger
}

func NewForwarder(config ForwarderConfig, db *sql.DB) *Forwarder {
    return &Forwarder{
        config: config,
        db:     db,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

// Start begins the forwarder background workers
func (f *Forwarder) Start(ctx context.Context) {
    // Worker pool for deliveries
    for i := 0; i < f.config.WorkerCount; i++ {
        go f.deliveryWorker(ctx, i)
    }

    // Main loop: poll for new logs
    go f.pollLoop(ctx)

    // Retry loop: retry failed deliveries
    go f.retryLoop(ctx)
}

func (f *Forwarder) pollLoop(ctx context.Context) {
    ticker := time.NewTicker(f.config.PollInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            f.processNewLogs(ctx)
        }
    }
}

func (f *Forwarder) processNewLogs(ctx context.Context) {
    // 1. Find enabled destinations
    destinations, _ := f.getEnabledDestinations(ctx)

    // 2. For each destination, find unprocessed logs
    for _, dest := range destinations {
        logs, _ := f.getUnprocessedLogs(ctx, dest)

        // 3. Create delivery records
        for _, log := range logs {
            if f.matchesFilter(log, dest) {
                f.queueDelivery(ctx, dest, log)
            }
        }
    }
}

func (f *Forwarder) deliveryWorker(ctx context.Context, id int) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            delivery, err := f.claimNextDelivery(ctx)
            if err != nil || delivery == nil {
                time.Sleep(100 * time.Millisecond)
                continue
            }

            f.processDelivery(ctx, delivery)
        }
    }
}

func (f *Forwarder) processDelivery(ctx context.Context, d *Delivery) {
    dest, _ := f.getDestination(ctx, d.DestinationID)
    log, _ := f.getLog(ctx, d.LogID)

    // Build payload
    payload := f.buildPayload(log, dest)

    // Make HTTP request
    req, _ := http.NewRequestWithContext(ctx, "POST", dest.URL, bytes.NewReader(payload))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Ubik-Signature", f.sign(payload, dest.SigningSecret))
    req.Header.Set("X-Ubik-Event", log.EventType)

    // Add auth
    f.addAuth(req, dest)

    // Send
    resp, err := f.httpClient.Do(req)

    // Update delivery status
    if err != nil || resp.StatusCode >= 400 {
        f.markFailed(ctx, d, err, resp)
    } else {
        f.markDelivered(ctx, d, resp)
    }
}
```

### Webhook Payload Format

```json
{
  "id": "evt_abc123",
  "type": "tool_call",
  "timestamp": "2025-12-22T17:30:00Z",
  "organization": {
    "id": "8b58e482-737e-4145-b0e8-69162a6b5db1",
    "name": "Acme Corp"
  },
  "user": {
    "email": "sarah@acme.com",
    "id": "ae848cb1-7c8a-41eb-b164-bd176dd934e4"
  },
  "session_id": "c704df8e-0126-4814-a07f-334de83c017f",
  "data": {
    "tool_name": "Bash",
    "tool_id": "toolu_01ABC123",
    "tool_input": {
      "command": "rm -rf /tmp/*"
    },
    "blocked": true,
    "block_reason": "Destructive commands blocked by policy"
  }
}
```

### Signature Verification

Customers can verify webhooks came from Ubik:

```python
# Customer's code to verify signature
import hmac
import hashlib

def verify_signature(payload: bytes, signature: str, secret: str) -> bool:
    expected = hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(f"sha256={expected}", signature)

# Usage
@app.route('/webhook', methods=['POST'])
def handle_webhook():
    signature = request.headers.get('X-Ubik-Signature')
    if not verify_signature(request.data, signature, UBIK_WEBHOOK_SECRET):
        return 'Invalid signature', 401

    event = request.json
    # Process event...
```

---

## CLI Commands (Primary Interface)

**Webhook management is CLI-only.** No web UI required. Users configure webhooks via CLI, which is:
- Faster to implement (no UI work)
- Scriptable (automation-friendly)
- Secure (secrets stay in terminal)
- Preferred by security teams

### Permission-Based Access Control

Commands check user role at runtime. No "admin" prefix needed:

```bash
# Same command, different permissions
$ ubik webhooks list          # Works for all roles (read-only)
$ ubik webhooks add ...       # Requires: admin, manager
$ ubik webhooks delete ...    # Requires: admin, manager
```

**How it works:**
1. CLI reads user role from `~/.ubik/config.json` (set at login)
2. Command checks required permission
3. Blocks with clear error if insufficient:

```bash
$ ubik webhooks add test https://example.com
Error: insufficient permissions

This command requires 'admin' or 'manager' role.
Your role: developer

Contact your organization admin to request access.
```

**Role hierarchy:**
| Role | webhooks list | webhooks add | webhooks delete |
|------|---------------|--------------|-----------------|
| admin | ✅ | ✅ | ✅ |
| manager | ✅ | ✅ | ✅ |
| developer | ✅ | ❌ | ❌ |

### Three Ways to Add Webhooks

| Method | Best For | Example |
|--------|----------|---------|
| **Interactive** | First-time users | `ubik webhooks add` |
| **File-based** | Automation/IaC | `ubik webhooks add -f webhook.yaml` |
| **Quick** | Simple webhooks | `ubik webhooks add <name> <url>` |

---

### Method 1: Interactive (Default)

Human-friendly prompts guide you through setup:

```bash
$ ubik webhooks add

? Webhook name: prod-siem
? Endpoint URL: https://siem.acme.com/api/events
? Authentication type:
  ❯ None (URL contains secret)
    Bearer Token
    Custom Header
    Basic Auth
? Bearer token: **********************
? Which events to send? (space to select, enter to confirm)
    ◉ tool_call
    ◉ policy_violation
    ○ api_request
    ○ api_response
? Only send blocked events? No

✓ Webhook 'prod-siem' created

  Signing secret: ubik_whsec_abc123def456...
  ⚠ Save this secret - it won't be shown again!

  Test: ubik webhooks test prod-siem
  Docs: https://docs.ubik.dev/webhooks/verify
```

---

### Method 2: File-Based (Recommended for Teams)

Define webhook in YAML, version control it, apply:

```bash
$ ubik webhooks add -f webhook.yaml
```

**webhook.yaml:**
```yaml
# Webhook destination configuration
name: prod-siem
url: https://siem.acme.com/api/events

# Authentication (choose one)
auth:
  type: bearer          # none | bearer | header | basic
  token: ${SIEM_TOKEN}  # Environment variable for secrets

# Which events to forward
events:
  - tool_call
  - policy_violation

# Optional: only forward matching events
filter:
  blocked: true         # Only blocked tool calls

# Optional: delivery settings
settings:
  batch_size: 1         # Events per request (1 = real-time)
  timeout: 5s
  retries: 3
```

**Environment variable for secrets:**
```bash
$ export SIEM_TOKEN="sk_live_xxx"
$ ubik webhooks add -f webhook.yaml

✓ Webhook 'prod-siem' created from webhook.yaml
```

**Example files for common integrations:**

```yaml
# datadog.yaml
name: datadog
url: https://http-intake.logs.datadoghq.com/api/v2/logs
auth:
  type: header
  name: DD-API-KEY
  value: ${DD_API_KEY}
events: [tool_call]
```

```yaml
# slack-alerts.yaml
name: slack-alerts
url: ${SLACK_WEBHOOK_URL}  # URL is the secret
auth:
  type: none
events: [tool_call]
filter:
  blocked: true
```

```yaml
# splunk.yaml
name: splunk-hec
url: https://splunk.acme.com:8088/services/collector
auth:
  type: bearer
  token: ${SPLUNK_HEC_TOKEN}
events: [tool_call, policy_violation]
```

---

### Method 3: Quick One-Liner

For simple webhooks, minimal syntax:

```bash
# Prompts only for auth
$ ubik webhooks add prod-siem https://siem.acme.com/api/events
? Authentication type: Bearer Token
? Token: ****

✓ Webhook 'prod-siem' created
```

---

### Other Commands

```bash
# List all webhooks
$ ubik webhooks list
┌──────────────┬────────────────────────────────┬─────────────────────┬──────────┐
│ NAME         │ URL                            │ EVENTS              │ STATUS   │
├──────────────┼────────────────────────────────┼─────────────────────┼──────────┤
│ prod-siem    │ https://siem.acme.com/...      │ tool_call, policy   │ ✓ active │
│ slack-alerts │ https://hooks.slack.com/...    │ tool_call (blocked) │ ✓ active │
│ datadog      │ https://http-intake.logs...    │ tool_call           │ ✗ paused │
└──────────────┴────────────────────────────────┴─────────────────────┴──────────┘

# Test webhook (sends sample event)
$ ubik webhooks test prod-siem
Sending test event...
✓ Success! 200 OK (142ms)

# View webhook details
$ ubik webhooks get prod-siem
Name:     prod-siem
URL:      https://siem.acme.com/api/events
Auth:     bearer (token: sk_...xxx)
Events:   tool_call, policy_violation
Status:   active
Created:  2025-12-22 10:30:00

# Export as YAML (for backup/copying)
$ ubik webhooks get prod-siem -o yaml > prod-siem.yaml

# Update from file
$ ubik webhooks update -f webhook.yaml

# Pause/resume
$ ubik webhooks pause prod-siem
$ ubik webhooks resume prod-siem

# Delete
$ ubik webhooks delete prod-siem
? Delete webhook 'prod-siem'? This cannot be undone. (y/N) y
✓ Deleted

# View delivery history
$ ubik webhooks deliveries prod-siem
┌─────────────┬───────────────────────┬──────────┬─────────┐
│ EVENT ID    │ TIME                  │ STATUS   │ LATENCY │
├─────────────┼───────────────────────┼──────────┼─────────┤
│ evt_abc123  │ 2025-12-22 10:31:05   │ ✓ sent   │ 145ms   │
│ evt_def456  │ 2025-12-22 10:31:02   │ ✓ sent   │ 132ms   │
│ evt_ghi789  │ 2025-12-22 10:30:58   │ ✗ failed │ timeout │
└─────────────┴───────────────────────┴──────────┴─────────┘

# View failed deliveries only
$ ubik webhooks deliveries prod-siem --status failed

# Rotate signing secret
$ ubik webhooks rotate-secret prod-siem
? Rotate signing secret? Existing integrations will break. (y/N) y
✓ New signing secret: ubik_whsec_newkey123...
```

### CLI Implementation

```go
// services/cli/internal/commands/webhooks/webhooks.go

var webhooksCmd = &cobra.Command{
    Use:   "webhooks",
    Short: "Manage webhook destinations",
}

// Permission check helper
func requireRole(c *container.Container, allowedRoles ...string) error {
    config, err := c.ConfigManager().Load()
    if err != nil {
        return err
    }
    for _, role := range allowedRoles {
        if config.Role == role {
            return nil
        }
    }
    return fmt.Errorf("insufficient permissions\n\nThis command requires '%s' role.\nYour role: %s\n\nContact your organization admin to request access.",
        strings.Join(allowedRoles, "' or '"), config.Role)
}

var webhooksListCmd = &cobra.Command{
    Use:   "list",
    Short: "List webhook destinations",
    RunE: func(cmd *cobra.Command, args []string) error {
        client := api.NewClient(cfg.PlatformURL, cfg.Token)
        destinations, err := client.ListWebhookDestinations(cmd.Context())
        if err != nil {
            return err
        }

        renderWebhooksTable(destinations)
        return nil
    },
}

var webhooksAddCmd = &cobra.Command{
    Use:   "add",
    Short: "Add webhook destination",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Permission check - only admin/manager can add webhooks
        if err := requireRole(c, "admin", "manager"); err != nil {
            return err
        }

        name, _ := cmd.Flags().GetString("name")
        url, _ := cmd.Flags().GetString("url")
        authType, _ := cmd.Flags().GetString("auth")
        token, _ := cmd.Flags().GetString("token")
        events, _ := cmd.Flags().GetStringSlice("events")

        client := api.NewClient(cfg.PlatformURL, cfg.Token)
        dest, err := client.CreateWebhookDestination(cmd.Context(), api.CreateWebhookRequest{
            Name:       name,
            URL:        url,
            AuthType:   authType,
            AuthConfig: map[string]string{"token": token},
            EventTypes: events,
        })
        if err != nil {
            return err
        }

        fmt.Printf("✓ Webhook destination '%s' created\n", dest.Name)
        fmt.Printf("  ID: %s\n", dest.ID)
        fmt.Printf("  Test with: ubik webhooks test %s\n", name)
        return nil
    },
}

var webhooksTestCmd = &cobra.Command{
    Use:   "test [name]",
    Short: "Send test event to webhook",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        name := args[0]

        client := api.NewClient(cfg.PlatformURL, cfg.Token)
        result, err := client.TestWebhookDestination(cmd.Context(), name)
        if err != nil {
            return err
        }

        if result.Success {
            fmt.Printf("✓ Success! Response: %d (%dms)\n",
                result.ResponseStatus, result.ResponseTimeMs)
        } else {
            fmt.Printf("✗ Failed: %s\n", result.Error)
        }
        return nil
    },
}
```

---

## Implementation Plan

### Phase 1: Database & API (Day 1-2)

- [ ] Add `webhook_destinations` table
- [ ] Add `webhook_deliveries` table
- [ ] Create SQLC queries
- [ ] Implement API endpoints (CRUD)
- [ ] Add tests

### Phase 2: Forwarder Service (Day 2-3)

- [ ] Implement `Forwarder` struct
- [ ] Poll loop for new logs
- [ ] Delivery workers
- [ ] Retry logic with backoff
- [ ] Signature generation
- [ ] Integration tests

### Phase 3: CLI Commands (Day 3-4)

- [ ] Permission check helper (`requireRole`)
- [ ] `ubik webhooks list` (all roles)
- [ ] `ubik webhooks add` (admin/manager only)
- [ ] `ubik webhooks test` (admin/manager only)
- [ ] `ubik webhooks delete` (admin/manager only)
- [ ] `ubik webhooks deliveries` (all roles)

### Phase 4: Testing & Polish (Day 4-5)

- [ ] End-to-end test with real webhook
- [ ] Kibana dashboard template
- [ ] Documentation
- [ ] Error handling improvements

---

## Testing Strategy

### Unit Tests

```go
func TestForwarder_MatchesFilter(t *testing.T) {
    // Test event_types filter
    // Test event_filter JSON matching
}

func TestForwarder_BuildPayload(t *testing.T) {
    // Test payload format
    // Test signature generation
}

func TestForwarder_RetryLogic(t *testing.T) {
    // Test exponential backoff
    // Test max retries
}
```

### Integration Tests

```go
func TestWebhookDelivery_EndToEnd(t *testing.T) {
    // 1. Start test HTTP server
    server := httptest.NewServer(...)

    // 2. Create webhook destination pointing to test server
    // 3. Create activity log
    // 4. Wait for delivery
    // 5. Verify test server received event
}
```

### Manual Testing

```bash
# 1. Start local webhook receiver
npx webhook-tester

# 2. Add destination
ubik webhooks add --name test --url http://localhost:9000/webhook

# 3. Generate activity (run ubik, trigger tool call)
ubik
> "list files"

# 4. Verify webhook received
```

---

## Security Considerations

| Risk | Mitigation |
|------|------------|
| Webhook URL injection | Validate URL format, block internal IPs |
| Secret exposure | Never return secrets in API responses |
| SSRF attacks | Allowlist external URLs only |
| Replay attacks | Include timestamp, customers verify freshness |
| Man-in-middle | HTTPS only, verify certificates |

### Enterprise Security Requirements

**What enterprises expect from webhook senders:**

#### 1. Request Headers (Required)

```http
POST /webhook HTTP/1.1
Content-Type: application/json
User-Agent: Ubik-Webhook/1.0

# Security
X-Ubik-Signature: sha256=HMAC(payload, secret)
X-Ubik-Timestamp: 1703258400
X-Ubik-Delivery-ID: del_abc123

# Event metadata
X-Ubik-Event-Type: tool_call
X-Ubik-Event-ID: evt_xyz789
```

#### 2. Signature Verification

Customers verify webhooks are from Ubik:

```python
import hmac, hashlib, time

def verify_webhook(payload: bytes, headers: dict, secret: str) -> bool:
    # Check timestamp (prevent replay attacks)
    timestamp = int(headers.get('X-Ubik-Timestamp', 0))
    if abs(time.time() - timestamp) > 300:  # 5 minute tolerance
        return False

    # Verify signature
    expected = hmac.new(
        secret.encode(),
        f"{timestamp}.".encode() + payload,  # Include timestamp in signature
        hashlib.sha256
    ).hexdigest()

    signature = headers.get('X-Ubik-Signature', '')
    return hmac.compare_digest(f"sha256={expected}", signature)
```

#### 3. Static Egress IPs

For enterprise IP allowlisting, we provide stable source IPs:

```
# Production webhook source IPs (Cloud NAT)
34.102.xxx.xxx/32
35.201.xxx.xxx/32

# Document these in admin console
```

#### 4. Rate Limit Handling

```go
// Respect customer rate limits
if resp.StatusCode == 429 {
    retryAfter := resp.Header.Get("Retry-After")
    if retryAfter != "" {
        delay, _ := strconv.Atoi(retryAfter)
        time.Sleep(time.Duration(delay) * time.Second)
    } else {
        time.Sleep(f.config.RetryBackoff)
    }
}
```

#### 5. Payload Envelope

```json
{
  "version": "1.0",
  "id": "evt_unique_id",
  "delivery_id": "del_attempt_id",
  "timestamp": "2025-12-22T17:30:00Z",
  "type": "tool_call",
  "source": "ubik",
  "organization": {
    "id": "org_xxx",
    "name": "Acme Corp"
  },
  "data": {
    // Event-specific payload
  }
}
```

### Common Integration Targets

| Target | Auth Method | Notes |
|--------|-------------|-------|
| **Logstash** | HMAC signature | Verify in filter |
| **AWS EventBridge** | IAM signature | Use SigV4 |
| **Datadog** | API key header | DD-API-KEY |
| **Splunk HEC** | Bearer token | Splunk format |
| **Slack** | None (URL is secret) | Specific format |
| **PagerDuty** | Routing key | Events API v2 |

---

## Monitoring & Observability

### Metrics to Track

```
ubik_webhook_deliveries_total{status="delivered|failed|dead"}
ubik_webhook_delivery_duration_ms
ubik_webhook_retry_count
ubik_webhook_queue_depth
```

### Alerts

- Delivery success rate < 95%
- Queue depth > 1000
- Destination failing for > 1 hour

---

## Open Questions

1. **Batching**: Should we support batch delivery (multiple events per request)?
2. **Ordering**: Guarantee event ordering or best-effort?
3. **Dead letter**: How long to keep failed deliveries before purging?
4. **Rate limiting**: Limit events per second to customer endpoints?

---

*Document created: 2025-12-22*
