# GitHub PR Workflow - Real-World Examples

Complete examples of PR workflows from creation to merge.

## Example 1: Simple Feature PR

**Scenario:** Add new API endpoint for listing agents

```bash
# ==========================================
# STEP 1: Create Feature Branch
# ==========================================
ISSUE_NUM=89
git checkout main
git pull
git checkout -b issue-${ISSUE_NUM}-list-agents-endpoint

# ==========================================
# STEP 2: Update Issue Status
# ==========================================
./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Progress"
gh issue edit $ISSUE_NUM --add-assignee "@me"

# ==========================================
# STEP 3: Implement Feature (TDD)
# ==========================================
# Write tests first
vim sqlc/queries/agents.sql
vim internal/service/agents_test.go

# Run tests (should fail)
make test-unit

# Implement feature
vim internal/service/agents.go
vim internal/handlers/agents.go

# Run tests (should pass)
make test

# ==========================================
# STEP 4: Create Pull Request
# ==========================================
ISSUE_TITLE=$(gh issue view $ISSUE_NUM --json title -q .title)
ISSUE_LABELS=$(gh issue view $ISSUE_NUM --json labels -q '.labels[].name' | grep -E '^area/' | head -1)

gh pr create \
  --title "feat: ${ISSUE_TITLE} (#${ISSUE_NUM})" \
  --label "$ISSUE_LABELS" \
  --body "$(cat <<EOF
## Summary
Implements GET /api/v1/agents endpoint to list available AI agents.

## Changes
- Added \`ListAgents\` SQL query with filters
- Implemented \`AgentsService.List()\` method
- Added \`AgentsHandler.List()\` HTTP handler
- Registered route in Chi router
- Added 15 unit tests + 3 integration tests

## Testing
- [x] Unit tests passing (98% coverage)
- [x] Integration tests passing
- [x] Manual testing via curl
- [x] No breaking changes

## Additional Notes
Response includes agent metadata (name, version, capabilities).
Pagination will be added in follow-up issue.

---

Closes #${ISSUE_NUM}

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"

PR_NUM=$(gh pr view --json number -q .number)
echo "✅ Created PR #${PR_NUM}"

# ==========================================
# STEP 5: Monitor CI Checks
# ==========================================
echo "🔍 Waiting for CI checks..."
gh pr checks $PR_NUM --watch --interval 10

# ==========================================
# STEP 6: Check CI Results
# ==========================================
CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE" or .state == "CANCELLED")) | length')

if [ "$CI_FAILED" -eq 0 ]; then
  echo "✅ All CI checks passed!"

  # Update issue status to In Review
  ./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Review"

  # Add success comment
  gh issue comment $ISSUE_NUM --body "✅ PR #${PR_NUM} created and all CI checks passing. Ready for review."
else
  echo "❌ Some CI checks failed:"
  gh pr checks $PR_NUM
  exit 1
fi

# ==========================================
# STEP 7: After Review Approval, Merge
# ==========================================
# (After reviewer approves)
gh pr merge $PR_NUM --squash --delete-branch

# Verify issue closed
gh issue view $ISSUE_NUM --json state -q .state
# Output: CLOSED

echo "✅ Feature complete! PR merged, issue closed."
```

---

## Example 2: PR with CI Failure

**Scenario:** CI fails due to linting error, fix and retry

