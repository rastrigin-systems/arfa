# Ubik Enterprise — AI Agent Management Platform

**Multi-tenant SaaS platform for centralized AI agent and MCP configuration management**

---

## Quick Navigation

**Working on a specific service?**
- 🔧 **API Development** → [services/api/CLAUDE.md](./services/api/CLAUDE.md)
- 💻 **CLI Development** → [services/cli/CLAUDE.md](./services/cli/CLAUDE.md)
- 🎨 **Web UI Development** → [services/web/CLAUDE.md](./services/web/CLAUDE.md)

**New to the project?**
1. [docs/QUICKSTART.md](./docs/QUICKSTART.md) - 5-minute setup
2. [docs/ERD.md](./docs/ERD.md) - Database schema visualization

**Common tasks:**
- [docs/QUICK_REFERENCE.md](./docs/QUICK_REFERENCE.md) - Command reference
- [docs/DEV_WORKFLOW.md](./docs/DEV_WORKFLOW.md) - PR workflow
- [docs/TESTING.md](./docs/TESTING.md) - Testing guide

---

## System Overview

### Purpose

Multi-tenant SaaS platform for companies to centrally manage AI agent (Claude Code, Cursor, Windsurf) and MCP server configurations for employees.

**Core Value:** Centralized control, policy enforcement, and visibility into AI agent usage across organizations.

### Architecture

```
Go Workspace Monorepo
├── services/api/       (REST API + WebSocket)
├── services/cli/       (Employee CLI tool)
├── services/web/       (Next.js admin UI)
├── pkg/types/          (Shared Go types)
├── platform/           (Schema, API spec, database)
└── generated/          (Auto-generated code, not committed)
```

**Code Generation Pipeline:**
```
platform/database/schema.sql → sqlc → generated/db/*.go
platform/api-spec/spec.yaml → oapi-codegen → generated/api/*.go
                            → openapi-typescript → services/web/lib/api/schema.ts
```

**Database:** PostgreSQL 15+ with 20 tables + 3 views, Row-Level Security for multi-tenancy

**Tech Stack:** Go 1.24+, Next.js 14, PostgreSQL, Docker, OpenAPI 3.0.3

---

## Essential Commands

```bash
# First-time setup
make db-up              # Start PostgreSQL
make install-tools      # Install code generation tools (one-time)
make install-hooks      # Install git hooks (one-time)
make generate           # Generate all code from schema/spec

# Development
make test               # Run all tests
make build              # Build all services
make clean              # Clean generated files

# Database
make db-reset           # Reset database (⚠️ deletes data)

# Code generation (after schema/API changes)
make generate           # Regenerate everything
make generate-api       # API code only
make generate-db        # Database code only
make generate-erd       # Documentation only
```

**See [docs/QUICK_REFERENCE.md](./docs/QUICK_REFERENCE.md) for complete command reference.**

---

## Critical Rules

### 1. Code Generation

**NEVER edit generated files** - they are completely regenerated!

**Generated code (NOT committed to git):**
- `generated/` directory (Go API + DB code)
- `services/web/lib/api/schema.ts` (TypeScript types)

**After changing schema or API spec:**
```bash
# 1. Edit source files
vim platform/database/schema.sql
vim platform/api-spec/spec.yaml

# 2. Regenerate everything
make generate

# 3. Commit source files + docs (not generated code)
git add platform/ docs/
git commit -m "feat: Add new endpoint"
```

**IMPORTANT:** CI/CD regenerates code automatically and FAILS if docs are stale.

---

### 2. Multi-Tenancy

**CRITICAL:** All database queries MUST be organization-scoped to prevent data leakage.

```go
// ✅ GOOD - Scoped to organization
employees, err := db.ListEmployees(ctx, orgID, status)

// ❌ BAD - Exposes all organizations!
employees, err := db.ListAllEmployees(ctx)
```

Row-Level Security (RLS) policies provide safety net, but queries MUST include `org_id`.

