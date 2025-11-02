# Complete Development Workflow Examples

Real-world examples demonstrating the complete development lifecycle from task start to merge.

## Example 1: Simple Feature Implementation

**Scenario:** Implement a new API endpoint for listing agents

```bash
# ==========================================
# STEP 1: START TASK (Workflow 1)
# ==========================================
ISSUE_NUM=89

# Update status to In Progress
./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Progress"

# Self-assign
gh issue edit $ISSUE_NUM --add-assignee "@me"

# Create worktree
git worktree add ../ubik-issue-89 -b issue-89-list-agents-endpoint
cd ../ubik-issue-89

# Verify setup
echo "Working on: $(git branch --show-current)"
echo "Location: $(pwd)"

# ==========================================
# STEP 2: IMPLEMENT (TDD)
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
# STEP 3: CREATE PR (Workflow 2)
# ==========================================

# Commit changes
git add .
git commit -m "feat: Add GET /api/v1/agents endpoint (#89)

Implements endpoint to list all available AI agents.

- Added ListAgents SQL query
- Implemented AgentsService.List() method
- Added HTTP handler with tests
- 95% test coverage

Closes #89

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Push to remote
git push -u origin issue-89-list-agents-endpoint

# Create PR
ISSUE_TITLE=$(gh issue view 89 --json title -q .title)
gh pr create \
  --title "feat: ${ISSUE_TITLE} (#89)" \
  --label "area/api" \
  --body "$(cat <<EOF
## Summary
Implements GET /api/v1/agents endpoint as described in #89.

## Changes
- Added \`ListAgents\` SQL query
- Implemented \`AgentsService.List()\` method
- Added HTTP handler with Chi routing
- Comprehensive tests (15 unit + 3 integration)

## Testing
- [x] Unit tests passing (95% coverage)
- [x] Integration tests passing
- [x] Manual testing via curl
- [x] No breaking changes

Closes #89
EOF
)"

PR_NUM=$(gh pr view --json number -q .number)
echo "✅ Created PR #${PR_NUM}"

# Wait for CI (MANDATORY!)
echo "🔍 Waiting for CI checks..."
gh pr checks $PR_NUM --watch --interval 10

# Check CI status and update
CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')

if [ "$CI_FAILED" -eq 0 ]; then
  echo "✅ All CI checks passed!"

  # Update to In Review
  ./scripts/update-project-status.sh --issue 89 --status "In Review"

  # Comment on issue
  gh issue comment 89 --body "✅ PR #${PR_NUM} created and all CI checks passing. Ready for review."

  echo "✅ Task ready for review"
else
  echo "❌ CI checks failed. Fix before review."
  exit 1
fi

# Return to main workspace
cd ../ubik-enterprise

# ==========================================
# STEP 4: MERGE PR (Workflow 3 - after approval)
# ==========================================

# Verify pre-merge conditions
CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')
MERGEABLE=$(gh pr view $PR_NUM --json mergeable -q .mergeable)

if [ "$CI_FAILED" -eq 0 ] && [ "$MERGEABLE" = "MERGEABLE" ]; then
  echo "✅ Ready to merge"

  # Merge PR (squash for clean history)
  gh pr merge $PR_NUM --squash --delete-branch

  # Verify issue closed
  ISSUE_STATE=$(gh issue view 89 --json state -q .state)
  echo "Issue state: $ISSUE_STATE"  # CLOSED

  # Clean up worktree
  git worktree remove ../ubik-issue-89

  echo "✅ Task complete! Issue #89 closed and merged."
fi
```

---

## Example 2: Bug Fix with CI Failure

**Scenario:** Fix authentication bug, encounter CI failure, fix and retry

```bash
# ==========================================
# START TASK
# ==========================================
ISSUE_NUM=101

./scripts/update-project-status.sh --issue 101 --status "In Progress"
gh issue edit 101 --add-assignee "@me"

git worktree add ../ubik-issue-101 -b issue-101-fix-auth-bug
cd ../ubik-issue-101

# ==========================================
# IMPLEMENT FIX
# ==========================================

# Write reproduction test
vim internal/auth/password_test.go

# Implement fix
vim internal/auth/password.go

# Tests pass locally
make test

# ==========================================
# CREATE PR
# ==========================================

git add .
git commit -m "fix: Resolve bcrypt password comparison issue (#101)

Root cause: Inconsistent bcrypt rounds between seed data (10)
and auth code (12).

Fix: Updated seed data to use 12 rounds consistently.

Fixes #101

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

git push -u origin issue-101-fix-auth-bug

gh pr create \
  --title "fix: Resolve authentication bug (#101)" \
  --label "area/api,type/bug" \
  --body "..."

PR_NUM=$(gh pr view --json number -q .number)

# Wait for CI
gh pr checks $PR_NUM --watch --interval 10

# ==========================================
# CI FAILED! (Linting error)
# ==========================================

CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')

if [ "$CI_FAILED" -gt 0 ]; then
  echo "❌ CI failed. Investigating..."

  # Check what failed
  gh pr checks $PR_NUM
  # Output: lint: FAILED - unused variable 'ctx'

  # Fix the issue
  vim internal/auth/password.go  # Remove unused variable

  # Test locally
  make lint

  # Commit fix
  git add .
  git commit -m "fix: Remove unused variable"
  git push

  # Add comment
  gh issue comment 101 --body "⚠️ CI failed due to linting. Fixed and pushed."

  # Wait for CI again
  echo "🔍 Waiting for CI (retry)..."
  gh pr checks $PR_NUM --watch --interval 10

  # Check again
  CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')

  if [ "$CI_FAILED" -eq 0 ]; then
    echo "✅ CI passed on second attempt"

    # NOW update to In Review
    ./scripts/update-project-status.sh --issue 101 --status "In Review"
    gh issue comment 101 --body "✅ CI issues resolved. PR ready for review."
  fi
fi

cd ../ubik-enterprise

# ==========================================
# MERGE PR
# ==========================================

# After approval, merge
gh pr merge $PR_NUM --squash --delete-branch
git worktree remove ../ubik-issue-101

echo "✅ Bug fix complete and merged"
```

