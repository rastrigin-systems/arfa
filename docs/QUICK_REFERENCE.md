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
# After changing schema.sql
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
cd services/cli && go build -o ../../bin/ubik ./cmd/ubik

# Build with specific tags
go build -tags integration -o bin/ubik-test ./services/cli/cmd/ubik
```

---

### Run Services

```bash
# Start dev server (once implemented)
make dev

# Run API server manually
./bin/ubik-api

# Run CLI manually
./bin/ubik <command>
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

---

## Git Workflow

### Feature Development

```bash
# Create feature branch
git checkout main && git pull
git checkout -b feature/my-feature

# Make changes, commit (Git hook auto-generates code)
git add .
git commit -m "feat: Add new feature (#<issue>)"

# Push and create PR
git push -u origin feature/my-feature
gh pr create --title "feat: My Feature (#<issue>)" --body "..."

# Wait for CI/CD checks
gh pr checks --watch

# Merge when ready
gh pr merge --squash
```

---

### Skip Code Generation

```bash
# Skip pre-commit hook (not recommended)
git commit --no-verify -m "your message"
```

**See:** [DEV_WORKFLOW.md](./DEV_WORKFLOW.md) for complete workflow.

---

## Docker Commands

### Container Management

```bash
# List running containers
docker ps

# List all containers (including stopped)
docker ps -a

# View container logs
docker logs <container-name> -f

# Execute command in container
docker exec -it <container-name> sh

# Restart container
docker restart <container-name>

# Remove container
docker rm -f <container-name>

# Remove all stopped containers
docker container prune
```

---

### Network Management

```bash
# List networks
docker network ls

# Inspect network
docker network inspect ubik-network

# Create network
docker network create ubik-network

# Remove network
docker network rm ubik-network
```

---

### Image Management

```bash
# List images
docker images

# Pull image
docker pull <image-name>

# Remove image
docker rmi <image-name>

# Remove unused images
docker image prune
```

---

## MCP Servers

### Management

```bash
# List configured servers
claude mcp list

# Get details about a server
claude mcp get <server-name>

# Add a new MCP server
claude mcp add <name> -- <command>

# Remove an MCP server
claude mcp remove <name> -s local
```

---

### GitHub MCP

```bash
# Verify GitHub MCP is connected
claude mcp list | grep github

# Re-add if needed
claude mcp add github \
  -e GITHUB_PERSONAL_ACCESS_TOKEN=$(gh auth token) \
  -- docker run -i --rm -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server

# Check container
docker ps | grep github-mcp-server
docker images | grep github-mcp-server
```

---

### PostgreSQL MCP

```bash
# Add PostgreSQL MCP for Ubik database
claude mcp add postgres \
  -- docker run -i --rm mcp/postgres \
  postgresql://ubik:ubik_dev_password@host.docker.internal:5432/ubik

# Verify connection
claude mcp list | grep postgres

# Check database is running
docker ps | grep ubik-postgres
make db-up  # If not running
```

---

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

---

## Helpful Aliases

Add these to your `~/.zshrc` or `~/.bashrc`:

```bash
# Ubik aliases
alias ubik-db-up='cd ~/Projects/ubik-enterprise && make db-up'
alias ubik-db-down='cd ~/Projects/ubik-enterprise && make db-down'
alias ubik-db-reset='cd ~/Projects/ubik-enterprise && make db-reset'
alias ubik-test='cd ~/Projects/ubik-enterprise && make test'
alias ubik-gen='cd ~/Projects/ubik-enterprise && make generate'
alias ubik-psql='docker exec -it ubik-postgres psql -U ubik -d ubik'

# Docker aliases
alias d='docker'
alias dc='docker-compose'
alias dps='docker ps'
alias dlogs='docker logs -f'

# Git aliases
alias gs='git status'
alias gp='git pull'
alias gpo='git push origin'
alias gco='git checkout'
alias gcb='git checkout -b'
```

---

## Keyboard Shortcuts

### psql

| Shortcut | Action |
|----------|--------|
| `\q` | Quit |
| `\l` | List databases |
| `\dt` | List tables |
| `\d <table>` | Describe table |
| `\du` | List users |
| `\?` | Help |
| `\x` | Toggle expanded output |

---

### Make

```bash
# Show all available make targets
make help

# Run target with verbose output
make <target> VERBOSE=1
```

---

## See Also

- [QUICKSTART.md](./QUICKSTART.md) - Detailed setup guide
- [DEVELOPMENT.md](./DEVELOPMENT.md) - Development workflow
- [TESTING.md](./TESTING.md) - Testing guide
- [DATABASE.md](./DATABASE.md) - Database operations
- [MCP_SERVERS.md](./MCP_SERVERS.md) - MCP server setup
- [DEV_WORKFLOW.md](./DEV_WORKFLOW.md) - Git workflow
- [DEBUGGING.md](./DEBUGGING.md) - Debugging guide
