# GitHub Task Manager - Real-World Workflow Examples

This document provides complete examples of common task management workflows using the GitHub Task Manager skill.

## Example 1: Creating a Feature Task

**Scenario:** You need to implement a new API endpoint for user authentication.

```bash
# Create the main feature task
gh issue create \
  --title "Implement JWT authentication for API" \
  --label "type/feature,area/api,priority/p0,size/m" \
  --body "$(cat <<'EOF'
## Description
Add JWT-based authentication to secure API endpoints.

## Acceptance Criteria
- [ ] Generate JWT tokens on successful login
- [ ] Validate JWT tokens on protected endpoints
- [ ] Handle token expiration gracefully
- [ ] Add authentication middleware
- [ ] Write integration tests

## Technical Notes
- Use `github.com/golang-jwt/jwt` library
- Token expiry: 24 hours
- Refresh token support (future)
EOF
)"

# Capture issue number
ISSUE_NUM=42

# Add to GitHub Project
gh project item-add 3 --owner sergei-rastrigin \
  --url "https://github.com/sergei-rastrigin/ubik-enterprise/issues/$ISSUE_NUM"

# Set initial status
./scripts/update-project-status.sh --issue $ISSUE_NUM --status "Todo"
```

## Example 2: Breaking Down a Large Task

**Scenario:** Issue #50 is too large (size/xl) and needs to be split into subtasks.

```bash
# Parent issue: "Implement Agent Management System" (size/xl)
PARENT_NUM=50

# Get parent node ID for GraphQL linking
PARENT_NODE_ID=$(gh api graphql -f query='
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    issue(number: $number) {
      id
    }
  }
}' -f owner='sergei-rastrigin' -f repo='ubik-enterprise' -F number=$PARENT_NUM -q .data.repository.issue.id)

# Define subtasks
declare -a SUBTASKS=(
  "Database: Create agent_catalog and agent_configs tables"
  "API: Implement agent CRUD endpoints"
  "API: Implement agent configuration endpoints"
  "CLI: Add 'ubik agents list' command"
  "CLI: Add 'ubik agents configure' command"
  "Web: Agent catalog page UI"
  "Web: Agent configuration page UI"
  "Tests: E2E test for agent management workflow"
)

# Create each subtask
SUBTASK_NUMS=()
for task_title in "${SUBTASKS[@]}"; do
  # Extract area from title
  if [[ $task_title == Database:* ]]; then
    AREA="area/db"
  elif [[ $task_title == API:* ]]; then
    AREA="area/api"
  elif [[ $task_title == CLI:* ]]; then
    AREA="area/cli"
  elif [[ $task_title == Web:* ]]; then
    AREA="area/web"
  elif [[ $task_title == Tests:* ]]; then
    AREA="area/testing"
  else
    AREA="area/api"
  fi

  # Create subtask
  ISSUE_URL=$(gh issue create \
    --title "$task_title (Part of #$PARENT_NUM)" \
    --label "type/feature,$AREA,priority/p0,size/m,subtask" \
    --body "$(cat <<EOF
Part of #$PARENT_NUM

## Description
${task_title#*: }

## Parent Task
This subtask is part of the larger "Implement Agent Management System" feature tracked in #$PARENT_NUM.

## Acceptance Criteria
- [ ] Implementation complete
- [ ] Tests passing
- [ ] Documentation updated
EOF
)" | tail -1)

  # Extract issue number
  SUB_NUM=$(echo "$ISSUE_URL" | grep -oE '[0-9]+$')
  SUBTASK_NUMS+=("$SUB_NUM")

  # Add to project
  gh project item-add 3 --owner sergei-rastrigin --url "$ISSUE_URL"

  # Get subtask node ID
  SUB_NODE_ID=$(gh api graphql -f query='
  query($owner: String!, $repo: String!, $number: Int!) {
    repository(owner: $owner, name: $repo) {
      issue(number: $number) {
        id
      }
    }
  }' -f owner='sergei-rastrigin' -f repo='ubik-enterprise' -F number=$SUB_NUM -q .data.repository.issue.id)

  # Link to parent via GraphQL
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

  echo "Created subtask #$SUB_NUM"
done

# Update parent issue with subtask checklist
CHECKLIST=""
for num in "${SUBTASK_NUMS[@]}"; do
  CHECKLIST="${CHECKLIST}- [ ] #${num}
"
done

gh issue comment $PARENT_NUM --body "$(cat <<EOF
## Subtasks Created

This large task has been broken down into manageable pieces:

$CHECKLIST

Each subtask can be worked on independently. The parent task will be closed when all subtasks are complete.
EOF
)"

# Update parent status
./scripts/update-project-status.sh --issue $PARENT_NUM --status "In Progress"

echo "✅ Split issue #$PARENT_NUM into ${#SUBTASK_NUMS[@]} subtasks"
```

