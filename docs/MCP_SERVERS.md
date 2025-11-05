# MCP Servers Configuration Guide

**Last Updated:** 2025-11-05

This guide covers all Model Context Protocol (MCP) servers used in the Ubik Enterprise project.

---

## Table of Contents

- [Overview](#overview)
- [Currently Configured Servers](#currently-configured-servers)
- [MCP Server Details](#mcp-server-details)
  - [GitHub MCP](#github-mcp)
  - [Playwright MCP](#playwright-mcp)
  - [Qdrant MCP](#qdrant-mcp)
  - [PostgreSQL MCP](#postgresql-mcp)
  - [Railway MCP](#railway-mcp)
- [Management Commands](#management-commands)
- [Troubleshooting](#troubleshooting)

---

## Overview

**Claude Code** uses Model Context Protocol (MCP) servers for enhanced capabilities. MCP servers provide specialized tools and integrations that extend Claude's abilities.

**Key Concepts:**
- MCP servers are Docker containers that auto-start when needed
- Configuration stored in `~/.claude.json` (project-specific) or global config
- Each server provides a set of tools accessible via `mcp__<server>__<tool>` pattern

---

## Currently Configured Servers

| Server | Purpose | Status | Container |
|--------|---------|--------|-----------|
| **github** | GitHub operations (issues, PRs, code search) | ✅ Active | `ghcr.io/github/github-mcp-server` |
| **playwright** | Browser automation and web interaction | ✅ Active | Official Playwright MCP |
| **qdrant** | Vector search and knowledge management | ⚠️ Manual | `qdrant/qdrant:latest` |
| **postgres** | Database operations and queries | ⚠️ Manual | `mcp/postgres` |
| **railway** | Cloud deployment management | ⚠️ Manual | `@railway/mcp-server` |

**Legend:**
- ✅ Active = Auto-configured with Claude Code
- ⚠️ Manual = Requires manual setup

---

## MCP Server Details

### GitHub MCP

**Purpose:** GitHub operations including issues, PRs, repositories, and code search.

**Capabilities:**
- Create/update/list issues and pull requests
- Search code across repositories
- Manage branches and files
- View repository details
- Monitor CI/CD workflows
- Code security scanning

**Setup:**

```bash
# Add GitHub MCP server
claude mcp add github \
  -e GITHUB_PERSONAL_ACCESS_TOKEN=$(gh auth token) \
  -- docker run -i --rm -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server
```

**Configuration in ~/.claude.json:**

```json
{
  "mcpServers": {
    "github": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-e",
        "GITHUB_PERSONAL_ACCESS_TOKEN",
        "ghcr.io/github/github-mcp-server"
      ],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "<your-token>"
      }
    }
  }
}
```

**Verification:**

```bash
# Check GitHub MCP is connected
claude mcp list | grep github

# If disconnected, restart Claude Code
# The Docker container auto-starts with Claude Code

# If container issue, check Docker
docker ps | grep github-mcp-server
docker images | grep github-mcp-server

# Re-pull image if needed
docker pull ghcr.io/github/github-mcp-server
```

**Troubleshooting:**
- **Container not running**: MCP containers auto-start when Claude Code needs them (not persistent)
- **Connection failed**: Check Docker is running: `docker ps`
- **Image missing**: Re-pull image: `docker pull ghcr.io/github/github-mcp-server`
- **Token expired**: Update token: `gh auth refresh` then `claude mcp remove github -s local` and re-add
- **Config location**: `~/.claude.json` (project-specific) or global config

---

### Playwright MCP

**Purpose:** Browser automation and web interaction for testing and web scraping.

**Capabilities:**
- Navigate web pages
- Interact with page elements
- Take screenshots
- Execute JavaScript in browser context
- Handle forms and user input
- Monitor network requests

**Setup:**

```bash
# Add Playwright MCP server
claude mcp add playwright -- <command>
```

**Usage:**
- See Playwright MCP documentation for detailed usage

---

### Qdrant MCP

**Purpose:** Vector search and knowledge management for storing and retrieving project knowledge.

**⚠️ MANDATORY: Use Qdrant MCP for all knowledge operations**

**When to Use:**
- **Before any task**: Search using `mcp__code-search__qdrant-find`
- **During work**: Store findings using `mcp__code-search__qdrant-store`
- **After completion**: Update stored knowledge

**What to Store in Qdrant:**
- Solutions to specific problems you solved
- "Why we chose X over Y" decisions
- Performance lessons ("approach X was 10x faster")
- Failed attempts and why they didn't work
- Code patterns that work well in this codebase

**What to Keep in .md Files:**
- Architecture overviews
- Getting started guides
- API references
- Comprehensive feature docs

**Workflow:**
1. Index important findings in Qdrant as you work
2. Use Qdrant for search/discovery → then read full .md file
3. Update stored knowledge after task completion

**Setup:**

```bash
# Start Qdrant container
docker run -d --name claude-qdrant -p 6333:6333 qdrant/qdrant:latest

# Verify running
docker ps | grep claude-qdrant

# Add to Claude Code (if needed)
claude mcp add qdrant -- <command>
```

**Configuration:**

```json
{
  "mcpServers": {
    "qdrant": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "-p",
        "6333:6333",
        "qdrant/qdrant:latest"
      ]
    }
  }
}
```

---

### PostgreSQL MCP

**Purpose:** Direct database operations and queries for development/debugging.

**⚠️ MANDATORY: Use PostgreSQL MCP for all database operations**

**When to Use PostgreSQL MCP:**
- ✅ Schema inspection and exploration
- ✅ Ad-hoc query execution
- ✅ Data validation and verification
- ✅ Migration testing
- ✅ Performance analysis
- ✅ Troubleshooting database issues

**When NOT to Use:**
- ❌ Production database access (security risk)
- ❌ Schema migrations (use sqlc/migrations instead)
- ❌ Automated tests (use testcontainers)
- ❌ Application queries (use generated sqlc code)

**Why PostgreSQL MCP?**
- Direct SQL query execution without psql CLI
- Schema introspection and analysis
- Table and column browsing
- Query result formatting
- Transaction support
- Multi-database connection support

**Setup:**

```bash
# Add PostgreSQL MCP server to Claude Code config
claude mcp add postgres \
  -- docker run -i --rm mcp/postgres \
  postgresql://host.docker.internal:5432/ubik

# For Ubik development database
claude mcp add postgres \
  -- docker run -i --rm mcp/postgres \
  postgresql://ubik:ubik_dev_password@host.docker.internal:5432/ubik
```

**Configuration in ~/.claude.json:**

```json
{
  "mcpServers": {
    "postgres": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "mcp/postgres",
        "postgresql://ubik:ubik_dev_password@host.docker.internal:5432/ubik"
      ]
    }
  }
}
```

**Common Operations:**

```bash
# Verify PostgreSQL MCP is connected
claude mcp list | grep postgres

# If not running, add it (container auto-starts when needed)
claude mcp add postgres -- docker run -i --rm mcp/postgres \
  postgresql://ubik:ubik_dev_password@host.docker.internal:5432/ubik

# Check database is accessible
docker ps | grep ubik-postgres

# If database not running, start it
make db-up
```

**Note:** PostgreSQL MCP is for development/debugging only. Always use type-safe sqlc queries in application code.

---

### Railway MCP

**Purpose:** Cloud deployment management for Railway platform.

**⚠️ MANDATORY: Use Railway MCP for all deployment operations**

**When to Use Railway MCP:**
- ✅ Creating and configuring services
- ✅ Managing environment variables
- ✅ Deploying from GitHub
- ✅ Monitoring build and deployment status
- ✅ Managing databases (PostgreSQL, Redis, etc.)
- ✅ Configuring networking and domains

**When NOT to Use:**
- ❌ Local development (use docker-compose)
- ❌ Running tests (use make test)
- ❌ Code generation (use make generate)

**Why Railway MCP?**
- Create and manage services programmatically
- Deploy from GitHub with automatic builds
- Manage environment variables across services
- Monitor deployments and logs
- Configure databases and networking
- Zero-configuration monorepo support

**Setup:**

```bash
# Add Railway MCP server to Claude Code config
# Requires Railway API token from https://railway.app/account/tokens
claude mcp add railway -- npx -y @railway/mcp-server
```

**Configuration in ~/.claude.json:**

```json
{
  "mcpServers": {
    "railway-mcp-server": {
      "command": "npx",
      "args": ["-y", "@railway/mcp-server"]
    }
  }
}
```

**Common Operations:**

```bash
# Verify Railway MCP is connected
claude mcp list | grep railway

# If not running, add it (requires Railway API token)
claude mcp add railway -- npx -y @railway/mcp-server
```

**Note:** Railway MCP requires a Railway API token. Create one at https://railway.app/account/tokens and set it as an environment variable or provide it when prompted.

**See Also:** [RAILWAY_DEPLOYMENT.md](./RAILWAY_DEPLOYMENT.md) for complete deployment guide.

---

## Management Commands

### List Configured Servers

```bash
# List all MCP servers
claude mcp list

# List with details
claude mcp list --verbose
```

### Get Server Details

```bash
# Get details about a specific server
claude mcp get <server-name>

# Example
claude mcp get github
```

### Add New MCP Server

```bash
# Add a new MCP server
claude mcp add <name> -- <command>

# Example: Add custom server
claude mcp add myserver -- docker run -i --rm myorg/myserver
```

### Remove MCP Server

```bash
# Remove an MCP server
claude mcp remove <name> -s local

# Example
claude mcp remove github -s local
```

### Update Server Configuration

```bash
# Remove old configuration
claude mcp remove <name> -s local

# Add with new configuration
claude mcp add <name> -- <new-command>
```

---

## Troubleshooting

### Common Issues

#### Container Not Running

**Problem:** MCP container shows as not running

**Solution:**
- MCP containers auto-start when Claude Code needs them (they're not persistent)
- No action needed unless there's an error

#### Connection Failed

**Problem:** MCP server connection fails

**Solution:**
```bash
# Check Docker is running
docker ps

# Restart Docker if needed
# On macOS: Docker Desktop → Restart
```

#### Image Missing

**Problem:** Docker image not found

**Solution:**
```bash
# Re-pull the image
docker pull <image-name>

# Example for GitHub MCP
docker pull ghcr.io/github/github-mcp-server
```

#### Token Expired

**Problem:** Authentication token expired (GitHub, Railway)

**Solution:**
```bash
# Refresh token
gh auth refresh  # For GitHub

# Remove old MCP config
claude mcp remove github -s local

# Re-add with new token
claude mcp add github \
  -e GITHUB_PERSONAL_ACCESS_TOKEN=$(gh auth token) \
  -- docker run -i --rm -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server
```

#### Configuration Not Found

**Problem:** MCP server configuration lost

**Solution:**
- Configuration stored in `~/.claude.json` (project-specific) or global config
- Re-add the server using setup commands above

### Getting Help

1. Check Claude Code documentation
2. Check MCP server-specific documentation
3. Search GitHub issues for the MCP server
4. Ask in Claude Code community channels

---

## See Also

- [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - Command reference
- [DEVELOPMENT.md](./DEVELOPMENT.md) - Development workflow
- [RAILWAY_DEPLOYMENT.md](./RAILWAY_DEPLOYMENT.md) - Railway deployment guide
- [DATABASE.md](./DATABASE.md) - Database operations
