# Enterprise AI Agent Management Platform - Migration Plan

## Project Codename: "Pivot"

**Goal**: Transform Ubik from single-user task automation to multi-tenant enterprise AI agent configuration management platform.

**Branch**: `pivot` (current)  
**Status**: 🟢 Planning Complete - Ready for Implementation  
**Started**: 2025-10-28

---

## 📋 Executive Summary

### What We're Building

Centralized management platform where companies can:
- Manage employees and teams
- Configure which AI agents (Claude Code, Cursor, Windsurf, etc.) employees can use
- Configure which MCP servers employees have access to
- Set policies and approval workflows
- Track usage and costs
- Employees sync configurations to their local machines via CLI

### What We're NOT Building (Removed Complexity)

- ❌ Mission orchestration and task execution
- ❌ Agent-to-agent communication
- ❌ Central artifact storage
- ❌ Event sourcing with full audit trail (simplified to activity logs)

### Architecture Philosophy

**Hybrid Documentation Approach**:
```
schema.sql (DB source of truth) → Auto-generated ERD (tbls)
                                    ↓
openapi.yaml (API source of truth) → Auto-generated Go code (oapi-codegen)
                                    ↓
                            Drift detection (CI checks)
```

---

## 🎯 Success Metrics

- [ ] Complete OpenAPI spec with 100% endpoint coverage
- [ ] Auto-generated ERD matches schema.sql
- [ ] Zero manual API validation code (all from OpenAPI)
- [ ] Admin UI can manage 1000+ employees
- [ ] Employee CLI syncs configs in <2s
- [ ] All code generated via `make generate`

---

## 📁 New Project Structure

```
ubik-enterprise/
├── MIGRATION_PLAN.md          # This file
├── DATABASE_SCHEMA.md         # ERD + schema documentation
├── schema.sql                 # PostgreSQL schema (DB source of truth)
│
├── openapi/
│   ├── spec.yaml              # OpenAPI 3.1 spec (API source of truth)
│   └── oapi-codegen.yaml      # oapi-codegen config
│
├── sqlc/
│   ├── queries/
│   │   ├── employees.sql      # SQL queries for employees
│   │   ├── agents.sql         # SQL queries for agents
│   │   ├── mcps.sql           # SQL queries for MCPs
│   │   └── usage.sql          # SQL queries for analytics
│   └── sqlc.yaml              # sqlc config
│
├── generated/                 # ⚠️ Never edit manually - regenerated via make
│   ├── api/                   # From oapi-codegen
│   │   ├── types.gen.go       # OpenAPI models
│   │   ├── server.gen.go      # HTTP server interface
│   │   └── spec.gen.go        # Embedded spec
│   └── db/                    # From sqlc
│       ├── models.go          # DB models
│       ├── querier.go         # DB interface
│       └── *.sql.go           # Type-safe queries
│
├── internal/
│   ├── handlers/              # HTTP request handlers
│   │   ├── employees.go       # Employee CRUD
│   │   ├── agents.go          # Agent configuration
│   │   ├── mcps.go            # MCP configuration
│   │   ├── approvals.go       # Approval workflows
│   │   └── analytics.go       # Usage analytics
│   │
│   ├── service/               # Business logic layer
│   │   ├── employee_service.go
│   │   ├── agent_service.go
│   │   ├── mcp_service.go
│   │   └── auth_service.go
│   │
│   ├── mapper/                # OpenAPI ↔ DB type conversion
│   │   └── mapper.go
│   │
│   ├── middleware/            # HTTP middleware
│   │   ├── auth.go            # JWT authentication
│   │   ├── rls.go             # Row-level security (set org_id)
│   │   ├── cors.go            # CORS headers
│   │   └── logging.go         # Request logging
│   │
│   └── validation/            # Custom validators
│       └── validators.go
│
├── cmd/
│   ├── server/                # API server
│   │   └── main.go
│   └── cli/                   # Employee CLI client
│       └── main.go
│
├── scripts/
│   ├── check-drift.js         # Detect OpenAPI ↔ DB drift
│   ├── seed-data.sh           # Load test data
│   └── reset-db.sh            # Drop and recreate DB
│
├── docs/                      # Auto-generated documentation
│   ├── schema.md              # ERD (from tbls)
│   ├── api.html               # API docs (from Redocly)
│   └── README.md              # Index
│
├── docker-compose.yml         # Local dev environment
├── Dockerfile                 # API server container
├── Makefile                   # Automation (make generate, make run)
├── go.mod
└── go.sum
```

