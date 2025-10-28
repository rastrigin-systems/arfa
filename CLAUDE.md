# Ubik Enterprise — AI Agent Management Platform

## System Purpose

Multi-tenant SaaS platform for companies to centrally manage AI agent (Claude Code, Cursor, Windsurf, etc.) and MCP server configurations for their employees.

**Core Value**: Centralized control, policy enforcement, and visibility into AI agent usage across an organization.

**Current Status**: 🟢 **Phase 1 Complete** - Database schema, code generation, and documentation ready (as of 2025-10-28)

---

## Quick Start

```bash
# Start database
cd pivot
make db-up

# Install tools (one-time)
make install-tools

# Generate all code
make generate

# View documentation
open docs/ERD.md
open docs/INDEX.md
```

**See [QUICKSTART.md](./docs/setup/QUICKSTART.md) for complete setup instructions.**

---

## What This Platform Does

### For Companies
- ✅ Manage employees, teams, and roles
- ✅ Control which AI agents employees can use (Claude Code, Cursor, Windsurf, Continue, Copilot)
- ✅ Configure MCP servers and access per employee
- ✅ Set usage policies (path restrictions, rate limits, cost limits)
- ✅ Approve/reject employee requests for new agents or MCPs
- ✅ Track usage, costs, and activity across the organization
- ✅ Enforce compliance and security policies

### For Employees
- ✅ Sync agent configurations to local machines via CLI
- ✅ Request access to new agents or MCP servers
- ✅ View their assigned agents and policies
- ✅ Use AI agents with centrally-managed configurations

---

## Architecture Overview

```
PostgreSQL Schema (DB source of truth)
    ↓
    ├─→ tbls → ERD documentation (auto-generated)
    └─→ sqlc → Type-safe Go database code

OpenAPI Spec (API source of truth)
    ↓
    └─→ oapi-codegen → Go API types + Chi server

Employee CLI Client (future)
    ↓
    └─→ Syncs configs from central server
```

**Hybrid Approach**: Database schema and API spec maintained separately, both generate code automatically.

---

## 📚 Documentation Map

### 🚀 Getting Started
- **[QUICKSTART.md](./docs/setup/QUICKSTART.md)** - 5-minute setup guide
- **[README.md](./README.md)** - Project overview and common commands
- **[INIT_COMPLETE.md](./docs/setup/INIT_COMPLETE.md)** - Phase 1 completion summary
- **[SETUP_COMPLETE.md](./docs/setup/SETUP_COMPLETE.md)** - What's done and what's next
- **[DOCUMENTATION_COMPLETE.md](./docs/setup/DOCUMENTATION_COMPLETE.md)** - Documentation overview

### 📋 Planning & Architecture
- **[MIGRATION_PLAN.md](./docs/planning/MIGRATION_PLAN.md)** - Complete 10-week roadmap
- **[DATABASE_SCHEMA.md](./docs/planning/DATABASE_SCHEMA.md)** - Original ERD with table definitions

### 📊 Database Documentation
- **[docs/ERD.md](./docs/ERD.md)** ⭐ - Mermaid ERD diagram (start here!)
- **[docs/INDEX.md](./docs/INDEX.md)** - Complete documentation index
- **[docs/README.md](./docs/README.md)** - Auto-generated table index
- **[docs/schema.json](./docs/schema.json)** - Machine-readable schema

### 🧪 Testing & Development
- **[docs/DEVELOPMENT_APPROACH.md](./docs/DEVELOPMENT_APPROACH.md)** ⭐ - TDD vs Implementation-First strategy + Week 1 schedule
- **[docs/TESTING_QUICKSTART.md](./docs/TESTING_QUICKSTART.md)** ⭐ - 5-minute testing setup guide
- **[docs/TESTING_STRATEGY.md](./docs/TESTING_STRATEGY.md)** - Complete testing guide with mock generation
- **[docs/TESTING_ANALYSIS.md](./docs/TESTING_ANALYSIS.md)** - Codebase testing analysis

