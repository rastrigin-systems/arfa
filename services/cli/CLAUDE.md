# CLI Development Guide

**You are working on the Ubik CLI** - the AI Agent Security Proxy for enterprise environments.

---

## Quick Context

**What is this service?**
Self-contained Go CLI that provides an HTTPS proxy for intercepting, logging, and enforcing policies on AI agent (Claude Code, Cursor, Windsurf) API traffic.

**Key capabilities:**
- **Security Proxy** - MITM proxy for LLM API traffic interception
- **Activity Logging** - Log all AI agent tool calls and API requests
- **Policy Enforcement** - Block dangerous tool calls based on org policies
- **Client Detection** - Auto-detect AI client from User-Agent headers
- **Enterprise Deployment** - PAC file setup for transparent routing

**Architecture principle:** Self-contained module with minimal dependencies. NO database code, NO generated API code. Only depends on `pkg/types` for shared data structures.

---

## CLI Commands

### Core Commands

```bash
# Start the proxy (default action)
ubik                      # Same as 'ubik proxy start'
ubik proxy start          # Start HTTPS proxy on localhost:8082

# Authentication
ubik login                # Authenticate with platform
ubik logout               # Clear credentials

# Monitoring
ubik logs stream          # Stream real-time activity logs
ubik policies list        # View active security policies
```

### Setup Commands

```bash
# For CI pipelines - output env vars
eval $(ubik proxy env)              # Set HTTPS_PROXY, SSL_CERT_FILE
ubik proxy env --format=github      # GitHub Actions format

# For local/enterprise - permanent system config
sudo ubik setup system              # Install PAC file + auto-start
ubik setup status                   # Check setup status
sudo ubik setup uninstall           # Remove system config
```

### Proxy Commands

```bash
ubik proxy start          # Start proxy (foreground)
ubik proxy stop           # Stop proxy
ubik proxy status         # Check if running
ubik proxy health         # Health check
ubik proxy env            # Output environment variables
```

---

## Deployment Modes

### 1. CI Pipelines (Temporary)

```yaml
# GitHub Actions
- name: Setup Ubik
  run: |
    ubik proxy start --background
    ubik proxy env --format=github >> $GITHUB_ENV

- name: Run AI Agent
  run: claude "fix the tests"
```

```bash
# Generic CI
ubik proxy start --background
eval $(ubik proxy env)
claude "fix the bug"
```

### 2. Local Development (Per-Session)

```bash
# Terminal 1: Start proxy
ubik proxy start

# Terminal 2: Configure shell and run agent
eval $(ubik proxy env)
claude
```

### 3. Enterprise (Permanent via PAC)

```bash
# One-time admin setup (requires sudo)
sudo ubik setup system

# Creates:
# - PAC file routing only LLM domains through proxy
# - CA certificate in system trust store
# - Auto-start daemon (launchd/systemd)

# Employees just run their AI agents normally
claude  # Transparently proxied
```

**Proxied domains (via PAC):**
- `api.anthropic.com`
- `api.openai.com`
- `generativelanguage.googleapis.com`

All other traffic goes direct - no performance impact.

---

## Architecture

### Control Service Pipeline

```
Request → ClientDetector → PolicyHandler → LoggerHandler → Forward to LLM API
                ↓                ↓               ↓
         Detect client     Check policies    Log to queue
         from User-Agent   Block if needed   Upload async
```

**Handlers (priority order):**
1. `ClientDetectorHandler` (200) - Detect client from User-Agent
2. `PolicyHandler` (110) - Enforce tool blocking policies
3. `LoggerHandler` (100) - Log requests/responses

### Directory Structure

