# Ubik CLI Service

**Command-line interface for syncing AI agent configurations from the Ubik Enterprise platform.**

---

## Overview

The Ubik CLI is a self-contained Go application that allows employees to sync their AI agent configurations from the centralized Ubik Enterprise platform to their local development machines. It handles authentication, configuration synchronization, Docker container management, and provides an interactive mode for easy use.

### Key Features

- 🔐 **Authentication**: Secure login with JWT tokens, session management
- 🔄 **Config Sync**: Sync agent configurations and MCP server settings
- 🐳 **Docker Integration**: Manage MCP server containers automatically
- 💻 **Interactive Mode**: User-friendly interactive interface
- 📊 **Activity Logging**: Track CLI usage and sync operations
- 🎯 **Agent Management**: List and inspect configured agents

---

## Architecture

```
services/cli/
├── cmd/ubik/          # Main CLI entry point
│   └── main.go        # Cobra commands, CLI interface
│
├── internal/          # Internal packages (not importable by other services)
│   ├── commands/      # Command implementations
│   ├── logging/       # Activity logging to platform
│   ├── auth.go        # Authentication (login, logout, tokens)
│   ├── sync.go        # Configuration sync logic
│   ├── agents.go      # Agent management
│   ├── docker.go      # Docker SDK integration
│   ├── container.go   # Container lifecycle management
│   ├── proxy.go       # MCP proxy server
│   ├── config.go      # Local config management
│   ├── workspace.go   # Workspace detection
│   └── skills.go      # Skills integration
│
├── tests/
│   ├── integration/   # Integration tests
│   ├── e2e/          # End-to-end tests
│   └── unit/         # (Unit tests live alongside code in internal/)
│
├── build/            # Build and deployment configs (empty for CLI)
├── docs/             # Service-specific documentation
├── scripts/          # CLI-specific scripts
│
├── Makefile          # Service-specific build, test, install
├── go.mod            # Go module definition
└── README.md         # This file
```

### Design Decisions

**Self-Contained Module**: The CLI is a separate Go module to keep dependencies minimal and binary size small. It does not depend on `generated/` code or database packages - only on `pkg/types` for shared data structures.

**Internal Package**: All implementation is in `internal/` to prevent external imports and maintain encapsulation.

**Test Organization**: Following Go best practices:
- Unit tests live alongside code in `internal/*_test.go`
- Integration tests in `tests/integration/`
- E2E tests in `tests/e2e/`

**No Docker**: CLI is installed locally, not containerized (unlike the API service).

---

## Installation

### From Source

```bash
# From CLI service directory
cd services/cli
make build
make install

# Or from repository root
make build-cli
make install-cli
```

This installs the `ubik` command to `/usr/local/bin/`.

### Uninstall

```bash
# From CLI service directory
cd services/cli
make uninstall

# Or from repository root
make uninstall-cli
```

---

## Usage

### Authentication

```bash
# Login to platform
ubik login

# Login with specific API URL
ubik login --api-url https://api.ubik.dev

# Logout
ubik logout
```

### Configuration Sync

```bash
# Sync all agent configurations
ubik sync

# Sync specific agent
ubik sync --agent claude

# Dry run (show what would be synced)
ubik sync --dry-run
```

### Agent Management

```bash
# List configured agents
ubik agents

# Show detailed info for specific agent
ubik agents show claude-code

# List available agents in catalog
ubik agents list
```

### Interactive Mode

```bash
# Launch interactive interface
ubik

# Interactive mode provides:
# - Guided configuration sync
# - Agent status overview
# - Quick access to common operations
```

### Docker Container Management

```bash
# Start MCP server containers
ubik sync  # Automatically starts required containers

# Stop containers
ubik sync --stop

# List running containers
ubik agents  # Shows container status
```

---

## Development

### Prerequisites

- Go 1.24+
- Docker (for integration tests and MCP containers)
- Access to Ubik Enterprise platform

### Setup

```bash
# Install dependencies
go mod download

# Run code generation (if needed)
cd ../.. && make generate

# Build binary
make build
```

### Testing

```bash
# Run all tests
make test

# Run unit tests only (fast)
make test-unit

# Run integration tests
make test-integration

# Run e2e tests
make test-e2e

# Show coverage report
make coverage
```

### Building

```bash
# Build binary to ../../bin/ubik-cli
make build

# Install to system
make install

# Clean build artifacts
make clean
```

### Project Structure

**Command Structure** (in `cmd/ubik/main.go`):
```
ubik                  # Interactive mode
├── login            # Authenticate with platform
├── logout           # Clear local session
├── sync             # Sync agent configurations
├── agents           # Agent management
│   ├── list        # List available agents
│   └── show <name> # Show agent details
├── version          # Show version
└── help             # Show help
```

**Internal Packages**:
- `auth.go`: JWT token management, session storage
- `sync.go`: Configuration download, file writing
- `agents.go`: Agent listing, filtering, display
- `docker.go`: Docker SDK wrapper
- `container.go`: MCP container lifecycle
- `proxy.go`: Local MCP proxy server
- `config.go`: Local config file management (~/.ubik/)
- `workspace.go`: Workspace detection (Git, env vars)
- `logging/`: Activity logging to platform API

---

## Configuration

### Local Config Files