```bash
# ==========================================
# Create PR (steps 1-4 same as Example 1)
# ==========================================
ISSUE_NUM=92
PR_NUM=45

# ==========================================
# CI Fails!
# ==========================================
gh pr checks $PR_NUM --watch --interval 10

CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')

if [ "$CI_FAILED" -gt 0 ]; then
  echo "❌ CI checks failed. Investigating..."

  # View which checks failed
  gh pr checks $PR_NUM --json name,state \
    -q '.[] | select(.state == "FAILURE") | "\(.name): FAILED"'

  # Output:
  # lint: FAILED

  # Get logs
  RUN_ID=$(gh run list --limit 1 --json databaseId -q '.[0].databaseId')
  gh run view $RUN_ID --log | grep -A 10 "Error"

  # Output shows: "Error: unused variable 'ctx' in handlers/agents.go:45"
fi

# ==========================================
# Fix the Issue
# ==========================================
# Remove unused variable
vim internal/handlers/agents.go

# Test locally
make lint

# Commit fix
git add .
git commit -m "fix: Remove unused variable causing lint failure"
git push

# Add comment to issue
gh issue comment $ISSUE_NUM --body "⚠️ CI failed due to linting error. Fixed and pushed."

# ==========================================
# Wait for CI Again
# ==========================================
echo "🔍 Waiting for CI checks (retry)..."
gh pr checks $PR_NUM --watch --interval 10

CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')

if [ "$CI_FAILED" -eq 0 ]; then
  echo "✅ All CI checks passed!"

  # NOW update status to In Review
  ./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Review"

  gh issue comment $ISSUE_NUM --body "✅ CI issues resolved. PR #${PR_NUM} ready for review."
fi

# ==========================================
# Merge after approval
# ==========================================
gh pr merge $PR_NUM --squash --delete-branch
```

---

## Example 3: PR with Review Feedback

**Scenario:** Reviewer requests changes, address feedback, re-review

```bash
# ==========================================
# Initial PR Created (steps 1-6 same)
# ==========================================
ISSUE_NUM=95
PR_NUM=48

# Status: "In Review"
# CI passed, waiting for review

# ==========================================
# Reviewer Requests Changes
# ==========================================
# Reviewer comments:
# - "Please add error handling for empty response"
# - "Add validation for negative page numbers"
# - "Extract magic number 100 to constant"

# View review comments
gh pr view $PR_NUM --comments

# ==========================================
# Address Feedback
# ==========================================
# Make requested changes
vim internal/handlers/agents.go
# - Add error handling
# - Add input validation
# - Extract constant: const DefaultPageSize = 100

# Update tests
vim internal/handlers/agents_test.go
# - Add test for empty response
# - Add test for negative page number

# Run tests locally
make test

# ==========================================
# Commit Changes
# ==========================================
git add .
git commit -m "fix: Address review feedback

- Add error handling for empty response
- Add validation for negative page numbers
- Extract magic number to constant DefaultPageSize"

git push

# ==========================================
# Wait for CI Again
# ==========================================
gh pr checks $PR_NUM --watch --interval 10

if [ "$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')" -eq 0 ]; then
  echo "✅ CI passed after changes"

  # Comment on PR
  gh pr comment $PR_NUM --body "✅ Addressed all review feedback. Changes:
- Added error handling for empty response
- Added input validation for negative page numbers
- Extracted DefaultPageSize constant

All tests passing. Ready for re-review @reviewer."

  # Request re-review
  gh pr review $PR_NUM --approve=false --request-changes=false
fi

# ==========================================
# After Re-Review Approval
# ==========================================
# Reviewer approves
gh pr merge $PR_NUM --squash --delete-branch

echo "✅ PR merged after addressing feedback"
```

---

## Example 4: PR with Merge Conflicts

**Scenario:** Another PR merged to main, causing conflicts

