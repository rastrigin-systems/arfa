# CLI Client Development Guide

**You are working on the Ubik CLI** - the command-line interface for employee configuration sync.

---

## Quick Context

**What is this service?**
Self-contained Go CLI that allows employees to sync AI agent configurations from the Ubik Enterprise platform to their local development machines.

**Key capabilities:**
- Secure authentication (login/logout with JWT)
- Configuration sync (download agent configs, MCP settings)
- Docker integration (manage MCP server containers)
- Interactive mode (user-friendly interface)
- Agent management (list, inspect configured agents)
- Activity logging (track CLI usage)

**Architecture principle:** Self-contained module with minimal dependencies. NO database code, NO generated API code. Only depends on `pkg/types` for shared data structures.

---

## Essential Commands

```bash
# From services/cli/ directory
make build              # Build CLI to ../../bin/ubik-cli
make test               # Run all tests
make test-unit          # Unit tests only (fast)
make test-integration   # Integration tests
make coverage           # View coverage report
make install            # Install to /usr/local/bin/ubik
make uninstall          # Remove from system

# From repository root
make build-cli          # Build CLI
make install-cli        # Install CLI to system
make uninstall-cli      # Remove from system
```

### Using the CLI

```bash
# Authentication
ubik login              # Login to platform
ubik logout             # Clear session

# Configuration sync
ubik sync               # Sync all configurations
ubik sync --agent claude  # Sync specific agent
ubik sync --dry-run     # Show what would be synced

# Agent management
ubik agents             # List configured agents
ubik agents show claude-code  # Show agent details

# Interactive mode
ubik                    # Launch interactive interface
```

---

## Architecture

### Design Decisions

**Self-Contained Module:**
- Separate Go module from main workspace
- Minimal dependencies (no DB drivers, no generated code)
- Small binary size (~10MB vs ~25MB for API)
- Faster build times

**What CLI Depends On:**
- ✅ `pkg/types` - Shared data structures
- ✅ `github.com/docker/docker` - Docker SDK
- ✅ `github.com/spf13/cobra` - CLI framework
- ❌ `generated/db` - NO database code
- ❌ `generated/api` - NO generated API code
- ❌ PostgreSQL drivers - NO direct DB access

### Directory Structure

```
services/cli/
├── cmd/ubik/main.go    # CLI entry point, Cobra commands
├── internal/           # Implementation (not importable)
│   ├── commands/       # Command implementations
│   ├── logging/        # Activity logging to platform
│   ├── auth.go         # JWT token management
│   ├── sync.go         # Configuration sync
│   ├── agents.go       # Agent management
│   ├── docker.go       # Docker SDK wrapper
│   ├── container.go    # Container lifecycle
│   ├── proxy.go        # MCP proxy server
│   ├── config.go       # Local config management
│   └── workspace.go    # Workspace detection
└── tests/
    ├── integration/    # Integration tests
    └── e2e/           # End-to-end tests
```

### Request Flow

```
CLI Command → Cobra → Internal Package → HTTP Request → API Server
                           ↓
                     Local Config (~/.ubik/)
                           ↓
                     Docker Container Management
```

---

## Testing Strategy

**CRITICAL: ALWAYS follow strict TDD (Test-Driven Development)**

### TDD Workflow (Mandatory)
1. ✅ Write failing test FIRST
2. ✅ Implement minimal code to pass test
3. ✅ Refactor with tests passing
4. ❌ NEVER write implementation before tests

### Test Types

**Unit Tests** (`internal/*_test.go`):
- Test individual functions
- Mock HTTP client, Docker SDK, filesystem
- Fast execution (<1s)
- Target: 85%+ coverage

**Integration Tests** (`tests/integration/`):
- Test CLI + API interaction
- May use test containers
- Test file operations
- Slower execution (~5-10s)

**E2E Tests** (`tests/e2e/`):
- Test complete user workflows
- Require running API server
- Test Docker container management
- Slowest execution (~10-30s)

### Running Tests

```bash
# Fast feedback loop
make test-unit          # ~1 second

# Full test suite
make test               # ~10-20 seconds

# Integration tests
make test-integration   # ~5-10 seconds

# E2E tests (requires API)
make test-e2e          # ~10-30 seconds

# Coverage report
make coverage          # Opens HTML report
```

