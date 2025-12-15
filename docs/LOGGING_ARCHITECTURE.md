# Ubik CLI Logging Architecture

## Overview

The Ubik CLI implements a multi-layer logging system that captures agent activity, API interactions, and sessions.

**Current Issue:** Logs appear with `session_id: "00000000-0000-0000-0000-000000000000"` instead of the actual session UUID.

---

## Architecture Diagram

```
┌────────────────────────────────────────────────────────────────────┐
│                   INTERACTIVE SESSION PROCESS                       │
│                                                                      │
│  main.go:runInteractiveMode()                                       │
│  └─ Logger created: sessionID = uuid.New() ← REAL UUID             │
│     └─ SetAgentID(selectedAgent.AgentID)                           │
│     └─ StartSession() → logs "session_start"                       │
│                                                                      │
│  NativeRunner.Start()                                               │
│  └─ Env: UBIK_SESSION_ID="550e8400-e29b-..." ← PASSED TO AGENT    │
│  └─ Env: UBIK_AGENT_ID="agent-uuid"                                │
│  └─ Exec: claude (agent binary)                                    │
│                                                                      │
│  Agent Process (claude-code)                                        │
│  └─ Makes HTTP requests to api.anthropic.com                       │
│  └─ Uses HTTP_PROXY=localhost:8082                                 │
│  └─ ⚠️ Does NOT add X-Ubik-Session header                         │
│                                                                      │
└────────────────────────────────────────────────────────────────────┘
                              │
                              │ HTTP via proxy
                              ▼
┌────────────────────────────────────────────────────────────────────┐
│              BACKGROUND PROXY DAEMON PROCESS                        │
│              (Persistent, runs across sessions)                     │
│                                                                      │
│  daemon.go:RunDaemon()                                              │
│  └─ Creates SessionManager                                         │
│  └─ Creates ProxyServer with logger                                │
│     └─ ⚠️ This logger has sessionID = ZERO UUID                   │
│     └─ Because StartSession() was never called on daemon logger   │
│                                                                      │
│  ProxyServer.logRequest()                                          │
│  └─ Tries: r.Header.Get("X-Ubik-Session") → "" (not set)          │
│  └─ Tries: getSessionInfo() → looks up SessionManager              │
│  └─ ⚠️ SessionManager may be empty (no session registered)        │
│  └─ Falls back to: logger.sessionID → ZERO UUID                   │
│                                                                      │
│  Logger.LogEvent()                                                  │
│  └─ SessionID: l.sessionID.String() → "00000000-0000-..."          │
│                                                                      │
└────────────────────────────────────────────────────────────────────┘
                              │
                              │ POST /api/v1/logs
                              ▼
┌────────────────────────────────────────────────────────────────────┐
│                     PLATFORM API + DATABASE                         │
│                                                                      │
│  activity_logs table:                                               │
│  └─ session_id: "00000000-0000-0000-0000-000000000000" ⚠️          │
│                                                                      │
└────────────────────────────────────────────────────────────────────┘
```

---

## The Zero UUID Problem

### Root Cause

There are **TWO separate logger instances**:

1. **Interactive Session Logger** (main.go)
   - Created when user runs `ubik`
   - Calls `StartSession()` → generates real UUID
   - Logs session_start and session_end events ✓

2. **Proxy Daemon Logger** (daemon.go → newProxyRunCommand)
   - Created when proxy daemon starts (`ubik proxy start`)
   - Never calls `StartSession()` → sessionID stays as zero UUID
   - Logs ALL api_request and api_response events ✗

### Why Sessions Are Not Linked

```
┌─────────────────────┐          ┌─────────────────────┐
│ Interactive Process │          │ Proxy Daemon Process │
│                     │          │                      │
│ Logger A:           │   ≠      │ Logger B:            │
│ sessionID = "550e8" │          │ sessionID = "00000"  │
│                     │          │                      │
│ NativeRunner passes │──────────│ SessionManager has   │
│ sessionID via ENV   │    ?     │ session registered   │
│                     │          │ but logger doesn't   │
│                     │          │ use it!              │
└─────────────────────┘          └─────────────────────┘
```

---

## Data Flow: Where Session ID Gets Lost

### Step 1: Session Created (WORKS)
```go
// main.go:runInteractiveMode()
logger, _ := logging.NewLogger(loggerConfig, apiClient)
logger.SetAgentID(selectedAgent.AgentID)
sessionID := logger.StartSession()  // → "550e8400-e29b-..."
```

### Step 2: Passed to Agent via ENV (WORKS)
```go
// native_runner.go:Start()
env = append(env, fmt.Sprintf("UBIK_SESSION_ID=%s", config.SessionID))
env = append(env, fmt.Sprintf("UBIK_AGENT_ID=%s", config.AgentID))
```

