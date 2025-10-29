# Enterprise AI Agent Management Platform - Entity Relationship Diagram

## Database Schema Overview

This Mermaid ERD diagram visualizes all tables and their relationships in the Ubik Enterprise platform.

**⚠️ AUTO-GENERATED FILE - DO NOT EDIT MANUALLY**
- Generated from: `docs/schema.json`
- Script: `scripts/generate-erd-overview.py`
- To update: `make generate-erd`

```mermaid
erDiagram
    %% Core Organization Structure
    organizations ||--o{ subscriptions : "has"
    organizations ||--o{ teams : "has"
    organizations ||--o{ employees : "has"
    teams ||--o{ employees : "has"
    roles ||--o{ employees : "has"
    employees ||--o{ sessions : "has"
    agents ||--o{ agent_tools : "has"
    tools ||--o{ agent_tools : "has"
    agents ||--o{ agent_policies : "has"
    policies ||--o{ agent_policies : "has"
    teams ||--o{ team_policies : "has"
    policies ||--o{ team_policies : "has"
    organizations ||--o{ org_agent_configs : "has"
    agents ||--o{ org_agent_configs : "has"
    teams ||--o{ team_agent_configs : "has"
    agents ||--o{ team_agent_configs : "has"
    employees ||--o{ employee_agent_configs : "has"
    agents ||--o{ employee_agent_configs : "has"
    agents ||--o{ system_prompts : "has"
    employees ||--o{ employee_policies : "has"
    policies ||--o{ employee_policies : "has"
    mcp_categories ||--o{ mcp_catalog : "has"
    employees ||--o{ employee_mcp_configs : "has"
    mcp_catalog ||--o{ employee_mcp_configs : "has"
    employees ||--o{ agent_requests : "has"
    employees ||--o{ approvals : "has"
    agent_requests ||--o{ approvals : "has"
    organizations ||--o{ activity_logs : "has"
    employees ||--o{ activity_logs : "has"
    organizations ||--o{ usage_records : "has"
    employees ||--o{ usage_records : "has"
    employee_agent_configs ||--o{ usage_records : "has"

    %% Table Definitions
    organizations {
        uuid id PK
        varchar255 name
        varchar100 slug UK
        varchar50 plan
        jsonb settings
        int max_employees
        int max_agents_per_employee
        text claude_api_token
        timestamp created_at
        timestamp updated_at
    }
    
    subscriptions {
        uuid id PK
        uuid org_id FK
        varchar50 plan_type
        decimal monthly_budget_usd
        decimal current_spending_usd
        timestamp billing_period_start
        timestamp billing_period_end
        varchar50 status
        timestamp created_at
        timestamp updated_at
    }
    
    teams {
        uuid id PK
        uuid org_id FK
        varchar255 name UK
        text description
        timestamp created_at
        timestamp updated_at
    }
    
    roles {
        uuid id PK
        varchar100 name UK
        text description
        jsonb permissions
        timestamp created_at
        timestamp updated_at
    }
    
    employees {
        uuid id PK
        uuid org_id FK
        uuid team_id FK
        uuid role_id FK
        varchar255 email UK
        varchar255 full_name
        varchar255 password_hash
        varchar50 status
        jsonb preferences
        text personal_claude_token
        timestamp last_login_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }
    
    sessions {
        uuid id PK
        uuid employee_id FK
        varchar255 token_hash UK
        varchar45 ip_address
        text user_agent
        timestamp expires_at
        timestamp created_at
    }
    
    tools {
        uuid id PK
        varchar100 name UK
        varchar50 type
        text description
        jsonb schema
        boolean requires_approval
        timestamp created_at
        timestamp updated_at
    }
    
    policies {
        uuid id PK
        varchar100 name UK
        varchar50 type
        jsonb rules
        varchar20 severity
        timestamp created_at
        timestamp updated_at
    }
    
    agent_tools {
        uuid agent_id PK
        uuid tool_id PK
        jsonb config
        timestamp created_at
    }
    
    agent_policies {
        uuid agent_id PK
        uuid policy_id PK
        timestamp created_at
    }
    
    team_policies {
        uuid id PK
        uuid team_id FK
        uuid policy_id FK
        jsonb overrides
        timestamp created_at
    }
    
    employee_agent_configs {
        uuid id PK
        uuid employee_id FK
        uuid agent_id FK
        jsonb config_override
        boolean is_enabled
        varchar255 sync_token UK
        timestamp last_synced_at
        timestamp created_at
        timestamp updated_at
    }
    
    mcp_categories {
        uuid id PK
        varchar100 name UK
        text description
        timestamp created_at
    }
    
    mcp_catalog {
        uuid id PK
        varchar255 name UK
        varchar255 provider
        varchar50 version
        text description
        jsonb connection_schema
        jsonb capabilities
        boolean requires_credentials
        boolean is_approved
        uuid category_id FK
        timestamp created_at
        timestamp updated_at
    }
    
    employee_mcp_configs {
        uuid id PK
        uuid employee_id FK
        uuid mcp_catalog_id FK
        varchar50 status
        jsonb connection_config
        text credentials_encrypted
        varchar255 sync_token UK
        timestamp last_sync_at
        timestamp created_at
        timestamp updated_at
    }
    
    agent_requests {
        uuid id PK
        uuid employee_id FK
        varchar50 request_type
        jsonb request_data
        varchar50 status
        text reason
        timestamp created_at
        timestamp resolved_at
    }
    
    approvals {
        uuid id PK
        uuid request_id FK
        uuid approver_id FK
        varchar50 status
        text comment
        timestamp created_at
        timestamp resolved_at
    }
    
    activity_logs {
        uuid id PK
        uuid org_id FK
        uuid employee_id FK
        varchar100 event_type
        varchar50 event_category
        jsonb payload
        timestamp created_at
    }
    
    usage_records {
        uuid id PK
        uuid org_id FK
        uuid employee_id FK
        uuid agent_config_id FK
        varchar50 resource_type
        bigint quantity
        decimal cost_usd
        timestamp period_start
        timestamp period_end
        jsonb metadata
        varchar20 token_source
        timestamp created_at
    }
    
```