---

## 🛠️ Technology Stack

### Core
- **Language**: Go 1.24+
- **Database**: PostgreSQL 15+ (replacing SQLite)
- **HTTP Router**: Chi (keep from Ubik)
- **Web UI**: Next.js 14 (reuse existing `/apps/web`)

### Code Generation
- **oapi-codegen**: Generate Go types, server stubs, validators from OpenAPI
- **sqlc**: Generate type-safe Go code from SQL queries
- **tbls**: Generate ERD diagrams from PostgreSQL schema

### Development Tools
- **Docker Compose**: Local PostgreSQL, Adminer
- **Air**: Live reload for Go development
- **Redocly CLI**: OpenAPI linting and doc generation

### Testing
- **testcontainers-go**: Spin up PostgreSQL for integration tests
- **httpexpect**: API testing

---

## 📝 Migration Steps

### Phase 1: Foundation (Week 1)

#### 1.1 Project Setup
- [x] Create `/pivot` directory structure
- [x] Create simplified database schema (17 tables)
- [x] Write migration plan (this document)
- [ ] Set up Docker Compose with PostgreSQL
- [ ] Initialize Go module for pivot API
- [ ] Install code generation tools

#### 1.2 Database & ERD
- [ ] Apply `schema.sql` to local PostgreSQL
- [ ] Configure tbls for ERD generation
- [ ] Generate initial ERD diagram
- [ ] Document all tables in `DATABASE_SCHEMA.md`

#### 1.3 Code Generation Setup
- [ ] Create `sqlc.yaml` configuration
- [ ] Write initial SQL queries (CRUD for employees)
- [ ] Create `oapi-codegen.yaml` configuration
- [ ] Write OpenAPI spec (employees, auth endpoints)
- [ ] Create `Makefile` with generation targets
- [ ] Test `make generate` workflow

**Deliverables**:
- ✅ Running PostgreSQL with schema applied
- ✅ Auto-generated ERD in `/pivot/docs/schema.md`
- ✅ Generated Go code in `/pivot/generated/`

---

### Phase 2: Core API (Week 2)

#### 2.1 Authentication & Authorization
- [ ] Implement JWT authentication
- [ ] Create login/logout endpoints
- [ ] Add auth middleware
- [ ] Implement Row-Level Security (RLS) middleware
- [ ] Session management

#### 2.2 Employee Management API
- [ ] `POST /api/v1/employees` - Create employee
- [ ] `GET /api/v1/employees/:id` - Get employee
- [ ] `GET /api/v1/employees` - List employees (org-scoped)
- [ ] `PATCH /api/v1/employees/:id` - Update employee
- [ ] `DELETE /api/v1/employees/:id` - Soft delete employee

#### 2.3 Organization & Team Management
- [ ] Organization CRUD endpoints
- [ ] Team CRUD endpoints
- [ ] Role management endpoints

**Deliverables**:
- ✅ Working authentication flow
- ✅ Employee management API with OpenAPI validation
- ✅ Integration tests with testcontainers

---

### Phase 3: Agent & MCP Configuration (Week 3)

#### 3.1 Agent Catalog & Configuration
- [ ] Agent catalog CRUD (admin only)
- [ ] Employee agent configuration endpoints
- [ ] Policy assignment endpoints
- [ ] Tool assignment endpoints

#### 3.2 MCP Catalog & Configuration
- [ ] MCP catalog CRUD (admin only)
- [ ] Employee MCP configuration endpoints
- [ ] Credential encryption/decryption
- [ ] MCP category management

#### 3.3 Sync Mechanism
- [ ] Generate sync tokens for configs
- [ ] `GET /api/v1/sync/agents` - Pull agent configs
- [ ] `GET /api/v1/sync/mcps` - Pull MCP configs
- [ ] `POST /api/v1/sync/heartbeat` - Update last_sync_at

**Deliverables**:
- ✅ Agent and MCP configuration APIs
- ✅ Sync endpoints for CLI client
- ✅ Encrypted credential storage

---

### Phase 4: Approval Workflows (Week 4)

#### 4.1 Request Submission
- [ ] `POST /api/v1/requests` - Submit agent/MCP request
- [ ] `GET /api/v1/requests` - List my requests
- [ ] `DELETE /api/v1/requests/:id` - Cancel request

#### 4.2 Approval Management
- [ ] `GET /api/v1/approvals/pending` - Pending approvals (manager)
- [ ] `POST /api/v1/approvals/:id/approve` - Approve request
- [ ] `POST /api/v1/approvals/:id/reject` - Reject request
- [ ] Email notifications (optional)

