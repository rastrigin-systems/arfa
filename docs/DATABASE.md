# Database Guide

**Last Updated:** 2025-11-05

Complete guide to the Ubik Enterprise database schema, operations, and best practices.

---

## Table of Contents

- [Overview](#overview)
- [Schema Structure](#schema-structure)
- [Database Access](#database-access)
- [Common Operations](#common-operations)
- [Schema Management](#schema-management)
- [Multi-Tenancy](#multi-tenancy)
- [Best Practices](#best-practices)

---

## Overview

**Database:** PostgreSQL 15+
**Schema:** 20 tables + 3 views
**Multi-Tenancy:** Org-based with Row-Level Security (RLS)
**Documentation:** Auto-generated via tbls

---

## Schema Structure

### Overview

**20 Tables + 3 Views**

**Core Organization (5 tables)**
- `organizations` - Top-level tenant
- `subscriptions` - Billing and budget tracking
- `teams` - Group employees
- `roles` - Define permissions
- `employees` - User accounts

**Agent Management (7 tables)**
- `agent_catalog` - Available AI agents (Claude Code, Cursor, etc.)
- `tools` - Available tools (fs, git, http, etc.)
- `policies` - Usage policies and restrictions
- `agent_tools` - Many-to-many: agents ↔ tools
- `agent_policies` - Many-to-many: agents ↔ policies
- `team_policies` - Team-specific policy overrides
- `employee_agent_configs` - Per-employee agent instances

**MCP Configuration (3 tables)**
- `mcp_categories` - Organize MCP servers
- `mcp_catalog` - Available MCP servers
- `employee_mcp_configs` - Per-employee MCP instances

**Authentication (1 table)**
- `sessions` - JWT session tracking

**Approvals (2 tables)**
- `agent_requests` - Employee requests for agents/MCPs
- `approvals` - Manager approval workflow

**Analytics (2 tables)**
- `activity_logs` - Audit trail
- `usage_records` - Cost and resource tracking

**Views (3)**
- `v_employee_agents` - Employee agents with catalog details
- `v_employee_mcps` - Employee MCPs with catalog details
- `v_pending_approvals` - Pending approval requests with context

**See [ERD.md](./ERD.md) for complete visual schema and relationships.**

---

## Database Access

### Connection Methods

#### 1. Adminer Web UI (Recommended for browsing)

```bash
# Start database
make db-up

# Open Adminer
open http://localhost:8080

# Login credentials:
# System: PostgreSQL
# Server: ubik-postgres
# Username: ubik
# Password: ubik_dev_password
# Database: ubik
```

---

#### 2. psql CLI (Recommended for queries)

```bash
# Connect to database
docker exec -it ubik-postgres psql -U ubik -d ubik

# One-line query
docker exec ubik-postgres psql -U ubik -d ubik -c "SELECT COUNT(*) FROM employees"

# Execute SQL file
docker exec -i ubik-postgres psql -U ubik -d ubik < platform/database/schema.sql
```

---

#### 3. PostgreSQL MCP (Recommended for development)

```bash
# Add PostgreSQL MCP server
claude mcp add postgres \
  -- docker run -i --rm mcp/postgres \
  postgresql://ubik:ubik_dev_password@host.docker.internal:5432/ubik

# Use via Claude Code
# MCP provides tools for queries, schema inspection, etc.
```

**See:** [MCP_SERVERS.md](./MCP_SERVERS.md#postgresql-mcp) for complete MCP setup.

---

#### 4. Direct Connection String

```bash
# Connection string
postgres://ubik:ubik_dev_password@localhost:5432/ubik

# Use with any PostgreSQL client:
# - DataGrip
# - pgAdmin
# - DBeaver
# - TablePlus
# - Postico
```

---

## Common Operations

### Database Management

```bash
# Start PostgreSQL
make db-up

# Stop PostgreSQL
make db-down

# Reset database (⚠️ deletes all data)
make db-reset

# Run migrations
make db-migrate

# Rollback migrations
make db-rollback
```

---

### Schema Inspection

```bash
# List all tables
docker exec ubik-postgres psql -U ubik -d ubik -c "\dt"

# Describe table structure
docker exec ubik-postgres psql -U ubik -d ubik -c "\d employees"

# List all views
docker exec ubik-postgres psql -U ubik -d ubik -c "\dv"

# List all functions
docker exec ubik-postgres psql -U ubik -d ubik -c "\df"

# List all indexes
docker exec ubik-postgres psql -U ubik -d ubik -c "\di"

# List all foreign keys
docker exec ubik-postgres psql -U ubik -d ubik -c "
  SELECT tc.table_name, kcu.column_name, ccu.table_name AS foreign_table_name
  FROM information_schema.table_constraints AS tc
  JOIN information_schema.key_column_usage AS kcu ON tc.constraint_name = kcu.constraint_name
  JOIN information_schema.constraint_column_usage AS ccu ON ccu.constraint_name = tc.constraint_name
  WHERE tc.constraint_type = 'FOREIGN KEY'
"
```

---

### Data Queries

```bash
# Count records in table
docker exec ubik-postgres psql -U ubik -d ubik -c "SELECT COUNT(*) FROM organizations"

# View recent records
docker exec ubik-postgres psql -U ubik -d ubik -c "
  SELECT id, name, created_at FROM organizations
  ORDER BY created_at DESC
  LIMIT 10
"

# Search records
docker exec ubik-postgres psql -U ubik -d ubik -c "
  SELECT * FROM employees
  WHERE email LIKE '%@example.com'
"

# Join tables
docker exec ubik-postgres psql -U ubik -d ubik -c "
  SELECT e.email, o.name as org_name
  FROM employees e
  JOIN organizations o ON e.org_id = o.id
"
```

---

### Data Manipulation

```bash
# Insert record
docker exec ubik-postgres psql -U ubik -d ubik -c "
  INSERT INTO organizations (id, name, email)
  VALUES (gen_random_uuid(), 'Test Org', 'test@example.com')
"

# Update record
docker exec ubik-postgres psql -U ubik -d ubik -c "
  UPDATE organizations
  SET name = 'Updated Org'
  WHERE email = 'test@example.com'
"

# Delete record
docker exec ubik-postgres psql -U ubik -d ubik -c "
  DELETE FROM organizations
  WHERE email = 'test@example.com'
"
```

---

## Schema Management

### Migrations

**Location:** `platform/database/migrations/`

**File naming:** `YYYYMMDDHHMMSS_description.up.sql` and `YYYYMMDDHHMMSS_description.down.sql`

**Create migration:**

```bash
# Create new migration files
migrate create -ext sql -dir platform/database/migrations -seq add_user_table
```

**Apply migrations:**

```bash
# Run all pending migrations
make db-migrate

# Or manually
migrate -path platform/database/migrations -database "postgres://ubik:ubik_dev_password@localhost:5432/ubik?sslmode=disable" up
```

**Rollback migrations:**

```bash
# Rollback last migration
make db-rollback

# Or manually
migrate -path platform/database/migrations -database "postgres://ubik:ubik_dev_password@localhost:5432/ubik?sslmode=disable" down 1
```

---

### Schema Documentation

**Auto-generated documentation:**

```bash
# Generate ERD and table docs
make generate-erd

# This creates:
# - docs/ERD.md (user-friendly overview)
# - docs/README.md (technical reference)
# - docs/schema.json (machine-readable)
# - docs/schema.svg (visual diagram)
# - docs/public.*.md (per-table docs)
```

**Update after schema changes:**
1. Modify `platform/database/schema.sql`
2. Run `make db-reset` to apply changes
3. Run `make generate-erd` to update docs
4. Commit schema + docs together

---

## Multi-Tenancy

### Organization Scoping

**All queries must be org-scoped to prevent data leaks.**

**✅ GOOD - Scoped to organization:**

```go
employees, err := db.ListEmployees(ctx, orgID, status)
```

**❌ BAD - Exposes all orgs:**

```go
employees, err := db.ListAllEmployees(ctx)
```

---

### Row-Level Security (RLS)

**RLS policies enforce org-scoping at database level.**

**Check RLS status:**

```bash
# List RLS policies
docker exec ubik-postgres psql -U ubik -d ubik -c "
  SELECT schemaname, tablename, policyname
  FROM pg_policies
  WHERE tablename = 'employees'
"

# Check if RLS is enabled
docker exec ubik-postgres psql -U ubik -d ubik -c "
  SELECT tablename, rowsecurity
  FROM pg_tables
  WHERE schemaname = 'public'
"
```

**Test RLS:**

```bash
# Set current org context
docker exec ubik-postgres psql -U ubik -d ubik -c "
  SET app.current_org_id = '<org-uuid>';
  SELECT * FROM employees;
"

# Should only return employees from that org
```

**Disable RLS (testing only):**

```bash
# ⚠️ Only for debugging!
docker exec ubik-postgres psql -U ubik -d ubik -c "
  ALTER TABLE employees DISABLE ROW LEVEL SECURITY
"

# Re-enable after testing
docker exec ubik-postgres psql -U ubik -d ubik -c "
  ALTER TABLE employees ENABLE ROW LEVEL SECURITY
"
```

---

## Best Practices

### Query Performance

**Use indexes:**

```sql
-- Check if index exists
SELECT indexname FROM pg_indexes WHERE tablename = 'employees';

-- Create index if needed
CREATE INDEX idx_employees_org_id ON employees(org_id);
CREATE INDEX idx_employees_email ON employees(email);
```

**Analyze query performance:**

```bash
# Explain query plan
docker exec ubik-postgres psql -U ubik -d ubik -c "
  EXPLAIN ANALYZE
  SELECT * FROM employees WHERE org_id = '<uuid>'
"
```

---

### Data Integrity

**Always use transactions:**

```sql
BEGIN;

-- Multiple operations
INSERT INTO organizations (...) VALUES (...);
INSERT INTO employees (...) VALUES (...);

-- Commit if all succeed
COMMIT;

-- Or rollback on error
-- ROLLBACK;
```

**Check foreign key constraints:**

```bash
# Find orphaned records
docker exec ubik-postgres psql -U ubik -d ubik -c "
  SELECT * FROM employee_agent_configs eac
  WHERE NOT EXISTS (
    SELECT 1 FROM org_agent_configs oac
    WHERE oac.id = eac.org_agent_config_id
  )
"
```

---

### Backup and Restore

**Backup database:**

```bash
# Backup to file
docker exec ubik-postgres pg_dump -U ubik -d ubik > backup.sql

# Backup with custom format (smaller, faster restore)
docker exec ubik-postgres pg_dump -U ubik -d ubik -Fc > backup.dump
```

**Restore database:**

```bash
# Restore from SQL file
docker exec -i ubik-postgres psql -U ubik -d ubik < backup.sql

# Restore from custom format
docker exec -i ubik-postgres pg_restore -U ubik -d ubik backup.dump
```

---

### Development vs Production

**Development:**
- Use `make db-reset` freely
- Test migrations both up and down
- Use seed data for testing
- Enable query logging
- Disable RLS for debugging (carefully!)

**Production:**
- Never `make db-reset`
- Test migrations in staging first
- Use real data (with privacy considerations)
- Minimize logging (performance)
- Always keep RLS enabled

---

## See Also

- [ERD.md](./ERD.md) - Visual schema and relationships
- [README.md](./README.md) - Technical reference (auto-generated)
- [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) - Command reference
- [MCP_SERVERS.md](./MCP_SERVERS.md) - PostgreSQL MCP setup
- [TESTING.md](./TESTING.md) - Database testing strategies
- [DEBUGGING.md](./DEBUGGING.md) - Database debugging