## Table Groups

### 🏢 Core Organization (5 tables)
- **organizations** - Top-level tenant
- **subscriptions** - Billing and budget tracking
- **teams** - Group employees
- **roles** - Define permissions
- **employees** - User accounts

### 🤖 Agent Management (7 tables)
- **agent_catalog** - Available AI agents (Claude Code, Cursor, etc.)
- **tools** - Available tools (fs, git, http, etc.)
- **policies** - Usage policies and restrictions
- **agent_tools** - Many-to-many: agents ↔ tools
- **agent_policies** - Many-to-many: agents ↔ policies
- **team_policies** - Team-specific policy overrides
- **employee_agent_configs** - Per-employee agent instances

### 🔌 MCP Configuration (3 tables)
- **mcp_categories** - Organize MCP servers
- **mcp_catalog** - Available MCP servers
- **employee_mcp_configs** - Per-employee MCP instances

### 🔐 Authentication (1 table)
- **sessions** - JWT session tracking

### ✅ Approvals (2 tables)
- **agent_requests** - Employee requests for agents/MCPs
- **approvals** - Manager approval workflow

### 📊 Analytics (2 tables)
- **activity_logs** - Audit trail
- **usage_records** - Cost and resource tracking

## Key Relationships

### Multi-Tenancy
```
organizations (1) ──→ (N) employees
organizations (1) ──→ (N) teams
organizations (1) ──→ (N) activity_logs
```

### Employee Configuration
```
employees (1) ──→ (N) employee_agent_configs
employees (1) ──→ (N) employee_mcp_configs
employees (1) ──→ (N) sessions
```

### Agent System
```
agent_catalog (1) ──→ (N) employee_agent_configs
agent_catalog (M) ←→ (N) tools (via agent_tools)
agent_catalog (M) ←→ (N) policies (via agent_policies)
```

### MCP System
```
mcp_catalog (1) ──→ (N) employee_mcp_configs
mcp_categories (1) ──→ (N) mcp_catalog
```

### Approval Workflow
```
agent_requests (1) ──→ (N) approvals
employees (1) ──→ (N) agent_requests (requester)
employees (1) ──→ (N) approvals (approver)
```

## Views

The schema also includes {len(views)} materialized views for common queries:

1. **v_employee_agents** - Employee agents with catalog details
2. **v_employee_mcps** - Employee MCPs with catalog details
3. **v_pending_approvals** - Pending approval requests with context

## Indexes

All tables have appropriate indexes on:
- Primary keys (id)
- Foreign keys
- Unique constraints (email, slug, sync_token)
- Frequently queried columns (status, org_id, created_at)

## Database Statistics

- **Total Tables**: 24
- **Junction Tables**: 3 (agent_tools, agent_policies, team_policies)
- **Views**: 3
- **Total Columns**: ~188
- **Foreign Keys**: 32+
- **Indexes**: 75+

## Legend

- **PK**: Primary Key
- **FK**: Foreign Key
- **UK**: Unique Key
- **(1) ──→ (N)**: One-to-Many relationship
- **(M) ←→ (N)**: Many-to-Many relationship

---

**Generated**: 2025-10-29 22:41:22
**Schema Version**: 1.0.0
**Database**: PostgreSQL 15+