## Example 3: Complete Development Workflow

**Scenario:** Working on issue #75 from start to finish.

```bash
ISSUE_NUM=75

# 1. Start working - create branch and update status
git checkout main
git pull
git checkout -b issue-$ISSUE_NUM-implement-feature

./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Progress"

# 2. Self-assign the issue
gh issue edit $ISSUE_NUM --add-assignee "@me"

# 3. Do the work (TDD approach)
# - Write failing tests
# - Implement feature
# - All tests pass

# 4. Create PR
gh pr create \
  --title "feat: Implement feature X (#$ISSUE_NUM)" \
  --body "$(cat <<'EOF'
## Summary
Implements feature X as described in #ISSUE_NUM.

## Changes
- Added new endpoint `/api/v1/feature`
- Implemented business logic in service layer
- Added comprehensive tests (unit + integration)

## Testing
- [x] Unit tests passing
- [x] Integration tests passing
- [x] Manual testing complete

## Screenshots
(if UI changes)

Closes #ISSUE_NUM
EOF
)" \
  --label "area/api"

# Capture PR number
PR_NUM=$(gh pr view --json number -q .number)

# 5. Wait for CI checks to pass
echo "Waiting for CI checks..."
gh pr checks $PR_NUM --watch --interval 10

# 6. Check if all CI checks passed
CI_STATUS=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE" or .state == "CANCELLED")) | length')

if [ "$CI_STATUS" -eq 0 ]; then
  echo "✅ All CI checks passed!"

  # Update status to In Review
  ./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Review"

  # Add comment to issue
  gh issue comment $ISSUE_NUM --body "✅ PR #$PR_NUM created and all CI checks passing. Ready for review."

else
  echo "❌ Some CI checks failed. Please investigate:"
  gh pr checks $PR_NUM
  gh issue comment $ISSUE_NUM --body "⚠️ PR #$PR_NUM created but CI checks failed. Investigating..."
fi

# 7. After review and approval, merge PR
# (Status will auto-update to "Done" when PR is merged and issue is closed)
```

## Example 4: Handling Blocked Tasks

**Scenario:** Issue #88 is blocked waiting for backend API to be implemented.

```bash
ISSUE_NUM=88
BLOCKING_ISSUE=87

# Update status to Blocked
./scripts/update-project-status.sh --issue $ISSUE_NUM --status "Blocked"

# Add comment explaining blocker
gh issue comment $ISSUE_NUM --body "$(cat <<EOF
## ⚠️ Blocked

This task is blocked by #$BLOCKING_ISSUE.

**Blocker:** Waiting for backend API endpoints to be implemented.

**Required endpoints:**
- GET /api/v1/agents
- POST /api/v1/agents/:id/configure

**Next steps:**
1. Wait for #$BLOCKING_ISSUE to be completed
2. Test API endpoints
3. Resume frontend implementation

Will update status to "In Progress" once blocker is resolved.
EOF
)"

# Link the dependency
gh issue comment $BLOCKING_ISSUE --body "⚠️ Issue #$ISSUE_NUM is blocked waiting for this to be completed."

# When blocker is resolved, update status
# ./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Progress"
```

## Example 5: Reporting a Bug

**Scenario:** User reports that login fails with valid credentials.

