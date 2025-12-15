# Ubik CLI Logging Architecture

## Overview

The Ubik CLI implements a multi-layer logging system that captures agent activity, API interactions, and sessions.

**Current Issue:** Logs appear with `session_id: "00000000-0000-0000-0000-000000000000"` instead of the actual session UUID.

**Status:** NOT FIXED - Multiple attempts have failed.

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        INTERACTIVE SESSION PROCESS                           │
│                        (ubik command - foreground)                           │
│                                                                              │
│  main.go:runInteractiveMode()                                                │
│  ├─ configManager.Load() → gets platform URL, auth token                    │
│  ├─ logging.NewLogger() → creates Logger A                                  │
│  │   └─ logger.sessionID = uuid.Nil (zero UUID initially)                   │
│  ├─ logger.SetAgentID(selectedAgent.AgentID)                                │
│  ├─ logger.StartSession() → generates REAL UUID "550e8400-..."              │
│  │   └─ logger.sessionID = uuid.New() ← NOW HAS REAL UUID                   │
│  │   └─ logs "session_start" event with real UUID ✓                         │
│  │                                                                           │
│  ├─ proxyDaemon.EnsureRunning(8082) → starts daemon if not running          │
│  │                                                                           │
│  ├─ NativeRunnerConfig{SessionID: sessionID} ← REAL UUID passed here        │
│  │                                                                           │
│  └─ runner.Run(ctx, config, stdin, stdout, stderr)                          │
│      │                                                                       │
│      ├─ RegisterWithSecurityGateway(config)                                 │
│      │   ├─ NewDefaultControlClient() → connects to ~/.ubik/proxy.sock      │
│      │   └─ client.RegisterSession(req) → POST /sessions to daemon          │
│      │       └─ req.SessionID = config.SessionID ← REAL UUID                │
│      │                                                                       │
│      └─ Start(ctx, config)                                                  │
│          └─ exec.Command(binaryPath) with env vars:                         │
│              ├─ UBIK_SESSION_ID="550e8400-..." ← REAL UUID                  │
│              ├─ UBIK_AGENT_ID="agent-uuid"                                  │
│              ├─ HTTP_PROXY=http://localhost:8082                            │
│              └─ NODE_EXTRA_CA_CERTS=~/.ubik/certs/ubik-ca.pem               │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ Agent process runs
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        AGENT PROCESS (claude-code)                           │
│                        (child process of ubik)                               │
│                                                                              │
│  Has environment variables:                                                  │
│  ├─ UBIK_SESSION_ID="550e8400-..." ← available but NOT USED by agent        │
│  └─ HTTP_PROXY=http://localhost:8082                                        │
│                                                                              │
│  Makes HTTP requests to api.anthropic.com                                    │
│  └─ Request goes through HTTP_PROXY                                          │
│  └─ ⚠️ Agent does NOT add X-Ubik-Session header (agent is unmodified)       │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ HTTP via proxy (port 8082)
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        PROXY DAEMON PROCESS                                  │
│                        (separate background process)                         │
│                                                                              │
│  Started by: ubik proxy start OR auto-started by EnsureRunning()            │
│  Binary: same ubik-cli binary with "proxy run" subcommand                    │
│                                                                              │
│  cmd/ubik/main.go:newProxyRunCommand()                                       │
│  ├─ configManager.Load() → gets platform URL, auth token                    │
│  ├─ logging.NewLogger() → creates Logger B                                  │
│  │   └─ logger.sessionID = uuid.Nil (ZERO UUID - never changes!)            │
│  │   └─ ⚠️ StartSession() is NEVER called on daemon logger                  │
│  │                                                                           │
│  └─ daemon.RunDaemon(ctx, port, logger)                                     │
│      ├─ sessionManager = NewSessionManager(8100, 8109)                      │
│      ├─ policyEngine = NewPolicyEngine()                                    │
│      ├─ server = NewProxyServer(logger) ← Logger B with ZERO UUID           │
│      ├─ server.SetSessionManager(sessionManager) ← SessionManager IS set    │
│      ├─ server.Start(port)                                                  │
│      └─ controlServer = NewControlServer(sockFile, sessionManager)          │
│          └─ controlServer.Start(ctx)                                        │
│                                                                              │
│  ═══════════════════════════════════════════════════════════════════════════│
│                                                                              │
│  Control API (Unix socket: ~/.ubik/proxy.sock)                               │
│  └─ POST /sessions → handleRegister()                                        │
│      └─ sessionManager.Register(req)                                         │
│          └─ sessions[req.SessionID] = Session{ID: req.SessionID, ...}       │
│          └─ ✓ Session IS stored with REAL UUID                              │
│                                                                              │
│  ═══════════════════════════════════════════════════════════════════════════│
│                                                                              │
│  Proxy Intercept (goproxy MITM on port 8082)                                 │
│  └─ OnRequest(hostRegex).DoFunc()                                            │
│      └─ logRequest(r)                                                        │
│          │                                                                   │
│          ├─ getSessionInfo(r) ← CRITICAL FUNCTION                           │
│          │   ├─ r.Header.Get("X-Ubik-Session") → "" (not set by agent)      │
│          │   ├─ r.Header.Get("X-Ubik-Agent") → "" (not set by agent)        │
│          │   │                                                               │
│          │   └─ if sessionID == "" && s.sessionManager != nil:              │
│          │       sessions := s.sessionManager.ListSessions()                 │
│          │       ├─ ⚠️ QUESTION: Does this return the registered session?  │
│          │       ├─ If len(sessions) == 1: use that session                 │
│          │       └─ If len(sessions) > 1: use most recent                   │
│          │                                                                   │
│          │   return sessionID, agentID                                       │
│          │                                                                   │
│          ├─ payload["session_id"] = sessionID ← Added to payload            │
│          ├─ payload["agent_id"] = agentID                                   │
│          │                                                                   │
│          └─ s.logger.LogEvent("api_request", "proxy", ..., payload)         │
│                                                                              │
│  ═══════════════════════════════════════════════════════════════════════════│
│                                                                              │
│  Logger B (daemon's logger)                                                  │
│  └─ LogEvent(eventType, category, content, metadata)                         │
│      │                                                                       │
│      │  PR #323 Fix (APPLIED):                                              │
│      │  ├─ Check metadata["session_id"] → use if provided                   │
│      │  └─ Check metadata["agent_id"] → use if provided                     │
│      │                                                                       │
│      │  BUT: If getSessionInfo() returns "", metadata has no session_id     │
│      │       and logger falls back to l.sessionID which is ZERO UUID        │
│      │                                                                       │
│      └─ entry := LogEntry{SessionID: sessionID, ...}                        │
│          └─ ⚠️ SessionID = "00000000-0000-0000-0000-000000000000"           │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    │ HTTP POST /api/v1/logs (batched)
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        PLATFORM API + DATABASE                               │
│                                                                              │
│  activity_logs table:                                                        │
│  └─ session_id: "00000000-0000-0000-0000-000000000000" ← WRONG              │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## The Problem Chain

### Two Separate Processes, Two Separate Loggers

```
┌─────────────────────────┐              ┌─────────────────────────┐
│   INTERACTIVE PROCESS   │              │   DAEMON PROCESS        │
│   (ubik command)        │              │   (ubik proxy run)      │
│                         │              │                         │
│   Logger A:             │              │   Logger B:             │
│   sessionID = "550e8"   │   ═══════    │   sessionID = "00000"   │
│   ✓ StartSession()      │   SEPARATE   │   ✗ No StartSession()   │
│   ✓ Logs session_start  │   PROCESSES  │   ✗ Logs api_request    │
│                         │              │     with ZERO UUID      │
│                         │              │                         │
│   NativeRunner:         │──────────────│   SessionManager:       │
│   Registers session     │   IPC via    │   Stores session        │
│   with daemon           │   Unix sock  │   with REAL UUID        │
│                         │              │                         │
│                         │              │   ProxyServer:          │
│                         │              │   getSessionInfo()      │
│                         │              │   ⚠️ Should find it!    │
└─────────────────────────┘              └─────────────────────────┘
```

---

## Fix Attempts

### Attempt 1: PR #322 - getSessionInfo() Helper

**Goal:** Look up session from SessionManager instead of relying on headers.

**Change:** Added `getSessionInfo()` function in `server.go`:
```go
func (s *ProxyServer) getSessionInfo(r *http.Request) (sessionID, agentID string) {
    // Try headers first
    sessionID = r.Header.Get("X-Ubik-Session")
    agentID = r.Header.Get("X-Ubik-Agent")

    // Fall back to SessionManager
    if sessionID == "" && s.sessionManager != nil {
        sessions := s.sessionManager.ListSessions()
        if len(sessions) == 1 {
            sessionID = sessions[0].ID
            agentID = sessions[0].AgentID
        }
    }
    return sessionID, agentID
}
```

**Result:** FAILED - Still seeing zero UUIDs.

---

### Attempt 2: PR #323 - Use session_id from Metadata

**Goal:** Make `LogEvent()` use session_id from payload metadata.

**Change:** Modified `logging/logger.go`:
```go
func (l *loggerImpl) LogEvent(..., metadata map[string]interface{}) {
    sessionID := l.sessionID.String()
    if metadata != nil {
        if sid, ok := metadata["session_id"].(string); ok && sid != "" {
            sessionID = sid  // Use session_id from payload
        }
    }
    // ...
}
```

**Result:** FAILED - Still seeing zero UUIDs.

---

## Root Cause Analysis

The fix chain should work:
1. Session registered → SessionManager stores it ✓
2. `getSessionInfo()` looks up SessionManager → should find it ?
3. `logRequest()` adds to payload → payload["session_id"] = sessionID ?
4. `LogEvent()` checks metadata → uses session_id if present ✓

**The break is likely in step 2 or 3:**
- Either `getSessionInfo()` is not finding the session
- Or the session is not being registered at all

---

## Debug Evidence Needed

### 1. Is Session Being Registered?

**Command:**
```bash
./bin/ubik-cli proxy sessions
```

**Expected:** Should show active session with real UUID.

**If empty:** Session registration is failing.

---

### 2. Is RegisterWithSecurityGateway() Succeeding?

**Look for in terminal output:**
```
Registered with security gateway on port 8100
```

**If you see:**
```
Note: Security gateway not available (...)
```
Then registration FAILED and that's the problem.

---

### 3. Add Debug Logging to getSessionInfo()

In `server.go`, add:
```go
func (s *ProxyServer) getSessionInfo(r *http.Request) (sessionID, agentID string) {
    sessionID = r.Header.Get("X-Ubik-Session")
    agentID = r.Header.Get("X-Ubik-Agent")

    fmt.Printf("DEBUG getSessionInfo: header sessionID=%q agentID=%q\n", sessionID, agentID)

    if sessionID == "" && s.sessionManager != nil {
        sessions := s.sessionManager.ListSessions()
        fmt.Printf("DEBUG getSessionInfo: sessionManager has %d sessions\n", len(sessions))
        for i, s := range sessions {
            fmt.Printf("DEBUG getSessionInfo: session[%d] ID=%s AgentID=%s\n", i, s.ID, s.AgentID)
        }
        // ... rest of logic
    }

    fmt.Printf("DEBUG getSessionInfo: returning sessionID=%q agentID=%q\n", sessionID, agentID)
    return sessionID, agentID
}
```

---

### 4. Check If Daemon is Using Correct Binary

The daemon is started by `exec.Command(execPath, "proxy", "run", ...)`.

If you rebuild the CLI but don't restart the daemon, it keeps running the OLD binary!

**Fix:**
```bash
./bin/ubik-cli proxy stop
make build-cli
./bin/ubik-cli proxy start
```

---

## Component Files Reference

| Component | File | Key Functions |
|-----------|------|---------------|
| Interactive Mode | `cmd/ubik/main.go` | `runInteractiveMode()`, `newProxyRunCommand()` |
| Native Runner | `internal/native_runner.go` | `Run()`, `RegisterWithSecurityGateway()` |
| Control Client | `internal/httpproxy/control_client.go` | `RegisterSession()`, `NewDefaultControlClient()` |
| Control Server | `internal/httpproxy/control.go` | `handleRegister()`, `handleSessions()` |
| Session Manager | `internal/httpproxy/session_manager.go` | `Register()`, `ListSessions()`, `GetByID()` |
| Proxy Server | `internal/httpproxy/server.go` | `logRequest()`, `logResponse()`, `getSessionInfo()` |
| Daemon | `internal/httpproxy/daemon.go` | `RunDaemon()`, `EnsureRunning()` |
| Logger | `internal/logging/logger.go` | `LogEvent()`, `StartSession()` |

---

## Hypothesis: Session Not Registered

Most likely, the session registration is failing silently:

```go
// native_runner.go:Run()
sessionResp, err := r.RegisterWithSecurityGateway(config)
if err != nil {
    // Security gateway not running - this is optional for now
    fmt.Fprintf(stderr, "Note: Security gateway not available (%v)\n", err)
    // ← CONTINUES WITHOUT SESSION REGISTRATION!
} else {
    config.ProxyPort = sessionResp.Port
    fmt.Fprintf(stderr, "Registered with security gateway on port %d\n", sessionResp.Port)
}
```

If you don't see "Registered with security gateway", the session is NOT being registered, and `getSessionInfo()` will never find it.

---

## Potential Fixes (Not Yet Implemented)

### Fix A: Make Registration Required (Fail-Closed)
```go
if err != nil {
    return fmt.Errorf("failed to register with security gateway: %w", err)
}
```

### Fix B: Set Session ID Directly on Daemon Logger
```go
// After registration, tell daemon to use this session ID
controlClient.SetActiveSessionID(sessionID)
```

### Fix C: Use Shared State File
Write session ID to `~/.ubik/current_session.json` and have daemon read it.

### Fix D: Inject Session Header in Proxy
Have the proxy inject `X-Ubik-Session` header based on registered session, so even if agent doesn't set it, proxy can use it.

---

## Summary

**Problem:** API request logs have zero session_id.

**Root Cause:** The session lookup chain is broken somewhere between:
1. Session registration (NativeRunner → ControlClient → ControlServer → SessionManager)
2. Session lookup (ProxyServer → getSessionInfo() → SessionManager)

**Next Step:** Add debug logging to identify exactly where the chain breaks.
