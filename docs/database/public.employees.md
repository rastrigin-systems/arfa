# public.employees

## Description

## Columns

| Name | Type | Default | Nullable | Children | Parents | Comment |
| ---- | ---- | ------- | -------- | -------- | ------- | ------- |
| id | uuid | uuid_generate_v4() | false | [public.sessions](public.sessions.md) [public.password_reset_tokens](public.password_reset_tokens.md) [public.tool_policies](public.tool_policies.md) [public.employee_policies](public.employee_policies.md) [public.activity_logs](public.activity_logs.md) [public.webhook_destinations](public.webhook_destinations.md) [public.invitations](public.invitations.md) |  |  |
| org_id | uuid |  | false |  | [public.organizations](public.organizations.md) |  |
| team_id | uuid |  | true |  | [public.teams](public.teams.md) |  |
| role_id | uuid |  | false |  | [public.roles](public.roles.md) |  |
| email | varchar(255) |  | false |  |  |  |
| full_name | varchar(255) |  | false |  |  |  |
| password_hash | varchar(255) |  | false |  |  |  |
| status | varchar(50) | 'active'::character varying | false |  |  |  |
| preferences | jsonb | '{}'::jsonb | false |  |  |  |
| last_login_at | timestamp without time zone |  | true |  |  |  |
| created_at | timestamp without time zone | now() | false |  |  |  |
| updated_at | timestamp without time zone | now() | false |  |  |  |
| deleted_at | timestamp without time zone |  | true |  |  |  |

## Constraints

| Name | Type | Definition |
| ---- | ---- | ---------- |
| employees_org_id_fkey | FOREIGN KEY | FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE |
| employees_team_id_fkey | FOREIGN KEY | FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL |
| employees_role_id_fkey | FOREIGN KEY | FOREIGN KEY (role_id) REFERENCES roles(id) |
| employees_pkey | PRIMARY KEY | PRIMARY KEY (id) |
| employees_email_key | UNIQUE | UNIQUE (email) |

## Indexes

| Name | Definition |
| ---- | ---------- |
| employees_pkey | CREATE UNIQUE INDEX employees_pkey ON public.employees USING btree (id) |
| employees_email_key | CREATE UNIQUE INDEX employees_email_key ON public.employees USING btree (email) |
| idx_employees_org_id | CREATE INDEX idx_employees_org_id ON public.employees USING btree (org_id) |
| idx_employees_team_id | CREATE INDEX idx_employees_team_id ON public.employees USING btree (team_id) |
| idx_employees_email | CREATE INDEX idx_employees_email ON public.employees USING btree (email) |
| idx_employees_status | CREATE INDEX idx_employees_status ON public.employees USING btree (status) |

## Triggers

| Name | Definition |
| ---- | ---------- |
| update_employees_updated_at | CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON public.employees FOR EACH ROW EXECUTE FUNCTION update_updated_at_column() |
| employee_revoke_trigger | CREATE TRIGGER employee_revoke_trigger AFTER UPDATE ON public.employees FOR EACH ROW EXECUTE FUNCTION notify_employee_revoke() |

## Relations

```mermaid
erDiagram

"public.sessions" }o--|| "public.employees" : "FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE"
"public.password_reset_tokens" }o--|| "public.employees" : "FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE"
"public.tool_policies" }o--o| "public.employees" : "FOREIGN KEY (created_by) REFERENCES employees(id) ON DELETE SET NULL"
"public.tool_policies" }o--o| "public.employees" : "FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE"
"public.employee_policies" }o--|| "public.employees" : "FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE"
"public.activity_logs" }o--o| "public.employees" : "FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE SET NULL"
"public.webhook_destinations" }o--o| "public.employees" : "FOREIGN KEY (created_by) REFERENCES employees(id) ON DELETE SET NULL"
"public.invitations" }o--o| "public.employees" : "FOREIGN KEY (accepted_by) REFERENCES employees(id) ON DELETE SET NULL"
"public.invitations" }o--|| "public.employees" : "FOREIGN KEY (inviter_id) REFERENCES employees(id) ON DELETE CASCADE"
"public.employees" }o--|| "public.organizations" : "FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE"
"public.employees" }o--o| "public.teams" : "FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL"
"public.employees" }o--|| "public.roles" : "FOREIGN KEY (role_id) REFERENCES roles(id)"

"public.employees" {
  uuid id
  uuid org_id FK
  uuid team_id FK
  uuid role_id FK
  varchar_255_ email
  varchar_255_ full_name
  varchar_255_ password_hash
  varchar_50_ status
  jsonb preferences
  timestamp_without_time_zone last_login_at
  timestamp_without_time_zone created_at
  timestamp_without_time_zone updated_at
  timestamp_without_time_zone deleted_at
}
"public.sessions" {
  uuid id
  uuid employee_id FK
  varchar_255_ token_hash
  varchar_45_ ip_address
  text user_agent
  timestamp_without_time_zone expires_at
  timestamp_without_time_zone created_at
}
"public.password_reset_tokens" {
  uuid id
  uuid employee_id FK
  varchar_64_ token
  timestamp_without_time_zone expires_at
  timestamp_without_time_zone used_at
  timestamp_without_time_zone created_at
}
"public.tool_policies" {
  uuid id
  uuid org_id FK
  uuid team_id FK
  uuid employee_id FK
  varchar_255_ tool_name
  jsonb conditions
  varchar_20_ action
  text reason
  uuid created_by FK
  timestamp_without_time_zone created_at
  timestamp_without_time_zone updated_at
}
"public.employee_policies" {
  uuid id
  uuid employee_id FK
  uuid policy_id FK
  jsonb overrides
  timestamp_without_time_zone created_at
}
"public.activity_logs" {
  uuid id
  uuid org_id FK
  uuid employee_id FK
  uuid proxy_session_id
  varchar_100_ client_name
  varchar_50_ client_version
  varchar_100_ event_type
  varchar_50_ event_category
  text content
  jsonb payload
  timestamp_without_time_zone created_at
}
"public.webhook_destinations" {
  uuid id
  uuid org_id FK
  varchar_100_ name
  text url
  varchar_50_ auth_type
  jsonb auth_config
  text__ event_types
  jsonb event_filter
  boolean enabled
  integer batch_size
  integer timeout_ms
  integer retry_max
  integer retry_backoff_ms
  varchar_255_ signing_secret
  uuid created_by FK
  timestamp_without_time_zone created_at
  timestamp_without_time_zone updated_at
}
"public.invitations" {
  uuid id
  uuid org_id FK
  uuid inviter_id FK
  varchar_255_ email
  uuid role_id FK
  uuid team_id FK
  varchar_64_ token
  varchar_50_ status
  timestamp_without_time_zone expires_at
  uuid accepted_by FK
  timestamp_without_time_zone accepted_at
  timestamp_without_time_zone created_at
  timestamp_without_time_zone updated_at
}
"public.organizations" {
  uuid id
  varchar_255_ name
  varchar_100_ slug
  jsonb settings
  integer max_employees
  timestamp_without_time_zone created_at
  timestamp_without_time_zone updated_at
}
"public.teams" {
  uuid id
  uuid org_id FK
  varchar_255_ name
  text description
  timestamp_without_time_zone created_at
  timestamp_without_time_zone updated_at
}
"public.roles" {
  uuid id
  varchar_100_ name
  text description
  jsonb permissions
  timestamp_without_time_zone created_at
  timestamp_without_time_zone updated_at
}
```

---

> Generated by [tbls](https://github.com/k1LoW/tbls)
