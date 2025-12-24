# Ubik Proxy Pivot Plan

**Goal**: Transform Ubik from an "agent wrapper" to a pure "security proxy" for AI agents.

**Core Value Proposition**: See and control what AI agents do, without managing the agents themselves.

---

## Executive Summary

| Aspect | Before | After |
|--------|--------|-------|
| **Model** | Ubik runs agents | Ubik proxies agent traffic |
| **Setup** | `ubik sync && ubik` | `ubik proxy start` |
| **Agent support** | Only synced agents | Any LLM client |
| **CLI commands** | 12+ commands | 6 commands |
| **Database tables** | 20 tables | 11 tables |
| **Codebase size** | ~15k lines | ~8k lines (estimate) |

---

## Phase 1: Database Cleanup

### Tables to DROP

```sql
-- Migration: 001_drop_agent_management.sql

-- Agent catalog (agents are external now)
DROP TABLE IF EXISTS agent_tools CASCADE;
DROP TABLE IF EXISTS agent_policies CASCADE;
DROP TABLE IF EXISTS agents CASCADE;
DROP TABLE IF EXISTS tools CASCADE;

-- Agent config hierarchy (not needed)
DROP TABLE IF EXISTS employee_agent_configs CASCADE;
DROP TABLE IF EXISTS team_agent_configs CASCADE;
DROP TABLE IF EXISTS org_agent_configs CASCADE;

-- MCP management (handled by agents themselves)
DROP TABLE IF EXISTS employee_mcp_configs CASCADE;
DROP TABLE IF EXISTS mcp_catalog CASCADE;
DROP TABLE IF EXISTS mcp_categories CASCADE;

-- Views that depend on dropped tables
DROP VIEW IF EXISTS v_employee_agents CASCADE;
DROP VIEW IF EXISTS v_employee_mcps CASCADE;
```

### Tables to KEEP

```
Core Organization (5 tables):
├── organizations
├── subscriptions
├── teams
├── roles
└── employees

Security & Control (3 tables):
├── policies
├── tool_policies      -- Rename from generic "policies"?
└── sessions

Telemetry (2 tables):
├── activity_logs
└── usage_records

Webhooks (2 tables):
├── webhook_destinations
└── webhook_deliveries

Approvals (2 tables - future use):
├── agent_requests     -- Rename to "access_requests"?
└── approvals
```

### Schema Changes

```sql
-- Migration: 002_simplify_activity_logs.sql

-- Make agent_id optional and add source detection
ALTER TABLE activity_logs
  ALTER COLUMN agent_id DROP NOT NULL;

-- Add client detection field
ALTER TABLE activity_logs
  ADD COLUMN client_name VARCHAR(100),      -- e.g., "claude-code", "cursor"
  ADD COLUMN client_version VARCHAR(50);    -- e.g., "1.0.25"

-- Add index for client filtering
CREATE INDEX idx_activity_logs_client ON activity_logs(client_name);
```

---

## Phase 2: CLI Cleanup

### Files to DELETE

```
services/cli/internal/
├── commands/
│   ├── agents/                 # DELETE - agent list, info, show
│   │   ├── agents.go
│   │   ├── info.go
│   │   ├── list.go
│   │   └── show.go
│   ├── sync/                   # DELETE - sync command
│   │   ├── sync.go
│   │   └── sync_test.go
│   └── interactive/            # DELETE - agent runner
│       ├── interactive.go
│       └── interactive_test.go
├── sync/                       # DELETE - entire sync package
│   ├── service.go
│   ├── service_test.go
│   ├── interfaces.go
│   ├── claude.go
│   ├── claude_test.go
│   ├── docker.go
│   ├── docker_test.go
│   ├── mock_test.go
│   └── network_test.go
├── docker/                     # DELETE - entire docker package
│   ├── runner.go              # Agent runner
│   ├── runner_test.go
│   ├── manager.go             # Container manager
│   ├── manager_test.go
│   ├── client.go              # Docker client wrapper
│   ├── client_test.go
│   ├── proxy.go               # Docker proxy (not control proxy)
│   ├── proxy_test.go
│   ├── interfaces.go
│   └── types.go
├── ui/                         # DELETE - agent picker UI
│   ├── agent_picker.go
│   └── agent_picker_test.go
└── workspace/                  # DELETE - workspace selection
    ├── workspace.go
    └── workspace_test.go
```

