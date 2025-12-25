# arfa

## Tables

| Name | Columns | Comment | Type |
| ---- | ------- | ------- | ---- |
| [public.organizations](public.organizations.md) | 7 |  | BASE TABLE |
| [public.teams](public.teams.md) | 6 |  | BASE TABLE |
| [public.roles](public.roles.md) | 6 |  | BASE TABLE |
| [public.employees](public.employees.md) | 13 |  | BASE TABLE |
| [public.sessions](public.sessions.md) | 7 |  | BASE TABLE |
| [public.password_reset_tokens](public.password_reset_tokens.md) | 6 |  | BASE TABLE |
| [public.policies](public.policies.md) | 7 |  | BASE TABLE |
| [public.tool_policies](public.tool_policies.md) | 11 |  | BASE TABLE |
| [public.team_policies](public.team_policies.md) | 5 |  | BASE TABLE |
| [public.employee_policies](public.employee_policies.md) | 5 |  | BASE TABLE |
| [public.activity_logs](public.activity_logs.md) | 11 |  | BASE TABLE |
| [public.webhook_destinations](public.webhook_destinations.md) | 17 |  | BASE TABLE |
| [public.webhook_deliveries](public.webhook_deliveries.md) | 12 |  | BASE TABLE |
| [public.invitations](public.invitations.md) | 13 |  | BASE TABLE |

## Stored procedures and functions

| Name | ReturnType | Arguments | Type |
| ---- | ------- | ------- | ---- |
| public.uuid_nil | uuid |  | FUNCTION |
| public.uuid_ns_dns | uuid |  | FUNCTION |
| public.uuid_ns_url | uuid |  | FUNCTION |
| public.uuid_ns_oid | uuid |  | FUNCTION |
| public.uuid_ns_x500 | uuid |  | FUNCTION |
| public.uuid_generate_v1 | uuid |  | FUNCTION |
| public.uuid_generate_v1mc | uuid |  | FUNCTION |
| public.uuid_generate_v3 | uuid | namespace uuid, name text | FUNCTION |
| public.uuid_generate_v4 | uuid |  | FUNCTION |
| public.uuid_generate_v5 | uuid | namespace uuid, name text | FUNCTION |
| public.digest | bytea | text, text | FUNCTION |
| public.digest | bytea | bytea, text | FUNCTION |
| public.hmac | bytea | text, text, text | FUNCTION |
| public.hmac | bytea | bytea, bytea, text | FUNCTION |
| public.crypt | text | text, text | FUNCTION |
| public.gen_salt | text | text | FUNCTION |
| public.gen_salt | text | text, integer | FUNCTION |
| public.encrypt | bytea | bytea, bytea, text | FUNCTION |
| public.decrypt | bytea | bytea, bytea, text | FUNCTION |
| public.encrypt_iv | bytea | bytea, bytea, bytea, text | FUNCTION |
| public.decrypt_iv | bytea | bytea, bytea, bytea, text | FUNCTION |
| public.gen_random_bytes | bytea | integer | FUNCTION |
| public.gen_random_uuid | uuid |  | FUNCTION |
| public.pgp_sym_encrypt | bytea | text, text | FUNCTION |
| public.pgp_sym_encrypt_bytea | bytea | bytea, text | FUNCTION |
| public.pgp_sym_encrypt | bytea | text, text, text | FUNCTION |
| public.pgp_sym_encrypt_bytea | bytea | bytea, text, text | FUNCTION |
| public.pgp_sym_decrypt | text | bytea, text | FUNCTION |
| public.pgp_sym_decrypt_bytea | bytea | bytea, text | FUNCTION |
| public.pgp_sym_decrypt | text | bytea, text, text | FUNCTION |
| public.pgp_sym_decrypt_bytea | bytea | bytea, text, text | FUNCTION |
| public.pgp_pub_encrypt | bytea | text, bytea | FUNCTION |
| public.pgp_pub_encrypt_bytea | bytea | bytea, bytea | FUNCTION |
| public.pgp_pub_encrypt | bytea | text, bytea, text | FUNCTION |
| public.pgp_pub_encrypt_bytea | bytea | bytea, bytea, text | FUNCTION |
| public.pgp_pub_decrypt | text | bytea, bytea | FUNCTION |
| public.pgp_pub_decrypt_bytea | bytea | bytea, bytea | FUNCTION |
| public.pgp_pub_decrypt | text | bytea, bytea, text | FUNCTION |
| public.pgp_pub_decrypt_bytea | bytea | bytea, bytea, text | FUNCTION |
| public.pgp_pub_decrypt | text | bytea, bytea, text, text | FUNCTION |
| public.pgp_pub_decrypt_bytea | bytea | bytea, bytea, text, text | FUNCTION |
| public.pgp_key_id | text | bytea | FUNCTION |
| public.armor | text | bytea | FUNCTION |
| public.armor | text | bytea, text[], text[] | FUNCTION |
| public.dearmor | bytea | text | FUNCTION |
| public.pgp_armor_headers | record | text, OUT key text, OUT value text | FUNCTION |
| public.update_updated_at_column | trigger |  | FUNCTION |
| public.generate_invitation_token | trigger |  | FUNCTION |
| public.expire_old_invitations | void |  | FUNCTION |

