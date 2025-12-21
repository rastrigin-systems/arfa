# CLI Client Development Guide

**You are working on the Ubik CLI** - the command-line interface for employee configuration sync.

---

## Quick Context

**What is this service?**
Self-contained Go CLI that allows employees to sync AI agent configurations from the Ubik Enterprise platform to their local development machines.

**Key capabilities:**
- Secure authentication (login/logout with JWT)
- Configuration sync (download agent configs)
- Docker integration (manage server containers)
- Interactive mode (user-friendly interface)
- Agent management (list, inspect configured agents)
- Activity logging (track CLI usage)
- **Control Service** - HTTPS proxy for LLM API interception, logging, and policy enforcement
- **Tool Blocking** (in progress) - Block tool calls based on org policies. See [TOOL_BLOCKING_DESIGN.md](../../docs/TOOL_BLOCKING_DESIGN.md)

**Architecture principle:** Self-contained module with minimal dependencies. NO database code, NO generated API code. Only depends on `pkg/types` for shared data structures.

---

## Essential Commands

Run `make` to see available commands (from services/cli/ or repository root).

---

## Critical: Pre-Commit Checks

**ALWAYS run these before committing Go files:**

```bash
# 1. Run go vet (required)
go vet ./...

# 2. Run tests
go test ./... -count=1
```

❌ **NEVER commit without running `go vet`** - this catches common bugs and issues.

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
├── cmd/ubik/main.go    # CLI entry point, creates DI container
├── internal/           # Implementation (not importable)
│   ├── commands/       # Command implementations (use container)
│   │   ├── root.go     # Root command, distributes container
│   │   ├── auth/       # Auth commands (login, logout)
│   │   ├── agents/     # Agent commands (list, info, show)
│   │   ├── sync/       # Sync command
│   │   └── ...         # Other command groups
│   ├── container/      # Dependency injection container
│   │   └── container.go
│   ├── mocks/          # Generated gomock interfaces
│   │   └── interfaces_mock.go
│   ├── interfaces.go   # Service interface definitions
│   ├── types_api.go    # API request/response types (auth, agent, sync, skill, log)
│   ├── auth.go         # AuthService implementation
│   ├── sync.go         # SyncService implementation
│   ├── agents.go       # AgentService implementation
│   ├── docker.go       # DockerClient implementation
│   ├── container_manager.go  # ContainerManager (Docker containers)
│   ├── config.go       # ConfigManager implementation
│   ├── platform.go     # PlatformClient (HTTP API client only)
│   ├── proxy.go        # ProxyService implementation
│   ├── native_runner.go # NativeRunner for agent processes
│   └── ...             # Other implementations
└── tests/
    ├── integration/    # Integration tests
    └── e2e/           # End-to-end tests
```

### Type Organization

**API types (`types_api.go`)** - All request/response types for platform API:
- Authentication: `LoginRequest`, `LoginResponse`, `LoginEmployeeInfo`, `EmployeeInfo`
- Agent configs: `AgentConfig`, `AgentConfigAPIResponse`, `MCPServerConfig`, etc.
- Sync types: `ClaudeCodeSyncResponse`, `AgentConfigSync`, `SkillConfigSync`, `MCPServerConfigSync`
- Token types: `ClaudeTokenStatusResponse`, `EffectiveClaudeTokenResponse`
- Skill types: `Skill`, `SkillFile`, `EmployeeSkill`, `ListSkillsResponse`
- Log types: `LogEntry`, `CreateLogRequest`

**Platform client (`platform.go`)** - HTTP API client methods only, no type definitions.

### Dependency Injection Architecture

**The CLI uses a DI container pattern for clean, testable code:**

```go
// main.go - Create container and pass to root command
func main() {
    c := container.New()
    defer c.Close()
    commands.NewRootCommand(version, c).Execute()
}

// Command receives container and gets services from it
func NewLoginCommand(c *container.Container) *cobra.Command {
    return &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            authService, err := c.AuthService()
            if err != nil {
                return err
            }
            return authService.Login(ctx, url, email, password)
        },
    }
}
```

**Key interfaces** (defined in `internal/interfaces.go`):
- `ConfigManagerInterface` - Local config (~/.ubik/config.json)
- `PlatformClientInterface` - HTTP API communication
- `AuthServiceInterface` - Authentication operations
- `SyncServiceInterface` - Configuration sync
- `AgentServiceInterface` - Agent management
- `DockerClientInterface` - Docker SDK wrapper
- `ContainerManagerInterface` - Container lifecycle

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

**Using gomock for service tests:**
```go
func TestAuthService_RequireAuth(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
    mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

    // Set expectations
    mockConfigManager.EXPECT().IsAuthenticated().Return(true, nil)
    mockConfigManager.EXPECT().IsTokenValid().Return(true, nil)
    mockConfigManager.EXPECT().Load().Return(&Config{Token: "test"}, nil)
    mockPlatformClient.EXPECT().SetToken("test")
    mockPlatformClient.EXPECT().SetBaseURL(gomock.Any())

    // Create service with interface-based constructor
    authService := NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

    // Test
    config, err := authService.RequireAuth()
    require.NoError(t, err)
    assert.Equal(t, "test", config.Token)
}
```

**Regenerating mocks:**
```bash
# From services/cli directory
make mocks

# Or directly
go generate ./internal/mocks/...
```

**See [../../docs/TESTING.md](../../docs/TESTING.md) for complete testing guide.**

---

## Common Tasks

### Adding New Command

1. **Create command file accepting container:**
   ```go
   // internal/commands/myfeature/myfeature.go
   func NewMyFeatureCommand(c *container.Container) *cobra.Command {
       return &cobra.Command{
           Use:   "myfeature",
           Short: "My new feature",
           RunE: func(cmd *cobra.Command, args []string) error {
               // Get services from container
               authService, err := c.AuthService()
               if err != nil {
                   return fmt.Errorf("failed to get auth service: %w", err)
               }

               // Use services
               config, err := authService.RequireAuth()
               if err != nil {
                   return err
               }

               // Implement feature logic
               return nil
           },
       }
   }
   ```

2. **Register command in root.go:**
   ```go
   // internal/commands/root.go
   rootCmd.AddCommand(myfeature.NewMyFeatureCommand(c))
   ```

3. **Write tests with mocks:**
   ```go
   // internal/commands/myfeature/myfeature_test.go
   func TestMyFeatureCommand(t *testing.T) {
       ctrl := gomock.NewController(t)
       defer ctrl.Finish()

       mockAuth := mocks.NewMockAuthServiceInterface(ctrl)
       mockAuth.EXPECT().RequireAuth().Return(&Config{}, nil)

       c := container.New(container.WithAuthService(mockAuth))
       cmd := NewMyFeatureCommand(c)
       // Execute and assert
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
import "github.com/rastrigin-systems/ubik-enterprise/generated/db"  # NO!

# ✅ Only import pkg/types
import "github.com/rastrigin-systems/ubik-enterprise/pkg/types"  # YES!
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

- [../../CLAUDE.md](../../CLAUDE.md) - Monorepo overview
- [../../docs/TESTING.md](../../docs/TESTING.md) - Testing guide
- [../../docs/DEV_WORKFLOW.md](../../docs/DEV_WORKFLOW.md) - PR workflow
- [../../docs/TOOL_BLOCKING_DESIGN.md](../../docs/TOOL_BLOCKING_DESIGN.md) - Tool blocking architecture (in progress)
- [../api/CLAUDE.md](../api/CLAUDE.md) - API development
- [../web/CLAUDE.md](../web/CLAUDE.md) - Web UI development