---

## Example 3: Feature with Merge Conflicts

**Scenario:** Implement feature, encounter merge conflicts during PR, resolve

```bash
# ==========================================
# START TASK
# ==========================================
ISSUE_NUM=112

./scripts/update-project-status.sh --issue 112 --status "In Progress"
gh issue edit 112 --add-assignee "@me"

git worktree add ../ubik-issue-112 -b issue-112-add-pagination
cd ../ubik-issue-112

# ==========================================
# IMPLEMENT FEATURE
# ==========================================

vim internal/handlers/agents.go  # Add pagination
vim internal/handlers/agents_test.go

make test

# ==========================================
# CREATE PR
# ==========================================

git add .
git commit -m "feat: Add pagination to agents endpoint (#112)

Implements cursor-based pagination for GET /api/v1/agents.

- Added pagination parameters (limit, cursor)
- Updated handler to support pagination
- Added tests for pagination logic

Closes #112

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

git push -u origin issue-112-add-pagination

gh pr create \
  --title "feat: Add pagination to agents endpoint (#112)" \
  --label "area/api" \
  --body "..."

PR_NUM=$(gh pr view --json number -q .number)

# Wait for CI
gh pr checks $PR_NUM --watch --interval 10

CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')

if [ "$CI_FAILED" -eq 0 ]; then
  ./scripts/update-project-status.sh --issue 112 --status "In Review"
  gh issue comment 112 --body "✅ PR ready for review"
fi

cd ../ubik-enterprise

# ==========================================
# ANOTHER PR MERGED TO MAIN!
# Now our PR has conflicts
# ==========================================

# Check merge status
MERGEABLE=$(gh pr view $PR_NUM --json mergeable -q .mergeable)
echo "Mergeable: $MERGEABLE"  # CONFLICTING

echo "⚠️ Merge conflicts detected. Resolving..."

# Navigate back to worktree
cd ../ubik-issue-112

# Fetch latest changes
git fetch origin

# Merge main into feature
git merge origin/main
# CONFLICT in internal/handlers/agents.go

# Resolve conflicts manually
vim internal/handlers/agents.go
# Fix conflicts, keep both changes

# Stage resolved files
git add internal/handlers/agents.go

# Commit merge
git commit -m "chore: Resolve merge conflicts with main"

# Push
git push

# Wait for CI again (IMPORTANT!)
echo "🔍 Waiting for CI after conflict resolution..."
gh pr checks $PR_NUM --watch --interval 10

# Verify no more conflicts
MERGEABLE=$(gh pr view $PR_NUM --json mergeable -q .mergeable)
CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')

if [ "$MERGEABLE" = "MERGEABLE" ] && [ "$CI_FAILED" -eq 0 ]; then
  echo "✅ Conflicts resolved, CI passing"

  # Comment on PR
  gh pr comment $PR_NUM --body "✅ Merge conflicts resolved. CI checks passing."
fi

cd ../ubik-enterprise

# ==========================================
# MERGE PR
# ==========================================

# After approval
gh pr merge $PR_NUM --squash --delete-branch
git worktree remove ../ubik-issue-112

echo "✅ Feature merged successfully after conflict resolution"
```

---

## Example 4: Multi-Commit Feature PR

**Scenario:** Large feature with multiple logical commits