### Test Patterns

**Mocking HTTP client:**
```go
func TestLogin(t *testing.T) {
    // Setup mock server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(types.LoginResponse{Token: "test-token"})
    }))
    defer server.Close()

    // Test login
    auth := NewAuthClient(server.URL)
    token, err := auth.Login("user@example.com", "password")

    assert.NoError(t, err)
    assert.Equal(t, "test-token", token)
}
```

**Mocking Docker SDK:**
```go
func TestStartContainer(t *testing.T) {
    ctrl := gomock.NewController(t)
    mockDocker := mocks.NewMockDockerClient(ctrl)

    // Expect container start
    mockDocker.EXPECT().
        ContainerStart(gomock.Any(), "container-id", gomock.Any()).
        Return(nil)

    // Test
    cm := NewContainerManager(mockDocker)
    err := cm.StartContainer("container-id")

    assert.NoError(t, err)
}
```

**See [../../docs/TESTING.md](../../docs/TESTING.md) for complete testing guide.**

---

## Common Tasks

### Adding New Command

1. **Define command in `cmd/ubik/main.go`:**
   ```go
   var statusCmd = &cobra.Command{
       Use:   "status",
       Short: "Show sync status",
       Run: func(cmd *cobra.Command, args []string) {
           // Call internal package
       },
   }
   ```

2. **Write tests FIRST:**
   ```go
   // internal/sync_test.go
   func TestGetSyncStatus(t *testing.T) {
       // Write failing test
   }
   ```

3. **Implement logic:**
   ```go
   // internal/sync.go
   func GetSyncStatus() (*SyncStatus, error) {
       // Implement to pass tests
   }
   ```

4. **Register command:**
   ```go
   // cmd/ubik/main.go
   func init() {
       rootCmd.AddCommand(statusCmd)
   }
   ```

### Adding Docker Integration

1. **Write tests with mock Docker client:**
   ```go
   func TestDockerOperation(t *testing.T) {
       mockDocker := setupMockDocker(t)
       // Test Docker operations
   }
   ```

2. **Implement using Docker SDK:**
   ```go
   // internal/docker.go
   func (d *DockerClient) PullImage(image string) error {
       // Use docker.Client API
   }
   ```

3. **Handle errors gracefully:**
   ```go
   if err != nil {
       return fmt.Errorf("failed to pull image %s: %w", image, err)
   }
   ```

### Configuration Management

**Local config location:** `~/.ubik/`

**Structure:**
```
~/.ubik/
├── config.json          # CLI configuration
├── auth.json           # Auth tokens (chmod 600)
├── agents/             # Synced agent configs
│   ├── claude-code/
│   │   ├── config.json
│   │   └── mcps/
│   └── cursor/
└── logs/               # Activity logs
```

**Reading config:**
```go
config, err := config.Load(config.DefaultConfigDir())
```

**Saving config:**
```go
err := config.Save(config.DefaultConfigDir(), cfg)
```

---

## Common Pitfalls

### 1. Stale CLI Binary
```bash
# ❌ Testing with old binary
ubik sync

# ✅ Rebuild and reinstall
make clean && make build && make install
ubik sync
```

### 2. Cache Issues
```bash
# ✅ Clear CLI cache
rm -rf ~/.ubik/

# ✅ Clear test cache
go clean -testcache
```

### 3. Docker Not Running
```bash
# ✅ Check Docker
docker ps

# ✅ Start Docker Desktop (macOS)
open -a Docker
```

### 4. API Connection Issues
```bash
# ✅ Verify API is running
curl http://localhost:8080/api/v1/health

# ✅ Check API URL in config
cat ~/.ubik/config.json
```

### 5. Import Errors
```bash
# ❌ NEVER import generated code
import "github.com/sergeirastrigin/ubik-enterprise/generated/db"  # NO!

# ✅ Only import pkg/types
import "github.com/sergeirastrigin/ubik-enterprise/pkg/types"  # YES!
```

**See [../../docs/DEBUGGING.md](../../docs/DEBUGGING.md) for debugging strategies.**

