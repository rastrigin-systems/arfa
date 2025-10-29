# Installation Guide - Ubik Enterprise

Complete installation and setup instructions for the Ubik Enterprise platform (Pivot).

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Platform API Setup](#platform-api-setup)
- [Employee CLI Setup](#employee-cli-setup)
- [Docker Images](#docker-images)
- [Configuration](#configuration)
- [Troubleshooting](#troubleshooting)
- [Next Steps](#next-steps)

---

## Prerequisites

### Required Software

- **Go 1.24+** - [Install Go](https://go.dev/doc/install)
- **PostgreSQL 15+** - [Install PostgreSQL](https://www.postgresql.org/download/)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Make** - Usually pre-installed on macOS/Linux

### Optional Tools

- **tbls** - For database documentation: `brew install k1LoW/tap/tbls`
- **sqlc** - For type-safe SQL: `brew install sqlc`
- **oapi-codegen** - For API generation: `go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest`

---

## Quick Start

### 1. Clone Repository

```bash
git clone https://github.com/yourusername/ubik-enterprise.git
cd ubik-enterprise/pivot
```

### 2. Start Database

```bash
# Start PostgreSQL via Docker Compose
make db-up

# Verify database is running
docker ps | grep pivot-postgres
```

### 3. Install Tools (One-Time)

```bash
make install-tools
```

### 4. Generate Code

```bash
# Generate all code (API, database, mocks, ERD)
make generate
```

### 5. Build Binaries

```bash
# Build both server and CLI
make build

# Or build individually
make build-server  # → bin/pivot-server
make build-cli     # → bin/ubik-cli
```

### 6. Run Server

```bash
# Start API server
./bin/pivot-server

# Server will start on http://localhost:3001
```

### 7. Install CLI (Optional)

```bash
# Copy CLI to system path
make install-cli

# Or manually
cp bin/ubik-cli /usr/local/bin/ubik
```

---

## Platform API Setup

### Local Development

**1. Start Database:**

```bash
cd pivot
make db-up
```

**2. Apply Schema:**

```bash
# Schema is applied automatically on first connection
# Or manually:
docker exec pivot-postgres psql -U pivot -d pivot -f /docker-entrypoint-initdb.d/schema.sql
```

**3. Seed Data (Optional):**

```bash
make db-seed
```

**4. Start Server:**

```bash
go run cmd/server/main.go

# Or use compiled binary
./bin/pivot-server
```

**5. Verify:**

```bash
curl http://localhost:3001/health
# Should return: {"status":"ok"}
```

### Docker Deployment

**1. Build Docker Image:**

```bash
make docker-build
```

**2. Run with Docker Compose:**

```bash
docker-compose up -d
```

**3. Check Logs:**

```bash
docker-compose logs -f pivot-api
```

### Environment Variables

Create `.env` file:

```bash
# Database
DATABASE_URL=postgres://pivot:pivot_dev_password@localhost:5432/pivot

# Server
PORT=3001
HOST=0.0.0.0

# JWT
JWT_SECRET=your-secret-key-here

# Environment
ENVIRONMENT=development
```

---

## Employee CLI Setup

### Installation

**Option 1: Build from Source**

```bash
cd pivot
make build-cli
cp bin/ubik-cli /usr/local/bin/ubik
```

**Option 2: Download Binary (Future)**

```bash
# Will be available at releases page
curl -LO https://github.com/ubik/pivot/releases/download/v0.2.0/ubik-cli-darwin-amd64
chmod +x ubik-cli-darwin-amd64
mv ubik-cli-darwin-amd64 /usr/local/bin/ubik
```

### First-Time Setup

**1. Login to Platform:**

```bash
ubik login --url https://api.yourcompany.com

# Or for local development:
ubik login --url http://localhost:3001

# Enter your email and password when prompted
```

**2. Sync Agent Configurations:**

```bash
ubik sync
```

**3. View Available Agents:**

```bash
ubik agents list
```

**4. Check Status:**

```bash
ubik status
```

### Docker Setup

The CLI requires Docker to run AI agents in containers.

**1. Ensure Docker is Running:**

```bash
docker ps
# Should not return an error
```

**2. Build Agent Images (First Time):**

```bash
cd pivot/docker
make build-all

# This will build:
# - ubik/claude-code:latest
# - ubik/mcp-filesystem:latest
# - ubik/mcp-git:latest
```

**3. Verify Images:**

```bash
docker images | grep ubik
```

### API Key Setup

For Claude Code agent, you need an Anthropic API key:

```bash
# Set environment variable
export ANTHROPIC_API_KEY="sk-ant-..."

# Or pass via flag
ubik start --api-key sk-ant-...
```

---

## Docker Images

### Building Agent Images

**All Images:**

```bash
cd pivot/docker
make build-all
```

**Individual Images:**

```bash
# Claude Code agent
cd agents/claude-code
docker build -t ubik/claude-code:latest .

# MCP Filesystem server
cd ../../mcp/filesystem
docker build -t ubik/mcp-filesystem:latest .

# MCP Git server
cd ../git
docker build -t ubik/mcp-git:latest .
```

### Testing Images

```bash
cd pivot/docker
make test-all
```

### Image Sizes

- **Claude Code**: ~400MB
- **MCP Filesystem**: ~150MB
- **MCP Git**: ~200MB
- **Total**: ~750MB

---

## Configuration

### Platform Configuration

**Database Connection:**

Edit `pivot/.env`:

```bash
DATABASE_URL=postgres://user:password@host:port/database
```

**Server Settings:**

```bash
PORT=3001
HOST=0.0.0.0
LOG_LEVEL=info
```

### CLI Configuration

Configuration stored in `~/.ubik/config.json`:

```json
{
  "platform_url": "https://api.yourcompany.com",
  "token": "eyJhbGci...",
  "employee_id": "uuid",
  "default_agent": "claude-code",
  "last_sync": "2025-10-29T10:00:00Z"
}
```

**View Configuration:**

```bash
ubik config
```

**Reset Configuration:**

```bash
ubik cleanup --remove-config
```

### Agent Configurations

Stored in `~/.ubik/agents/{agent-id}/config.json`:

```json
{
  "agent_id": "uuid",
  "agent_name": "Claude Code",
  "agent_type": "claude-code",
  "is_enabled": true,
  "configuration": {
    "model": "claude-3-5-sonnet-20241022",
    "temperature": 0.7,
    "max_tokens": 8192
  },
  "mcp_servers": []
}
```

---

## Troubleshooting

### Common Issues

#### 1. Database Connection Failed

```bash
# Check if PostgreSQL is running
docker ps | grep pivot-postgres

# Check logs
docker logs pivot-postgres

# Restart database
make db-down && make db-up
```

#### 2. Port Already in Use

```bash
# Find process using port 3001
lsof -i :3001

# Kill process
kill -9 <PID>

# Or use different port
PORT=3002 ./bin/pivot-server
```

#### 3. Docker Not Running

```bash
# Start Docker Desktop (macOS)
open -a Docker

# Or start Docker daemon (Linux)
sudo systemctl start docker
```

#### 4. Container Name Conflict

```bash
# Remove old containers
ubik cleanup --remove-containers

# Or manually
docker rm -f $(docker ps -aq --filter "label=com.ubik.managed=true")
```

#### 5. Terminal Corrupted After Crash

```bash
# Reset terminal
reset

# Or restore manually
stty sane
```

#### 6. Images Not Found

```bash
# Build Docker images
cd pivot/docker
make build-all

# Verify
docker images | grep ubik
```

### Getting Help

**Check Logs:**

```bash
# Server logs
docker-compose logs -f pivot-api

# Container logs
docker logs <container-id>

# CLI debug mode (future)
ubik --debug status
```

**Verify Installation:**

```bash
# Check versions
go version        # Should be 1.24+
docker --version  # Should be 20.10+
psql --version    # Should be 15+

# Check CLI
ubik --version

# Check server
./bin/pivot-server --version
```

---

## Next Steps

### For Administrators

1. **Setup Organization:**
   - Create organization via API
   - Add teams and roles
   - Configure initial agents

2. **Configure Agents:**
   ```bash
   # Via API or Web UI (future)
   curl -X POST http://localhost:3001/api/v1/organizations/current/agent-configs \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "agent_id": "agent-uuid",
       "config": {...},
       "is_enabled": true
     }'
   ```

3. **Add Employees:**
   ```bash
   curl -X POST http://localhost:3001/api/v1/employees \
     -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{
       "email": "employee@company.com",
       "full_name": "Employee Name",
       "password": "SecurePassword123!",
       "role_id": "role-uuid"
     }'
   ```

### For Employees

1. **Login:**
   ```bash
   ubik login
   ```

2. **Browse Agents:**
   ```bash
   ubik agents list
   ```

3. **Request Access:**
   ```bash
   ubik agents request <agent-id>
   ```

4. **Sync Configurations:**
   ```bash
   ubik sync
   ```

5. **Start Working:**
   ```bash
   # Interactive mode
   ubik

   # Or specific agent
   ubik --agent cursor
   ```

### For Developers

1. **Read Documentation:**
   - [CLAUDE.md](./CLAUDE.md) - Complete project documentation
   - [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) - Development workflow
   - [docs/TESTING.md](./docs/TESTING.md) - Testing guide

2. **Run Tests:**
   ```bash
   make test
   ```

3. **Generate Code:**
   ```bash
   make generate
   ```

4. **View Database:**
   ```bash
   # Adminer (web UI)
   open http://localhost:8080

   # Or psql
   docker exec -it pivot-postgres psql -U pivot -d pivot
   ```

---

## Additional Resources

- **Documentation**: [docs/](./docs/)
- **API Reference**: [openapi/spec.yaml](./openapi/spec.yaml)
- **Database Schema**: [schema.sql](./schema.sql)
- **ERD Diagram**: [docs/ERD.md](./docs/ERD.md)
- **CLI Documentation**: [docs/CLI_CLIENT.md](./docs/CLI_CLIENT.md)

---

## Support

- **Issues**: https://github.com/ubik/pivot/issues
- **Discussions**: https://github.com/ubik/pivot/discussions
- **Email**: support@ubik-enterprise.com

---

**Last Updated**: 2025-10-29
**Version**: 0.2.0