**Deliverables**:
- ✅ Complete approval workflow
- ✅ Manager dashboard endpoints

---

### Phase 5: Analytics & Usage Tracking (Week 5)

#### 5.1 Usage Tracking
- [ ] `POST /api/v1/usage` - Report usage from CLI
- [ ] Background job to aggregate usage
- [ ] Cost calculation logic

#### 5.2 Analytics Endpoints
- [ ] `GET /api/v1/analytics/org` - Org-level stats
- [ ] `GET /api/v1/analytics/team/:id` - Team-level stats
- [ ] `GET /api/v1/analytics/employee/:id` - Employee-level stats
- [ ] `GET /api/v1/analytics/spending` - Cost breakdown

#### 5.3 Activity Logs
- [ ] Middleware to log all API calls
- [ ] `GET /api/v1/activity-logs` - Audit trail viewer

**Deliverables**:
- ✅ Usage tracking system
- ✅ Analytics dashboard data
- ✅ Full audit trail

---

### Phase 6: Employee CLI Client (Week 6)

#### 6.1 CLI Core
- [ ] `ubik-cli login --org acme --email alice@acme.com`
- [ ] `ubik-cli logout`
- [ ] `ubik-cli whoami`
- [ ] JWT token storage in `~/.ubik-enterprise/auth.json`

#### 6.2 Configuration Sync
- [ ] `ubik-cli sync` - Pull all configs from server
- [ ] `ubik-cli status` - Show current configs
- [ ] `ubik-cli config list` - List agents and MCPs
- [ ] Local config storage in `~/.ubik-enterprise/config.json`

#### 6.3 Request Management
- [ ] `ubik-cli request agent <name>` - Request new agent
- [ ] `ubik-cli request mcp <name>` - Request new MCP
- [ ] `ubik-cli requests list` - Show my requests

#### 6.4 Usage Reporting (Optional)
- [ ] Background process to report local usage
- [ ] `ubik-cli usage report` - Manual usage sync

**Deliverables**:
- ✅ Functional CLI for employees
- ✅ Config sync working
- ✅ Request submission from CLI

---

### Phase 7: Admin Web UI (Week 7-8)

#### 7.1 Reuse Existing Next.js App
- [ ] Copy `/apps/web` to `/pivot/web`
- [ ] Strip out mission/task UI components
- [ ] Update API client to call new endpoints

#### 7.2 Core Pages
- [ ] Dashboard (org overview)
- [ ] Employees management table
- [ ] Teams management
- [ ] Agent catalog configuration
- [ ] MCP catalog configuration
- [ ] Approval queue

#### 7.3 Analytics Views
- [ ] Usage charts (by team, employee, agent)
- [ ] Cost tracking dashboard
- [ ] Activity log viewer

**Deliverables**:
- ✅ Admin dashboard
- ✅ Employee, agent, MCP management UI
- ✅ Analytics visualizations

---

### Phase 8: Production Readiness (Week 9-10)

#### 8.1 Security Hardening
- [ ] Implement rate limiting
- [ ] Add CSRF protection
- [ ] API key authentication (for CLI)
- [ ] Audit security policies

#### 8.2 Performance Optimization
- [ ] Add database indexes
- [ ] Implement caching (Redis)
- [ ] Connection pooling
- [ ] Query optimization

#### 8.3 Deployment
- [ ] Multi-stage Dockerfile
- [ ] Kubernetes manifests (optional)
- [ ] Database migration strategy
- [ ] Environment configuration

#### 8.4 Documentation
- [ ] API documentation (Swagger UI)
- [ ] CLI documentation
- [ ] Admin guide
- [ ] Employee onboarding guide

**Deliverables**:
- ✅ Production-ready API server
- ✅ Deployment artifacts
- ✅ Complete documentation

---

## 🔄 Daily Development Workflow

```bash
# 1. Make database schema changes
vim pivot/schema.sql

# 2. Apply to local database
make db-migrate

# 3. Regenerate ERD
make generate-erd

# 4. Update OpenAPI spec (if needed)
vim pivot/openapi/spec.yaml

# 5. Regenerate all code
make generate

# 6. Check for drift
make check-drift

# 7. Run tests
make test

# 8. Start dev server
make dev
```

---

## 🧪 Testing Strategy

### Unit Tests
- Test business logic in `/internal/service/`
- Mock database with sqlc interface
- Mock HTTP handlers