```bash
# ==========================================
# PR Created, CI Passed, In Review
# ==========================================
ISSUE_NUM=97
PR_NUM=50

# ==========================================
# Another PR Merges to Main
# ==========================================
# Now your PR has conflicts with main

# GitHub shows: "This branch has conflicts that must be resolved"

gh pr view $PR_NUM --json mergeable -q .mergeable
# Output: CONFLICTING

# ==========================================
# Option 1: Use GitHub's Update Branch
# ==========================================
gh pr update-branch $PR_NUM

# Wait for CI
gh pr checks $PR_NUM --watch --interval 10

# ==========================================
# Option 2: Manually Merge Main
# ==========================================
# If automatic merge fails:

git checkout issue-97-feature
git fetch origin
git merge origin/main

# Conflicts in: internal/handlers/router.go
# CONFLICT (content): Merge conflict in internal/handlers/router.go

# Resolve conflicts manually
vim internal/handlers/router.go
# Fix conflicts, keep both changes

git add internal/handlers/router.go
git commit -m "chore: Merge main into feature branch"
git push

# ==========================================
# Wait for CI After Conflict Resolution
# ==========================================
gh pr checks $PR_NUM --watch --interval 10

if [ "$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')" -eq 0 ]; then
  echo "✅ CI passed after merge"

  # Verify no conflicts
  gh pr view $PR_NUM --json mergeable -q .mergeable
  # Output: MERGEABLE

  # Comment on PR
  gh pr comment $PR_NUM --body "✅ Merged latest changes from main. Conflicts resolved. CI passing."
fi

# ==========================================
# Merge PR
# ==========================================
gh pr merge $PR_NUM --squash --delete-branch
```

---

## Example 5: Emergency Hotfix PR

**Scenario:** Production is down, need immediate fix

```bash
# ==========================================
# Critical Bug Discovered
# ==========================================
# Production API returning 500 errors
# Issue: Database connection pool exhausted

ISSUE_NUM=101

# ==========================================
# Create Hotfix Branch
# ==========================================
git checkout main
git pull
git checkout -b hotfix-db-connection-pool

# Update issue status immediately
./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Progress"
gh issue edit $ISSUE_NUM --add-label "priority/p0"

# ==========================================
# Make MINIMAL Fix
# ==========================================
# Only fix the immediate issue
vim internal/db/connection.go
# Change: MaxOpenConns from 10 to 50

# Test locally
make test

# ==========================================
# Create Hotfix PR
# ==========================================
gh pr create \
  --title "fix: 🚨 HOTFIX - Increase DB connection pool (#${ISSUE_NUM})" \
  --label "area/api,type/bug,priority/p0" \
  --body "$(cat <<EOF
## 🚨 HOTFIX - Database Connection Pool Exhaustion

**Severity:** Critical (P0)
**Impact:** Complete API outage - all endpoints returning 500
**Affected Users:** 100% of active users
**Duration:** 15 minutes

## Issue
Database connection pool exhausted due to traffic spike causing all requests to fail.

## Fix
Increased MaxOpenConns from 10 to 50

## Changes
- Modified \`internal/db/connection.go\`: MaxOpenConns = 50

## Testing
- [x] Tested in staging with 100 concurrent requests
- [x] Verified connection pool doesn't exhaust
- [x] Rollback plan: revert to v0.3.0

## Follow-up
- Issue #102: Investigate slow queries
- Issue #103: Add connection pool monitoring

---

Fixes #${ISSUE_NUM}

🚨 HOTFIX - Expedited merge required
EOF
)"

PR_NUM=$(gh pr view --json number -q .number)

# Request immediate review
gh pr review --request-reviewer @tech-lead

# ==========================================
# Monitor CI Closely (Poll Every 5s)
# ==========================================
gh pr checks $PR_NUM --watch --interval 5

# ==========================================
# Merge IMMEDIATELY When CI Passes
# ==========================================
if [ "$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')" -eq 0 ]; then
  echo "✅ CI passed! Merging hotfix immediately..."

  # Update status
  ./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Review"

  # Wait for approval (if required by branch protection)
  # Then merge
  gh pr merge $PR_NUM --squash --delete-branch

  echo "✅ Hotfix deployed! Monitoring production..."

  # Create follow-up issues
  gh issue create --title "Investigate slow queries causing connection pool exhaustion" \
    --label "type/research,area/api,priority/p1" --body "Follow-up from hotfix #${ISSUE_NUM}"

  gh issue create --title "Add connection pool metrics to monitoring" \
    --label "type/feature,area/infra,priority/p1" --body "Follow-up from hotfix #${ISSUE_NUM}"
fi
```

