# Common Production Queries

## Table of Contents
- [Employees](#employees)
- [Roles](#roles)
- [Organizations](#organizations)
- [Teams](#teams)
- [Sessions](#sessions)
- [Activity Logs](#activity-logs)
- [Agent Configurations](#agent-configurations)

## Employees

### List all employees with roles
```sql
SELECT e.email, e.full_name, r.name as role_name, e.status, e.created_at
FROM employees e
JOIN roles r ON e.role_id = r.id
ORDER BY e.full_name;
```

### Find employee by email
```sql
SELECT * FROM employees WHERE email = 'user@example.com';
```

### Count employees by role
```sql
SELECT r.name, COUNT(e.id) as count
FROM roles r
LEFT JOIN employees e ON r.id = e.role_id
GROUP BY r.name
ORDER BY count DESC;
```

### List employees by organization
```sql
SELECT e.email, e.full_name, o.name as org_name
FROM employees e
JOIN organizations o ON e.org_id = o.id
ORDER BY o.name, e.full_name;
```

## Roles

### List all roles with permissions
```sql
SELECT id, name, description, permissions FROM roles ORDER BY name;
```

### Find employees with specific role
```sql
SELECT e.email, e.full_name
FROM employees e
JOIN roles r ON e.role_id = r.id
WHERE r.name = 'admin';
```

## Organizations

### List all organizations
```sql
SELECT id, name, slug, subscription_tier, created_at
FROM organizations
ORDER BY name;
```

### Organization with employee count
```sql
SELECT o.name, o.slug, COUNT(e.id) as employee_count
FROM organizations o
LEFT JOIN employees e ON o.id = e.org_id
GROUP BY o.id, o.name, o.slug
ORDER BY employee_count DESC;
```

## Teams

### List teams with member count
```sql
SELECT t.name, o.name as org_name, COUNT(e.id) as member_count
FROM teams t
JOIN organizations o ON t.org_id = o.id
LEFT JOIN employees e ON t.id = e.team_id
GROUP BY t.id, t.name, o.name
ORDER BY o.name, t.name;
```

## Sessions

### Active sessions
```sql
SELECT s.id, e.email, s.created_at, s.expires_at
FROM sessions s
JOIN employees e ON s.employee_id = e.id
WHERE s.expires_at > NOW()
ORDER BY s.created_at DESC
LIMIT 20;
```

### Expired sessions cleanup check
```sql
SELECT COUNT(*) as expired_sessions
FROM sessions
WHERE expires_at < NOW();
```

## Activity Logs

### Recent activity
```sql
SELECT event_type, event_category, created_at
FROM activity_logs
ORDER BY created_at DESC
LIMIT 20;
```

### Activity by type
```sql
SELECT event_type, COUNT(*) as count
FROM activity_logs
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY event_type
ORDER BY count DESC;
```

## Agent Configurations

### List agent catalog
```sql
SELECT name, type, provider, description
FROM agents
ORDER BY name;
```

### Employee agent configs
```sql
SELECT e.email, a.name as agent_name, eac.enabled, eac.config
FROM employee_agent_configs eac
JOIN employees e ON eac.employee_id = e.id
JOIN agents a ON eac.agent_id = a.id
ORDER BY e.email, a.name;
```

## Dangerous Operations (Use With Caution)

### Update employee role
```sql
BEGIN;
UPDATE employees
SET role_id = (SELECT id FROM roles WHERE name = 'admin')
WHERE email = 'user@example.com';
-- Verify before commit
SELECT email, (SELECT name FROM roles WHERE id = role_id) as role FROM employees WHERE email = 'user@example.com';
COMMIT;
```

### Deactivate employee
```sql
BEGIN;
UPDATE employees SET status = 'inactive' WHERE email = 'user@example.com';
-- Verify
SELECT email, status FROM employees WHERE email = 'user@example.com';
COMMIT;
```
