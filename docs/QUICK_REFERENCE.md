# Quick Reference

**Last Updated:** 2025-11-05

Quick reference for common commands, operations, and configurations.

---

## Table of Contents

- [Quick Start](#quick-start)
- [Database Commands](#database-commands)
- [Code Generation](#code-generation)
- [Testing](#testing)
- [Development](#development)
- [Git Workflow](#git-workflow)
- [Docker Commands](#docker-commands)
- [Database Access](#database-access)
- [MCP Servers](#mcp-servers)

---

## Quick Start

### First-Time Setup

```bash
# Clone repository
git clone https://github.com/yourusername/ubik-enterprise.git
cd ubik-enterprise

# Start database
make db-up

# Install tools (one-time)
make install-tools

# Generate all code
make generate

# Run tests
make test

# View documentation
open docs/ERD.md
```

**See:** [QUICKSTART.md](./QUICKSTART.md) for detailed setup guide.

---

## Database Commands

### Basic Operations

```bash
# Start PostgreSQL
make db-up

# Stop PostgreSQL
make db-down

# Reset database (⚠️ deletes all data)
make db-reset

# Run migrations
make db-migrate

# Rollback migrations
make db-rollback
```

---

### Database Access

```bash
# PostgreSQL connection string
postgres://ubik:ubik_dev_password@localhost:5432/ubik

# Adminer web UI
open http://localhost:8080

# psql CLI
docker exec -it ubik-postgres psql -U ubik -d ubik

# Direct SQL query
docker exec ubik-postgres psql -U ubik -d ubik -c "SELECT COUNT(*) FROM employees"
```

---

### Common Queries

```bash
# List all tables
docker exec ubik-postgres psql -U ubik -d ubik -c "\dt"

# Describe table structure
docker exec ubik-postgres psql -U ubik -d ubik -c "\d employees"

# Count records in table
docker exec ubik-postgres psql -U ubik -d ubik -c "SELECT COUNT(*) FROM organizations"

# View all organizations
docker exec ubik-postgres psql -U ubik -d ubik -c "SELECT id, name, created_at FROM organizations"

# Check migrations status
docker exec ubik-postgres psql -U ubik -d ubik -c "SELECT version, dirty FROM schema_migrations"
```

**See:** [DATABASE.md](./DATABASE.md) for complete database guide.

---

## Code Generation

### Setup

```bash
# Install code generation tools (one-time)
make install-tools
```

---

### Generate Code

```bash
# Generate everything (ERD + API + DB + Mocks)
make generate

# Generate ERD docs only
make generate-erd

# Generate API code only
make generate-api

# Generate DB code only
make generate-db

# Generate mocks only
make generate-mocks
```

---

### When to Regenerate

**Manual:**
```bash
# After changing shared/schema/schema.sql
make db-reset && make generate-db && make generate-mocks

# After changing openapi/spec.yaml
make generate-api

# After changing SQL queries
make generate-db && make generate-mocks

# After changing interfaces (for mocks)
make generate-mocks

# Or regenerate everything
make generate
```

---

## Testing

### Run Tests

```bash
# Run all tests with coverage
make test

# Run unit tests only (fast, no Docker required)
make test-unit

# Run integration tests only (requires Docker)
make test-integration

# Generate HTML coverage report
make test-coverage

# Open coverage report
open coverage.html
```

---

### Test Specific Package

```bash
# Test single package
go test ./services/api/internal/handlers/...

# Test with verbose output
go test -v ./services/api/internal/handlers/...

# Test with coverage
go test -cover ./services/api/internal/handlers/...

# Test specific function
go test -run TestCreateEmployee ./services/api/internal/handlers/...
```

**See:** [TESTING.md](./TESTING.md) for complete testing guide.

---

## Development

### Build Binaries

```bash
# Build everything
make build

# Build API server
cd services/api && go build -o ../../bin/ubik-api ./cmd/server

# Build CLI
cd services/cli && go build -o ../../bin/ubik ./cmd/ubik-cli

# Build with specific tags
go build -tags integration -o bin/ubik-test ./services/cli/cmd/ubik-cli
```

---

### Run Services

```bash
# Start all services with docker compose
make dev

# Run API server manually
./bin/ubik-api

# Run CLI manually
./bin/ubik-cli <command>
```

---

### Clean Build Artifacts

```bash
# Clean generated files and binaries
make clean

# Clean Docker volumes
docker volume prune

# Clean Go build cache
go clean -cache

# Remove test databases
docker rm -f $(docker ps -aq --filter "name=postgres")
```

**See:** [DEV_WORKFLOW.md](./DEV_WORKFLOW.md) for complete workflow.

## Common Tasks

### Update Dependencies

```bash
# Update Go dependencies
go get -u ./...
go mod tidy

# Update specific dependency
go get -u github.com/org/package@latest

# Sync go.work
go work sync
```

---

### View Logs

```bash
# API server logs
tail -f logs/api.log

# CLI logs
tail -f ~/.ubik/logs/cli.log

# Docker container logs
docker logs ubik-postgres -f
docker logs ubik-api -f
```

---

### Environment Variables

```bash
# View all env vars
env | grep UBIK

# Set env var for session
export UBIK_API_URL=http://localhost:8080

# Set env var permanently (add to ~/.zshrc or ~/.bashrc)
echo 'export UBIK_API_URL=http://localhost:8080' >> ~/.zshrc
source ~/.zshrc
```

### Make

```bash
# Show all available make targets
make help

# Run target with verbose output
make <target> VERBOSE=1
```

---

## See Also
- [DEVELOPMENT.md](./DEVELOPMENT.md) - Development workflow
- [TESTING.md](./TESTING.md) - Testing guide
- [DATABASE.md](./DATABASE.md) - Database operations
- [MCP_SERVERS.md](./MCP_SERVERS.md) - MCP server setup
- [DEV_WORKFLOW.md](./DEV_WORKFLOW.md) - Git workflow
- [DEBUGGING.md](./DEBUGGING.md) - Debugging guide