### 🔧 Configuration Files
- **[openapi/spec.yaml](./openapi/spec.yaml)** - OpenAPI 3.0 specification
- **[sqlc/sqlc.yaml](./sqlc/sqlc.yaml)** - sqlc configuration
- **[sqlc/queries/*.sql](./sqlc/queries/)** - Type-safe SQL queries
- **[docker-compose.yml](./docker-compose.yml)** - Local development environment
- **[Makefile](./Makefile)** - Automation commands

---

## Project Structure

```
pivot/
├── CLAUDE.md                  # This file - main documentation hub
├── README.md                  # Quick reference
├── schema.sql                 # PostgreSQL schema (20 tables + 3 views)
├── Makefile                   # Automation commands
├── docker-compose.yml         # Local environment
├── go.mod                     # Go dependencies
│
├── docs/                      # All documentation
│   ├── planning/              # Planning documents
│   │   ├── MIGRATION_PLAN.md  # 10-week roadmap
│   │   └── DATABASE_SCHEMA.md # Original ERD documentation
│   ├── setup/                 # Setup guides
│   │   ├── QUICKSTART.md      # 5-minute setup guide
│   │   ├── INIT_COMPLETE.md   # Phase 1 completion
│   │   ├── SETUP_COMPLETE.md  # What's done/next
│   │   └── DOCUMENTATION_COMPLETE.md # Docs overview
│   ├── ERD.md                 # ⭐ Mermaid ERD (start here!)
│   ├── INDEX.md               # Documentation index
│   ├── README.md              # Auto-generated table index
│   ├── schema.svg             # Full schema diagram
│   ├── schema.json            # Machine-readable schema
│   └── public.*.md            # Table docs (24 files)
│
├── openapi/
│   ├── spec.yaml              # OpenAPI 3.0.3 spec (API source of truth)
│   └── oapi-codegen.yaml      # Generator config
│
├── sqlc/
│   ├── sqlc.yaml              # Generator config
│   └── queries/
│       ├── employees.sql      # Employee CRUD
│       ├── auth.sql           # Sessions
│       └── organizations.sql  # Org/team/roles
│
├── generated/                 # ⚠️ Auto-generated (don't edit!)
│   ├── api/                   # From OpenAPI spec
│   │   └── server.gen.go
│   └── db/                    # From SQL queries
│       ├── models.go
│       ├── employees.sql.go
│       └── ...
│
├── internal/                  # Your code goes here
│   ├── handlers/              # HTTP handlers
│   ├── service/               # Business logic
│   ├── middleware/            # Auth, RLS, logging
│   ├── mapper/                # Type conversion
│   └── validation/            # Custom validators
│
├── cmd/
│   ├── server/                # API server
│   └── cli/                   # Employee CLI (future)
│
└── scripts/                   # Utility scripts
```

---

## Database Schema Summary

### 20 Tables + 3 Views

**Core Organization** (5 tables)
- `organizations` - Top-level tenants
- `subscriptions` - Billing and budgets
- `teams` - Group employees
- `roles` - Permission definitions
- `employees` - User accounts

**Agent Management** (7 tables)
- `agent_catalog` - Available AI agents
- `tools` - Tool registry (fs, git, http, shell, docker)
- `policies` - Usage policies and restrictions
- `agent_tools` - Agent ↔ Tool mapping
- `agent_policies` - Agent ↔ Policy mapping
- `team_policies` - Team-specific overrides
- `employee_agent_configs` - Per-employee agent instances

**MCP Configuration** (3 tables)
- `mcp_categories` - Organize MCP servers
- `mcp_catalog` - Available MCP servers
- `employee_mcp_configs` - Per-employee MCP instances

**Authentication** (1 table)
- `sessions` - JWT session tracking

**Approvals** (2 tables)
- `agent_requests` - Employee requests
- `approvals` - Manager approval workflow

**Analytics** (2 tables)
- `activity_logs` - Audit trail
- `usage_records` - Cost and resource tracking

**Views** (3)
- `v_employee_agents` - Employee agents with details
- `v_employee_mcps` - Employee MCPs with details
- `v_pending_approvals` - Approval queue

**See [docs/ERD.md](./docs/ERD.md) for complete visual schema.**

---

## Key Features

### ✅ Phase 1: Foundation (COMPLETE)
1. **Database Schema** - PostgreSQL with 20 tables, RLS, seed data
2. **Code Generation** - oapi-codegen, sqlc, tbls all configured
3. **OpenAPI Spec** - Auth + Employee endpoints defined
4. **SQL Queries** - Type-safe queries for employees, auth, orgs
5. **Documentation** - 50+ docs including Mermaid ERD
6. **Automation** - Makefile with 20+ commands
7. **Local Environment** - Docker Compose with PostgreSQL + Adminer

### 📋 Phase 2: Core API (Next - Week of 2025-11-04)
1. **Authentication & Authorization** - JWT, sessions, RLS middleware
2. **Employee Management API** - CRUD endpoints
3. **Organization API** - Org, team, role management
4. **Integration Tests** - Full test coverage

### 📋 Phase 3-8: See [MIGRATION_PLAN.md](./docs/planning/MIGRATION_PLAN.md)

---

## Technology Stack

- **Language**: Go 1.24+
- **Database**: PostgreSQL 15+ (multi-tenant with RLS) - **20 tables + 3 views**
- **API Specification**: OpenAPI 3.0.3
- **Code Generation**: oapi-codegen, sqlc, tbls
- **HTTP Router**: Chi
- **Web UI**: Next.js 14 (future)
- **Testing**: testcontainers-go
- **Deployment**: Docker, Docker Compose

---

## Common Commands

```bash
# Database
make db-up              # Start PostgreSQL
make db-down            # Stop PostgreSQL
make db-reset           # Reset database (⚠️ deletes data)

# Code Generation
make generate           # Generate everything
make generate-erd       # Generate ERD docs only
make generate-api       # Generate API code only
make generate-db        # Generate DB code only

# Development
make dev                # Start dev server (once implemented)
make test               # Run tests
make build              # Build binaries
make clean              # Clean generated files

# Help
make help               # Show all commands
```

---

## Quick Reference

### Database Access
```bash
# PostgreSQL connection
postgres://pivot:pivot_dev_password@localhost:5432/pivot

# Adminer web UI
open http://localhost:8080

# psql CLI
docker exec pivot-postgres psql -U pivot -d pivot
```

### API Endpoints (OpenAPI)
```bash
# View spec
cat openapi/spec.yaml

# Once server is running:
# POST /api/v1/auth/login
# GET  /api/v1/auth/me
# GET  /api/v1/employees
# POST /api/v1/employees
# GET  /api/v1/employees/{id}
# PATCH /api/v1/employees/{id}
# DELETE /api/v1/employees/{id}
```

### Generated Code
```bash
# API types
cat generated/api/server.gen.go

# Database queries
cat generated/db/employees.sql.go

# Database models
cat generated/db/models.go
```

---

## Development Workflow

### Making Changes

```bash
# 1. Update database schema
vim schema.sql

# 2. Apply to database
make db-reset

# 3. Update OpenAPI spec (if API changes)
vim openapi/spec.yaml

# 4. Update SQL queries (if needed)
vim sqlc/queries/employees.sql

# 5. Regenerate all code
make generate

# 6. Implement handlers
vim internal/handlers/employees.go

# 7. Run tests
make test

# 8. Build and test locally
go run cmd/server/main.go
```

### Code Generation Pipeline

```
schema.sql → PostgreSQL → tbls → docs/ERD.md, docs/*.md
                        ↓
                       sqlc → generated/db/*.go

openapi/spec.yaml → oapi-codegen → generated/api/server.gen.go

Your code (internal/) → Uses generated types
```

---

## Critical Usage Notes

### ✅ Hybrid Architecture

**Two Sources of Truth**:
1. **schema.sql** - Database structure
2. **openapi/spec.yaml** - API contract

These are maintained separately because:
- DB tables ≠ API DTOs (different concerns)
- DB can have more tables than API exposes
- API can aggregate/transform DB data

**Keep in sync manually** or use drift detection script (future).

### ✅ Generated Code

**Never edit files in `generated/`**:
- They are completely regenerated on `make generate`
- Add `.gitignore` entry to exclude them from commits
- Treat as read-only artifacts

### ✅ Multi-Tenancy

**All queries must be org-scoped**:
```go
// GOOD
db.ListEmployees(ctx, org_id, status)

// BAD
db.ListAllEmployees(ctx) // Exposes all orgs!
```

Use Row-Level Security (RLS) policies as safety net.

---

## Success Metrics

### Phase 1 (Complete)
- ✅ 20 tables + 3 views in PostgreSQL
- ✅ 50+ documentation files generated
- ✅ OpenAPI spec with 10+ endpoints
- ✅ 15+ type-safe SQL queries
- ✅ Code generation working end-to-end
- ✅ Complete automation via Makefile

### Phase 2 Targets
- [ ] Authentication working (JWT + sessions)
- [ ] Employee CRUD endpoints functional
- [ ] Integration tests with >80% coverage
- [ ] API response time <100ms (p95)

---

## 📖 Documentation Standards

### When to Update Docs

**Always update when**:
- Adding new tables → Regenerate ERD: `make generate-erd`
- Adding API endpoints → Update `openapi/spec.yaml`
- Adding SQL queries → Add to `sqlc/queries/*.sql`
- Changing architecture → Update `docs/planning/MIGRATION_PLAN.md`

**How to update**:
```bash
# Database docs (auto-generated)
make generate-erd

# Manual docs (update manually)
vim MIGRATION_PLAN.md
vim docs/ERD.md  # If schema structure changes significantly
```

---

## Roadmap

### Phase 1: ✅ COMPLETED (Foundation)
- Database schema, code generation, documentation

### Phase 2: In Progress (Core API)
- Authentication, employee management, org management

### Phase 3-8: Planned
- Agent/MCP configuration APIs
- Approval workflows
- Analytics endpoints
- Employee CLI client
- Admin web UI
- Production deployment

**See [MIGRATION_PLAN.md](./docs/planning/MIGRATION_PLAN.md) for complete 10-week plan.**

---

## 🎯 Next Actions

### For New Developers

1. Read [QUICKSTART.md](./docs/setup/QUICKSTART.md)
2. Review [docs/ERD.md](./docs/ERD.md)
3. Explore generated code in `generated/`
4. Start implementing handlers in `internal/handlers/`

### For This Week (Phase 2)

1. Create `internal/middleware/auth.go`
2. Implement `internal/handlers/auth.go`
3. Set up JWT token generation
4. Add integration tests
5. Create `cmd/server/main.go`

---

## 📞 Important Links

### Planning
- [MIGRATION_PLAN.md](./docs/planning/MIGRATION_PLAN.md) - Complete roadmap
- [DATABASE_SCHEMA.md](./docs/planning/DATABASE_SCHEMA.md) - Original ERD
- [../CLAUDE.md](../CLAUDE.md) - Parent project (original Ubik)

### Code
- [openapi/spec.yaml](./openapi/spec.yaml) - API specification
- [sqlc/queries/](./sqlc/queries/) - SQL queries
- [generated/](./generated/) - Auto-generated code

### Documentation
- [docs/ERD.md](./docs/ERD.md) - Visual schema
- [docs/INDEX.md](./docs/INDEX.md) - Complete docs index
- [QUICKSTART.md](./docs/setup/QUICKSTART.md) - Setup guide

---

## Status Summary

**Last Updated**: 2025-10-28  
**Version**: 0.1.0 (Phase 1 Complete)  
**Status**: 🟢 Ready for Phase 2 Development

**Phase 1 Achievements**:
- ✅ Complete database schema (20 tables + 3 views)
- ✅ Code generation pipeline working
- ✅ 50+ documentation files
- ✅ OpenAPI spec for auth + employees
- ✅ Type-safe SQL queries
- ✅ Local development environment
- ✅ Comprehensive Mermaid ERD

**Next Milestone**: Phase 2 - Core API (Week of 2025-11-04)

---

**For detailed information, see the [Documentation Map](#-documentation-map) above.**
- save .md files inside /docs. Update cloud.md if .md files were updated or created