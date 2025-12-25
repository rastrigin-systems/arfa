# Arfa OSS v1.0 Scope

## Overview

This document defines what features are included in the open source release and what will be removed or deferred.

---

## Included in OSS v1.0

### Core Proxy Features
| Feature | Component | Description |
|---------|-----------|-------------|
| HTTPS Proxy | CLI | MITM proxy intercepting AI agent traffic (Claude Code, Cursor, Windsurf) |
| Activity Logging | CLI + API | Request/response capture, tool call logging |
| Tool Blocking | CLI | Pre-flight policy enforcement (block dangerous commands) |
| Control Service | CLI | Centralized config, extensible policy handlers |

### Platform Features
| Feature | Component | Description |
|---------|-----------|-------------|
| Webhooks | API + CLI | Event forwarding to external systems (Slack, SIEM, etc.) |
| Web UI | Web | Admin dashboard for logs, policies, employees |
| Multi-tenant API | API | Organization isolation with PostgreSQL RLS |
| Authentication | API | JWT-based auth, password reset, sessions |

### Data Model (Simplified)
```
Organization
├── Employees (users who use AI agents)
├── Teams (grouping for policy assignment)
├── Policies (tool blocking rules)
└── Activity Logs (audit trail)
```

---

## To Be Removed

### Billing/Subscriptions
**Reason:** OSS doesn't need billing tiers

| Table | Action |
|-------|--------|
| `subscriptions` | DELETE table |
| Subscription queries | DELETE |
| Subscription handlers | DELETE |

**Files to remove:**
- `platform/database/sqlc/queries/subscriptions.sql`
- `services/api/internal/handlers/subscriptions.go`
- `services/api/internal/handlers/subscriptions_test.go`

---

### Agent Configs (MCP, Skills, Config Cascade)
**Reason:** Over-engineered for v1. Tool blocking policies are sufficient.

| Table | Action |
|-------|--------|
| `skill_catalog` | DELETE table |
| `employee_skills` | DELETE table |
| `org_agent_configs` | DELETE table |
| `team_agent_configs` | DELETE table |
| `employee_agent_configs` | DELETE table |

**Files to remove:**
- `platform/database/sqlc/queries/skills.sql`
- `platform/database/migrations/001_skills_and_mcp.sql`
- `services/api/internal/handlers/skills.go`
- `services/api/internal/service/config_resolver.go`
- `services/api/internal/service/config_resolver_test.go`
- `services/cli/internal/commands/skills/`
- `services/cli/internal/skill/`
- Related integration tests

---

### Agent Catalog
**Reason:** Not needed if we're not managing agent configs

| Table | Action |
|-------|--------|
| `agent_catalog` | DELETE table |
| `agent_tools` | DELETE table |
| `agent_policies` | DELETE table |

---

### Agent Requests (Approval Workflow)
**Reason:** Was for requesting access to new agents - not needed without agent configs

| Table | Action |
|-------|--------|
| `agent_requests` | DELETE table |
| `approvals` | DELETE table |

**Files to remove:**
- `platform/database/sqlc/queries/agent_requests.sql`
- `services/api/internal/handlers/agent_requests.go`

---

### Claude Tokens
**Reason:** Was for managed API key distribution - keep auth simple

| Table | Action |
|-------|--------|
| `claude_tokens` | DELETE table |

**Files to remove:**
- `platform/database/sqlc/queries/claude_tokens.sql`
- `services/api/internal/handlers/claude_tokens.go`

---

### Usage Records
**Reason:** Billing/metering data - not needed for OSS. Activity logs provide audit trail.

| Table | Action |
|-------|--------|
| `usage_records` | DELETE table |

**Files to remove:**
- `platform/database/sqlc/queries/usage_records.sql`
- `services/api/internal/handlers/usage_stats.go`
- `services/api/internal/handlers/usage_stats_test.go`

---

## To Keep (Reconsidered)

### Teams
**Keep.** Still valuable for:
- Grouping employees for policy assignment (`team_policies`)
- Filtering logs by team
- Future: team-level reporting

### Employees
**Keep.** Core to multi-tenant model:
- Links to organization
- Links to activity logs
- Can be assigned to teams

### Policies & Tool Blocking
**Keep.** This IS the core value:
- `policies` table
- `team_policies` (assign policies to teams)
- `employee_policies` (individual overrides)

### Roles
**Keep.** Needed for authorization:
- admin, manager, user roles
- Controls who can manage policies, view logs, etc.

---

## Simplified Schema (Post-Cleanup)

```sql
-- Core
organizations
employees
teams
roles
sessions
password_reset_tokens

-- Policies (Tool Blocking)
policies
team_policies
employee_policies

-- Logging
activity_logs

-- Webhooks
webhook_destinations
webhook_deliveries

-- Invitations
invitations
```

**Removed (12 tables):**
- subscriptions
- skill_catalog
- employee_skills
- org_agent_configs
- team_agent_configs
- employee_agent_configs
- agent_catalog
- agent_tools
- agent_policies
- agent_requests
- approvals
- claude_tokens
- usage_records

---

## Implementation Order

### Phase 1: Database Cleanup ✅ COMPLETED
1. [x] Drop removed tables from schema.sql (no migrations, clean start)
2. [x] Remove related SQL queries
3. [x] Regenerate sqlc code

### Phase 2: API Cleanup ✅ COMPLETED
1. [x] Remove handlers for deleted features
2. [x] Remove services (config_resolver, etc.)
3. [x] Update routes
4. [x] Fix/remove broken tests

### Phase 3: CLI Cleanup ✅ COMPLETED
1. [x] Remove skills commands
2. [x] Remove skill service from container
3. [x] Proxy already focused on logging + tool blocking

### Phase 4: Web UI Cleanup
1. [ ] Remove agent config pages
2. [ ] Remove skills pages
3. [ ] Simplify navigation

### Phase 5: Documentation
1. [ ] Update README for simplified feature set
2. [ ] Update API spec
3. [ ] Create CONTRIBUTING.md

---

## Decisions Made

1. **Invitations** - KEEP, but show invite link in UI (email optional)
   - Admin creates invitation → UI displays copyable invite link
   - Email service remains optional enhancement
   - No email config required for self-hosted

## Decisions Made (continued)

2. **Usage Records** - REMOVE (billing data, not needed for OSS)
3. **Views** - REMOVE `v_pending_approvals` (part of approval workflow being removed)

---

*This document will be updated as decisions are made.*