```
~/.ubik/
├── config.json           # CLI configuration
├── auth.json            # Authentication tokens
├── agents/              # Synced agent configs
│   ├── claude-code/     # Per-agent configuration
│   │   ├── config.json
│   │   └── mcps/        # MCP server configs
│   └── cursor/
└── logs/                # CLI activity logs
```

### Environment Variables

```bash
# Override API URL
export UBIK_API_URL="https://api.ubik.dev"

# Override config directory
export UBIK_CONFIG_DIR="$HOME/.ubik"

# Enable debug logging
export UBIK_DEBUG=true

# Set log level (debug, info, warn, error)
export UBIK_LOG_LEVEL=debug
```

---

## Testing Strategy

### Unit Tests
- Test individual functions and methods
- Use mocks for external dependencies (HTTP, Docker, filesystem)
- Located alongside code in `internal/*_test.go`
- Run with: `make test-unit`

### Integration Tests
- Test interactions between components
- May use test containers or mock servers
- Located in `tests/integration/`
- Run with: `make test-integration`

### E2E Tests
- Test complete user workflows
- Require running platform API
- Located in `tests/e2e/`
- Run with: `make test-e2e`

---

## Docker Integration

### MCP Container Management

The CLI automatically manages Docker containers for MCP servers:

1. **Auto-Start**: When syncing, required MCP containers are started
2. **Health Checks**: Monitors container health and restarts if needed
3. **Network Management**: Creates isolated networks for MCP communication
4. **Volume Management**: Persists MCP data across restarts

### Container Naming

```
ubik-mcp-<employee-id>-<mcp-name>
```

Example: `ubik-mcp-emp123-postgres`

---

## Troubleshooting

### Login Issues

```bash
# Clear cached credentials
rm ~/.ubik/auth.json

# Try login with debug logging
UBIK_DEBUG=true ubik login
```

### Sync Issues

```bash
# Check current configuration
cat ~/.ubik/config.json

# Check agent configurations
ls -la ~/.ubik/agents/

# Re-sync with verbose output
UBIK_DEBUG=true ubik sync
```

### Docker Issues

```bash
# Check Docker is running
docker ps

# Check MCP containers
docker ps -f name=ubik-mcp

# View container logs
docker logs ubik-mcp-<employee-id>-<mcp-name>

# Restart containers
docker restart ubik-mcp-<employee-id>-<mcp-name>
```

### Debug Mode

```bash
# Enable debug logging for any command
UBIK_DEBUG=true ubik <command>

# View detailed error messages
UBIK_LOG_LEVEL=debug ubik <command>
```

---

## Contributing

### Code Style

- Follow Go best practices and conventions
- Run `gofmt` and `golangci-lint` before committing
- Write tests for new functionality (TDD preferred)
- Update documentation for user-facing changes

### Adding New Commands

1. Add command definition in `cmd/ubik/main.go`
2. Implement logic in appropriate `internal/*.go` file
3. Write tests in `internal/*_test.go`
4. Update this README with usage examples
5. Update CLI documentation in repository docs

### Testing Requirements

- All new code must have tests (target: 85% coverage)
- Unit tests must pass locally before PR
- Integration tests must pass in CI
- E2E tests must pass against development platform

---

## Release Process

The CLI is released independently from the API service:

1. Version bump in `cmd/ubik/main.go` (semantic versioning)
2. Update CHANGELOG.md with changes
3. Create git tag: `cli-v0.2.0`
4. Build binaries for all platforms
5. Create GitHub Release with binaries
6. Update installation instructions

---

## Dependencies

### Runtime Dependencies

- `github.com/spf13/cobra` - CLI framework
- `github.com/docker/docker` - Docker SDK
- `github.com/sergeirastrigin/ubik-enterprise/pkg/types` - Shared types

### Development Dependencies

- `github.com/stretchr/testify` - Testing utilities
- `github.com/google/uuid` - UUID generation

### No Database Dependencies

The CLI intentionally does NOT depend on:
- `generated/db` - Database code (API only)
- `generated/api` - API server code (API only)
- PostgreSQL drivers

This keeps the binary small (~10MB) and dependencies minimal.

---

## Performance

### Binary Size

```
Uncompressed: ~10MB
Compressed:   ~3MB
```

### Startup Time

- Cold start: ~50ms
- With Docker ops: ~200ms

### Sync Performance

- Initial sync: ~1-2s (downloads all configs)
- Incremental sync: ~200-500ms (checks for updates)

---

## Security

### Credential Storage

- JWT tokens stored in `~/.ubik/auth.json` (chmod 600)
- Tokens encrypted at rest (future enhancement)
- Session expiry enforced

### Network Security

- All API communication over HTTPS
- TLS certificate validation enforced
- No plaintext passwords stored

### Container Security

- MCP containers run with minimal privileges
- Isolated networks per employee
- No host network access

---

## Future Enhancements

- [ ] Cross-platform binary releases (macOS, Linux, Windows)
- [ ] Auto-update mechanism
- [ ] Config validation before sync
- [ ] Offline mode with cached configs
- [ ] Shell completion (bash, zsh, fish)
- [ ] Config rollback functionality
- [ ] Multi-workspace support

---

## Support

- 📖 Documentation: [docs/CLI_CLIENT.md](../../docs/CLI_CLIENT.md)
- 🐛 Issues: [GitHub Issues](https://github.com/ubik-enterprise/ubik/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/ubik-enterprise/ubik/discussions)

---

**Version**: 0.2.0
**Last Updated**: 2025-11-13
**Maintained by**: Ubik Enterprise Team
