<p align="center">
  <img src="services/web/public/logo.svg" alt="arfa" width="200" />
</p>

# Claude Code Security Gateway

**Enterprise security proxy for Claude Code**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15%2B-blue)](https://www.postgresql.org/)

---

## What is Arfa?

Arfa is an **open-source security gateway** for [Claude Code](https://docs.anthropic.com/en/docs/build-with-claude/claude-code/overview) that gives organizations **visibility and control** over AI agent usage.

**Key Features:**
- **Transparent HTTPS Proxy** - Intercept and log all Claude Code API traffic
- **Activity Monitoring** - Track every tool call, file access, and command execution
- **Policy Enforcement** - Block dangerous operations based on organizational policies
- **Zero Configuration** - Automatic client detection and policy application
- **Multi-Tenant SaaS** - Centralized management across teams and organizations

### Why Arfa?

**The Problem:**
- Companies want developers using Claude Code for productivity
- But IT/Security teams lack visibility into what the AI agent is doing
- No way to enforce policies (e.g., "don't write to production database", "don't access secrets")
- No audit trail for compliance

**The Solution:**
Arfa's transparent proxy architecture lets you:
- **See everything** - Every API call, tool execution, file access is logged
- **Control access** - Block dangerous operations via declarative policies
- **Audit usage** - Complete audit trail for security/compliance reviews
- **Enable safely** - Developers use Claude Code freely within guardrails

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Developer's Machine                                        │
│                                                             │
│  ┌──────────────┐           ┌──────────────┐                │
│  │ Claude Code  │──────────▶│ Arfa Proxy   │                │
│  │              │  HTTPS    │ (localhost)  │                │
│  └──────────────┘           └──────┬───────┘                │
│                           Intercept, Log, Enforce           │
│                                    │                        │
└────────────────────────────────────┼────────────────────────┘
                                     │
                            ┌────────▼─────────┐
                            │ api.anthropic.com│
                            │  (Claude API)    │
                            └────────┬─────────┘
                                     │
                            ┌────────▼─────────┐
                            │  Arfa Platform   │
                            │  (Activity Logs, │
                            │   Policies)      │
                            └──────────────────┘
```

**How it works:**
1. **CLI installs** local HTTPS proxy (localhost:8082) with self-signed CA certificate
2. **Claude Code** auto-configured via `HTTPS_PROXY` and `SSL_CERT_FILE` env vars
3. **Proxy intercepts** all API traffic to api.anthropic.com
4. **Policies enforced** - dangerous tool calls blocked before reaching Claude
5. **Logs uploaded** async to central platform for audit/analysis

---

## Quick Start

### Prerequisites

- **Go 1.24+** - [Install](https://go.dev/doc/install)
- **PostgreSQL 15+** - [Install](https://www.postgresql.org/download/) or use Docker
- **Claude Code** - [Install](https://docs.anthropic.com/en/docs/build-with-claude/claude-code/overview)

### 1. Start Database

```bash
# Clone repository
git clone https://github.com/rastrigin-systems/arfa.git
cd arfa

# Start PostgreSQL with Docker
docker compose up -d

# Database will auto-load schema and seed data
# Default admin: admin@acme.com / password
```

### 2. Generate Code

```bash
# Install code generation tools
make install-tools

# Generate API types and database code
make generate
```

### 3. Run API Server

```bash
cd services/api
go run cmd/server/main.go
```

API server runs at `http://localhost:8080`

### 4. Install CLI

```bash
# Build and install CLI
cd services/cli
make install

# CLI installed to /usr/local/bin/arfa
arfa version
```

### 5. Login and Start Proxy

```bash
# Login to platform (use admin@acme.com / password)
arfa login

# Start proxy
arfa proxy start
```

Proxy runs at `http://localhost:8082`

### 6. Configure Claude Code

```bash
# In a new terminal, set environment variables
eval $(arfa proxy env)

# Now run Claude Code - all traffic is logged!
claude "list files in current directory"

# View logs
arfa logs view
```

---

## Documentation

### Getting Started
- [Getting Started](docs/development/getting-started.md) - Setup instructions
- [Architecture Overview](docs/architecture/overview.md) - System design
- [Database Schema](docs/database/schema-reference.md) - Visual ERD + table reference

### Development
- [Contributing](docs/development/contributing.md) - Code generation, TDD, best practices
- [Testing Guide](docs/development/testing.md) - Unit, integration, E2E testing
- [PR Workflow](docs/development/workflows.md) - Git workflow, branch naming

### Service Documentation
- [API Server](services/api/README.md) - REST API, WebSocket, business logic
- [CLI Client](services/cli/README.md) - Proxy, commands, setup
- [Web UI](services/web/README.md) - Next.js admin panel

### API Reference
- [OpenAPI Spec](platform/api-spec/spec.yaml) - Complete API contract

---

## Project Structure

```
arfa/
├── services/                     # Self-contained services
│   ├── api/                      # REST API (Go)
│   ├── cli/                      # Proxy CLI (Go)
│   └── web/                      # Admin UI (Next.js)
│
├── platform/                     # Shared resources (source of truth)
│   ├── api-spec/                 # OpenAPI 3.0.3 spec
│   └── database/                 # PostgreSQL schema, queries, seeds
│
├── generated/                    # Auto-generated code (not in git)
│   ├── api/                      # From OpenAPI spec
│   └── db/                       # From SQL schema
│
├── docs/                         # Documentation
│   ├── architecture/             # System design
│   ├── development/              # Developer guides
│   └── database/                 # Schema docs (auto-generated)
│
├── go.work                       # Go workspace config
├── docker-compose.yml            # Local dev environment
└── Makefile                      # Build automation
```

---

## Tech Stack

### Backend
- **Language**: Go 1.24+
- **API Framework**: Chi router, OpenAPI 3.0.3
- **Database**: PostgreSQL 15+ (multi-tenant with RLS)
- **Proxy**: goproxy with custom handlers
- **Code Generation**: oapi-codegen, sqlc

### Frontend
- **Framework**: Next.js 14 (App Router)
- **Styling**: Tailwind CSS
- **API Client**: OpenAPI TypeScript types

---

## Development

### Common Commands

```bash
make help              # Show all available commands

# Database
make db-up             # Start PostgreSQL
make db-reset          # Reset schema + load seeds
make db-down           # Stop PostgreSQL

# Code generation
make generate          # Generate all code
make generate-api      # API types only
make generate-db       # Database code only

# Testing
make test              # Run all tests
make test-api          # API tests only
make test-cli          # CLI tests only

# Building
make build             # Build all services
```

---

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/development/contributing.md) for details.

### Development Workflow

1. **Fork** the repository
2. **Create** feature branch: `git checkout -b feature/my-feature`
3. **Make changes** following TDD
4. **Run tests**: `make test`
5. **Commit** with descriptive message
6. **Push** and create Pull Request

---

## Security

### Reporting Security Issues

**DO NOT** open public GitHub issues for security vulnerabilities.

Instead, email: **security@arfa.dev** (or contact maintainer directly)

### Security Features

- **JWT Authentication** - Secure API access
- **Row-Level Security** - Database-level multi-tenancy
- **HTTPS Proxy** - TLS interception with self-signed CA
- **Policy Enforcement** - Block dangerous tool calls
- **Audit Logging** - Complete activity trail

---

## License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

## Support

- **Documentation**: [docs/](docs/)
- **Bug Reports**: [GitHub Issues](https://github.com/rastrigin-systems/arfa/issues)
- **Discussions**: [GitHub Discussions](https://github.com/rastrigin-systems/arfa/discussions)

---

## Acknowledgments

Built with:
- [Go](https://go.dev/) - Systems programming language
- [PostgreSQL](https://www.postgresql.org/) - Powerful open source database
- [Next.js](https://nextjs.org/) - React framework
- [goproxy](https://github.com/elazarl/goproxy) - HTTP/HTTPS proxy library
- [sqlc](https://sqlc.dev/) - Type-safe SQL code generation

---

**Made with love for teams embracing Claude Code safely**
