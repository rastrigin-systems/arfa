---
sidebar_position: 4
---

# ADR-003: Multi-tenant Row-Level Security

**Status:** Accepted
**Date:** 2025-10

## Context

Arfa is a multi-tenant SaaS platform where multiple organizations share the same database. Requirements:

1. **Data isolation**: Organizations must never see each other's data
2. **Developer safety**: Hard to accidentally leak data
3. **Auditability**: Able to verify isolation works

## Decision

Implement **defense-in-depth multi-tenancy** with four layers:

1. **Application filtering**: Every query includes `org_id`
2. **sqlc enforcement**: Generated functions require `org_id`
3. **PostgreSQL RLS**: Database-level safety net
4. **Middleware context**: Set org context per request

```sql
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;

CREATE POLICY employees_org_isolation ON employees
    FOR ALL
    USING (org_id = current_setting('app.current_org_id')::uuid);
```

## Consequences

### Positive

- **Defense in depth**: Multiple protection layers
- **Compile-time safety**: sqlc prevents missing org_id
- **Database guarantee**: RLS enforces even if app has bugs

### Negative

- **Performance overhead**: RLS adds small query overhead
- **Complexity**: Four layers to maintain
- **Testing burden**: Need to test isolation per feature

## Alternatives Considered

1. **Separate databases** - Rejected: Operational complexity
2. **Schema-per-tenant** - Rejected: Migration complexity
3. **Application-only filtering** - Rejected: No defense in depth
