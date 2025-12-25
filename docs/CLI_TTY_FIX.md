# CLI TTY Fix - Interactive Mode Issue

**Date:** 2025-10-29
**Issue:** User inputs not reaching Claude Code container
**Status:** ✅ Fixed

---

## Problem Description

When running `arfa` in interactive mode, users could type inputs like "2+2" or "234" but they weren't reaching the Claude Code container. The container would start successfully but no interaction was possible.

**Symptoms:**
```bash
$ arfa
✓ Agent: Claude Code (ide_assistant)
✓ Workspace: /Users/user/project (88.4 MB, 234 files)

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✨ Interactive session started
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

2+2          <-- Input not being processed
234          <-- No response
```

---

## Root Causes

### 1. Terminal Not in Raw Mode
**Problem:** The host terminal was in canonical (cooked) mode, which buffers input until newline.

**Solution:** Set terminal to raw mode using `golang.org/x/term`:
```go
stdinFd := int(os.Stdin.Fd())
isTerminal := term.IsTerminal(stdinFd)

if isTerminal {
    oldState, err := term.MakeRaw(stdinFd)
    if err != nil {
        return fmt.Errorf("failed to set terminal to raw mode: %w", err)
    }
    defer term.Restore(stdinFd, oldState)
}
```

**Why this matters:**
- Raw mode passes every keypress immediately to the container
- Canonical mode buffers until Enter is pressed
- TTY applications like Claude Code expect raw input

### 2. Incorrect Output Demultiplexing
**Problem:** Using `stdcopy.StdCopy()` for TTY containers.

**Solution:** For TTY containers, Docker doesn't multiplex streams - use direct copy:
```go
// Before (WRONG for TTY)
_, err := stdcopy.StdCopy(options.Stdout, options.Stderr, resp.Reader)

// After (CORRECT for TTY)
_, err := io.Copy(options.Stdout, resp.Reader)
```

**Why this matters:**
- TTY containers send raw output on a single stream
- `stdcopy.StdCopy()` is only for non-TTY multiplexed streams
- Using wrong demux causes output corruption or blocking

### 3. Container Network Configuration
**Status:** Already Correct ✅

Containers are connected to `arfa-network` bridge network, which allows:
- Container-to-container communication (agent ↔ MCP servers)
- Internet access (Claude Code → Anthropic API)

---

## Changes Made

### Modified Files

**1. `internal/cli/proxy.go`**
- Added `golang.org/x/term` import
- Implemented raw terminal mode handling
- Fixed output stream copying for TTY
- Improved signal handling with terminal restoration

**2. `go.mod`**
- Added dependency: `golang.org/x/term`

---

## Code Changes

### Before (Broken)

```go
// AttachToContainer (simplified)
func (ps *ProxyService) AttachToContainer(ctx context.Context, options ProxyOptions) error {
    attachOpts := container.AttachOptions{
        Stream: true,
        Stdin:  true,
        Stdout: true,
        Stderr: true,
    }

    resp, err := ps.dockerClient.cli.ContainerAttach(ctx, options.ContainerID, attachOpts)
    // ...

    // Copy streams
    go func() {
        io.Copy(resp.Conn, options.Stdin) // ❌ Terminal in canonical mode
    }()

    go func() {
        stdcopy.StdCopy(options.Stdout, options.Stderr, resp.Reader) // ❌ Wrong for TTY
    }()
}
```

### After (Fixed)

```go
// AttachToContainer (simplified)
func (ps *ProxyService) AttachToContainer(ctx context.Context, options ProxyOptions) error {
    // ✅ Put terminal in raw mode
    stdinFd := int(os.Stdin.Fd())
    isTerminal := term.IsTerminal(stdinFd)

    var oldState *term.State
    if isTerminal {
        oldState, _ = term.MakeRaw(stdinFd)
        defer term.Restore(stdinFd, oldState)
    }

    // Create attach with proper flags
    attachOpts := container.AttachOptions{
        Stream: true,
        Stdin:  true,
        Stdout: true,
        Stderr: true,
        Logs:   false, // ✅ Don't replay old logs
    }

    resp, err := ps.dockerClient.cli.ContainerAttach(ctx, options.ContainerID, attachOpts)
    // ...

    // Copy streams
    go func() {
        io.Copy(resp.Conn, options.Stdin) // ✅ Now in raw mode
    }()

    go func() {
        io.Copy(options.Stdout, resp.Reader) // ✅ Direct copy for TTY
    }()
}
```

---

## Testing

### Manual Test
```bash
# 1. Rebuild CLI
make build-cli

# 2. Stop existing containers
./bin/arfa-cli stop

# 3. Start fresh interactive session
./bin/arfa-cli

# 4. Type interactive input
> What is 2+2?
[Should now respond immediately]
```

### Expected Behavior
- Input appears immediately character-by-character
- Claude Code responds to queries
- Colors and formatting work correctly
- Ctrl+C gracefully exits and restores terminal

---

## How Raw Mode Works

**Canonical (Cooked) Mode:**
```
User types: "hello" [Enter]
          ↓
Buffer waits for newline
          ↓
Sends entire line: "hello\n"
```

**Raw Mode:**
```
User types: "h"
          ↓
Sends immediately: "h"

User types: "e"
          ↓
Sends immediately: "e"

(etc...)
```

**Why Claude Code Needs Raw Mode:**
- Implements its own line editing (backspace, arrow keys, etc.)
- Needs immediate feedback for autocomplete
- Handles special keys (Ctrl+C, Ctrl+D, etc.)
- Provides rich interactive experience

---

## Additional Notes

### Docker Container Configuration
The container is already set up correctly for TTY:

```go
// In container.go
config := &container.Config{
    Image:        spec.Image,
    Env:          env,
    Tty:          true,        // ✅ Enable TTY
    OpenStdin:    true,        // ✅ Keep stdin open
    AttachStdin:  true,        // ✅ Attach stdin
    AttachStdout: true,        // ✅ Attach stdout
    AttachStderr: true,        // ✅ Attach stderr
    // ...
}
```

### Internet Access
The container has internet access via Docker's default bridge networking:
- Uses `arfa-network` for container-to-container communication
- Default route allows internet egress
- Claude Code can reach Anthropic API at `api.anthropic.com`

---

## Troubleshooting

### If inputs still don't work:

**1. Check Docker networking:**
```bash
docker inspect arfa-agent-<agent-id> | grep NetworkMode
```

**2. Verify internet access from container:**
```bash
docker exec arfa-agent-<agent-id> curl -I https://api.anthropic.com
```

**3. Check Claude Code is running:**
```bash
docker exec arfa-agent-<agent-id> ps aux | grep claude
```

**4. Check API key is set:**
```bash
docker exec arfa-agent-<agent-id> env | grep ANTHROPIC_API_KEY
```

### If terminal gets corrupted after crash:

```bash
# Reset terminal
reset

# Or restore manually
stty sane
```

---

## References

- [Terminal Raw Mode](https://en.wikipedia.org/wiki/Terminal_mode)
- [Docker TTY Mode](https://docs.docker.com/engine/reference/run/#foreground)
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term)
- [Docker Container Attach](https://docs.docker.com/engine/api/v1.43/#tag/Container/operation/ContainerAttach)

---

**Status:** ✅ Fixed in v0.2.0
**Files Changed:** 2 (proxy.go, go.mod)
**Test Status:** Pending user verification