```
services/cli/
├── cmd/ubik/main.go          # CLI entry point
├── internal/
│   ├── commands/             # Cobra command implementations
│   │   ├── root.go           # Root command
│   │   ├── auth/             # login, logout
│   │   ├── proxy/            # start, stop, status, env
│   │   ├── setup/            # system, status, uninstall
│   │   ├── logs/             # stream, view
│   │   └── policies/         # list
│   ├── control/              # Control service (core proxy logic)
│   │   ├── service.go        # Main service coordinator
│   │   ├── pipeline.go       # Handler pipeline
│   │   ├── handler.go        # Handler interface + context
│   │   ├── client_detector.go      # User-Agent parsing
│   │   ├── client_detector_handler.go
│   │   ├── policy_handler.go       # Policy enforcement
│   │   ├── logger_handler.go       # Request/response logging
│   │   ├── tool_logger_handler.go  # Tool call logging
│   │   ├── queue.go          # Disk-based log queue
│   │   ├── uploader.go       # Async log upload
│   │   └── proxy.go          # goproxy wrapper
│   ├── proxy/                # Simplified MITM proxy
│   ├── logging/              # Legacy logging (being replaced)
│   ├── logparser/            # Anthropic API response parser
│   ├── api/                  # Platform API client
│   ├── auth/                 # Authentication service
│   ├── config/               # Local config management
│   └── container/            # DI container
└── tests/
    └── integration/          # Integration tests
```

### Key Data Types

**HandlerContext** - Passed through pipeline:
```go
type HandlerContext struct {
    EmployeeID    string
    OrgID         string
    SessionID     string
    ClientName    string    // e.g., "claude-code"
    ClientVersion string    // e.g., "1.0.25"
    Metadata      map[string]interface{}
}
```

**LogEntry** - Queued for upload:
```go
type LogEntry struct {
    EmployeeID    string
    OrgID         string
    SessionID     string
    ClientName    string
    ClientVersion string
    EventType     string    // "api_request", "tool_call", etc.
    EventCategory string    // "proxy", "classified"
    Timestamp     time.Time
    Payload       map[string]interface{}
}
```

---

## Essential Commands

```bash
# Build and test
go build ./...
go test ./... -count=1

# Format (required before commit)
gofmt -w .
go vet ./...

# Run locally
go run ./cmd/ubik proxy start
```

---

## Configuration

**Local config location:** `~/.ubik/`

```
~/.ubik/
├── config.json           # CLI configuration (API URL, credentials)
├── certs/
│   ├── ubik-ca.pem       # CA certificate (trust this)
│   └── ubik-ca-key.pem   # CA private key
├── proxy.pac             # PAC file (created by setup system)
├── log_queue/            # Pending log uploads
└── logs/                 # Daemon logs
```

---

## Client Detection

The proxy auto-detects AI clients from User-Agent headers:

| Client | User-Agent Pattern | Detected As |
|--------|-------------------|-------------|
| Claude Code | `claude-code/1.0.25` | `claude-code` |
| Cursor | `Cursor/0.43.0` | `cursor` |
| Continue | `Continue/1.0.0` | `continue` |
| Windsurf | `Windsurf/1.0.0` | `windsurf` |
| Aider | `aider/0.50.0` | `aider` |
| GitHub Copilot | `GithubCopilot/1.0` | `copilot` |

Detection happens in `ClientDetectorHandler` at the start of the pipeline.

---

## Testing

```bash
# Unit tests (fast)
go test ./internal/control/... -count=1

# All tests
go test ./... -count=1

# Integration tests
go test ./tests/integration/... -count=1

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Common Pitfalls

### 1. Stale Binary
```bash
# ✅ Rebuild before testing
go build ./... && go run ./cmd/ubik proxy start
```

### 2. Certificate Not Trusted
```bash
# ✅ Check if cert exists
ls ~/.ubik/certs/ubik-ca.pem

# ✅ Start proxy once to generate cert
ubik proxy start

# ✅ Trust cert (macOS)
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain ~/.ubik/certs/ubik-ca.pem
```

### 3. Port Already in Use
```bash
# ✅ Check what's using 8082
lsof -i :8082

# Proxy will auto-try 8082-8091
```

### 4. Proxy Not Working
```bash
# ✅ Verify proxy is running
curl -x http://127.0.0.1:8082 https://api.anthropic.com/v1/messages

# ✅ Check env vars
echo $HTTPS_PROXY
echo $SSL_CERT_FILE
```

---

## Related Documentation

- [../../CLAUDE.md](../../CLAUDE.md) - Monorepo overview
- [../../docs/TESTING.md](../../docs/TESTING.md) - Testing guide
- [../../docs/DEV_WORKFLOW.md](../../docs/DEV_WORKFLOW.md) - PR workflow
- [../api/CLAUDE.md](../api/CLAUDE.md) - API development
- [../web/CLAUDE.md](../web/CLAUDE.md) - Web UI development