---

## Docker Integration

### Container Naming Convention
```
ubik-mcp-<employee-id>-<mcp-name>
```

Example: `ubik-mcp-emp123-postgres`

### Container Lifecycle

**Starting containers:**
```go
func (cm *ContainerManager) StartMCPContainer(mcp types.MCPConfig) error {
    // 1. Pull image if needed
    // 2. Create container with proper config
    // 3. Start container
    // 4. Wait for health check
}
```

**Stopping containers:**
```go
func (cm *ContainerManager) StopMCPContainer(containerID string) error {
    // 1. Stop container gracefully
    // 2. Remove container
    // 3. Clean up volumes
}
```

### Health Checks

```go
func (cm *ContainerManager) WaitForHealthy(containerID string, timeout time.Duration) error {
    // Poll container health status
    // Return error if timeout
}
```

---

## Debugging

**Golden Rule: Check the cache, not just the code**

### Quick Debug Checklist
1. ✅ Rebuild CLI: `make clean && make build`
2. ✅ Clear cache: `rm -rf ~/.ubik/`
3. ✅ Check Docker: `docker ps`
4. ✅ Verify API: `curl http://localhost:8080/api/v1/health`
5. ✅ Check logs: `cat ~/.ubik/logs/*.log`

### Debug Logging

```bash
# Enable debug logging
export UBIK_DEBUG=true
ubik sync

# Set log level
export UBIK_LOG_LEVEL=debug
ubik sync

# View debug output
tail -f ~/.ubik/logs/ubik.log
```

### Testing with Local API

```bash
# Terminal 1: Start API
cd ../../ && make dev-api

# Terminal 2: Test CLI
cd services/cli
make build
../../bin/ubik-cli login --api-url http://localhost:8080
```

---

## Installation & Distribution

### Local Installation

```bash
# Build and install
make build && make install

# Verify
ubik version
which ubik  # Should show /usr/local/bin/ubik
```

### Uninstallation

```bash
# Remove binary
make uninstall

# Remove config (optional)
rm -rf ~/.ubik/
```

### Cross-Platform Builds (Future)

```bash
# Build for all platforms
make build-all

# Outputs:
# bin/ubik-darwin-amd64
# bin/ubik-darwin-arm64
# bin/ubik-linux-amd64
# bin/ubik-windows-amd64.exe
```

---

## Performance

**Binary size:** ~10MB (uncompressed)
**Startup time:** ~50ms (cold start)
**Sync time:** ~1-2s (initial), ~200-500ms (incremental)

**Optimization tips:**
- Use connection pooling for HTTP client
- Cache Docker client instance
- Minimize filesystem operations
- Batch API requests when possible

---

## Related Documentation

**Root Documentation:**
- [../../CLAUDE.md](../../CLAUDE.md) - Monorepo overview, critical rules
- [../../docs/QUICKSTART.md](../../docs/QUICKSTART.md) - First-time setup
- [../../docs/QUICK_REFERENCE.md](../../docs/QUICK_REFERENCE.md) - Command reference

**Development:**
- [../../docs/DEVELOPMENT.md](../../docs/DEVELOPMENT.md) - Development workflow
- [../../docs/DEV_WORKFLOW.md](../../docs/DEV_WORKFLOW.md) - PR workflow (mandatory)
- [../../docs/TESTING.md](../../docs/TESTING.md) - Complete testing guide
- [../../docs/DEBUGGING.md](../../docs/DEBUGGING.md) - Debugging strategies

**CLI Specific:**
- [../../docs/CLI_CLIENT.md](../../docs/CLI_CLIENT.md) - CLI architecture
- [../../docs/CLI_PHASE4_COMPLETE.md](../../docs/CLI_PHASE4_COMPLETE.md) - Latest phase details

**Other Services:**
- [../api/CLAUDE.md](../api/CLAUDE.md) - API server development
- [../web/CLAUDE.md](../web/CLAUDE.md) - Web UI development

---

**Quick Links:**
- API Docs (local): http://localhost:8080/api/docs
- CLI Config: `~/.ubik/config.json`
- CLI Logs: `~/.ubik/logs/`
