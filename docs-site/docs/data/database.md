---
sidebar_position: 1
---

# Database Architecture

Arfa uses PostgreSQL 15+ with Row-Level Security for multi-tenant data isolation.

## Schema Overview

20 tables + 3 views organized into domains:

| Domain | Tables |
|--------|--------|
| **Organization** | organizations, subscriptions, teams, roles, employees |
| **Policy** | policies, team_policies, employee_policies |
| **Activity** | activity_logs, usage_records, sessions |
| **Webhooks** | webhook_destinations, webhook_deliveries |

## Entity Relationships

```
organizations (root)
    │
    ├── employees ─────────────────┬── sessions
    │       │                      └── activity_logs
    │       └── employee_policies
    │
    ├── teams ─────────────────────┬── employees
    │       └── team_policies
    │
    ├── policies
    │
    └── webhook_destinations
            └── webhook_deliveries
```

## Key Tables

### organizations

```sql
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### activity_logs

```sql
CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id),
    employee_id UUID REFERENCES employees(id),
    event_type VARCHAR(50) NOT NULL,
    content TEXT,
    payload JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Indexes

```sql
-- Activity log queries
CREATE INDEX idx_activity_logs_org_created
    ON activity_logs(org_id, created_at DESC);

-- Policy lookups
CREATE INDEX idx_policies_org_enabled
    ON policies(org_id) WHERE enabled = true;

-- Session validation
CREATE INDEX idx_sessions_token_hash
    ON sessions(token_hash);
```

## Schema Management

Schema is defined in `platform/database/schema.sql`:

```bash
make db-reset  # Drop and recreate with schema
```