### Integration Tests
- Use testcontainers-go for PostgreSQL
- Test full API flows (auth → CRUD → sync)
- Test RLS policies

### End-to-End Tests
- Test CLI + API + Web UI together
- Playwright for web UI testing
- Shell scripts for CLI testing

---

## 📊 Key Metrics to Track

### Development Metrics
- Lines of generated code vs hand-written
- API endpoint coverage by OpenAPI spec
- Test coverage (target: >80%)
- Build time
- Time from schema change → working API

### Runtime Metrics
- API response times (p50, p95, p99)
- Database query performance
- CLI sync time
- Memory usage
- Active sessions

---

## ⚠️ Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| OpenAPI ↔ DB drift | High | Automated drift detection in CI |
| Schema changes break API | High | Versioned API (`/api/v1/`) |
| Code generation brittleness | Medium | Pin tool versions in go.mod |
| PostgreSQL migration complexity | Medium | Use golang-migrate with rollback |
| Learning curve for new tools | Low | Comprehensive docs + examples |

---

## 🎓 Learning Resources

### Code Generation
- [oapi-codegen docs](https://github.com/deepmap/oapi-codegen)
- [sqlc docs](https://docs.sqlc.dev/)
- [tbls docs](https://github.com/k1LoW/tbls)

### OpenAPI
- [OpenAPI 3.1 Spec](https://spec.openapis.org/oas/v3.1.0)
- [OpenAPI Guide](https://swagger.io/docs/specification/about/)

### PostgreSQL
- [PostgreSQL RLS](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)
- [Indexes](https://www.postgresql.org/docs/current/indexes.html)

---

## 📞 Decision Log

### Decision 1: PostgreSQL over SQLite
**Date**: 2025-10-28  
**Reason**: Multi-tenancy requires RLS, better concurrency, and production scalability  
**Alternative Considered**: Continue with SQLite  
**Outcome**: PostgreSQL selected

### Decision 2: Hybrid Documentation (ERD + OpenAPI)
**Date**: 2025-10-28  
**Reason**: DB schema and API DTOs serve different purposes, auto-generate where possible  
**Alternative Considered**: Single source of truth (OpenAPI or DB)  
**Outcome**: schema.sql → ERD, openapi.yaml → API code

### Decision 3: oapi-codegen over go-swagger
**Date**: 2025-10-28  
**Reason**: Better OpenAPI 3.1 support, cleaner generated code, active maintenance  
**Alternative Considered**: go-swagger, ogen  
**Outcome**: oapi-codegen selected

### Decision 4: Remove Mission/Task Execution
**Date**: 2025-10-28  
**Reason**: Pivot to configuration management, not orchestration  
**Alternative Considered**: Keep dual functionality  
**Outcome**: Simplified to 17 tables, focus on config sync

---

## 📅 Timeline

| Phase | Duration | Target Completion |
|-------|----------|------------------|
| Phase 1: Foundation | 1 week | Week of 2025-10-28 |
| Phase 2: Core API | 1 week | Week of 2025-11-04 |
| Phase 3: Agent/MCP Config | 1 week | Week of 2025-11-11 |
| Phase 4: Approval Workflows | 1 week | Week of 2025-11-18 |
| Phase 5: Analytics | 1 week | Week of 2025-11-25 |
| Phase 6: Employee CLI | 1 week | Week of 2025-12-02 |
| Phase 7: Admin Web UI | 2 weeks | Week of 2025-12-16 |
| Phase 8: Production Readiness | 2 weeks | Week of 2025-12-30 |

**Total**: ~10 weeks to production-ready MVP

---

## 🚀 Next Immediate Actions

1. ✅ Create this migration plan
2. ⏳ Set up Docker Compose with PostgreSQL
3. ⏳ Install code generation tools (tbls, oapi-codegen, sqlc)
4. ⏳ Generate initial ERD from schema.sql
5. ⏳ Create Makefile for automation
6. ⏳ Write initial OpenAPI spec (auth + employees)
7. ⏳ Update CLAUDE.md with pivot documentation link

---

## 📖 Related Documentation

- [Database Schema](./DATABASE_SCHEMA.md) - Complete ERD and table definitions
- [OpenAPI Spec](./openapi/spec.yaml) - API contract (to be created)
- [Project README](../README.md) - Original Ubik documentation
- [Main Project Instructions](../CLAUDE.md) - Project context

---

**Last Updated**: 2025-10-28  
**Status**: 🟢 Ready to Start Implementation