### Step 3: Session Registered with Proxy (SHOULD WORK)
```go
// native_runner.go:registerWithProxy()
client.RegisterSession(RegisterSessionRequest{
    SessionID:  r.sessionID,
    AgentID:    r.agentID,
    AgentName:  config.AgentName,
    Workspace:  config.Workspace,
})
```

### Step 4: Proxy Receives Request (FAILS HERE)
```go
// server.go:getSessionInfo()
func (s *ProxyServer) getSessionInfo(r *http.Request) (sessionID, agentID string) {
    // Try headers first
    sessionID = r.Header.Get("X-Ubik-Session")  // → "" (agent doesn't set)
    agentID = r.Header.Get("X-Ubik-Agent")      // → "" (agent doesn't set)

    // Try SessionManager
    if sessionID == "" && s.sessionManager != nil {
        sessions := s.sessionManager.ListSessions()
        // ⚠️ IS THIS RETURNING SESSIONS?
        // ⚠️ OR IS SessionManager EMPTY?
    }

    return sessionID, agentID  // → "", "" if nothing found
}
```

### Step 5: Logger Uses Zero UUID
```go
// logging/logger.go:LogEvent()
entry := LogEntry{
    SessionID: l.sessionID.String(),  // → "00000000-..." (daemon's logger)
    // ...
}
```

---

## Components Involved

| Component | File | Role |
|-----------|------|------|
| Interactive Mode | `cmd/ubik/main.go` | Creates session, starts agent |
| Native Runner | `internal/native_runner.go` | Launches agent, registers session |
| Control Client | `internal/httpproxy/control_client.go` | IPC to daemon |
| Control Server | `internal/httpproxy/control.go` | Handles session registration |
| Session Manager | `internal/httpproxy/session_manager.go` | Stores active sessions |
| Proxy Server | `internal/httpproxy/server.go` | Intercepts & logs requests |
| Daemon | `internal/httpproxy/daemon.go` | Manages proxy lifecycle |
| Logger | `internal/logging/logger.go` | Buffers & sends logs to API |

---

## Session Registration Flow

```
ubik (interactive mode)
    │
    ├─1─► NativeRunner.Start()
    │         │
    │         ├─2─► registerWithProxy()
    │         │         │
    │         │         └─3─► ControlClient.RegisterSession()
    │         │                   │
    │         │                   └─4─► HTTP POST unix://~/.ubik/proxy.sock
    │         │                              │
    │         │                              ▼
    │         │                   ControlServer.handleRegister()
    │         │                              │
    │         │                              └─5─► SessionManager.Register()
    │         │                                        │
    │         │                                        └─► sessions["uuid"] = Session{...}
    │         │
    │         └─6─► Start agent process
    │
    └─► Agent makes API requests through proxy
              │
              ▼
        ProxyServer.logRequest()
              │
              └─► getSessionInfo()
                      │
                      └─► s.sessionManager.ListSessions()
                              │
                              └─► Should return the registered session!
```

---

## Debug Checklist

### 1. Is Session Being Registered?
```bash
# Check active sessions
./bin/ubik-cli proxy sessions
```

### 2. Is SessionManager Passed to ProxyServer?
```go
// daemon.go:RunDaemon()
server := NewProxyServer(logger)
server.SetSessionManager(d.sessionManager)  // ← Is this called?
```

### 3. Is getSessionInfo() Finding Sessions?
Add debug logging to `server.go:getSessionInfo()`:
```go
fmt.Printf("DEBUG: sessions = %d\n", len(sessions))
for _, s := range sessions {
    fmt.Printf("DEBUG: session %s agent %s\n", s.ID, s.AgentID)
}
```

### 4. Is registerWithProxy() Being Called?
Check `native_runner.go` - is registration happening?

---

## Potential Fixes

### Fix 1: Verify Session Registration Chain
Ensure the full chain works:
1. NativeRunner calls registerWithProxy()
2. ControlClient sends request to daemon
3. ControlServer receives and processes
4. SessionManager stores session
5. ProxyServer can access SessionManager

### Fix 2: Pass Logger's Session ID to Daemon
When starting interactive mode, update daemon's session context:
```go
// After registering session
controlClient.SetActiveSession(sessionID, agentID)
```

### Fix 3: Use Logger Per-Request Context
Instead of one daemon logger, pass session context per-request.

### Fix 4: Inject Headers via Proxy
Proxy could inject `X-Ubik-Session` header based on registered session.

---

## Key Insight

The `getSessionInfo()` fix should work IF:
1. Session is properly registered with daemon
2. SessionManager is passed to ProxyServer
3. ListSessions() returns the active session

**Most likely issue:** One of these links in the chain is broken.
