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
- [Next Tasks](./IMPLEMENTATION_ROADMAP.md) - What to build next

---

## Architecture

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
1. `schema.sql` - Database structure
2. `openapi/spec.yaml` - API contract

---

## Project Structure

```
ubik-enterprise/
├── CLAUDE.md              # ⭐ Start here
├── schema.sql             # Database schema
├── openapi/spec.yaml      # API contract
├── docs/                  # All documentation
├── generated/             # ⚠️ Auto-generated
├── internal/              # Your code here
│   ├── handlers/
│   ├── auth/
│   └── middleware/
├── tests/
│   ├── integration/
│   └── testutil/
└── cmd/server/            # API server
```

---

## Common Commands

```bash
make help              # Show all commands
make db-up            # Start PostgreSQL
make generate         # Generate all code (run after pulling changes!)
make test             # Run tests
make test-coverage    # View coverage report
make clean            # Clean generated files
```

**Note:** The `generated/` directory is NOT committed to git. Always run `make generate` after pulling changes that modify `schema.sql`, `openapi/spec.yaml`, or SQL queries.

---

## Tech Stack

- **Language**: Go 1.24+
- **Database**: PostgreSQL 15+ (multi-tenant with RLS)
- **API**: OpenAPI 3.0.3, Chi router
- **Code Generation**: oapi-codegen, sqlc, tbls
- **Testing**: testcontainers-go, gomock

---

## Current Status

### Phase 1: ✅ COMPLETE
- Database schema (20 tables + 3 views)
- Code generation pipeline
- Documentation (60+ files)

### Phase 2: ✅ COMPLETE
- JWT authentication system
- 43/43 tests passing
- ~88% code coverage
- Login, Logout, GetMe endpoints
- JWT middleware

### Phase 3: 🎯 IN PROGRESS
- Employee CRUD endpoints
- Organization management

**See [IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md) for detailed plan.**

---

## Contributing

See [CLAUDE.md](./CLAUDE.md) for complete development guide.

---

**Last Updated**: 2025-10-29
