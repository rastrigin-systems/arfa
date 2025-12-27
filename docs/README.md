# Arfa Documentation

Welcome to the Arfa documentation. This guide covers everything you need to develop, deploy, and contribute to the Arfa AI Agent Security Gateway.

## Quick Links

| I want to... | Go to |
|--------------|-------|
| Set up development environment | [Getting Started](./development/getting-started.md) |
| Understand the architecture | [Architecture Overview](./architecture/overview.md) |
| Learn the PR workflow | [PR Workflow](./development/workflows.md) |
| Write tests | [Testing Guide](./development/testing.md) |
| Contribute code | [Contributing](./development/contributing.md) |
| Debug issues | [Debugging Guide](./development/debugging.md) |

## Documentation Structure

```
docs/
├── README.md                    # This file - documentation index
├── architecture/                # System architecture
│   ├── overview.md              # High-level architecture and vision
│   ├── control-service.md       # LLM traffic interception design
│   ├── logging.md               # Logging and telemetry
│   └── monorepo-structure.md    # Project organization
├── database/                    # Database documentation (auto-generated)
│   ├── README.md                # Database overview
│   ├── schema-reference.md      # Complete schema reference
│   └── public.*.md              # Per-table documentation
├── design/                      # Design documents
│   └── realtime-policies.md     # Real-time policy updates design
├── development/                 # Development guides
│   ├── getting-started.md       # Environment setup
│   ├── contributing.md          # Contribution guidelines
│   ├── workflows.md             # PR and Git workflow
│   ├── project-workflows.md     # Milestone and release planning
│   ├── testing.md               # Testing strategy and TDD
│   ├── debugging.md             # Debugging techniques
│   └── docker-testing.md        # Docker testing checklist
├── features/                    # Feature documentation
│   ├── authorization.md         # Role-based access control
│   ├── email-service.md         # Email notifications
│   ├── mcp-servers.md           # MCP server configuration
│   ├── tool-blocking.md         # Tool call policy enforcement
│   └── webhooks.md              # Webhook integrations
└── releases/                    # Release information
    └── RELEASES.md              # Version history and changelog
```

## Architecture

Arfa is a multi-tenant SaaS platform providing visibility and control for AI coding agents.

```
┌─────────────────────────────────────────────────────────────────┐
│                         SIEM                                    │
│              (Kibana / Splunk / Datadog / etc.)                 │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ Webhooks / OpenTelemetry
                              │
┌─────────────────────────────────────────────────────────────────┐
│                     ARFA SECURITY GATEWAY                       │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────────┐    │
│  │    CAPTURE    │  │    ENFORCE    │  │     FORWARD       │    │
│  │  Tool calls   │  │   Policies    │  │   To SIEM         │    │
│  └───────────────┘  └───────────────┘  └───────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │
    ┌─────────────────────────┼─────────────────────────┐
    │                         │                         │
┌───────────┐           ┌───────────┐           ┌───────────┐
│ Claude    │           │  Cursor   │           │ Windsurf  │
│ Code      │           │           │           │           │
└───────────┘           └───────────┘           └───────────┘
```

Learn more: [Architecture Overview](./architecture/overview.md)

## Services

| Service | Description | Documentation |
|---------|-------------|---------------|
| API | Go REST API server | [services/api/README.md](../services/api/README.md) |
| CLI | Go CLI tool (arfa-cli) | [services/cli/README.md](../services/cli/README.md) |
| Web | Next.js admin UI | [services/web/README.md](../services/web/README.md) |

## Getting Started

### Prerequisites

- Go 1.24+
- Node.js 18+ and pnpm
- Docker and Docker Compose
- PostgreSQL 15+ (via Docker)

### Quick Setup

```bash
# Clone repository
git clone https://github.com/rastrigin-systems/arfa.git
cd arfa

# Start database
make db-up

# Install tools and generate code
make install-tools
make generate

# Run tests
make test

# Start development servers
make dev-api        # Terminal 1: API on :8080
cd services/web && pnpm dev  # Terminal 2: Web on :3000
```

See [Getting Started](./development/getting-started.md) for detailed instructions.

## Contributing

We welcome contributions! Please read:

1. [Contributing Guidelines](./development/contributing.md) - Code standards and process
2. [PR Workflow](./development/workflows.md) - Branch naming, commits, PRs
3. [Testing Guide](./development/testing.md) - TDD workflow and test patterns

### Quick Contribution Workflow

```bash
# Create feature branch
git checkout -b feature/your-feature

# Make changes following TDD
# Write tests first, then implement

# Run tests
make test

# Create PR
gh pr create --title "feat: Your feature (#issue)" --body "..."
```

## Database

The database schema is auto-documented. See:

- [Database Overview](./database/README.md)
- [Schema Reference](./database/schema-reference.md)

To regenerate database docs after schema changes:
```bash
make generate-erd
```

## Features

| Feature | Status | Documentation |
|---------|--------|---------------|
| Authentication | Implemented | [authorization.md](./features/authorization.md) |
| Tool Blocking | Implemented | [tool-blocking.md](./features/tool-blocking.md) |
| Webhooks | Implemented | [webhooks.md](./features/webhooks.md) |
| MCP Servers | Implemented | [mcp-servers.md](./features/mcp-servers.md) |
| Email Service | Implemented | [email-service.md](./features/email-service.md) |
| Real-time Policies | Design | [realtime-policies.md](./design/realtime-policies.md) |

## License

MIT License - see [LICENSE](../LICENSE) for details.
