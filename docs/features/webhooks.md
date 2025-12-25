# Webhook Export

Forward activity logs in real-time to external systems like Splunk, Datadog, Kibana, or Slack.

---

## Overview

Webhook export allows you to send activity logs to external endpoints as they happen. Use cases:

- **SIEM Integration** - Forward security events to Splunk, Elastic, or Datadog
- **Alerting** - Send blocked tool calls to Slack or PagerDuty
- **Compliance** - Archive events to your own data warehouse
- **Custom Processing** - Build custom analytics on your infrastructure

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ Activity    │────▶│ Event Forwarder │────▶│ Your Endpoint   │
│ Logs        │     │ (background)    │     │ (webhook URL)   │
└─────────────┘     └─────────────────┘     └─────────────────┘
```

**Key features:**
- Real-time delivery (typically < 10 seconds)
- HMAC-SHA256 signature verification
- Automatic retries with exponential backoff
- Event type filtering
- Multiple authentication methods

---

## Quick Start

### 1. Create a Webhook

```bash
# Login first
arfa login

# Create a webhook destination
arfa webhooks create \
  --name "my-siem" \
  --url "https://siem.example.com/api/events" \
  --auth-type bearer \
  --bearer-token "sk_live_xxx"
```

### 2. Test the Webhook

```bash
arfa webhooks test my-siem
```

Output:
```
Testing webhook 'my-siem'...

Test successful!
  Response Status: 200
  Response Time:   142ms
```

### 3. Verify Events Are Flowing

Generate some activity (run CLI commands), then check deliveries:

```bash
arfa webhooks list
```

---

## CLI Commands

### List Webhooks

```bash
arfa webhooks list
```

Output:
```
WEBHOOKS
ID                                    NAME        URL                              STATUS
6c83ef27-5296-459c-809c-424a3f4d148d  prod-siem   https://siem.example.com/events  enabled
a1b2c3d4-e5f6-7890-abcd-ef1234567890  slack       https://hooks.slack.com/...      enabled
```

### Create Webhook

```bash
arfa webhooks create \
  --name "webhook-name" \
  --url "https://your-endpoint.com/webhook" \
  --auth-type bearer \
  --bearer-token "your-token"
```

**Options:**

| Flag | Description | Default |
|------|-------------|---------|
| `--name` | Unique name for the webhook (required) | - |
| `--url` | Destination URL (required) | - |
| `--auth-type` | Authentication: `none`, `bearer`, `header`, `basic` | `none` |
| `--bearer-token` | Bearer token (when auth-type=bearer) | - |
| `--event-types` | Filter events: `tool_call,permission_denied` | all |
| `--json` | Output as JSON | false |

### Test Webhook

Send a test event to verify the endpoint is working:

```bash
arfa webhooks test <webhook-id>
```

### Delete Webhook

```bash
arfa webhooks delete <webhook-id>
```

Add `--force` to skip confirmation:

```bash
arfa webhooks delete <webhook-id> --force
```

---

## Webhook Payload Format

Each webhook delivery contains a JSON payload:

```json
{
  "id": "37e7c9c3-e286-4e27-8abc-4999fe50a6a5",
  "event_type": "tool_call",
  "event_category": "agent_activity",
  "timestamp": "2025-12-23T16:02:19.890Z",
  "org_id": "e5d10009-0988-44b6-b313-67ffbbbb1ef8",
  "employee_id": "6a41beee-cf2d-4c59-affd-80e3f58466d6",
  "session_id": "c704df8e-0126-4814-a07f-334de83c017f",
  "content": "Called Bash tool",
  "payload": {
    "tool_name": "Bash",
    "command": "ls -la"
  }
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | UUID | Unique log entry ID |
| `event_type` | string | Event type (see below) |
| `event_category` | string | Category grouping |
| `timestamp` | ISO8601 | When the event occurred |
| `org_id` | UUID | Organization ID |
| `employee_id` | UUID | Employee who triggered the event (optional) |
| `session_id` | UUID | CLI session ID (optional) |
| `agent_id` | UUID | Agent configuration ID (optional) |
| `content` | string | Human-readable description |
| `payload` | object | Event-specific data |

### Event Types

| Event Type | Category | Description |
|------------|----------|-------------|
| `tool_call` | agent_activity | Tool was invoked |
| `permission_denied` | security | Tool call was blocked by policy |
| `api_request` | agent_activity | API request made |
| `session_start` | auth | CLI session started |
| `session_end` | auth | CLI session ended |

---

## HTTP Headers

Each webhook request includes these headers:

```http
POST /webhook HTTP/1.1
Content-Type: application/json
User-Agent: Arfa-Webhook/1.0
X-Arfa-Event-Type: tool_call
X-Arfa-Delivery-ID: 3623d0e4-5ed6-4a8b-a5ed-915845df446d
X-Arfa-Signature: sha256=a1b2c3d4e5f6...
```

| Header | Description |
|--------|-------------|
| `X-Arfa-Event-Type` | The event type being delivered |
| `X-Arfa-Delivery-ID` | Unique delivery attempt ID |
| `X-Arfa-Signature` | HMAC-SHA256 signature for verification |

---

## Signature Verification

Verify webhooks came from Arfa using the signing secret:

### Python

```python
import hmac
import hashlib

def verify_signature(payload: bytes, signature: str, secret: str) -> bool:
    expected = hmac.new(
        secret.encode(),
        payload,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(f"sha256={expected}", signature)

# Usage in Flask
@app.route('/webhook', methods=['POST'])
def handle_webhook():
    signature = request.headers.get('X-Arfa-Signature')
    if not verify_signature(request.data, signature, ARFA_WEBHOOK_SECRET):
        return 'Invalid signature', 401

    event = request.json
    print(f"Received event: {event['event_type']}")
    return 'OK', 200
```

### Node.js

```javascript
const crypto = require('crypto');

function verifySignature(payload, signature, secret) {
    const expected = 'sha256=' + crypto
        .createHmac('sha256', secret)
        .update(payload)
        .digest('hex');
    return crypto.timingSafeEqual(
        Buffer.from(signature),
        Buffer.from(expected)
    );
}

// Usage in Express
app.post('/webhook', (req, res) => {
    const signature = req.headers['x-arfa-signature'];
    if (!verifySignature(req.rawBody, signature, ARFA_WEBHOOK_SECRET)) {
        return res.status(401).send('Invalid signature');
    }

    const event = req.body;
    console.log(`Received event: ${event.event_type}`);
    res.send('OK');
});
```

### Go

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func verifySignature(payload []byte, signature, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expected))
}
```

---

## Authentication Methods

### None (URL contains secret)

For services like Slack where the URL itself is the secret:

```bash
arfa webhooks create \
  --name "slack" \
  --url "https://hooks.slack.com/services/T00/B00/xxxx" \
  --auth-type none