```bash
gh issue create \
  --title "Bug: Login fails with valid credentials" \
  --label "type/bug,area/api,priority/p0" \
  --body "$(cat <<'EOF'
## Bug Description
Login endpoint returns 401 Unauthorized even when providing valid email and password.

## Steps to Reproduce
1. Start server with `make dev`
2. Create test user: Alice (alice@acme.com)
3. Attempt login via API:
   ```bash
   curl -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"alice@acme.com","password":"password123"}'
   ```
4. Observe 401 error

## Expected Behavior
Should return 200 OK with JWT token and employee data.

## Actual Behavior
Returns 401 Unauthorized with error message "Invalid credentials".

## Environment
- Version: v0.2.0
- OS: macOS 14.0
- Go version: 1.24
- Database: PostgreSQL 15

## Logs
```
ERROR: Authentication failed for user alice@acme.com
ERROR: bcrypt password comparison failed
```

## Impact
- **Severity:** Critical (P0)
- **Users affected:** All users
- **Workaround:** None
EOF
)"

# Capture issue number
BUG_NUM=$(echo "$ISSUE_URL" | grep -oE '[0-9]+$')

# Add to project and prioritize
gh project item-add 3 --owner sergei-rastrigin \
  --url "https://github.com/sergei-rastrigin/ubik-enterprise/issues/$BUG_NUM"

./scripts/update-project-status.sh --issue $BUG_NUM --status "Todo"

# Assign immediately if P0
gh issue edit $BUG_NUM --add-assignee "@me"
```

## Example 6: Querying Tasks for Daily Standup

**Scenario:** Prepare for daily standup - what did I work on, what am I working on, any blockers?

```bash
echo "=== My Tasks ==="

echo -e "\n📋 **What I did yesterday (Recently Closed):**"
gh issue list \
  --assignee "@me" \
  --state closed \
  --search "closed:>=$(date -v-1d +%Y-%m-%d)" \
  --json number,title \
  --jq '.[] | "- Completed #\(.number): \(.title)"'

echo -e "\n🚧 **What I'm working on today (In Progress):**"
gh issue list \
  --assignee "@me" \
  --state open \
  --label "status/in-progress" \
  --json number,title \
  --jq '.[] | "- Working on #\(.number): \(.title)"'

echo -e "\n⚠️ **Blockers:**"
gh issue list \
  --assignee "@me" \
  --state open \
  --label "blocked" \
  --json number,title \
  --jq '.[] | "- Blocked #\(.number): \(.title)"'

echo -e "\n📝 **Ready to Start (Todo):**"
gh issue list \
  --assignee "@me" \
  --state open \
  --label "priority/p0,priority/p1" \
  --json number,title,labels \
  --jq '.[] | select(.labels | map(.name) | index("status/in-progress") | not) | "- Todo #\(.number): \(.title)"' \
  | head -3
```

## Example 7: Sprint Planning

**Scenario:** Planning sprint v0.4.0 - identify and prioritize tasks.

```bash
MILESTONE="v0.4.0"

echo "=== Sprint Planning: $MILESTONE ==="

# View all tasks in milestone
echo -e "\n📊 **All Tasks:**"
gh issue list \
  --milestone "$MILESTONE" \
  --json number,title,labels \
  --jq '.[] | "#\(.number): \(.title) [\(.labels | map(select(.name | startswith("size/"))) | .[].name)]"'

# Count by priority
echo -e "\n🔥 **Priority Breakdown:**"
for priority in p0 p1 p2 p3; do
  count=$(gh issue list --milestone "$MILESTONE" --label "priority/$priority" --json number -q 'length')
  echo "- Priority $priority: $count tasks"
done

# Count by area
echo -e "\n🎯 **Area Breakdown:**"
for area in api cli web db infra testing; do
  count=$(gh issue list --milestone "$MILESTONE" --label "area/$area" --json number -q 'length')
  echo "- $area: $count tasks"
done

# Estimate total effort
echo -e "\n⏱️ **Effort Estimation:**"
for size in xs s m l xl; do
  count=$(gh issue list --milestone "$MILESTONE" --label "size/$size" --json number -q 'length')
  echo "- Size $size: $count tasks"
done
```

## Best Practices Demonstrated

1. **Always update status** - Move tasks through workflow states
2. **Comment on changes** - Explain blockers, progress, decisions
3. **Link related issues** - Show dependencies and relationships
4. **Use labels consistently** - Enable querying and reporting
5. **Wait for CI** - Only mark "In Review" after checks pass
6. **Close via PR** - Use "Closes #123" in PR description
7. **Query for insights** - Use `gh` CLI for reporting and planning

## Tips

- **Save common commands** - Create shell aliases for frequently used commands
- **Use templates** - Refer to `templates/` directory for consistent issue creation
- **Automate status updates** - Use `update-project-status.sh` script
- **Check CI before review** - `gh pr checks --watch` to monitor CI
- **Document decisions** - Add comments explaining "why" not just "what"

---

**These examples are production-ready workflows used in ubik-enterprise development.**
