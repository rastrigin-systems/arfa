# Pivot - Enterprise AI Agent Management Platform

**Status**: 🚧 In Development (Phase 1: Foundation)

Centralized platform for companies to manage AI agent and MCP configurations for their employees.

## Quick Start

```bash
# 1. Install dependencies
make install-tools

# 2. Start database
make db-up

# 3. Generate code
make generate

# 4. Start development server
make dev
```

**Access**:
- API Server: http://localhost:3001
- Adminer (DB UI): http://localhost:8080
- API Docs: http://localhost:3001/docs (once implemented)

## What This Does

Pivot allows companies to:
- ✅ Manage employees, teams, and roles
- ✅ Configure which AI agents (Claude Code, Cursor, Windsurf) employees can use
- ✅ Configure which MCP servers employees have access to
- ✅ Set policies (path restrictions, rate limits, cost limits)
- ✅ Approve/reject employee requests for new agents or MCPs
- ✅ Track usage and costs across the organization
- ✅ Employees sync configurations to their local machines via CLI

## Architecture

```
PostgreSQL Schema (source of truth)
    ↓
    ├─→ tbls → Mermaid ERD (auto-generated)
    └─→ sqlc → Go database code (type-safe)

OpenAPI Spec (source of truth)
    ↓
    └─→ oapi-codegen → Go API types + validators
```

## Project Structure

```
pivot/
├── schema.sql                 # PostgreSQL schema (20 tables + 3 views)
├── openapi/spec.yaml          # OpenAPI 3.0.3 spec
├── sqlc/queries/              # SQL queries
├── docs/                      # Documentation
│   ├── planning/              # Planning documents
│   ├── setup/                 # Setup guides
│   ├── ERD.md                 # Mermaid ERD diagram
│   └── public.*.md            # Table documentation (24 files)
├── generated/                 # ⚠️ Auto-generated code
│   ├── api/                   # From OpenAPI
│   └── db/                    # From SQL
├── internal/
│   ├── handlers/              # HTTP request handlers
│   ├── service/               # Business logic
│   └── middleware/            # Auth, RLS, logging
└── cmd/
    ├── server/                # API server
    └── cli/                   # Employee CLI client
```

## Documentation

- **[CLAUDE.md](./CLAUDE.md)** ⭐ - Main documentation hub (start here!)
- **[Migration Plan](./docs/planning/MIGRATION_PLAN.md)** - Complete implementation roadmap
- **[Database Schema](./docs/planning/DATABASE_SCHEMA.md)** - ERD and table definitions
- **[Quickstart](./docs/setup/QUICKSTART.md)** - 5-minute setup guide
- **[ERD Diagram](./docs/ERD.md)** - Visual schema with Mermaid
- **[OpenAPI Spec](./openapi/spec.yaml)** - API contract (OpenAPI 3.0.3)

## Development Workflow

```bash
# Make schema changes
vim schema.sql

# Apply to database
make db-reset

# Regenerate everything
make generate

# Check for drift between OpenAPI and DB
make check-drift

# Run tests
make test

# Start dev server
make dev
```

## Tech Stack

- **Language**: Go 1.24+
- **Database**: PostgreSQL 15+
- **Code Generation**: oapi-codegen, sqlc, tbls
- **HTTP Router**: Chi
- **Testing**: testcontainers-go

## Common Commands

```bash
make help              # Show all commands
make install-tools     # Install code generators
make db-up            # Start PostgreSQL
make generate         # Generate all code
make dev              # Start dev server
make test             # Run tests
make clean            # Clean generated files
```

## Environment Variables

```bash
DATABASE_URL=postgres://pivot:pivot_dev_password@localhost:5432/pivot?sslmode=disable
JWT_SECRET=your_secret_here
PORT=3001
```

## Current Status

### ✅ Completed (Phase 1)
- [x] Database schema design (20 tables + 3 views)
- [x] Project structure setup
- [x] Docker Compose configuration
- [x] Makefile automation
- [x] Migration plan documentation
- [x] ERD generation with tbls (50+ docs generated)
- [x] OpenAPI spec creation (OpenAPI 3.0.3)
- [x] Code generation setup (oapi-codegen, sqlc, tbls)

### 📋 Upcoming
- [ ] Authentication & authorization
- [ ] Employee management API
- [ ] Agent configuration API
- [ ] MCP configuration API
- [ ] Approval workflows
- [ ] Employee CLI client
- [ ] Admin web UI

See [MIGRATION_PLAN.md](./docs/planning/MIGRATION_PLAN.md) for complete roadmap.

## Contributing

This is a pivot from the original Ubik project. See [../CLAUDE.md](../CLAUDE.md) for project context.

---

**Last Updated**: 2025-10-28