```

### Bearer Token

Standard Authorization header with Bearer token:

```bash
arfa webhooks create \
  --name "splunk" \
  --url "https://splunk.example.com:8088/services/collector" \
  --auth-type bearer \
  --bearer-token "your-hec-token"
```

Sends: `Authorization: Bearer your-hec-token`

### Custom Header

For services requiring a specific header name:

```yaml
# Coming soon: --header-name and --header-value flags
```

### Basic Auth

HTTP Basic authentication:

```yaml
# Coming soon: --username and --password flags
```

---

## Integration Examples

### Splunk HEC

```bash
arfa webhooks create \
  --name "splunk-hec" \
  --url "https://splunk.example.com:8088/services/collector/event" \
  --auth-type bearer \
  --bearer-token "YOUR_HEC_TOKEN" \
  --event-types "tool_call,permission_denied"
```

### Datadog Logs

```bash
arfa webhooks create \
  --name "datadog" \
  --url "https://http-intake.logs.datadoghq.com/api/v2/logs" \
  --auth-type bearer \
  --bearer-token "YOUR_DD_API_KEY"
```

### Slack (Blocked Events Only)

```bash
arfa webhooks create \
  --name "slack-alerts" \
  --url "https://hooks.slack.com/services/T00/B00/xxxx" \
  --auth-type none \
  --event-types "permission_denied"
```

### Elasticsearch

```bash
arfa webhooks create \
  --name "elastic" \
  --url "https://elastic.example.com:9200/arfa-logs/_doc" \
  --auth-type bearer \
  --bearer-token "YOUR_API_KEY"
```

---

## Retry Behavior

Webhooks automatically retry on failure:

| Attempt | Delay | Total Time |
|---------|-------|------------|
| 1 | Immediate | 0s |
| 2 | 1s | 1s |
| 3 | 2s | 3s |
| 4 (max) | 4s | 7s |

After max retries, the delivery is marked as `dead` and won't be retried.

**Retry conditions:**
- Network errors
- Timeout (default: 5 seconds)
- HTTP 5xx responses
- HTTP 429 (rate limited)

**No retry:**
- HTTP 2xx (success)
- HTTP 4xx (client error, except 429)

---

## Troubleshooting

### Webhook Not Receiving Events

1. **Check webhook is enabled:**
   ```bash
   arfa webhooks list
   ```
   Status should show "enabled"

2. **Test the endpoint:**
   ```bash
   arfa webhooks test <webhook-id>
   ```

3. **Check API server logs** for forwarder errors

4. **Verify URL is accessible** from the server

### Invalid Signature Errors

1. **Check you're using the correct signing secret**
2. **Verify you're using the raw request body** (not parsed JSON)
3. **Ensure no middleware is modifying the payload**

### Events Delayed

The forwarder processes events every 10 seconds. Events should arrive within 10-20 seconds of being generated.

### Getting the Signing Secret

The signing secret is generated when you create a webhook. If you lost it, you'll need to delete and recreate the webhook, or contact support for secret rotation.

---

## API Reference

Webhooks can also be managed via the REST API:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/webhooks` | List all webhooks |
| POST | `/api/v1/webhooks` | Create webhook |
| GET | `/api/v1/webhooks/{id}` | Get webhook details |
| PATCH | `/api/v1/webhooks/{id}` | Update webhook |
| DELETE | `/api/v1/webhooks/{id}` | Delete webhook |
| POST | `/api/v1/webhooks/{id}/test` | Test webhook |
| GET | `/api/v1/webhooks/{id}/deliveries` | List delivery history |

See the [OpenAPI spec](../platform/api-spec/spec.yaml) for full details.

---

## Best Practices

1. **Always verify signatures** - Don't trust webhook payloads without verification
2. **Respond quickly** - Return 200 OK within 5 seconds; process asynchronously if needed
3. **Handle duplicates** - Use `X-Arfa-Delivery-ID` to deduplicate
4. **Filter events** - Only subscribe to events you need to reduce noise
5. **Monitor deliveries** - Check for failed deliveries regularly
6. **Rotate secrets** - Periodically rotate signing secrets for security

---

## Limits

| Limit | Value |
|-------|-------|
| Webhooks per organization | 10 |
| Payload size | 1 MB max |
| Request timeout | 30 seconds |
| Retry attempts | 3 |
| Delivery history retention | 7 days |