### Docker & Platform Files to DELETE

```
# Agent Docker images
platform/docker-images/
├── agents/                     # DELETE - entire agents directory
│   ├── claude-code/
│   │   └── Dockerfile
│   └── gemini/
│       └── Dockerfile
├── mcp/                        # DELETE - entire mcp directory
│   ├── filesystem/
│   │   └── Dockerfile
│   └── git/
│       └── Dockerfile
├── Makefile                    # DELETE
└── README.md                   # DELETE

# Legacy docker directory (duplicate)
docker/                         # DELETE - entire directory
├── mcp/
│   ├── filesystem/
│   │   └── Dockerfile
│   └── git/
│       └── Dockerfile
├── Makefile
└── README.md

# Update docker-compose.yml - remove agent/mcp services if any
docker-compose.yml              # MODIFY - review and simplify
```

### Go Dependencies to Remove

```go
// go.mod - remove these dependencies (after deleting docker package):
// github.com/docker/docker
// github.com/docker/go-connections
// github.com/moby/term
```

# Also delete from api package:
services/cli/internal/api/
├── types.go                    # MODIFY - remove agent-related types
└── client.go                   # MODIFY - remove agent-related methods
```

### Files to KEEP & MODIFY

```
services/cli/internal/
├── commands/
│   ├── root.go                 # MODIFY - remove agent commands
│   ├── auth/                   # KEEP - login/logout
│   ├── logs/                   # KEEP - logs view/stream
│   ├── policies/               # KEEP - policies list
│   └── proxy/                  # NEW - proxy start/stop/status
│       ├── proxy.go
│       ├── start.go
│       ├── stop.go
│       └── status.go
├── control/                    # KEEP - core proxy functionality
│   ├── service.go              # MODIFY - standalone proxy mode
│   ├── proxy.go
│   ├── policy_handler.go
│   ├── tool_logger_handler.go
│   └── ...
├── auth/                       # KEEP
├── config/                     # KEEP - simplify
└── api/                        # KEEP - simplify
```

### New CLI Structure

```
ubik
├── login                       # Authenticate with platform
├── logout                      # Clear credentials
├── proxy
│   ├── start                   # Start HTTPS proxy (default command)
│   ├── stop                    # Stop running proxy
│   ├── status                  # Show proxy status
│   └── health                  # Health check
├── logs
│   ├── view                    # View recent logs
│   └── stream                  # Stream logs in real-time
├── policies
│   └── list                    # List active policies
└── version                     # Show version
```

### Default Command Change

```go
// Current: ubik → interactive agent picker
// New: ubik → ubik proxy start

func NewRootCommand() *cobra.Command {
    rootCmd := &cobra.Command{
        Use:   "ubik",
        Short: "AI Agent Security Proxy",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Default action: start proxy
            return runProxyStart(cmd, args)
        },
    }
    return rootCmd
}
```

---

## Phase 3: API Cleanup

### Handlers to DELETE

```
services/api/internal/handlers/
├── org_agent_configs.go        # DELETE
├── org_agent_configs_test.go   # DELETE
├── sync.go                     # DELETE
├── sync_test.go                # DELETE
├── agents.go                   # DELETE (if exists)
└── mcps.go                     # DELETE (if exists)
```

### Handlers to KEEP

```
services/api/internal/handlers/
├── auth.go                     # KEEP
├── employees.go                # KEEP - simplify
├── logs.go                     # KEEP - enhance
├── policies.go                 # KEEP
├── webhooks.go                 # KEEP
├── health.go                   # KEEP
└── organizations.go            # KEEP - simplify
```

### Service Layer Cleanup

```
services/api/internal/service/
├── config_resolver.go          # DELETE - no more config resolution
├── config_resolver_test.go     # DELETE
└── webhook_forwarder.go        # KEEP
```

### Routes to Remove

```go
// DELETE these route groups:
r.Route("/api/v1/agents", ...)
r.Route("/api/v1/org-agent-configs", ...)
r.Route("/api/v1/employees/{id}/agent-configs", ...)
r.Route("/api/v1/sync", ...)
r.Route("/api/v1/mcps", ...)

