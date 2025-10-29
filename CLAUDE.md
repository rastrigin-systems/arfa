# Ubik Enterprise — AI Agent Management Platform

**Multi-tenant SaaS platform for centralized AI agent and MCP configuration management**

---

## 📑 Table of Contents

### Foundation (Rarely Changes)
- [System Overview](#system-overview)
- [Architecture](#architecture)
- [Database Schema](#database-schema)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)

### Documentation
- [Documentation Map](#-documentation-map)
- [Quick Reference](#quick-reference)

### Development
- [Common Commands](#common-commands)
- [Development Workflow](#development-workflow)
- [Critical Notes](#critical-notes)

### Status & Next Steps
- [Current Status](#current-status)
- [Success Metrics](#success-metrics)
- [Roadmap](#roadmap)

---

# FOUNDATION

*This section describes the stable foundation of the system that rarely changes.*

---

## System Overview

### System Purpose

Multi-tenant SaaS platform for companies to centrally manage AI agent (Claude Code, Cursor, Windsurf, etc.) and MCP server configurations for their employees.

**Core Value**: Centralized control, policy enforcement, and visibility into AI agent usage across an organization.

### What This Platform Does

**For Companies:**
- Manage employees, teams, and roles
- Control which AI agents employees can use
- Configure MCP servers and access per employee
- Set usage policies (path restrictions, rate limits, cost limits)
- Approve/reject employee requests for new agents or MCPs
- Track usage, costs, and activity across the organization
- Enforce compliance and security policies

**For Employees:**
- Sync agent configurations to local machines via CLI
- Request access to new agents or MCP servers
- View their assigned agents and policies
- Use AI agents with centrally-managed configurations

### Design Decisions

**Authentication Model**: Self-service employees with passwords (no separate admin tier)
- Each employee manages their own configs
- Roles define permissions (member, approver)
- No admin/user distinction - everyone is an employee
- Multi-tenant via org_id scoping

---

## Architecture

### System Design

```
PostgreSQL Schema (DB source of truth)
    ↓
    ├─→ tbls → schema.json, README.md, public.*.md, schema.svg (auto-generated)
    ├─→ Python script → ERD.md (from schema.json, auto-generated)
    └─→ sqlc → Type-safe Go database code

OpenAPI Spec (API source of truth)
    ↓
    └─→ oapi-codegen → Go API types + Chi server

Employee CLI Client (future)
    ↓
    └─→ Syncs configs from central server
```

**Hybrid Approach**: Database schema and API spec maintained separately, both generate code automatically.

**Why Hybrid?**
- DB tables ≠ API DTOs (different concerns)
- DB can have more tables than API exposes
- API can aggregate/transform DB data

---

## Database Schema

### Overview

**20 Tables + 3 Views**

### Core Organization (5 tables)
- `organizations` - Top-level tenant
- `subscriptions` - Billing and budget tracking
- `teams` - Group employees
- `roles` - Define permissions
- `employees` - User accounts

### Agent Management (7 tables)
- `agent_catalog` - Available AI agents (Claude Code, Cursor, etc.)
- `tools` - Available tools (fs, git, http, etc.)
- `policies` - Usage policies and restrictions
- `agent_tools` - Many-to-many: agents ↔ tools
- `agent_policies` - Many-to-many: agents ↔ policies
- `team_policies` - Team-specific policy overrides
- `employee_agent_configs` - Per-employee agent instances

### MCP Configuration (3 tables)
- `mcp_categories` - Organize MCP servers
- `mcp_catalog` - Available MCP servers
- `employee_mcp_configs` - Per-employee MCP instances

### Authentication (1 table)
- `sessions` - JWT session tracking

### Approvals (2 tables)
- `agent_requests` - Employee requests for agents/MCPs
- `approvals` - Manager approval workflow

### Analytics (2 tables)
- `activity_logs` - Audit trail
- `usage_records` - Cost and resource tracking

### Views (3)
- `v_employee_agents` - Employee agents with catalog details
- `v_employee_mcps` - Employee MCPs with catalog details
- `v_pending_approvals` - Pending approval requests with context

**See [docs/ERD.md](./docs/ERD.md) for complete visual schema.**

---

## Technology Stack

- **Language:** Go 1.24+
- **Database:** PostgreSQL 15+ (multi-tenant with RLS) - 20 tables + 3 views
- **API Specification:** OpenAPI 3.0.3
- **Code Generation:** oapi-codegen, sqlc, tbls
- **HTTP Router:** Chi
- **Testing:** testcontainers-go, gomock
- **Web UI:** Next.js 14 (future)
- **Deployment:** Docker, Docker Compose

---

## Project Structure

```
pivot/
├── CLAUDE.md                  # This file - documentation root
├── README.md                  # Quick overview
├── IMPLEMENTATION_ROADMAP.md  # Next endpoints to build
├── schema.sql                 # PostgreSQL schema (source of truth)
├── Makefile                   # Automation commands
├── docker-compose.yml         # Local environment
├── go.mod                     # Go dependencies
│
├── docs/
│   ├── QUICKSTART.md          # 5-minute setup
│   ├── TESTING.md             # Testing guide
│   ├── DEVELOPMENT.md         # Development workflow
│   │
│   ├── ERD.md                 # ⭐ Auto-generated ERD
│   ├── README.md              # Auto-generated table index (tbls)
│   ├── public.*.md            # Auto-generated per-table docs (27 files)
│   ├── schema.json            # Machine-readable schema
│   ├── schema.svg             # Visual diagram
│   │
│   └── archive/               # Historical documentation
│       ├── MIGRATION_PLAN.md
│       ├── INIT_COMPLETE.md
│       ├── SETUP_COMPLETE.md
│       └── DOCUMENTATION_COMPLETE.md
│
├── openapi/
│   ├── spec.yaml              # OpenAPI 3.0.3 spec (source of truth)
│   └── oapi-codegen.yaml      # Generator config
│
├── sqlc/
│   ├── sqlc.yaml              # Generator config
│   └── queries/
│       ├── employees.sql      # Employee CRUD
│       ├── auth.sql           # Sessions
│       └── organizations.sql  # Org/team/roles
│
├── generated/                 # ⚠️ AUTO-GENERATED (don't edit!)
│   ├── api/                   # From OpenAPI spec
│   ├── db/                    # From SQL queries
│   └── mocks/                 # From interfaces
│
├── internal/                  # Your code goes here
│   ├── handlers/              # HTTP handlers
│   ├── auth/                  # JWT utilities
│   ├── middleware/            # Auth, RLS, logging
│   ├── mapper/                # Type conversion
│   └── validation/            # Custom validators
│
├── tests/
│   ├── integration/           # Full stack tests
│   └── testutil/              # Test helpers
│
├── cmd/
│   └── server/                # API server
│
└── scripts/                   # Utility scripts
    └── generate-erd-overview.py  # Auto-generates ERD.md
```

---

# DOCUMENTATION

*Links to all project documentation organized by purpose.*

---

## 📚 Documentation Map

### 🔥 START HERE

**New to the project?**
1. **[docs/QUICKSTART.md](./docs/QUICKSTART.md)** - 5-minute setup guide
2. **[docs/ERD.md](./docs/ERD.md)** - Visual database schema
3. **[IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md)** ⭐ - Next endpoints to build

### 📊 Database Documentation

- **[docs/ERD.md](./docs/ERD.md)** ⭐ - **Start here!** User-friendly ERD with categories (auto-generated)
- **[docs/README.md](./docs/README.md)** - Technical reference with table index and functions (auto-generated by tbls)
- **[docs/public.*.md](./docs/)** - Per-table documentation (27 files, auto-generated by tbls)
- **[docs/schema.json](./docs/schema.json)** - Machine-readable schema (auto-generated by tbls)
- **[docs/schema.svg](./docs/schema.svg)** - Visual ERD diagram (auto-generated by tbls)

**Note:** Both ERD.md and README.md contain full Mermaid ERD diagrams but serve different purposes:
- **ERD.md** = Human-friendly overview with grouping (🏢 Core, 🤖 Agent, 🔌 MCP, etc.)
- **README.md** = Technical reference with table index and function list

### 🧪 Development & Testing

- **[docs/TESTING.md](./docs/TESTING.md)** ⭐ - Complete testing guide (TDD workflow, patterns, commands)
- **[docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)** ⭐ - Development workflow and best practices
- **[IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md)** ⭐⭐⭐ - **PRIORITY ORDER** for next endpoints

### 🔧 Configuration Files

- **[openapi/spec.yaml](./openapi/spec.yaml)** - OpenAPI 3.0 specification
- **[sqlc/sqlc.yaml](./sqlc/sqlc.yaml)** - sqlc configuration
- **[sqlc/queries/*.sql](./sqlc/queries/)** - Type-safe SQL queries
- **[docker-compose.yml](./docker-compose.yml)** - Local development environment
- **[Makefile](./Makefile)** - Automation commands

### 📦 Archived Documentation

- **[docs/archive/MIGRATION_PLAN.md](./docs/archive/MIGRATION_PLAN.md)** - Original 10-week plan (historical reference)
- **[docs/archive/INIT_COMPLETE.md](./docs/archive/INIT_COMPLETE.md)** - Phase 1 completion summary
- **[docs/archive/SETUP_COMPLETE.md](./docs/archive/SETUP_COMPLETE.md)** - Initial setup notes
- **[docs/archive/DOCUMENTATION_COMPLETE.md](./docs/archive/DOCUMENTATION_COMPLETE.md)** - Docs overview

---

## Quick Reference

### Quick Start

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
```

### Common Commands

```bash
# Database
make db-up              # Start PostgreSQL
make db-down            # Stop PostgreSQL
make db-reset           # Reset database (⚠️ deletes data)

# Code Generation
make generate           # Generate everything (ERD + API + DB + Mocks)
make generate-erd       # Generate ERD docs (ERD.md + README.md + schema.json + SVG)
make generate-api       # Generate API code only
make generate-db        # Generate DB code only
make generate-mocks     # Generate mocks only

# Testing
make test               # Run all tests with coverage
make test-unit          # Run unit tests only (fast)
make test-integration   # Run integration tests (requires Docker)
make test-coverage      # Generate HTML coverage report

# Development
make dev                # Start dev server (once implemented)
make build              # Build binaries
make clean              # Clean generated files

# Help
make help               # Show all commands
```

### Database Access

```bash
# PostgreSQL connection
postgres://pivot:pivot_dev_password@localhost:5432/pivot

# Adminer web UI
open http://localhost:8080

# psql CLI
docker exec pivot-postgres psql -U pivot -d pivot
```

---

# DEVELOPMENT

*Essential information for working with the codebase.*

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
schema.sql → PostgreSQL → tbls → schema.json, README.md, public.*.md, schema.svg
                        ↓         ↓
                       sqlc      Python script → ERD.md (user-friendly)
                        ↓
                  generated/db/*.go

openapi/spec.yaml → oapi-codegen → generated/api/server.gen.go

Your code (internal/) → Uses generated types
```

**ERD Documentation:**
- `make generate-erd` creates **both** README.md (tbls) and ERD.md (custom script)
- README.md = Technical reference with table index
- ERD.md = User-friendly overview with categories

**See [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) for complete development guide.**

---

## Critical Notes

### Code Generation

**Never edit files in `generated/`** - they are completely regenerated!

**Always regenerate after changes:**
```bash
# After changing schema.sql
make db-reset && make generate-db && make generate-mocks

# After changing openapi/spec.yaml
make generate-api

# After changing SQL queries
make generate-db && make generate-mocks
```

### Multi-Tenancy

**All queries must be org-scoped:**

```go
// ✅ GOOD - Scoped to organization
employees, err := db.ListEmployees(ctx, orgID, status)

// ❌ BAD - Exposes all orgs!
employees, err := db.ListAllEmployees(ctx)
```

Use Row-Level Security (RLS) policies as safety net.

### Testing Strategy

**⚠️ CRITICAL: ALWAYS FOLLOW STRICT TDD (Test-Driven Development)**

**Mandatory TDD Workflow:**
```
✅ 1. Write failing tests FIRST
✅ 2. Implement minimal code to pass tests
✅ 3. Refactor with tests passing
❌ NEVER write implementation before tests
```

**Example:**
```
✅ CORRECT: Write router tests → Implement router wiring → Tests pass
❌ WRONG:   Implement router wiring → Write tests later
❌ WRONG:   Implement handler → Write tests after
```

**Target Coverage:** 85% overall (excluding generated code)

**Why TDD is Mandatory:**
- Prevents regression bugs
- Forces good API design
- Ensures all code is testable
- Provides immediate feedback
- Builds confidence in changes

**See [docs/TESTING.md](./docs/TESTING.md) for complete testing guide with TDD workflow.**

---

# STATUS & ROADMAP

*Current progress and next steps.*

---

## Current Status

**Last Updated:** 2025-10-29
**Version:** 0.1.0 🎉
**Status:** 🟢 **Milestone v0.1 - Foundation Complete**
**Git Tag:** `v0.1.0`

### 🎉 Milestone v0.1.0 Released!

**39 API endpoints implemented** | **144+ tests passing** | **73-88% coverage**

See **[docs/MILESTONE_v0.1.md](./docs/MILESTONE_v0.1.md)** for complete release notes.

### Phase 1 Achievements ✅
- Complete database schema (20 tables + 3 views)
- Code generation pipeline working
- 60+ documentation files
- OpenAPI spec for all endpoints
- Type-safe SQL queries
- Local development environment
- Comprehensive Mermaid ERD

### Phase 2 Achievements ✅ (v0.1.0)
- **Complete authentication system** with JWT + sessions
- **Employee CRUD** - Full lifecycle management (5 endpoints)
- **Organization management** - Get/Update current org (2 endpoints)
- **Team management** - Full CRUD (5 endpoints)
- **Role management** - Full CRUD (5 endpoints)
- **Agent catalog** - List and get agents (2 endpoints)
- **Agent configurations** - Org/Team/Employee configs (16 endpoints)
- **144+ tests passing** (119 unit + 25+ integration)
- **73-88% code coverage** across all modules
- Multi-tenancy verified with integration tests
- TDD workflow throughout

**Test Coverage by Module:**
- `internal/handlers`: 73.3%
- `internal/auth`: 88.2%
- `internal/middleware`: 82.2%
- `internal/service`: 77.8%

### Phase 3 - CLI Development 🎯 (In Progress)

**Current Focus:** Employee CLI Client (v0.2.0)

**✅ Phase 0 - Docker Images (Complete)**
- Docker images for Claude Code and MCP servers built
- See [docker/README.md](./docker/README.md)

**✅ CLI Phase 1 - Foundation (Complete)**
- CLI project structure with cobra
- Authentication (`ubik login`, `ubik logout`)
- Platform API client
- Config management (`~/.ubik/config.json`)
- Sync service (`ubik sync` - fetch configs)
- 13 unit tests passing (100%)
- See **[docs/CLI_PHASE1_COMPLETE.md](./docs/CLI_PHASE1_COMPLETE.md)** for details

**✅ CLI Phase 2 - Docker Integration (Complete)**
- Docker SDK integration & client wrapper
- Container lifecycle management (start/stop/status)
- Network management (`ubik-network`)
- MCP server orchestration
- Enhanced `ubik sync --start-containers`
- New commands: `ubik start`, `ubik stop`
- **42 tests passing (24 unit + 18 integration)** ✅
- Coverage: ~22% (unit only), ~60-70% (with Docker)
- Comprehensive error handling and edge case coverage
- See **[docs/CLI_PHASE2_COMPLETE.md](./docs/CLI_PHASE2_COMPLETE.md)** for details

**🎯 CLI Phase 3 - Interactive Mode (Next)**
- Interactive workspace selection
- I/O proxying to agent container
- TTY mode for interactive sessions
- Agent switching on-the-fly
- Session management

**See [docs/CLI_CLIENT.md](./docs/CLI_CLIENT.md) for complete architecture.**

**Future Phases:**
- Phase 3: Interactive mode & I/O proxying
- Phase 4: Agent management & approvals
- Phase 5: Polish & telemetry
- v0.3+: System prompts, MCP management, Web UI

**See [IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md) for detailed plan.**

---

## Success Metrics

### Phase 1 (Complete) ✅
- ✅ 20 tables + 3 views in PostgreSQL
- ✅ 60+ documentation files (streamlined)
- ✅ OpenAPI spec with 10+ endpoints
- ✅ 15+ type-safe SQL queries
- ✅ Code generation working end-to-end
- ✅ Complete automation via Makefile

### Phase 2 (Complete) ✅
- ✅ Authentication working (JWT + sessions)
- ✅ JWT Middleware implemented
- ✅ 43/43 tests passing
- ✅ ~88% code coverage
- ✅ Integration tests with real PostgreSQL

### Phase 3 (In Progress) 🎯
- [ ] Employee CRUD endpoints (0/5 complete)
- [ ] Organization management endpoints
- [ ] API response time <100ms (p95)
- [ ] 80%+ test coverage overall

---

## Roadmap

### Phase 1: ✅ COMPLETED (Foundation)
- Database schema, code generation, documentation

### Phase 2: ✅ COMPLETED (Authentication)
- JWT authentication, sessions, middleware

### Phase 3: In Progress (Employee Management)
- Employee CRUD, organization management

### Phase 4-8: Planned
- Agent/MCP configuration APIs
- Approval workflows
- Analytics endpoints
- Employee CLI client
- Admin web UI
- Production deployment

**See [docs/archive/MIGRATION_PLAN.md](./docs/archive/MIGRATION_PLAN.md) for original 10-week plan (historical reference).**

---

## Documentation Standards

### When to Update Docs

**Always update when:**
- Adding new tables → Regenerate ERD: `make generate-erd`
- Adding API endpoints → Update `openapi/spec.yaml`
- Adding SQL queries → Add to `sqlc/queries/*.sql`
- Changing architecture → Update this file (CLAUDE.md)

**How to update:**
```bash
# Database docs (auto-generated)
make generate-erd

# Manual docs (update manually)
vim CLAUDE.md
vim docs/TESTING.md
vim docs/DEVELOPMENT.md
```

---

**Foundation documentation is above this line. For implementation details, see linked documents.**

**Key Links:**
- 🚀 [Get Started](./docs/QUICKSTART.md)
- 🎯 [Next Tasks](./IMPLEMENTATION_ROADMAP.md)
- 🧪 [Testing Guide](./docs/TESTING.md)
- 🔧 [Development Guide](./docs/DEVELOPMENT.md)
- 📊 [Database ERD](./docs/ERD.md)

## Critical Usage Notes

### ✅ Qdrant MCP for Knowledge Management

**MANDATORY: Use Qdrant MCP for all knowledge operations**

- **Before any task**: Search using `mcp__code-search__qdrant-find`
- **During work**: Store findings using `mcp__code-search__qdrant-store`
- **After completion**: Update stored knowledge

**Setup:**
```bash
docker run -d --name claude-qdrant -p 6333:6333 qdrant/qdrant:latest
```