---

## Example 6: Multi-Commit PR Workflow

**Scenario:** Feature requires multiple logical commits

```bash
# ==========================================
# Feature: Implement Agent Approval Workflow
# ==========================================
ISSUE_NUM=105

git checkout -b issue-${ISSUE_NUM}-agent-approval

./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Progress"

# ==========================================
# Commit 1: Database Schema
# ==========================================
vim schema/schema.sql
# Add agent_requests table

make db-reset
make generate-db

git add schema/schema.sql generated/db/
git commit -m "feat(db): Add agent_requests table for approval workflow"

# ==========================================
# Commit 2: Service Layer
# ==========================================
vim internal/service/agent_requests.go
vim internal/service/agent_requests_test.go

make test-unit

git add internal/service/
git commit -m "feat(service): Implement AgentRequestsService

- Create agent request
- List pending requests
- Approve/reject requests"

# ==========================================
# Commit 3: API Handlers
# ==========================================
vim internal/handlers/agent_requests.go
vim internal/handlers/agent_requests_test.go

make test

git add internal/handlers/
git commit -m "feat(api): Add agent request endpoints

- POST /api/v1/agent-requests
- GET /api/v1/agent-requests
- PUT /api/v1/agent-requests/:id/approve
- PUT /api/v1/agent-requests/:id/reject"

# ==========================================
# Commit 4: Integration Tests
# ==========================================
vim tests/integration/agent_requests_test.go

make test-integration

git add tests/
git commit -m "test: Add integration tests for agent approval workflow"

# ==========================================
# Push All Commits
# ==========================================
git push -u origin issue-${ISSUE_NUM}-agent-approval

# ==========================================
# Create PR
# ==========================================
gh pr create \
  --title "feat: Implement agent approval workflow (#${ISSUE_NUM})" \
  --label "area/api" \
  --body "$(cat <<EOF
## Summary
Implements approval workflow for employee agent access requests.

## Changes
- **Database**: Added \`agent_requests\` table
- **Service**: Implemented AgentRequestsService with CRUD operations
- **API**: Added 4 new endpoints for request management
- **Tests**: 25 unit tests + 8 integration tests

## Testing
- [x] Unit tests passing (96% coverage)
- [x] Integration tests passing
- [x] Manual testing complete
- [x] No breaking changes

## Commit Structure
1. Database schema changes
2. Service layer implementation
3. API endpoint handlers
4. Integration tests

Each commit is independently reviewable.

---

Closes #${ISSUE_NUM}
EOF
)"

PR_NUM=$(gh pr view --json number -q .number)

# ==========================================
# Monitor CI
# ==========================================
gh pr checks $PR_NUM --watch --interval 10

if [ "$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')" -eq 0 ]; then
  ./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Review"
  gh issue comment $ISSUE_NUM --body "✅ PR #${PR_NUM} ready. Feature implemented with 4 logical commits for easier review."
fi

# ==========================================
# Merge with Merge Commit (Preserve History)
# ==========================================
# Use --merge to keep all commits
gh pr merge $PR_NUM --merge --delete-branch
```

---

## Best Practices Demonstrated

### 1. Always Wait for CI
- Never skip CI checks
- Use `--watch` to poll automatically
- Only update to "In Review" when CI passes

### 2. Clear Communication
- Add comments when making changes
- Explain reasoning in commit messages
- Update issue status at key points

### 3. Proper Status Management
- `In Progress` → Working on feature
- `In Review` → PR created, CI passed
- `Done` → PR merged, issue closed

### 4. Handle Failures Gracefully
- Investigate failures immediately
- Fix and push again
- Document what went wrong

### 5. Merge Strategies
- **Squash**: Feature PRs, clean history
- **Merge**: Multi-commit PRs, preserve history
- **Rebase**: Linear history (use cautiously)

---

**These examples demonstrate production-ready PR workflows used in ubik-enterprise development.**
