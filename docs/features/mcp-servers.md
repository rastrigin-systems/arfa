# MCP Servers Configuration Guide

**Last Updated:** 2025-11-05

This guide covers all Model Context Protocol (MCP) servers used in the Arfa project.

---

## Table of Contents

- [Overview](#overview)
- [Currently Configured Servers](#currently-configured-servers)
- [MCP Server Details](#mcp-server-details)
  - [GitHub MCP](#github-mcp)
  - [Playwright MCP](#playwright-mcp)
  - [Qdrant MCP](#qdrant-mcp)
  - [Google Cloud Platform MCP](#google-cloud-platform-mcp)
  - [Google Cloud Observability MCP](#google-cloud-observability-mcp)
  - [PostgreSQL MCP](#postgresql-mcp)
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
| **qdrant** | Vector search and knowledge management | ✅ Active | `better-qdrant-mcp-server` (npm) |
| **gcloud** | Google Cloud Platform operations | ✅ Active | `@google-cloud/gcloud-mcp` (npm) |
| **observability** | Google Cloud monitoring and logging | ✅ Active | `@google-cloud/observability-mcp` (npm) |
| **postgres** | Database operations and queries | ⚠️ Manual | `mcp/postgres` |

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

**Setup (Completed):**

```bash
# 1. Start Qdrant container (already running)
docker run -d --name claude-qdrant -p 6333:6333 -v $(pwd)/.qdrant:/qdrant/storage qdrant/qdrant:latest

# 2. Install better-qdrant-mcp-server (already installed)
npm install -g better-qdrant-mcp-server

# 3. Add to Claude Code (already configured)
claude mcp add qdrant -- npx better-qdrant-mcp-server

# Verify running
docker ps | grep claude-qdrant
claude mcp list | grep qdrant
```

**Configuration in ~/.claude.json:**

```json
{
  "mcpServers": {
    "qdrant": {
      "command": "npx",
      "args": ["better-qdrant-mcp-server"]
    }
  }
}
```

**Current Status:**
- ✅ Container running on localhost:6333
- ✅ MCP server connected and active
- ✅ Ready for knowledge storage and retrieval
- 📁 Persistent storage at `.qdrant/` directory

---


### Google Cloud Platform MCP

**Purpose:** Google Cloud Platform operations including projects, compute, storage, and services.

**Capabilities:**
- Manage GCP projects and resources
- Deploy to Cloud Run, App Engine, Compute Engine
- Configure Cloud Storage buckets
- Set up networking and load balancing
- Manage IAM and service accounts
- Work with Cloud SQL databases
- Configure Cloud Build CI/CD

**Setup:**

```bash
# GCP MCP is installed via npm and auto-runs when needed
# No manual installation required - it's configured in Claude Code
```

**Configuration in claude_desktop_config.json:**

```json
{
  "mcpServers": {
    "gcloud": {
      "command": "npx",
      "args": ["-y", "@google-cloud/gcloud-mcp"]
    }
  }
}
```

**Prerequisites:**

```bash
# Install Google Cloud SDK if not already installed
brew install google-cloud-sdk

# Authenticate with Google Cloud
gcloud auth login

# Set default project (optional)
gcloud config set project your-project-id
```

**Verification:**

```bash
# Check gcloud authentication
gcloud auth list

# Check current project
gcloud config get-value project

# Test access
gcloud projects list
```

**Common Operations:**
- Create and manage GCP projects
- Deploy applications to Cloud Run
- Configure Cloud Storage
- Set up Cloud SQL databases
- Manage IAM permissions
- View billing information

---

### Google Cloud Observability MCP

**Purpose:** Google Cloud monitoring, logging, and observability for production systems.

**Capabilities:**
- View Cloud Logging logs
- Monitor Cloud Monitoring metrics
- Set up alerts and notifications
- Trace distributed requests
- Profile application performance
- Debug production issues
- Analyze error reports

**Setup:**

```bash
# Observability MCP is installed via npm and auto-runs when needed
# No manual installation required - it's configured in Claude Code
```

**Configuration in claude_desktop_config.json:**

```json
{
  "mcpServers": {
    "observability": {
      "command": "npx",
      "args": ["-y", "@google-cloud/observability-mcp"]
    }
  }
}
```

**Prerequisites:**

```bash
# Same as GCP MCP - requires Google Cloud SDK
# Ensure you're authenticated
gcloud auth login

# Grant necessary observability permissions
# Your account needs roles:
# - Logging Viewer (roles/logging.viewer)
# - Monitoring Viewer (roles/monitoring.viewer)
# - Error Reporting Viewer (roles/errorreporting.viewer)
```

**Verification:**

```bash
# Check authentication
gcloud auth list

# Test log access
gcloud logging read --limit 5

# Test metrics access
gcloud monitoring dashboards list
```

**Common Operations:**
- Query Cloud Logging logs
- View application metrics
- Monitor service health
- Debug production errors
- Analyze request traces
- Set up alerting policies

**Use Cases:**
- Production debugging
- Performance analysis
- Cost monitoring
- Security auditing
- SLA monitoring

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
  postgresql://host.docker.internal:5432/arfa

# For Arfa development database
claude mcp add postgres \
  -- docker run -i --rm mcp/postgres \
  postgresql://arfa:arfa_dev_password@host.docker.internal:5432/arfa
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
        "postgresql://arfa:arfa_dev_password@host.docker.internal:5432/arfa"
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
  postgresql://arfa:arfa_dev_password@host.docker.internal:5432/arfa

# Check database is accessible
docker ps | grep arfa-postgres

# If database not running, start it
make db-up
```

**Note:** PostgreSQL MCP is for development/debugging only. Always use type-safe sqlc queries in application code.

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

**Problem:** Authentication token expired (GitHub)

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
- [DATABASE.md](./DATABASE.md) - Database operations