**See [docs/DATABASE.md](./docs/DATABASE.md#multi-tenancy) for details.**

---

### 3. Test-Driven Development (TDD)

**CRITICAL: ALWAYS write tests BEFORE implementation**

**Mandatory TDD workflow:**
1. ✅ Write failing test FIRST
2. ✅ Implement minimal code to pass test
3. ✅ Refactor with tests passing
4. ❌ NEVER write implementation before tests

**Target coverage:** 85% overall (excluding generated code)

**See [docs/TESTING.md](./docs/TESTING.md) for complete testing guide.**

---

### 4. Pull Request Workflow

**CRITICAL: ALL code changes MUST go through Pull Requests**

Branch protection BLOCKS direct commits to `main`.

**Required steps for EVERY change:**
1. ✅ Create feature branch: `feature/{issue-number}-{description}`
2. ✅ Implement changes following TDD
3. ✅ Run tests locally: `make test`
4. ✅ Commit with descriptive message
5. ✅ Push branch
6. ✅ Create PR with proper title: `feat: Description (#123)`
7. ✅ Wait for CI/CD checks to pass
8. ✅ Merge when approved

**IMPORTANT:** PR title MUST include issue number for automatic linking: `feat: Add login endpoint (#123)`

**See [docs/DEV_WORKFLOW.md](./docs/DEV_WORKFLOW.md) for complete workflow.**

---

### 5. UI Development

**CRITICAL: Wireframes required for ALL UI changes**

**Mandatory UI workflow:**
1. ✅ Request wireframes from **product-designer agent** FIRST
2. ✅ Wait for wireframes approval
3. ✅ Implement UI matching wireframes exactly
4. ❌ NEVER implement UI without wireframes

**Wireframe location:** `docs/wireframes/`

**See [services/web/CLAUDE.md](./services/web/CLAUDE.md) for Web UI development details.**
**See [.claude/agents/product-designer.md](./.claude/agents/product-designer.md) for designer agent.**

---

### 6. Debugging

**Golden Rule: Check the Data, Not Just the Code**

When tests or operations fail unexpectedly:
1. ✅ Add request/response logging FIRST
2. ✅ Verify database state (foreign keys, seed data)
3. ✅ Check for stale cache (`~/.ubik/`, binaries)
4. ✅ Rebuild binaries

**Common pitfalls:**
- ❌ Assuming code is wrong when data is wrong
- ❌ Testing with stale binaries
- ❌ Not checking cache invalidation

**See [docs/DEBUGGING.md](./docs/DEBUGGING.md) for complete debugging guide.**

---

### 7. Docker Testing

**CRITICAL: ALWAYS test Docker builds locally before deploying**

```bash
# 1. Build image
docker build -f services/api/Dockerfile.gcp -t ubik-api-test .

# 2. Verify files in image
docker run --rm ubik-api-test ls -la /app/platform/api-spec/

# 3. Test container
docker run --rm -p 8080:8080 ubik-api-test

# 4. Verify endpoints
curl http://localhost:8080/api/v1/health
```

**When Docker testing is required:**
- ✅ ANY Dockerfile change
- ✅ ANY cloudbuild.yaml change
- ✅ New API endpoints
- ✅ New file resources needed

**See [docs/DOCKER_TESTING_CHECKLIST.md](./docs/DOCKER_TESTING_CHECKLIST.md) for complete guide.**

---

### 8. Tool Selection

**CRITICAL: Choose the right tool for the task**

**Context cost awareness:**
- Playwright snapshot: ~12,000 tokens
- Screenshot: ~1,000 tokens
- Curl: ~100 tokens

**Golden rule: Use the simplest tool that accomplishes the task**

**API testing:**
```bash
# ✅ GOOD - Efficient
curl http://localhost:8080/api/v1/health

# ❌ BAD - Wastes ~50k tokens
# Using Playwright to test API endpoints
```

**Best practices:**
1. Test APIs with curl, not Playwright
2. Use screenshots for visual verification
3. Minimize page snapshots
4. Monitor token usage

---

## Documentation Map

### Getting Started
- **[docs/QUICKSTART.md](./docs/QUICKSTART.md)** - 5-minute setup guide
- **[docs/ERD.md](./docs/ERD.md)** - Database schema visualization

### Development
- **[docs/QUICK_REFERENCE.md](./docs/QUICK_REFERENCE.md)** - Command cheat sheet
- **[docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)** - Development workflow
- **[docs/DEV_WORKFLOW.md](./docs/DEV_WORKFLOW.md)** ⭐ - PR workflow (mandatory)
- **[docs/TESTING.md](./docs/TESTING.md)** ⭐ - Testing guide (TDD workflow)
- **[docs/DEBUGGING.md](./docs/DEBUGGING.md)** - Debugging strategies

### Database
- **[docs/DATABASE.md](./docs/DATABASE.md)** - Database operations
- **[docs/ERD.md](./docs/ERD.md)** - Visual schema (auto-generated)
- **[docs/README.md](./docs/README.md)** - Technical reference (auto-generated)
- **[docs/public.*.md](./docs/)** - Per-table docs (auto-generated)

### Service Documentation
- **[services/api/CLAUDE.md](./services/api/CLAUDE.md)** - API server development
- **[services/cli/CLAUDE.md](./services/cli/CLAUDE.md)** - CLI client development
- **[services/web/CLAUDE.md](./services/web/CLAUDE.md)** - Web UI development

### CLI Documentation
- **[docs/CLI_CLIENT.md](./docs/CLI_CLIENT.md)** - CLI architecture
- **[docs/CLI_PHASE1_COMPLETE.md](./docs/CLI_PHASE1_COMPLETE.md)** - Phase 1 details
- **[docs/CLI_PHASE2_COMPLETE.md](./docs/CLI_PHASE2_COMPLETE.md)** - Phase 2 details
- **[docs/CLI_PHASE3_COMPLETE.md](./docs/CLI_PHASE3_COMPLETE.md)** - Phase 3 details
- **[docs/CLI_PHASE4_COMPLETE.md](./docs/CLI_PHASE4_COMPLETE.md)** - Phase 4 details

### Operations
- **[docs/WORKFLOWS.md](./docs/WORKFLOWS.md)** - Milestone planning
- **[.claude/skills/release-manager/SKILL.md](./.claude/skills/release-manager/SKILL.md)** - Release workflow
- **[docs/RELEASES.md](./docs/RELEASES.md)** - Release history

### AI Agents
- **[.claude/agents/go-backend-developer.md](./.claude/agents/go-backend-developer.md)** - Backend development
- **[.claude/agents/frontend-developer.md](./.claude/agents/frontend-developer.md)** - Frontend development
- **[.claude/agents/product-designer.md](./.claude/agents/product-designer.md)** - UI/UX design
- **[.claude/agents/README.md](./.claude/agents/README.md)** - Complete agent documentation

---

## Database Schema

**20 Tables + 3 Views**

| Category | Tables | Count |
|----------|--------|-------|
| Core Organization | organizations, subscriptions, teams, roles, employees | 5 |
| Agent Management | agent_catalog, tools, policies, agent_tools, agent_policies, team_policies, employee_agent_configs | 7 |
| MCP Configuration | mcp_categories, mcp_catalog, employee_mcp_configs | 3 |
| Authentication | sessions | 1 |
| Approvals | agent_requests, approvals | 2 |
| Analytics | activity_logs, usage_records | 2 |
| **Views** | v_employee_agents, v_employee_mcps, v_pending_approvals | 3 |

**See [docs/ERD.md](./docs/ERD.md) for complete visual schema.**

---

## Current Status

**Version:** 0.2.0
**Last Updated:** 2025-11-05
**Status:** CLI Phase 4 Complete - Ready for v0.2.0 Release

### Completed
- ✅ **Phase 1:** Database schema, code generation, documentation
- ✅ **Phase 2:** API (v0.1.0) - 39 endpoints, 144+ tests, 73-88% coverage
- ✅ **Phase 3:** CLI (v0.2.0) - Authentication, sync, Docker integration, 79 tests

### In Progress
- 🎯 **v0.3.0:** Web UI Foundation - Next.js 14, authentication, agent catalog

### Planned
- **v0.4.0:** Analytics & Approvals
- **v0.5.0:** System Prompts & MCP Management
- **v1.0.0:** Production Release

**See [docs/WORKFLOWS.md](./docs/WORKFLOWS.md) for milestone planning.**
**See [docs/RELEASES.md](./docs/RELEASES.md) for release history.**

---

## Documentation Standards

### When to Update Docs

**Always update when:**
- Adding tables → `make generate-erd`
- Adding API endpoints → Update `platform/api-spec/spec.yaml`
- Adding SQL queries → Add to `platform/database/sqlc/queries/*.sql`
- Changing architecture → Update CLAUDE.md files

**How to update:**
```bash
# Auto-generated docs
make generate-erd       # Regenerate ERD and schema docs

# Manual docs
vim CLAUDE.md
vim services/*/CLAUDE.md
vim docs/*.md
```

---

**For service-specific details, see:**
- [services/api/CLAUDE.md](./services/api/CLAUDE.md) - API server development
- [services/cli/CLAUDE.md](./services/cli/CLAUDE.md) - CLI client development
- [services/web/CLAUDE.md](./services/web/CLAUDE.md) - Web UI development
