# Subtask Template

Use this template when creating child tasks linked to a parent issue.

## Usage

```bash
# Step 1: Get parent issue node ID
PARENT_NUM=123
PARENT_NODE_ID=$(gh api graphql -f query='
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    issue(number: $number) {
      id
    }
  }
}' -f owner='sergei-rastrigin' -f repo='ubik-enterprise' -F number=$PARENT_NUM -q .data.repository.issue.id)

# Step 2: Create subtask
SUB_NUM=$(gh issue create \
  --title "Subtask: Specific component or step" \
  --label "INHERITED_LABELS,subtask,size/SIZE" \
  --body "$(cat <<'EOF'
Part of #PARENT_NUM

## Description

[What this subtask accomplishes as part of the larger feature]

## Parent Task

This subtask is part of the larger feature tracked in #PARENT_NUM

## Scope

[Specific scope of this subtask - what's included and what's not]

## Implementation Steps

- [ ] Step 1
- [ ] Step 2
- [ ] Step 3

## Acceptance Criteria

- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Technical Notes

[Any specific technical considerations for this subtask]

## Dependencies

[Dependencies within the parent task context]
- Depends on: #OTHER_SUBTASK
- Blocks: #ANOTHER_SUBTASK
EOF
)" | grep -oE '#[0-9]+' | cut -c2-)

# Step 3: Get subtask node ID and link to parent
SUB_NODE_ID=$(gh api graphql -f query='
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    issue(number: $number) {
      id
    }
  }
}' -f owner='sergei-rastrigin' -f repo='ubik-enterprise' -F number=$SUB_NUM -q .data.repository.issue.id)

gh api graphql -f query='
mutation($parentId: ID!, $childId: ID!) {
  updateIssue(input: {
    id: $childId,
    trackedInIssues: [$parentId]
  }) {
    issue {
      id
    }
  }
}' -f parentId="$PARENT_NODE_ID" -f childId="$SUB_NODE_ID"

# Step 4: Add to project
gh project item-add 3 --owner sergei-rastrigin --url "https://github.com/rastrigin-org/ubik-enterprise/issues/$SUB_NUM"

# Step 5: Update parent issue
gh issue comment $PARENT_NUM --body "Created subtask #$SUB_NUM"
```

## Example: API Subtask

```bash
PARENT_NUM=123  # Parent: "Implement Agent Configuration Management"

SUB_NUM=$(gh issue create \
  --title "Subtask: Create agent_configs API endpoints" \
  --label "type/feature,area/api,subtask,size/m" \
  --body "$(cat <<'EOF'
Part of #123

## Description

Implement the API endpoints for managing agent configurations:
- POST /api/v1/agent-configs
- GET /api/v1/agent-configs
- GET /api/v1/agent-configs/:id
- PUT /api/v1/agent-configs/:id
- DELETE /api/v1/agent-configs/:id

## Parent Task

This subtask is part of the larger "Implement Agent Configuration Management" feature tracked in #123

## Scope

**Included:**
- CRUD endpoints for agent configurations
- Multi-tenancy enforcement (org-scoped)
- Request validation
- Error handling
- Unit and integration tests

**Not Included:**
- Web UI (separate subtask)
- CLI integration (separate subtask)
- Background processing (if needed, separate subtask)

## Implementation Steps

- [ ] Add OpenAPI spec definitions
- [ ] Create SQL queries for CRUD operations
- [ ] Implement handlers following TDD
- [ ] Add JWT middleware for auth
- [ ] Write unit tests for each endpoint
- [ ] Write integration tests for workflow
- [ ] Update API documentation

## Acceptance Criteria

- [ ] All 5 CRUD endpoints implemented
- [ ] Endpoints properly scoped to organization
- [ ] Request/response validation working
- [ ] Test coverage > 85%
- [ ] API documentation updated
- [ ] Integration tests pass

## Technical Notes

- Follow existing endpoint patterns from employees/teams
- Ensure RLS policies enforce org scoping
- Use existing pagination for list endpoint
- Consider caching for frequently accessed configs

## Dependencies

- Depends on: #124 (agent_configs table migration)
- Blocks: #125 (CLI agent commands)
- Blocks: #126 (Web agent config UI)
EOF
)" | grep -oE '#[0-9]+' | cut -c2-)

echo "Created subtask #$SUB_NUM for parent #$PARENT_NUM"
```

## Example: Database Subtask