## Relations

```mermaid
erDiagram

"public.teams" }o--|| "public.organizations" : "FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE"
"public.employees" }o--|| "public.organizations" : "FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE"
"public.employees" }o--o| "public.teams" : "FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL"
"public.employees" }o--|| "public.roles" : "FOREIGN KEY (role_id) REFERENCES roles(id)"
"public.sessions" }o--|| "public.employees" : "FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE"
"public.password_reset_tokens" }o--|| "public.employees" : "FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE"
"public.tool_policies" }o--|| "public.organizations" : "FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE"
"public.tool_policies" }o--o| "public.teams" : "FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE"
"public.tool_policies" }o--o| "public.employees" : "FOREIGN KEY (created_by) REFERENCES employees(id) ON DELETE SET NULL"
"public.tool_policies" }o--o| "public.employees" : "FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE"
"public.team_policies" }o--|| "public.teams" : "FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE"
"public.team_policies" }o--|| "public.policies" : "FOREIGN KEY (policy_id) REFERENCES policies(id) ON DELETE CASCADE"
"public.employee_policies" }o--|| "public.employees" : "FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE"
"public.employee_policies" }o--|| "public.policies" : "FOREIGN KEY (policy_id) REFERENCES policies(id) ON DELETE CASCADE"
"public.activity_logs" }o--|| "public.organizations" : "FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE"
"public.activity_logs" }o--o| "public.employees" : "FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE SET NULL"
"public.webhook_destinations" }o--|| "public.organizations" : "FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE"
"public.webhook_destinations" }o--o| "public.employees" : "FOREIGN KEY (created_by) REFERENCES employees(id) ON DELETE SET NULL"
"public.webhook_deliveries" }o--|| "public.activity_logs" : "FOREIGN KEY (log_id) REFERENCES activity_logs(id) ON DELETE CASCADE"
"public.webhook_deliveries" }o--|| "public.webhook_destinations" : "FOREIGN KEY (destination_id) REFERENCES webhook_destinations(id) ON DELETE CASCADE"
"public.invitations" }o--|| "public.organizations" : "FOREIGN KEY (org_id) REFERENCES organizations(id) ON DELETE CASCADE"
"public.invitations" }o--o| "public.teams" : "FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE SET NULL"
"public.invitations" }o--|| "public.roles" : "FOREIGN KEY (role_id) REFERENCES roles(id)"
"public.invitations" }o--o| "public.employees" : "FOREIGN KEY (accepted_by) REFERENCES employees(id) ON DELETE SET NULL"
"public.invitations" }o--|| "public.employees" : "FOREIGN KEY (inviter_id) REFERENCES employees(id) ON DELETE CASCADE"

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
"public.policies" {
  uuid id
  varchar_100_ name
  varchar_50_ type
  jsonb rules
  varchar_20_ severity
  timestamp_without_time_zone created_at
  timestamp_without_time_zone updated_at
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
"public.team_policies" {
  uuid id
  uuid team_id FK
  uuid policy_id FK
  jsonb overrides
  timestamp_without_time_zone created_at
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
  uuid session_id
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
"public.webhook_deliveries" {
  uuid id
  uuid destination_id FK
  uuid log_id FK
  varchar_50_ status
  integer attempts
  timestamp_without_time_zone last_attempt_at
  timestamp_without_time_zone next_retry_at
  integer response_status
  text response_body
  text error_message
  timestamp_without_time_zone created_at
  timestamp_without_time_zone delivered_at
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
```

---

> Generated by [tbls](https://github.com/k1LoW/tbls)
