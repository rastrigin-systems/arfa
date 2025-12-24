# Ubik Documentation

Complete documentation for the Ubik AI Agent Security Platform.

---

## Quick Start

**New to Ubik?** Start here:
1. [Architecture Overview](architecture/overview.md) - Understand the system
2. [Getting Started](development/getting-started.md) - Set up your environment
3. [Database Schema](database/erd.md) - Explore the data model

---

## 🏗️ Architecture

High-level system design and technical decisions.

- **[System Overview](architecture/overview.md)** - AI Agent Security Gateway vision
- **[Control Service](architecture/control-service.md)** - Proxy architecture and pipeline
- **[Logging Architecture](architecture/logging.md)** - Activity logging system
- **[Monorepo Structure](architecture/monorepo-structure.md)** - Repository organization

---

## 💻 Development

Guides for contributing to Ubik.

### Getting Started
- **[Getting Started](development/getting-started.md)** - Quick reference for common tasks
- **[Contributing Guide](development/contributing.md)** - Development workflow and standards

### Testing & Debugging
- **[Testing Guide](development/testing.md)** - Unit, integration, and E2E testing
- **[Debugging Guide](development/debugging.md)** - Troubleshooting strategies
- **[Docker Testing](development/docker-testing.md)** - Testing Docker builds locally

### Workflows
- **[Git Workflow](development/workflows.md)** - PR process and branching strategy
- **[Project Workflows](development/project-workflows.md)** - Milestone planning and releases

---

## 🗄️ Database

Database schema, ERD, and table documentation.

- **[Schema Overview](database/schema.md)** - Database operations and multi-tenancy
- **[ERD Diagram](database/erd.md)** - Visual schema with relationships
- **[Schema Reference](database/schema-reference.md)** - Complete table listing
- **[Schema JSON](database/schema.json)** - Machine-readable schema
- **[Table Docs](database/tables/)** - Auto-generated per-table documentation

### Key Tables
- [organizations](database/tables/public.organizations.md) - Multi-tenant organizations
- [employees](database/tables/public.employees.md) - Organization users
- [agent_catalog](database/tables/public.agent_catalog.md) - Available AI agents
- [tool_policies](database/tables/public.tool_policies.md) - Security policies
- [activity_logs](database/tables/public.activity_logs.md) - Audit trail

---

## ✨ Features

Feature-specific documentation.

- **[Webhooks](features/webhooks.md)** - Webhook event forwarding
- **[Tool Blocking](features/tool-blocking.md)** - Policy-based tool blocking
- **[MCP Servers](features/mcp-servers.md)** - Model Context Protocol integration
- **[Authorization](features/authorization.md)** - Auth and permission system
- **[Email Service](features/email-service.md)** - Email notifications

---

## 📦 Releases

- **[Release History](releases/RELEASES.md)** - Version history and changelogs

---

## Service-Specific Documentation

Each service has detailed documentation in its directory:

- **[API Server](../services/api/CLAUDE.md)** - REST API development
- **[CLI Client](../services/cli/CLAUDE.md)** - CLI development
- **[Web UI](../services/web/CLAUDE.md)** - Next.js UI development

---

## External Resources

- **[GitHub Repository](https://github.com/rastrigin-systems/ubik)** - Source code
- **[Issues](https://github.com/rastrigin-systems/ubik/issues)** - Bug reports and feature requests
- **[Discussions](https://github.com/rastrigin-systems/ubik/discussions)** - Community Q&A

---

## Documentation Index

### By Topic

**Getting Started:**
- [Quick Reference](development/getting-started.md)
- [Contributing](development/contributing.md)
- [Architecture Overview](architecture/overview.md)

**Architecture & Design:**
- [System Overview](architecture/overview.md)
- [Control Service](architecture/control-service.md)
- [Logging](architecture/logging.md)
- [Monorepo Structure](architecture/monorepo-structure.md)

**Development:**
- [Testing](development/testing.md)
- [Debugging](development/debugging.md)
- [Git Workflow](development/workflows.md)
- [Docker Testing](development/docker-testing.md)

**Database:**
- [Schema](database/schema.md)
- [ERD](database/erd.md)
- [Tables](database/tables/)

**Features:**
- [Webhooks](features/webhooks.md)
- [Tool Blocking](features/tool-blocking.md)
- [MCP Servers](features/mcp-servers.md)
- [Authorization](features/authorization.md)
- [Email Service](features/email-service.md)

---

## Contributing to Docs

Found a typo? Want to improve documentation? See [Contributing Guide](development/contributing.md).

**Auto-generated docs:**
- Database table docs are generated from schema - edit `platform/database/schema.sql`
- ERD is generated from schema - run `make generate-docs`

**Manual docs:**
- All other docs are manually maintained
- Follow [Markdown style guide](development/contributing.md#documentation-style)
