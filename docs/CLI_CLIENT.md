# ubik CLI Client - Architecture & Design

**Version**: v0.2.0 (Planned)
**Status**: Design Phase
**Target Agent**: Claude Code
**Last Updated**: 2025-10-29

---

## 📋 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [User Experience](#user-experience)
- [Technical Design](#technical-design)
- [Implementation Plan](#implementation-plan)
- [Open Questions](#open-questions)

---

## Overview

### Purpose

The `ubik` CLI client enables employees to use AI coding agents (Claude Code, Aider, etc.) with centrally-managed configurations from the platform. The CLI acts as a **container orchestrator**, managing Docker containers that run the actual agents with injected configs and MCP servers.

### Core Value Proposition

1. **Zero Manual Setup** - Employee types `ubik sync`, everything is configured
2. **Central Policy Enforcement** - Org policies applied automatically
3. **Usage Tracking** - All agent usage tracked and attributed
4. **Multi-Agent Support** - Switch between agents seamlessly
5. **Transparent UX** - Feels like native CLI, Docker is invisible

---

## Architecture

### System Diagram

```
┌─────────────────────────────────────────────────────────────┐
│ Host Machine (Employee's Computer)                          │
│                                                              │
│  ┌────────────────────────────────────────────────────┐    │
│  │ ubik CLI (Native Go Process)                       │    │
│  │  • Authenticates with platform                     │    │
│  │  • Fetches resolved configs                        │    │
│  │  • Manages Docker containers                       │    │
│  │  • Proxies stdin/stdout to agent                   │    │
│  │  • Sends usage telemetry to platform               │    │
│  └────────────────────────────────────────────────────┘    │
│                      ↕                                       │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Docker                                             │    │
│  │                                                     │    │
│  │  ┌──────────────────────────────────────────────┐ │    │
│  │  │ Container: claude-code                       │ │    │
│  │  │  • Claude Code CLI binary                    │ │    │
│  │  │  • Injected config from platform             │ │    │
│  │  │  • Connected to MCP servers                  │ │    │
│  │  │  • Workspace mounted (/workspace)            │ │    │
│  │  └──────────────────────────────────────────────┘ │    │
│  │                      ↕ Network                     │    │
│  │  ┌──────────────────────────────────────────────┐ │    │
│  │  │ Container: MCP Servers (separate)            │ │    │
│  │  │  • mcp-server-filesystem                     │ │    │
│  │  │  • mcp-server-git                            │ │    │
│  │  │  • mcp-server-postgres                       │ │    │
│  │  │  • Auto-configured from platform             │ │    │
│  │  └──────────────────────────────────────────────┘ │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                      ↕ HTTPS
┌─────────────────────────────────────────────────────────────┐
│ Platform (Central Server)                                   │
│  • Authentication (JWT)                                     │
│  • Config resolution (org → team → employee)               │
│  • Agent approval workflows                                │
│  • Usage tracking API (TBD)                                │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| **Agent Execution** | Docker containers | Isolation, reproducibility, MCP server bundling |
| **CLI Process** | Native Go binary | Fast, no Docker for CLI itself, portable |
| **MCP Servers** | **Separate containers** | Better isolation, resource management, reusability |
| **Config Injection** | Environment variables | Simple, secure, standard Docker pattern |
| **I/O Proxy** | stdin/stdout streaming | Transparent user experience |
| **First Agent** | **Claude Code** | Most requested, good CLI interface |
| **Workspace** | **Ask user, default to CWD** | Flexible, intuitive default |

### MCP Server Separation - Pros & Cons

**✅ Pros of Separate MCP Containers:**
- Independent lifecycle (restart MCP without restarting agent)
- Shared MCPs across multiple agents
- Better resource isolation
- Easier debugging (separate logs)
- Security: MCP crashes don't affect agent

**❌ Cons:**
- More containers to manage (3-5 instead of 1)
- Network complexity (container-to-container communication)
- Slightly slower startup (multiple containers)

**Decision**: Use separate containers - benefits outweigh complexity.

---

## User Experience

### Installation

```bash
# macOS
brew install ubik

# Linux
curl -fsSL https://get.ubik.io | sh

# Or download binary
wget https://github.com/ubik/cli/releases/download/v0.2.0/ubik-linux-amd64
chmod +x ubik-linux-amd64
sudo mv ubik-linux-amd64 /usr/local/bin/ubik
```

### First Time Setup

```bash
# Step 1: Login
$ ubik login
Platform URL [https://api.ubik.io]:
Email: alice@acme.com
Password: ****
✓ Authenticated successfully
✓ Employee ID: 550e8400-e29b-41d4-a716-446655440000

# Step 2: Check Docker
✓ Docker is running
✓ Docker version: 24.0.6

# Step 3: Sync configs and start
$ ubik sync
✓ Fetching configs from platform...
✓ Resolved configs for: claude-code
✓ Pulling Docker images...
  • ubik/claude-code:latest (142 MB)
  • ubik/mcp-filesystem:latest (12 MB)
  • ubik/mcp-git:latest (15 MB)
✓ Starting MCP servers...
  • mcp-filesystem (container: ubik-mcp-fs-abc123)
  • mcp-git (container: ubik-mcp-git-def456)
✓ Configuring claude-code...
✓ Starting claude-code container...

🎉 Ready! Run 'ubik' to start coding.
```

### Daily Usage

```bash
# Start agent (prompts for workspace, defaults to current directory)
$ ubik
Workspace [/Users/alice/projects/myapp]:
# User presses Enter to accept default or types path
✓ Workspace: /Users/alice/projects/myapp (2.3 GB, 1,234 files)
✓ Agent: claude-code (v1.2.3)
✓ MCP Servers: filesystem, git

claude-code> Fix the authentication bug in login.go

Analyzing login.go...
Found issue on line 42: password comparison not constant-time
Suggested fix:
  - Use crypto/subtle.ConstantTimeCompare

Apply this fix? (y/n) y
✓ Applied changes to login.go

claude-code> exit
✓ Session duration: 5m 23s
✓ Tokens used: 1,234 input / 567 output
✓ Usage synced to platform

# Start agent in specific directory
$ ubik --workspace /path/to/project
✓ Workspace: /path/to/project

# Use different agent (if configured)
$ ubik --agent aider
✓ Agent: aider (v0.15.0)
```

### Requesting Agent Access

```bash
# Request access to agent not in your config
$ ubik config agent cursor
⚠ Agent 'cursor' not in your approved list
→ Creating approval request...
✓ Request created (ID: req-abc123)
⏳ Waiting for manager approval...
  Manager: bob@acme.com
  Status: https://platform.ubik.io/requests/req-abc123

# Check request status
$ ubik requests
ID          Agent      Status    Requested      Manager
req-abc123  cursor     pending   2 hours ago    bob@acme.com
req-def456  aider      approved  1 day ago      bob@acme.com

# After approval
$ ubik sync
✓ New agent available: cursor
✓ Run 'ubik --agent cursor' to use it
```

---

## Technical Design

### CLI Project Structure

```
pivot/
├── cmd/
│   └── ubik-cli/
│       └── main.go           # CLI entry point
│
├── internal/cli/
│   ├── auth.go               # Login, token management
│   ├── sync.go               # Fetch configs, start containers
│   ├── agent.go              # Agent lifecycle management
│   ├── container.go          # Docker API wrapper
│   ├── proxy.go              # I/O proxy to container
│   ├── telemetry.go          # Usage tracking (TBD)
│   ├── config.go             # Local config management
│   └── workspace.go          # Workspace selection logic
│
├── pkg/cli/
│   ├── docker/
│   │   ├── compose.go        # docker-compose.yml generation
│   │   ├── client.go         # Docker SDK wrapper
│   │   └── network.go        # Container networking
│   │
│   └── platform/
│       ├── client.go         # Platform API client
│       ├── auth.go           # JWT handling
│       └── models.go         # API response types
│
└── configs/agents/
    ├── claude-code.yaml      # Agent image config
    └── aider.yaml
```

### Local Configuration Storage

```
~/.ubik/
├── config.json               # CLI configuration
│   {
│     "platform_url": "https://api.ubik.io",
│     "token": "eyJhbGc...",
│     "employee_id": "uuid",
│     "default_agent": "claude-code",
│     "last_sync": "2025-10-29T10:30:00Z"
│   }
│
└── agents/
    └── claude-code/
        ├── config.json       # Resolved agent config
        └── mcp-servers.json  # MCP configuration
```

---

## Implementation Plan

### Phase 1: Foundation (Week 1 - 3-4 days)

**Tasks**:
- [ ] Project setup (Go module, structure)
- [ ] Authentication (`ubik login`)
- [ ] Platform API client
- [ ] Config fetching (`ubik sync` - fetch only)
- [ ] Unit tests

**Deliverables**: `ubik login` and basic config sync working

---

### Phase 2: Docker Integration (Week 2 - 4-5 days)

**Tasks**:
- [ ] Docker client integration
- [ ] Docker Compose generation
- [ ] Container lifecycle management
- [ ] Complete `ubik sync` (start containers)
- [ ] Integration tests

**Deliverables**: Containers start with configs, MCP servers accessible

---

### Phase 3: Interactive Mode (Week 3 - 3-4 days)

**Tasks**:
- [ ] Workspace selection (ask user, default CWD)
- [ ] I/O proxying to container
- [ ] Agent switching
- [ ] Session management

**Deliverables**: `ubik` command works transparently

---

### Phase 4: Agent Management (Week 4 - 3-4 days)

**Tasks**:
- [ ] Agent listing/info commands
- [ ] Agent request/approval workflow
- [ ] Update mechanism
- [ ] Cleanup commands

**Deliverables**: Multi-agent support, approval workflow complete

---

### Phase 5: Polish & Documentation (Week 5 - 4-5 days)

**Tasks**:
- [ ] Error handling & logging
- [ ] Installation scripts
- [ ] User documentation
- [ ] Usage telemetry (design TBD)
- [ ] Beta testing

**Deliverables**: Production-ready CLI with docs

---

**Total Timeline**: 4-5 weeks

---

## Open Questions

### 🔴 High Priority (Blocking)

1. **Claude Code CLI Access**
   - Is Claude Code CLI publicly available?
   - Do we have API keys/licenses?
   - Docker base image available?

2. **Platform API Endpoints**
   - Is GET /employees/{id}/agent-configs/resolved implemented?
   - Need agent approval endpoints (POST/GET /agent-requests)

### 🟡 Medium Priority

3. **Telemetry Design** ⚠️ **TBD - DISCUSS WITH USER LATER**
   - What metrics to track?
     - Session duration
     - Commands executed?
     - Token usage (input/output)?
     - Error rates?
     - Agent performance?
   - Frequency?
     - Real-time streaming?
     - Batched every N minutes?
     - End of session?
   - Privacy considerations?
     - What NOT to track (code content)?
     - Anonymization?
   - Storage format?
     - Platform database?
     - Time-series DB?

   **Action**: Schedule design discussion with user before Phase 5

4. **MCP Server Images**
   - Build our own or use community images?
   - Which MCPs to support initially?
     - filesystem (required)
     - git (required)
     - postgres (optional?)

---

## Dependencies

### Platform API Requirements

1. **Authentication** ✅ (Already exists)
   - POST /auth/login
   - GET /auth/me

2. **Resolved Configs** ❓ (Need to verify endpoint exists)
   - GET /employees/{id}/agent-configs/resolved

3. **Agent Approval** ❌ (Need to build)
   - POST /agent-requests
   - GET /agent-requests/{id}

4. **Usage Tracking** ❌ (TBD - Design needed, discuss later)
   - POST /usage/sessions (design TBD)

---

## Success Metrics (v0.2.0)

- [ ] Employee can install CLI in < 5 minutes
- [ ] First sync completes in < 2 minutes
- [ ] Agent responds in < 3 seconds
- [ ] Zero manual configuration required
- [ ] Support Claude Code + filesystem/git MCPs
- [ ] Works on macOS
- [ ] Complete documentation

---

**Status**: Design Complete, Ready for Implementation
**Next Step**: Phase 1 - Foundation (Week 1)
