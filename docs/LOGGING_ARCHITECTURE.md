# Logging Architecture

This document describes how the CLI captures and sends logs to the Ubik platform.

## Flow Diagram

```mermaid
flowchart TB
    subgraph CLI["CLI Process"]
        subgraph Interactive["Interactive Command"]
            IC[Start Interactive Session]
            IC --> InitLogger[Initialize Logger]
            IC --> InitProxy[Initialize Proxy]
        end

        subgraph Logger["Logger (logging/)"]
            L[Logger Instance]
            Buffer[Event Buffer]
            BatchTimer[Batch Timer<br/>5 seconds]
            RetryQueue[Retry Queue<br/>Exponential Backoff]
            DiskQueue[Disk Queue<br/>~/.ubik/log_queue/]
        end

        subgraph Proxy["In-Process Proxy (proxy/)"]
            P[MITM Proxy<br/>goproxy]
            CA[CA Certificate<br/>~/.ubik/certs/]
            Parser[Log Parser<br/>logparser/anthropic.go]
        end

        subgraph Agent["Native Agent Process"]
            A[Claude Code / Cursor / Windsurf]
        end
    end

    subgraph External["External Services"]
        LLM[LLM API<br/>api.anthropic.com<br/>api.openai.com<br/>generativelanguage.googleapis.com]
    end

    subgraph Platform["Ubik Platform"]
        API[API Server<br/>/api/v1/logs]
        DB[(PostgreSQL<br/>activity_logs<br/>usage_records)]
        Web[Web UI<br/>Logs Dashboard]
    end

    %% Initialization Flow
    InitLogger --> L
    InitProxy --> P
    P --> CA

    %% Agent Configuration
    IC -->|"HTTP_PROXY=127.0.0.1:PORT<br/>NODE_EXTRA_CA_CERTS"| A

    %% Request Flow
    A -->|"API Request"| P
    P -->|"Intercept & Log"| Parser
    Parser -->|"ClassifiedLogEntry"| L
    P -->|"Forward Request"| LLM

    %% Response Flow
    LLM -->|"API Response"| P
    P -->|"Intercept & Log"| Parser
    P -->|"Forward Response"| A

    %% Logging Flow
    L --> Buffer
    Buffer -->|"100 entries OR"| BatchTimer
    BatchTimer -->|"Flush"| API

    %% Error Handling
    API -->|"Success 201"| DB
    API -->|"Failure 400/5xx"| RetryQueue
    RetryQueue -->|"Max 5 retries"| DiskQueue
    DiskQueue -->|"Background worker<br/>every 10s"| API

    %% Display
    DB --> Web
```

## Components

| Component | Location | Purpose |
|-----------|----------|---------|
| **Proxy** | `services/cli/internal/proxy/proxy.go` | MITM intercepts LLM API calls |
| **Logger** | `services/cli/internal/logging/logger.go` | Batches & sends logs to API |
| **Parser** | `services/cli/internal/logparser/anthropic.go` | Extracts tokens, tools from JSON |
| **API Client** | `services/cli/internal/api/client.go` | HTTP calls to platform |

## Data Flow

### 1. Initialization

When `ubik interactive` starts:
1. Logger instance created with batching config (100 entries, 5s interval)
2. Proxy starts on available port (8082-8091)
3. CA certificate loaded/generated at `~/.ubik/certs/`
4. Agent process spawned with proxy environment variables

### 2. Request Interception

```
Agent makes API call (e.g., to api.anthropic.com)
    |
Proxy intercepts via MITM
    |
Parser extracts: model, tokens, tool calls
    |
Logger.LogClassified() called
    |
Request forwarded to LLM
```

### 3. Response Interception

```
LLM returns response
    |
Proxy intercepts
    |
Parser extracts: usage stats, tool results, errors
    |
Logger.LogClassified() called
    |
Response forwarded to Agent
```

### 4. Log Batching & Sending

```
Events accumulate in buffer
    |
Flush triggered by:
  - Buffer reaches 100 entries
  - 5 second timer fires
    |
POST /api/v1/logs with batch
    |
Success: Logs stored in DB
Failure: Retry with exponential backoff (1s, 2s, 4s, 8s, 16s)
    |
After 5 retries: Queue to disk (~/.ubik/log_queue/)
    |
Background worker retries every 10 seconds
```

## Event Types

| Event Type | Category | Description |
|------------|----------|-------------|
| `session_start` | session | CLI session began |
| `session_end` | session | CLI session ended |
| `api_request` | proxy | Outgoing LLM API request |
| `api_response` | proxy | Incoming LLM API response |
| `user_prompt` | classified | Parsed user message |
| `ai_text` | classified | Parsed AI response |
| `tool_call` | classified | AI invoked a tool |
| `tool_result` | classified | Tool execution result |

## Configuration

### Logger Config

```go
loggerConfig := &logging.Config{
    Enabled:       true,
    BatchSize:     100,              // Entries per batch
    BatchInterval: 5 * time.Second,  // Max wait time
    MaxRetries:    5,                // Retry attempts
    RetryBackoff:  1 * time.Second,  // Initial backoff
}
```

### Proxy Environment Variables

```bash
HTTP_PROXY=http://127.0.0.1:8082
HTTPS_PROXY=http://127.0.0.1:8082
NODE_EXTRA_CA_CERTS=~/.ubik/certs/ubik-ca.pem
UBIK_SESSION_ID=<uuid>
UBIK_AGENT_ID=<agent-id>
```

## Local Storage

| Path | Purpose |
|------|---------|
| `~/.ubik/certs/ubik-ca.pem` | CA certificate for HTTPS interception |
| `~/.ubik/certs/ubik-ca-key.pem` | CA private key |
| `~/.ubik/log_queue/logs_*.json` | Failed logs queued for retry |
| `~/.ubik/config.json` | CLI configuration with API token |

## Troubleshooting

### Logs Not Appearing in UI

**Check the disk queue:**
```bash
ls -la ~/.ubik/log_queue/
```

If files exist, logs are failing to send. Common causes:
- API returning 400 (invalid event_type enum)
- API returning 401 (expired token)
- Network issues

**Check queued log content:**
```bash
head -c 500 ~/.ubik/log_queue/logs_*.json
```

**Clear queue after fixing issues:**
```bash
rm ~/.ubik/log_queue/*.json
```
