# Enterprise AI Agent Management Platform - Entity Relationship Diagram

## Database Schema Overview

This Mermaid ERD diagram visualizes all tables and their relationships in the Ubik Enterprise platform.

```mermaid
erDiagram
    %% Core Organization Structure
    organizations ||--o{ subscriptions : "has"
    organizations ||--o{ teams : "has"
    organizations ||--o{ employees : "has"
    organizations ||--o{ activity_logs : "tracks"
    organizations ||--o{ usage_records : "tracks"
    
    teams ||--o{ employees : "contains"
    teams ||--o{ team_policies : "enforces"
    
    roles ||--o{ employees : "assigned_to"
    
    employees ||--o{ sessions : "creates"
    employees ||--o{ employee_agent_configs : "has"
    employees ||--o{ employee_mcp_configs : "has"
    employees ||--o{ agent_requests : "submits"
    employees ||--o{ approvals : "approves"
    employees ||--o{ activity_logs : "performs"
    employees ||--o{ usage_records : "generates"
    
    %% Agent Configuration
    agent_catalog ||--o{ agent_tools : "includes"
    agent_catalog ||--o{ agent_policies : "enforces"
    agent_catalog ||--o{ employee_agent_configs : "instantiates"
    
    tools ||--o{ agent_tools : "used_in"
    
    policies ||--o{ agent_policies : "applied_to"
    policies ||--o{ team_policies : "overridden_by"
    
    employee_agent_configs ||--o{ usage_records : "generates"
    
    %% MCP Configuration
    mcp_categories ||--o{ mcp_catalog : "contains"
    mcp_catalog ||--o{ employee_mcp_configs : "provides"
    
    %% Approval Workflow
    agent_requests ||--o{ approvals : "requires"
    
    %% Table Definitions
    organizations {
        uuid id PK
        varchar name
        varchar slug UK
        varchar plan
        jsonb settings
        int max_employees
        int max_agents_per_employee
        timestamp created_at
        timestamp updated_at
    }
    
    subscriptions {
        uuid id PK
        uuid org_id FK
        varchar plan_type
        decimal monthly_budget_usd
        decimal current_spending_usd
        timestamp billing_period_start
        timestamp billing_period_end
        varchar status
        timestamp created_at
        timestamp updated_at
    }
    
    teams {
        uuid id PK
        uuid org_id FK
        varchar name
        text description
        timestamp created_at
        timestamp updated_at
    }
    
    roles {
        uuid id PK
        varchar name UK
        text description
        jsonb permissions
        timestamp created_at
    }
    
    employees {
        uuid id PK
        uuid org_id FK
        uuid team_id FK
        uuid role_id FK
        varchar email UK
        varchar full_name
        varchar status
        jsonb preferences
        timestamp last_login_at
        timestamp created_at
        timestamp updated_at
    }
    
    sessions {
        uuid id PK
        uuid employee_id FK
        varchar token_hash UK
        varchar ip_address
        text user_agent
        timestamp expires_at
        timestamp created_at
    }
    
    agent_catalog {
        uuid id PK
        varchar name UK
        varchar type
        text description
        varchar provider
        jsonb default_config
        jsonb capabilities
        varchar llm_provider
        varchar llm_model
        boolean is_public
        timestamp created_at
        timestamp updated_at
    }
    
    tools {
        uuid id PK
        varchar name UK
        varchar type
        text description
        jsonb schema
        boolean requires_approval
        timestamp created_at
        timestamp updated_at
    }
    
    policies {
        uuid id PK
        varchar name UK
        varchar type
        jsonb rules
        varchar severity
        timestamp created_at
        timestamp updated_at
    }
    
    agent_tools {
        uuid agent_id FK
        uuid tool_id FK
        jsonb config
        timestamp created_at
    }
    
    agent_policies {
        uuid agent_id FK
        uuid policy_id FK
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
        uuid agent_catalog_id FK
        varchar name
        varchar status
        jsonb config_override
        varchar sync_token UK
        timestamp last_sync_at
        timestamp created_at
        timestamp updated_at
    }
    
    mcp_categories {
        uuid id PK
        varchar name UK
        text description
        timestamp created_at
    }
    
    mcp_catalog {
        uuid id PK
        varchar name UK
        varchar provider
        varchar version
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
        varchar status
        jsonb connection_config
        text credentials_encrypted
        varchar sync_token UK
        timestamp last_sync_at
        timestamp created_at
        timestamp updated_at
    }
    
    agent_requests {
        uuid id PK
        uuid employee_id FK
        varchar request_type
        jsonb request_data
        varchar status
        text reason
        timestamp created_at
        timestamp resolved_at
    }
    
    approvals {
        uuid id PK
        uuid request_id FK
        uuid approver_id FK
        varchar status
        text comment
        timestamp created_at
        timestamp resolved_at
    }
    
    activity_logs {
        uuid id PK
        uuid org_id FK
        uuid employee_id FK
        varchar event_type
        varchar event_category
        jsonb payload
        timestamp created_at
    }
    
    usage_records {
        uuid id PK
        uuid org_id FK
        uuid employee_id FK
        uuid agent_config_id FK
        varchar resource_type
        bigint quantity
        decimal cost_usd
        timestamp period_start
        timestamp period_end
        jsonb metadata
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

The schema also includes 3 materialized views for common queries:

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

- **Total Tables**: 20
- **Junction Tables**: 3 (agent_tools, agent_policies, team_policies)
- **Views**: 3
- **Total Columns**: ~150
- **Foreign Keys**: 25+
- **Indexes**: 50+

## Legend

- **PK**: Primary Key
- **FK**: Foreign Key
- **UK**: Unique Key
- **(1) ──→ (N)**: One-to-Many relationship
- **(M) ←→ (N)**: Many-to-Many relationship

---

**Generated**: 2025-10-28  
**Schema Version**: 1.0.0  
**Database**: PostgreSQL 15+