// KEEP these route groups:
r.Route("/api/v1/auth", ...)
r.Route("/api/v1/logs", ...)
r.Route("/api/v1/policies", ...)
r.Route("/api/v1/webhooks", ...)
r.Route("/api/v1/employees", ...)  // Simplified
r.Route("/api/v1/organizations", ...) // Simplified
```

---

## Phase 4: OpenAPI Spec Cleanup

### Endpoints to Remove

```yaml
# DELETE from platform/api-spec/spec.yaml:

paths:
  # Agent management
  /api/v1/agents:                           # DELETE
  /api/v1/agents/{agent_id}:                # DELETE

  # Agent configs
  /api/v1/org-agent-configs:                # DELETE
  /api/v1/org-agent-configs/{id}:           # DELETE
  /api/v1/employees/{id}/agent-configs:     # DELETE
  /api/v1/employees/{id}/agent-configs/resolved: # DELETE
  /api/v1/employees/me/agent-configs/resolved:   # DELETE

  # Sync
  /api/v1/sync/claude-code:                 # DELETE

  # MCPs
  /api/v1/mcps:                             # DELETE
  /api/v1/mcp-categories:                   # DELETE

components:
  schemas:
    # DELETE these schemas:
    Agent:
    AgentConfig:
    OrgAgentConfig:
    ResolvedAgentConfig:
    MCPServer:
    MCPCategory:
    ClaudeCodeSyncResponse:
```

---

## Phase 5: Local Storage Cleanup

### Directories to Remove

```bash
# User's ~/.ubik/ cleanup
~/.ubik/
├── config/
│   └── agents/          # DELETE - no more agent configs
├── agents/              # DELETE - legacy location
└── ...                  # Keep: config.json, certs/, policies.json
```

### Simplified Config

```json
// ~/.ubik/config.json - BEFORE
{
  "platform_url": "https://api.ubik.dev",
  "token": "...",
  "employee_id": "...",
  "org_id": "...",
  "default_agent": "claude-code",  // DELETE
  "last_sync": "..."               // DELETE
}

// ~/.ubik/config.json - AFTER
{
  "platform_url": "https://api.ubik.dev",
  "token": "...",
  "employee_id": "...",
  "org_id": "...",
  "proxy_port": 8082,              // NEW
  "auto_start": false              // NEW - start proxy on login?
}
```

---

## Phase 6: Client Detection

### Detect AI Client from Traffic

Instead of requiring `agent_id`, detect the client from HTTP headers:

```go
// services/cli/internal/control/client_detector.go

type ClientInfo struct {
    Name    string // "claude-code", "cursor", "continue", etc.
    Version string
}

func DetectClient(req *http.Request) ClientInfo {
    // Claude Code sets these headers:
    // User-Agent: claude-code/1.0.25
    // X-Client-Name: claude-code

    userAgent := req.Header.Get("User-Agent")

    // Parse patterns
    patterns := []struct {
        prefix string
        name   string
    }{
        {"claude-code/", "claude-code"},
        {"cursor/", "cursor"},
        {"continue/", "continue"},
        {"windsurf/", "windsurf"},
        {"copilot/", "copilot"},
    }

    for _, p := range patterns {
        if strings.HasPrefix(userAgent, p.prefix) {
            version := strings.TrimPrefix(userAgent, p.prefix)
            return ClientInfo{Name: p.name, Version: version}
        }
    }

    // Fallback: check request URL patterns
    if strings.Contains(req.URL.Host, "anthropic.com") {
        return ClientInfo{Name: "anthropic-client", Version: "unknown"}
    }
    if strings.Contains(req.URL.Host, "openai.com") {
        return ClientInfo{Name: "openai-client", Version: "unknown"}
    }

    return ClientInfo{Name: "unknown", Version: "unknown"}
}
```

### Update Log Entry

```go
// Modify tool_logger_handler.go

func (h *ToolCallLoggerHandler) logToolCall(ctx *HandlerContext, call *pendingToolCall) {
    entry := LogEntry{
        EmployeeID:    ctx.EmployeeID,
        OrgID:         ctx.OrgID,
        SessionID:     ctx.SessionID,
        // AgentID removed - use ClientName instead
        ClientName:    ctx.ClientName,    // NEW
        ClientVersion: ctx.ClientVersion, // NEW
        EventType:     "tool_call",
        // ...
    }
}
```

---

## Phase 7: New User Experience

### Setup Flow

```bash
# 1. Install
brew install ubik  # or: go install github.com/ubik/cli@latest

