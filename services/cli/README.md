# Arfa CLI

Command-line interface for syncing AI agent configurations from the Arfa platform.

## Installation

```bash
# From project root
make build-cli
make install-cli    # Installs to /usr/local/bin/arfa-cli

# Or from this directory
make build
make install
```

## Usage

```bash
arfa-cli login                    # Authenticate with platform
arfa-cli logout                   # Clear session
arfa-cli sync                     # Sync agent configurations
arfa-cli agents list              # List available agents
arfa-cli proxy start              # Start MCP proxy server
arfa-cli proxy stop               # Stop MCP proxy server
arfa-cli                          # Interactive mode
```

## Project Structure

```
services/cli/
├── cmd/arfa/           Main CLI entrypoint (Cobra commands)
├── internal/
│   ├── commands/       Command implementations
│   ├── logging/        Activity logging
│   ├── auth.go         JWT token management
│   ├── sync.go         Configuration sync
│   ├── proxy.go        MCP proxy server
│   ├── config.go       Local config (~/.arfa/)
│   └── ...
├── tests/
│   ├── integration/    Integration tests
│   └── e2e/            End-to-end tests
└── Makefile            Build commands
```

## Commands

```bash
make build              # Build binary to ../../bin/arfa-cli
make test               # Run all tests
make test-unit          # Unit tests only
make test-integration   # Integration tests
make coverage           # Coverage report
make install            # Install to /usr/local/bin
make uninstall          # Remove from /usr/local/bin
```

## Configuration

Local config stored in `~/.arfa/`:
```
~/.arfa/
├── config.json         CLI configuration
├── auth.json           Authentication tokens
└── agents/             Synced agent configs
```

Environment variables:
```bash
ARFA_API_URL=https://api.arfa.dev   # Override API URL
ARFA_DEBUG=true                      # Enable debug logging
ARFA_LOG_LEVEL=debug                 # Log level
```

## Design

- Self-contained Go module (no dependency on `generated/` or database code)
- All implementation in `internal/` for encapsulation
- Unit tests alongside code in `internal/*_test.go`
- Binary size: ~10MB

## Documentation

- [Architecture](../../docs/architecture/overview.md)
- [Testing](../../docs/development/testing.md)
- [Contributing](../../docs/development/contributing.md)