```bash
PARENT_NUM=123  # Parent: "Implement Agent Configuration Management"

SUB_NUM=$(gh issue create \
  --title "Subtask: Add agent_configs database table" \
  --label "type/feature,area/db,subtask,size/s" \
  --body "$(cat <<'EOF'
Part of #123

## Description

Create the agent_configs table to store per-employee agent configurations with proper relationships and constraints.

## Parent Task

This subtask is part of the larger "Implement Agent Configuration Management" feature tracked in #123

## Scope

**Included:**
- Table schema design
- Migration script
- RLS policies
- Indexes for performance
- Foreign key constraints
- Documentation

**Not Included:**
- API endpoints (separate subtask)
- Seed data (can add later if needed)

## Implementation Steps

- [ ] Design table schema
- [ ] Create migration in shared/schema/migrations/
- [ ] Add RLS policies for multi-tenancy
- [ ] Add indexes for common queries
- [ ] Update ERD documentation
- [ ] Run migration locally and verify
- [ ] Add rollback migration

## Acceptance Criteria

- [ ] Table created with all required columns
- [ ] Foreign keys properly constrained
- [ ] RLS policies enforce org scoping
- [ ] Indexes added for performance
- [ ] ERD updated (run make generate-erd)
- [ ] Migration tested (up and down)

## Technical Notes

Schema draft:
```sql
CREATE TABLE agent_configs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL REFERENCES organizations(id),
  employee_id UUID NOT NULL REFERENCES employees(id),
  agent_id UUID NOT NULL REFERENCES agent_catalog(id),
  config JSONB NOT NULL,
  is_enabled BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT now(),
  updated_at TIMESTAMP DEFAULT now(),
  UNIQUE(employee_id, agent_id)
);

CREATE INDEX idx_agent_configs_employee ON agent_configs(employee_id);
CREATE INDEX idx_agent_configs_org ON agent_configs(org_id);
```

## Dependencies

- Blocks: #124 (API endpoints - needs this table)
- Blocks: #125 (CLI - needs this schema)
EOF
)" | grep -oE '#[0-9]+' | cut -c2-)

echo "Created subtask #$SUB_NUM for parent #$PARENT_NUM"
```

## Subtask Principles

1. **Reference Parent**: Always start body with "Part of #PARENT_NUM"
2. **Inherit Labels**: Copy area labels from parent
3. **Add Subtask Label**: Always include `subtask` label
4. **Clear Scope**: Explicitly state what's included and excluded
5. **Right-Sized**: Aim for size/s or size/m (2 hours to 2 days)
6. **Linked**: Use GraphQL to create proper parent-child relationship
7. **Independent**: Each subtask should be workable independently
8. **Update Parent**: Add comment to parent when subtask created

## Common Subtask Patterns

### Vertical Slice (Full Stack)
- Subtask 1: Database schema
- Subtask 2: API endpoints
- Subtask 3: CLI commands
- Subtask 4: Web UI
- Subtask 5: E2E tests

### Layer by Layer (Horizontal)
- Subtask 1: Data layer (DB + queries)
- Subtask 2: Business logic (services)
- Subtask 3: API layer (handlers)
- Subtask 4: Client layer (CLI/Web)
- Subtask 5: Tests (unit + integration)

### Dependency Order
- Subtask 1: Foundation (DB, core types)
- Subtask 2: Core functionality (API, services)
- Subtask 3: Integrations (CLI, Web)
- Subtask 4: Polish (docs, error handling)
- Subtask 5: Testing (E2E, performance)

## Updating Parent Issue

After creating all subtasks, update parent with checklist:

```bash
gh issue comment $PARENT_NUM --body "$(cat <<'EOF'
## Subtasks

Breaking this down into manageable pieces:

### Implementation
- [ ] #124 - Add agent_configs table (DB)
- [ ] #125 - Create agent_configs API endpoints (API)
- [ ] #126 - Implement `ubik agents` CLI commands (CLI)
- [ ] #127 - Build agent config UI (Web)

### Testing & Polish
- [ ] #128 - E2E agent configuration workflow tests
- [ ] #129 - Documentation and examples

Each subtask can be worked on independently (respecting dependencies).
EOF
)"
```

## Linking Multiple Subtasks

```bash
# Create all subtasks first, collect issue numbers
SUBTASKS=(124 125 126 127 128)

# Link each to parent
for SUB_NUM in "${SUBTASKS[@]}"; do
  SUB_NODE_ID=$(gh api graphql -f query='...' -F number=$SUB_NUM ...)
  gh api graphql -f query='mutation ...' -f parentId="$PARENT_NODE_ID" -f childId="$SUB_NODE_ID"
done

# Update parent with full checklist
gh issue comment $PARENT_NUM --body "Created subtasks: #${SUBTASKS[*]}"
```
