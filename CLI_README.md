# ubik CLI - Employee Agent Management

**Version:** v0.2.0-dev
**Status:** Phase 2 Complete ✅ (Docker Integration)
**Location:** `pivot/cmd/cli/`

---

## Quick Start

### Build & Test

```bash
# Build CLI
make build-cli

# Run tests
make test-cli

# Try it out
./bin/ubik-cli --help
./bin/ubik-cli --version
```

### Usage

```bash
# 1. Login to platform
./bin/ubik-cli login
Platform URL [https://api.ubik.io]: http://localhost:8080
Email: alice@acme.com
Password: ****

# 2. Check status
./bin/ubik-cli status

# 3. Sync configs from platform
./bin/ubik-cli sync

# 4. View local config
./bin/ubik-cli config

# 5. Logout
./bin/ubik-cli logout
```

---

## Available Commands

| Command | Description | Status |
|---------|-------------|--------|
| `ubik login` | Authenticate with platform | ✅ Working |
| `ubik logout` | Clear stored credentials | ✅ Working |
| `ubik sync` | Fetch configs from platform | ✅ Working |
| `ubik sync --start-containers` | Fetch configs and start Docker containers | ✅ Working |
| `ubik start` | Start Docker containers | ✅ Working |
| `ubik stop` | Stop Docker containers | ✅ Working |
| `ubik config` | View local configuration | ✅ Working |
| `ubik status` | Show status with container info | ✅ Working |
| `ubik --version` | Show CLI version | ✅ Working |
| `ubik --help` | Show help | ✅ Working |

---

## What's Implemented (Phases 1 & 2)

✅ **Authentication** (Phase 1)
- Interactive login (prompts for credentials)
- Non-interactive login (via flags)
- JWT token management
- Session persistence

✅ **Configuration Management** (Phase 1)
- Local config storage (`~/.ubik/config.json`)
- Agent config sync from platform
- Multi-agent support

✅ **Platform Integration** (Phase 1)
- REST API client
- Employee authentication
- Resolved config fetching

✅ **Docker Integration** (Phase 2)
- Docker SDK integration
- Container lifecycle management (start/stop)
- Network management (`ubik-network`)
- MCP server orchestration
- Agent container management
- Volume mounting for workspaces
- Environment variable injection
- Container status display

✅ **Testing** (Phases 1 & 2)
- 24 unit tests (fast, no Docker required)
- 18 integration tests (require Docker)
- **42 total tests** (100% pass rate)
- ~22% coverage (unit only), ~60-70% coverage (with Docker)
- Comprehensive error handling and edge cases
- See **[docs/CLI_TEST_SUMMARY.md](./docs/CLI_TEST_SUMMARY.md)** for details

---

## What's Coming Next (Phase 3)

🎯 **Interactive Mode & I/O Proxying** (3-4 days)
- Interactive workspace selection
- I/O proxying to agent container (stdin/stdout)
- TTY mode for interactive sessions
- Agent switching on-the-fly
- Session management

See [docs/CLI_CLIENT.md](./docs/CLI_CLIENT.md) for complete roadmap.

---

## Configuration Files

### Local CLI Config
```
~/.ubik/config.json
{
  "platform_url": "https://api.ubik.io",
  "token": "eyJhbGc...",
  "employee_id": "uuid",
  "default_agent": "claude-code",
  "last_sync": "2025-10-29T10:30:00Z"
}
```

### Agent Configs
```
~/.ubik/agents/{agent-id}/
├── config.json           # Agent configuration
└── mcp-servers.json      # MCP server configs
```

---

## Development

### Project Structure
```
pivot/
├── cmd/cli/
│   └── main.go              # CLI entry point
│
├── internal/cli/
│   ├── auth.go              # Authentication service
│   ├── auth_test.go         # Auth tests
│   ├── config.go            # Config manager
│   ├── config_test.go       # Config tests
│   ├── platform.go          # Platform API client
│   ├── sync.go              # Sync service
│   └── sync_test.go         # Sync tests
│
└── bin/
    └── ubik-cli             # Compiled binary
```

### Testing
```bash
# Run unit tests only (fast, no Docker)
make test-cli
go test ./internal/cli/... -short -v

# Run all tests including Docker integration
go test ./internal/cli/... -v

# Run specific test
go test -v ./internal/cli/ -run TestAuthService_Login

# Run with coverage (unit tests only)
go test -short -race -coverprofile=coverage.out ./internal/cli/...
go tool cover -html=coverage.out

# Run with coverage (all tests including Docker)
go test -race -coverprofile=coverage-full.out ./internal/cli/...
go tool cover -html=coverage-full.out
```

### Building
```bash
# Build CLI only
make build-cli

# Build all binaries (server + CLI)
make build

# Clean build artifacts
make clean
```

---

## Dependencies

- `github.com/spf13/cobra` - CLI framework
- `golang.org/x/term` - Password input masking
- Standard library (net/http, encoding/json, etc.)

---

## Platform API Requirements

The CLI requires these API endpoints to be implemented on the server:

1. ✅ `POST /auth/login` - User authentication
2. ✅ `GET /employees/{id}` - Employee info
3. ⚠️ `GET /employees/{id}/agent-configs/resolved` - Resolved configs (may not be fully implemented yet)

---

## Troubleshooting

### Authentication Issues

**Problem:** Login fails with "authentication failed"
**Solution:**
- Verify platform URL is correct
- Check if server is running
- Confirm email/password are correct

### Config Issues

**Problem:** Sync fails with "not authenticated"
**Solution:** Run `ubik login` first

**Problem:** Sync returns empty configs
**Solution:**
- Verify you have agent configs assigned in the platform
- Check with admin if you need approval for agents

### Build Issues

**Problem:** Build fails with missing dependencies
**Solution:** Run `go mod tidy`

---

## Documentation

- 📘 [CLI_CLIENT.md](./docs/CLI_CLIENT.md) - Complete architecture & design
- 📗 [CLI_PHASE1_COMPLETE.md](./docs/CLI_PHASE1_COMPLETE.md) - Phase 1 summary
- 📙 [CLAUDE.md](./CLAUDE.md) - Main project documentation

---

## Contributing

When working on the CLI:

1. **Add tests first** (TDD approach)
2. **Keep commands simple** - one responsibility per command
3. **Handle errors gracefully** - provide helpful messages
4. **Update docs** - keep README and CLI_CLIENT.md in sync

---

## Status Summary

**Phase 0:** ✅ Docker Images Complete
**Phase 1:** ✅ Foundation Complete (Authentication, Config, Sync)
**Phase 2:** ✅ Docker Integration Complete (Containers, Networks, Orchestration)
**Testing:** ✅ 42 tests (24 unit + 18 integration, 100% passing)
**Phase 3:** 🎯 Interactive Mode (Next)
**Phase 4:** 📅 Agent Management (Planned)
**Phase 5:** 📅 Polish & Telemetry (Planned)

---

**Built with ❤️ for efficient AI agent management**
