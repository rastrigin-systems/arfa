---
name: prod-database
description: Connect to and query Ubik Enterprise production PostgreSQL database on Google Cloud SQL. Use when you need to query production data for debugging, update production records, check database state after deployments, or verify data migrations. Triggers on requests involving production database, Cloud SQL, checking production data, or running SQL against prod.
---

# Production Database

Connect to the Ubik Enterprise production PostgreSQL database via Cloud SQL Proxy.

## Connection Details

- **Project:** `ubik-enterprise-prod`
- **Instance:** `ubik-postgres`
- **Database:** `ubik`
- **User:** `ubik`
- **Region:** `us-central1`

## Quick Query

Run the query script with your SQL:

```bash
.claude/skills/prod-database/scripts/query_prod.sh "SELECT COUNT(*) FROM employees;"
```

## Multi-Statement Queries

For complex queries, use heredoc:

```bash
.claude/skills/prod-database/scripts/query_prod.sh << 'EOF'
SELECT e.email, r.name as role
FROM employees e
JOIN roles r ON e.role_id = r.id
ORDER BY e.email;
EOF
```

## Common Queries

See `references/common_queries.md` for frequently used queries including:
- Employee and role queries
- Organization data
- Activity logs
- Agent configurations

## Safety Guidelines

1. **Use transactions** for UPDATE/DELETE:
   ```sql
   BEGIN;
   UPDATE employees SET status = 'inactive' WHERE id = '...';
   SELECT * FROM employees WHERE id = '...';  -- verify
   COMMIT;  -- or ROLLBACK if wrong
   ```

2. **Never DROP** tables without explicit user approval

3. **Use LIMIT** when exploring data to avoid huge result sets

4. **Backup first** for bulk updates:
   ```sql
   CREATE TABLE employees_backup AS SELECT * FROM employees;
   ```

## Troubleshooting

- **Port in use:** Script auto-kills existing proxies on port 9473
- **Auth failed:** Password fetched fresh from Secret Manager each time
- **IPv6 error:** Script uses Cloud SQL Proxy to avoid IPv6 issues
