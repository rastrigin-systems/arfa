---
sidebar_position: 1
slug: /
---

# Arfa Architecture

Arfa is an AI Agent Security Platform that provides visibility and control over AI coding assistant usage.

## System Overview

```
┌─────────────────────────────────────────────────────────────┐
│  Developer's Machine                                        │
│  ┌──────────────┐           ┌──────────────┐                │
│  │ Claude Code  │──────────▶│ Arfa Proxy   │                │
│  │ Cursor       │  HTTPS    │ (localhost)  │                │
│  │ Windsurf     │           └──────┬───────┘                │
│  └──────────────┘                  │                        │
│                           Intercept, Log, Enforce           │
└────────────────────────────────────┼────────────────────────┘
                                     │
                            ┌────────▼─────────┐
                            │   LLM APIs       │
                            │  (Anthropic,     │
                            │   OpenAI)        │
                            └────────┬─────────┘
                                     │
                            ┌────────▼─────────┐
                            │  Arfa Platform   │
                            │  (Logs, Policy)  │
                            └──────────────────┘
```

## Core Components

| Component | Technology | Purpose |
|-----------|------------|---------|
| **API Server** | Go + PostgreSQL | REST API, data persistence, policy storage |
| **CLI Proxy** | Go + goproxy | HTTPS interception, tool call capture |
| **Web Dashboard** | Next.js | Admin interface, log visualization |

## Key Design Decisions

This documentation covers the architectural decisions that shaped Arfa:

- **[Go Workspace Monorepo](adr/001-go-workspace-monorepo)** - Why we use Go workspaces
- **[Transparent HTTPS Proxy](adr/002-transparent-https-proxy)** - How we intercept LLM traffic
- **[Multi-tenant RLS](adr/003-multi-tenant-rls)** - How we isolate organization data
- **[Code Generation](adr/004-code-generation-pipeline)** - Why we generate code from specs

## Getting Started

**Building the services:**
```bash
make build       # Build all services
make test        # Run tests
make generate    # Regenerate code from specs
```

**Running locally:**
```bash
make db-up       # Start PostgreSQL
make dev         # Start all services
```

## Source Code

- [GitHub Repository](https://github.com/rastrigin-systems/arfa)
- [OpenAPI Specification](https://github.com/rastrigin-systems/arfa/blob/main/platform/api-spec/spec.yaml)
- [Database Schema](https://github.com/rastrigin-systems/arfa/blob/main/platform/database/schema.sql)