# 2. Login (one-time)
$ ubik login
Email: admin@acme.com
Password: ********
✓ Logged in as admin@acme.com
✓ Organization: Acme Corp
✓ Policies synced: 3 rules

# 3. Start proxy
$ ubik
✓ Proxy started on localhost:8082
✓ Certificate: ~/.ubik/certs/ca.pem
✓ Session: a1b2c3d4

Press Ctrl+C to stop

# 4. Use any AI agent (in another terminal)
$ HTTPS_PROXY=http://localhost:8082 claude
# Or add to shell profile:
# export HTTPS_PROXY=http://localhost:8082

# 5. View activity
$ ubik logs stream
[12:34:56] tool_call  | Bash      | ls -la        | allowed
[12:34:57] tool_call  | Write     | /tmp/test.txt | allowed
[12:34:58] tool_call  | Bash      | rm -rf /      | BLOCKED
```

### Simplified Commands

```bash
# Core commands
ubik                    # Start proxy (default)
ubik login              # Authenticate
ubik logout             # Clear credentials

# Proxy management
ubik proxy start        # Explicit start
ubik proxy stop         # Stop daemon
ubik proxy status       # Show status

# Monitoring
ubik logs view          # Recent logs
ubik logs stream        # Real-time stream
ubik logs stream -j     # JSON format

# Policies
ubik policies list      # Show active policies
```

---

## Phase 7b: Tool Auto-Configuration

### Problem
Requiring `HTTPS_PROXY=http://localhost:8082 claude` is poor UX.

### Solution
Auto-configure AI tools to use Ubik proxy.

### New Commands

| Command | Purpose |
|---------|---------|
| `ubik setup` | One-time: detect & configure all AI tools |
| `ubik start` | Start proxy as background daemon |
| `ubik stop` | Stop proxy daemon |
| `ubik status` | Show proxy status + configured tools |
| `ubik <tool>` | Run tool with proxy (fallback wrapper) |

### Supported Tools

| Tool | Config Location | Proxy Setting |
|------|-----------------|---------------|
| Claude Code | `~/.claude/settings.json` | `env.HTTPS_PROXY` |
| Cursor | `~/.cursor/settings.json` | `http.proxy` |
| Continue | `~/.continue/config.json` | `proxy` |
| Windsurf | `~/.windsurf/settings.json` | TBD |
| VS Code | `~/.config/Code/User/settings.json` | `http.proxy` |

### Implementation

```go
// services/cli/internal/setup/tools.go

type AITool struct {
    Name       string
    ConfigPath string
    Detect     func() bool           // Check if installed
    Configure  func(proxyURL string) error
    Revert     func() error          // Undo configuration
}

var supportedTools = []AITool{
    {
        Name:       "Claude Code",
        ConfigPath: "~/.claude/settings.json",
        Detect:     func() bool { return commandExists("claude") },
        Configure:  configureClaudeCode,
    },
    // ... other tools
}

func configureClaudeCode(proxyURL string) error {
    settingsPath := expandPath("~/.claude/settings.json")

    settings := loadOrCreateJSON(settingsPath)

    // Add proxy to env
    env := getOrCreateMap(settings, "env")
    env["HTTPS_PROXY"] = proxyURL
    env["NODE_EXTRA_CA_CERTS"] = expandPath("~/.ubik/certs/ca.pem")

    return saveJSON(settingsPath, settings)
}
```

### User Flow (After Pivot)

```bash
# First time setup (once)
$ ubik login
✓ Logged in as admin@acme.com

$ ubik setup
Detecting AI tools...
  ✓ Claude Code found
  ✓ Cursor found
  ✗ Continue not found

Configuring proxy settings...
  ✓ Claude Code: Updated ~/.claude/settings.json
  ✓ Cursor: Updated ~/.cursor/settings.json

Done! Run 'ubik start' to begin monitoring.

# Start proxy
$ ubik start
✓ Proxy started on localhost:8082
✓ Policies synced: 3 rules
✓ Session: a1b2c3d4

# Now just use your tools normally
$ claude   # Works - already configured to use proxy
$ cursor   # Works - already configured to use proxy

# View activity
$ ubik logs stream
```