```bash
# ==========================================
# START TASK
# ==========================================
ISSUE_NUM=125

./scripts/update-project-status.sh --issue 125 --status "In Progress"
gh issue edit 125 --add-assignee "@me"

git worktree add ../ubik-issue-125 -b issue-125-agent-approval-workflow
cd ../ubik-issue-125

# ==========================================
# IMPLEMENT FEATURE (Multiple Commits)
# ==========================================

# Commit 1: Database schema
vim schema/schema.sql  # Add agent_requests table
make db-reset && make generate-db

git add schema/ generated/db/
git commit -m "feat(db): Add agent_requests table for approval workflow (#125)"

# Commit 2: Service layer
vim internal/service/agent_requests.go
vim internal/service/agent_requests_test.go

make test-unit

git add internal/service/
git commit -m "feat(service): Implement AgentRequestsService (#125)

- Create agent request
- List pending requests
- Approve/reject requests"

# Commit 3: API handlers
vim internal/handlers/agent_requests.go
vim internal/handlers/agent_requests_test.go

make test

git add internal/handlers/
git commit -m "feat(api): Add agent request endpoints (#125)

- POST /api/v1/agent-requests
- GET /api/v1/agent-requests
- PUT /api/v1/agent-requests/:id/approve
- PUT /api/v1/agent-requests/:id/reject"

# Commit 4: Integration tests
vim tests/integration/agent_requests_test.go

make test-integration

git add tests/
git commit -m "test: Add integration tests for agent approval workflow (#125)"

# ==========================================
# CREATE PR
# ==========================================

# Push all commits
git push -u origin issue-125-agent-approval-workflow

# Create PR
gh pr create \
  --title "feat: Implement agent approval workflow (#125)" \
  --label "area/api" \
  --body "$(cat <<EOF
## Summary
Implements complete approval workflow for employee agent access requests.

## Changes
- **Database**: Added \`agent_requests\` table
- **Service**: Implemented AgentRequestsService with CRUD
- **API**: Added 4 new endpoints
- **Tests**: 25 unit tests + 8 integration tests

## Commit Structure
1. Database schema changes
2. Service layer implementation
3. API endpoint handlers
4. Integration tests

Each commit is independently reviewable.

## Testing
- [x] Unit tests passing (96% coverage)
- [x] Integration tests passing
- [x] Manual testing complete

Closes #125
EOF
)"

PR_NUM=$(gh pr view --json number -q .number)

# Wait for CI
gh pr checks $PR_NUM --watch --interval 10

if [ "$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')" -eq 0 ]; then
  ./scripts/update-project-status.sh --issue 125 --status "In Review"
  gh issue comment 125 --body "✅ PR #${PR_NUM} ready. Feature implemented with 4 logical commits."
fi

cd ../ubik-enterprise

# ==========================================
# MERGE PR (Preserve Commit History)
# ==========================================

# Use --merge to keep all commits
gh pr merge $PR_NUM --merge --delete-branch

git worktree remove ../ubik-issue-125

echo "✅ Feature merged with commit history preserved"
```

---

## Example 5: Emergency Hotfix

**Scenario:** Production bug, need immediate fix

```bash
# ==========================================
# START TASK (HOTFIX)
# ==========================================
ISSUE_NUM=150

./scripts/update-project-status.sh --issue 150 --status "In Progress"
gh issue edit 150 --add-assignee "@me" --add-label "priority/p0"

# Create hotfix worktree from main
git worktree add ../ubik-issue-150 -b hotfix-db-connection-pool
cd ../ubik-issue-150

# ==========================================
# IMPLEMENT MINIMAL FIX
# ==========================================

# Only change what's necessary
vim internal/db/connection.go
# MaxOpenConns: 10 → 50

# Test
make test

# ==========================================
# CREATE PR (EXPEDITED)
# ==========================================

git add .
git commit -m "fix: 🚨 HOTFIX - Increase DB connection pool (#150)

Critical production issue: API returning 500 due to connection
pool exhaustion.

Fix: Increased MaxOpenConns from 10 to 50.

Fixes #150

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

git push -u origin hotfix-db-connection-pool

gh pr create \
  --title "fix: 🚨 HOTFIX - Database connection pool exhaustion (#150)" \
  --label "area/api,type/bug,priority/p0" \
  --body "..."

PR_NUM=$(gh pr view --json number -q .number)

# Request immediate review
gh pr review --request-reviewer @tech-lead

# Monitor CI closely (poll every 5s, not 10s)
gh pr checks $PR_NUM --watch --interval 5

if [ "$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')" -eq 0 ]; then
  ./scripts/update-project-status.sh --issue 150 --status "In Review"
  gh issue comment 150 --body "🚨 HOTFIX ready for immediate merge. CI passing."

  # Merge ASAP after approval
  gh pr merge $PR_NUM --squash --delete-branch

  # Clean up
  cd ../ubik-enterprise
  git worktree remove ../ubik-issue-150

  echo "✅ Hotfix deployed!"

  # Create follow-up issues
  gh issue create --title "Investigate slow queries causing connection pool issues" \
    --label "type/research,area/api,priority/p1"
fi
```

---

## Best Practices Demonstrated

### 1. Always Follow the Three Workflows
- Start Task → Create PR → Merge PR
- Never skip steps
- Never skip CI checks

### 2. Status Updates at Every Step
- "In Progress" when starting
- "In Review" when CI passes
- "Done" when merged

### 3. Git Worktrees for Parallel Work
- One worktree per issue
- Clean separation
- Easy cleanup

### 4. Proper Commit Messages
- Type prefix (feat/fix/chore)
- Issue number
- Detailed description
- Attribution

### 5. CI Quality Gates
- Always wait for CI
- Never merge with failures
- Re-run CI after changes

### 6. Clean History
- Squash for feature PRs
- Merge for multi-commit PRs
- Rebase cautiously

---

**These examples demonstrate production-ready workflows used in ubik-enterprise development.**
