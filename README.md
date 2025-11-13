# Ubik Enterprise — AI Agent Management Platform

**Multi-tenant SaaS for centralized AI agent and MCP configuration management**

🟢 **Status**: Phase 2 Complete - Authentication System Working

---

## What is Ubik Enterprise?

Centralized platform for companies to manage AI agent configurations (Claude Code, Cursor, Windsurf, etc.) and MCP servers for their employees.

**For Companies:**
- Control which AI agents employees can use
- Configure MCP access and policies
- Track usage and costs
- Approve/reject access requests

**For Employees:**
- Sync configurations to local machines via CLI
- Request access to new agents/MCPs
- View usage and policies

---

## Quick Start

```bash
# 1. Start database
make db-up

# 2. Install tools (one-time)
make install-tools

# 3. Generate code
make generate

# 4. View documentation
open docs/ERD.md
```

**Access:**
- Adminer (DB UI): http://localhost:8080
- Database: `postgres://ubik:ubik_dev_password@localhost:5432/ubik`

---

## Documentation

**Start here:** [CLAUDE.md](./CLAUDE.md) ⭐ - Complete documentation hub

**Quick Links:**
- [5-Minute Setup](./docs/QUICKSTART.md) - Get started
- [Database ERD](./docs/ERD.md) - Visual schema
- [Testing Guide](./docs/TESTING.md) - How to test
- [Development Guide](./docs/DEVELOPMENT.md) - Development workflow

---

## Architecture

**Go Workspace Monorepo** with self-contained services:

```
PostgreSQL (20 tables + 3 views)
    ↓
    ├─→ tbls → ERD docs (auto-generated)
    └─→ sqlc → Type-safe Go code

OpenAPI 3.0.3 Spec
    ↓
    └─→ oapi-codegen → API types + router
```

**Two sources of truth:**
1. `platform/database/schema.sql` - Database structure
2. `platform/api-spec/spec.yaml` - API contract

**Services:**
- **API Server** - REST API (see [services/api/README.md](./services/api/README.md))
- **CLI Client** - Command-line tool (see [services/cli/README.md](./services/cli/README.md))
- **Web UI** - Next.js frontend (see [services/web/README.md](./services/web/README.md))

---

## Project Structure

```
ubik-enterprise/                  # 🌟 Monorepo Root
├── CLAUDE.md                     # ⭐ Complete documentation hub
├── README.md                     # This file
│
├── services/                     # 🎯 Self-contained services
│   ├── api/                      # API Server (Go)
│   ├── cli/                      # CLI Client (Go)
│   └── web/                      # Web UI (Next.js)
│
├── platform/                     # 🔧 Shared platform resources
│   ├── api-spec/                 # OpenAPI 3.0.3 spec
│   ├── database/                 # PostgreSQL schema & sqlc queries
│   └── docker-images/            # Docker image definitions
│
├── pkg/types/                    # 📦 Shared Go types
├── generated/                    # ⚠️ AUTO-GENERATED (don't edit!)
├── docs/                         # 📚 Documentation
└── scripts/                      # 🛠️ Build & utility scripts
```

**Each service is self-contained** with its own:
- `README.md` - Service-specific documentation
- `go.mod` - Independent dependencies
- `internal/` - Service implementation
- `tests/` - Service tests
- Build & deployment configs

---

## Common Commands

```bash
make help              # Show all commands
```

**Note:** The `generated/` directory is NOT committed to git. Always run `make generate` after pulling changes that modify `platform/database/schema.sql`, `platform/api-spec/spec.yaml`, or SQL queries.

---

## Tech Stack

- **Language**: Go 1.24+
- **Database**: PostgreSQL 15+ (multi-tenant with RLS)
- **API**: OpenAPI 3.0.3, Chi router
- **WEB**: Next.js, Tailwind CSS
- **Code Generation**: oapi-codegen, sqlc, tbls
- **Testing**: testcontainers-go, gomock