### Files to Create

```
services/cli/internal/
├── setup/
│   ├── setup.go           # Main setup logic
│   ├── tools.go           # Tool detection & configuration
│   ├── claude.go          # Claude Code specific
│   ├── cursor.go          # Cursor specific
│   └── continue.go        # Continue specific
└── commands/
    └── setup/
        └── setup.go       # ubik setup command
```

---

## Phase 8: Implementation Order

### Week 1: Database & Core
- [ ] Create database migration to drop unused tables
- [ ] Update schema with client detection fields
- [ ] Create client detector in control package
- [ ] Update log entry structure

### Week 2: CLI Cleanup
- [ ] Delete agent-related commands
- [ ] Delete sync package
- [ ] Delete docker/runner package
- [ ] Delete ui package
- [ ] Simplify root command to start proxy

### Week 3: API Cleanup
- [ ] Delete agent config handlers
- [ ] Delete sync handler
- [ ] Delete config resolver service
- [ ] Update OpenAPI spec
- [ ] Regenerate API types

### Week 4: Polish
- [ ] Update documentation
- [ ] Update CLAUDE.md files
- [ ] Clean up tests
- [ ] Update CI/CD pipelines
- [ ] Write migration guide for users

---

## Migration Script

```bash
#!/bin/bash
# scripts/pivot-to-proxy.sh

echo "=== Ubik Proxy Pivot Migration ==="

# 1. Database migration
echo "Running database migrations..."
psql $DATABASE_URL -f migrations/001_drop_agent_management.sql
psql $DATABASE_URL -f migrations/002_simplify_activity_logs.sql

# 2. Clean up user's local config
echo "Cleaning up local config..."
rm -rf ~/.ubik/config/agents/
rm -rf ~/.ubik/agents/

# 3. Update config.json
echo "Updating config.json..."
jq 'del(.default_agent, .last_sync) | .proxy_port = 8082' \
    ~/.ubik/config.json > ~/.ubik/config.json.tmp
mv ~/.ubik/config.json.tmp ~/.ubik/config.json

echo "✓ Migration complete"
echo ""
echo "New usage:"
echo "  ubik              # Start proxy"
echo "  ubik logs stream  # View activity"
```

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Breaking existing users | Version bump (v2.0), migration script |
| Loss of agent_id in logs | Client detection + backfill script |
| Orphaned database data | Soft delete first, hard delete after 30 days |
| Documentation stale | Update all CLAUDE.md files in same PR |

---

## Success Metrics

After pivot:
- [ ] CLI binary size reduced by 30%+
- [ ] Setup time < 1 minute (was 5+ minutes)
- [ ] Works with any LLM client without config
- [ ] Database schema: 11 tables (was 20)
- [ ] CLI commands: 6 (was 12+)

---

## Files Summary

### To DELETE (approximate)

```
# CLI packages
services/cli/internal/commands/agents/     ~400 lines
services/cli/internal/commands/sync/       ~200 lines
services/cli/internal/commands/interactive/ ~300 lines
services/cli/internal/sync/                ~1500 lines
services/cli/internal/docker/              ~800 lines (10 files)
services/cli/internal/ui/                  ~200 lines
services/cli/internal/workspace/           ~150 lines

# API packages
services/api/internal/handlers/org_agent_configs* ~800 lines
services/api/internal/handlers/sync*       ~300 lines
services/api/internal/service/config_resolver* ~400 lines

# Docker & Platform directories
platform/docker-images/agents/             # 2 Dockerfiles + configs
platform/docker-images/mcp/                # 2 Dockerfiles + configs
platform/docker-images/Makefile
platform/docker-images/README.md
docker/                                    # Entire legacy directory

─────────────────────────────────────────────────
Total: ~5,050+ lines of Go code
       + 4 Dockerfiles
       + 2 Makefiles
       + 2 READMEs
       + associated configs
```

### To CREATE

```
services/cli/internal/commands/proxy/      ~300 lines
services/cli/internal/control/client_detector.go ~100 lines
migrations/001_drop_agent_management.sql   ~50 lines
migrations/002_simplify_activity_logs.sql  ~20 lines
─────────────────────────────────────────────────
Total: ~470 lines to create
```

**Net reduction: ~4,200 lines of code**

---

*Created: 2024-12-23*
