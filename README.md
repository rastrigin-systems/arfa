# Arfa — AI Agent Security Platform

**Enterprise-grade security proxy for AI coding assistants**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-blue)](https://go.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15%2B-blue)](https://www.postgresql.org/)

---

## What is Arfa?

Arfa is a **security-first platform** that gives enterprises **visibility and control** over AI coding assistant usage (Claude Code, Cursor, Windsurf, GitHub Copilot).

Instead of blocking AI tools, Arfa enables **safe adoption** through:
- 🔒 **Transparent HTTPS Proxy** - Intercept and log all AI agent API traffic
- 📊 **Activity Monitoring** - Track every tool call, file access, and command execution
- 🛡️ **Policy Enforcement** - Block dangerous operations based on organizational policies
- 🎯 **Zero Configuration** - Automatic client detection and policy application
- 🏢 **Multi-Tenant SaaS** - Centralized management across teams and organizations

### Why Arfa?

**The Problem:**
- Companies want developers using AI coding assistants for productivity
- But IT/Security teams lack visibility into what AI agents are doing
- No way to enforce policies (e.g., "don't write to prod database")
- No audit trail for compliance

**The Solution:**
Arfa's transparent proxy architecture lets you:
- ✅ **See everything** - Every API call, tool execution, file access logged
- ✅ **Control access** - Block dangerous operations via declarative policies
- ✅ **Audit usage** - Complete audit trail for security/compliance reviews
- ✅ **Enable safely** - Developers use AI tools freely within guardrails

---

## Architecture

### High-Level Overview

```
┌─────────────────────────────────────────────────────────────┐
│  Developer's Machine                                        │
│                                                             │
│  ┌──────────────┐           ┌──────────────┐                │
│  │ Claude Code  │──────────▶│ Arfa Proxy   │                │
│  │ Cursor       │  HTTPS    │ (localhost)  │                │
│  │ Windsurf     │           └──────┬───────┘                │
│  └──────────────┘                  │                        │
│                           Intercept, Log, Enforce           │
│                                    │                        │
└────────────────────────────────────┼────────────────────────┘
                                     │
                            ┌────────▼─────────┐
                            │   LLM APIs       │
                            │  (Anthropic,     │
                            │   OpenAI, etc.)  │
                            └────────┬─────────┘
                                     │
                            ┌────────▼─────────┐
                            │  Arfa Platform   │
                            │  (Activity Logs, │
                            │   Policies)      │
                            └──────────────────┘
```

### Transparent Proxy Design

**How it works:**
1. **CLI installs** local HTTPS proxy (localhost:8082) with self-signed CA certificate
2. **AI clients** auto-configured via `HTTPS_PROXY` env var or PAC file
3. **Proxy intercepts** all API traffic, decrypts, analyzes, logs
4. **Policies enforced** - dangerous tool calls blocked before reaching LLM
5. **Logs uploaded** async to central platform for audit/analysis

**Deployment modes:**
- 🔧 **CI Pipelines** - Temporary proxy via `eval $(arfa proxy env)`
- 💻 **Local Development** - Per-session proxy in terminal
- 🏢 **Enterprise** - Permanent system-wide setup via PAC file + auto-start daemon

**Supported AI Clients:**
- Claude Code, Cursor, Continue, Windsurf, Aider, GitHub Copilot
- Auto-detected via User-Agent header

---

## Quick Start

### Prerequisites

- **Go 1.24+** - [Install](https://go.dev/doc/install)
- **PostgreSQL 15+** - [Install](https://www.postgresql.org/download/) or use Docker
- **Docker** (optional) - For running PostgreSQL locally

### 1. Start Database

```bash
# Clone repository
git clone https://github.com/yourusername/arfa.git
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

### 6. Configure AI Client

```bash
# In a new terminal, set environment variables
eval $(arfa proxy env)

# Now run your AI client (e.g., Claude Code)
claude "list files in current directory"

# All API traffic is logged to Arfa platform!
```

---

## Documentation

### Getting Started
- 📖 **[Quick Start Guide](docs/quickstart.md)** - Detailed setup instructions
- 🏗️ **[Architecture Overview](docs/architecture/monorepo-structure.md)** - System design
- 🗄️ **[Database Schema](docs/database/schema-reference.md)** - Visual ERD + table reference

### Development
- 🔨 **[Development Workflow](docs/development/contributing.md)** - Code generation, TDD, best practices
- 🧪 **[Testing Guide](docs/testing/strategy.md)** - Unit, integration, E2E testing
- 🚀 **[Deployment](docs/deployment/)** - Docker, GCP Cloud Run

### Service Documentation
- **[API Server](services/api/README.md)** - REST API, WebSocket, business logic
- **[CLI Client](services/cli/README.md)** - Proxy, commands, setup
- **[Web UI](services/web/README.md)** - Next.js admin panel

### API Reference
- **[OpenAPI Spec](platform/api-spec/spec.yaml)** - Complete API contract
- **[Postman Collection](docs/api/)** - Example requests

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
│   ├── database/                 # PostgreSQL schema, queries, seeds
│   └── docker-images/            # Docker images for MCP servers
│
├── generated/                    # Auto-generated code (not in git)
│   ├── api/                      # From OpenAPI spec
│   └── db/                       # From SQL schema
│
├── docs/                         # Documentation
│   ├── architecture/             # System design
│   ├── development/              # Developer guides
│   ├── testing/                  # Testing strategy
│   └── database/                 # Schema docs (auto-generated)
│
├── go.work                       # Go workspace config
├── docker-compose.yml            # Local dev environment
└── Makefile                      # Build automation
```

**Key Principles:**
- **Service Independence** - Each service is a complete Go/Node module
- **Clear Boundaries** - Services never import from each other's internals
- **Generated Code** - Types and DB code auto-generated from source of truth
- **Multi-Tenant** - All queries organization-scoped with Row-Level Security

---

## Tech Stack

### Backend
- **Language**: Go 1.24+
- **API Framework**: Chi router, OpenAPI 3.0.3
- **Database**: PostgreSQL 15+ (multi-tenant with RLS)
- **Proxy**: goproxy with custom handlers
- **Code Generation**: oapi-codegen, sqlc, tbls

### Frontend
- **Framework**: Next.js 14 (App Router)
- **Styling**: Tailwind CSS
- **API Client**: OpenAPI TypeScript types

### Infrastructure
- **Containerization**: Docker, Docker Compose
- **Deployment**: Google Cloud Run (API), Standalone binaries (CLI)
- **Testing**: testcontainers-go, gomock

---

## Development

### Prerequisites

```bash
# Required
brew install go@1.24      # Go 1.24+
brew install postgresql   # PostgreSQL 15+

# Code generation tools (installed via make install-tools)
# - oapi-codegen (API types)
# - sqlc (database code)
# - tbls (ERD documentation)
# - mockgen (test mocks)
```

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
make generate-erd      # Database docs only

# Testing
make test              # Run all tests
make test-api          # API tests only
make test-cli          # CLI tests only

# Building
make build             # Build all services
make build-cli         # CLI binary only
make install-cli       # Install CLI to /usr/local/bin/
```

### Code Generation Workflow

**Two sources of truth:**
1. `platform/database/schema.sql` - Database schema
2. `platform/api-spec/spec.yaml` - API contract

**After changing either:**
```bash
# 1. Reset database (if schema changed)
make db-reset

# 2. Regenerate all code
make generate

# 3. Commit source files + docs (NOT generated code)
git add platform/ docs/
git commit -m "feat: Add new endpoint"
```

**Important:** `generated/` is NOT committed to git. CI regenerates and validates.

---

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/development/contributing.md) for details.

### Development Workflow

1. **Fork** the repository
2. **Create** feature branch: `git checkout -b feature/my-feature`
3. **Make changes** following TDD:
   - Write failing test first
   - Implement minimal code to pass
   - Refactor with tests passing
4. **Run tests**: `make test`
5. **Commit** with descriptive message: `git commit -m "feat: Add feature"`
6. **Push** to your fork: `git push origin feature/my-feature`
7. **Create** Pull Request

### Code Style

- **Go**: Follow [Effective Go](https://go.dev/doc/effective_go)
- **Tests**: Target 85% coverage (excluding generated code)
- **Formatting**: Run `gofmt` before commit
- **Linting**: Run `go vet` before commit

---

## Database Schema

**20 Tables + 3 Views** organized into:

| Category | Tables | Purpose |
|----------|--------|---------|
| **Organization** | organizations, subscriptions, teams, roles, employees | Multi-tenant structure |
| **Agent Management** | agent_catalog, tools, policies, agent_tools, agent_policies, team_policies, employee_agent_configs | Agent configuration |
| **MCP Servers** | mcp_categories, mcp_catalog, employee_mcp_configs | MCP server management |
| **Authentication** | sessions | JWT session tracking |
| **Approvals** | agent_requests, approvals | Access request workflow |
| **Analytics** | activity_logs, usage_records | Usage tracking |

**Views:**
- `v_employee_agents` - Resolved agent configs per employee
- `v_employee_mcps` - Resolved MCP configs per employee
- `v_pending_approvals` - Pending access requests

See [Database Documentation](docs/database/schema-reference.md) for complete visual ERD.

---

## Roadmap

### Current Status (v0.3.0)
✅ API Server with authentication
✅ CLI proxy with policy enforcement
✅ Activity logging and upload
✅ Multi-tenant database with RLS
✅ Client auto-detection

### Upcoming (v1.0.0)
- [ ] Web UI admin panel
- [ ] Advanced policy DSL
- [ ] Real-time log streaming
- [ ] Usage analytics dashboard
- [ ] MCP server management
- [ ] Approval workflows
- [ ] SSO integration (OAuth, SAML)

---

## Security

### Reporting Security Issues

**DO NOT** open public GitHub issues for security vulnerabilities.

Instead, email: **security@arfa.dev** (or contact maintainer directly)

We'll respond within 48 hours and work with you to address the issue.

### Security Features

- ✅ **JWT Authentication** - Secure API access
- ✅ **Row-Level Security** - Database-level multi-tenancy
- ✅ **HTTPS Proxy** - TLS interception with self-signed CA
- ✅ **Policy Enforcement** - Block dangerous operations
- ✅ **Audit Logging** - Complete activity trail
- 🔄 **SSO Integration** - Coming soon

---

## License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

Individual services may have their own licenses - check each service directory.

---

## Support

- 📖 **Documentation**: [docs/](docs/)
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/yourusername/arfa/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/yourusername/arfa/discussions)
- 📧 **Email**: support@arfa.dev

---

## Acknowledgments

Built with:
- [Go](https://go.dev/) - Systems programming language
- [PostgreSQL](https://www.postgresql.org/) - Powerful open source database
- [Next.js](https://nextjs.org/) - React framework
- [Chi](https://github.com/go-chi/chi) - Lightweight Go router
- [goproxy](https://github.com/elazarl/goproxy) - HTTP/HTTPS proxy library
- [sqlc](https://sqlc.dev/) - Type-safe SQL code generation
- [oapi-codegen](https://github.com/deepmap/oapi-codegen) - OpenAPI code generation

---

**Made with ❤️ for enterprises embracing AI coding assistants safely**
