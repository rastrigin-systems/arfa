# Agent Configuration Updates

**Date:** 2025-11-01
**Purpose:** Ensure `go-backend-developer` and `frontend-developer` agents always follow the standard PR workflow

---

## Overview

Both specialized agents (`go-backend-developer` and `frontend-developer`) must be configured to **automatically follow the standard PR workflow** for all code changes. This ensures consistency, quality gates, and proper CI/CD integration.

---

## Required Agent Configuration Changes

### Location

Agent configurations are stored in:
```
~/.claude/agents/go-backend-developer.md
~/.claude/agents/frontend-developer.md
```

### What to Add

Add the following section to **BOTH** agent configuration files:

```markdown
## Mandatory PR Workflow

**CRITICAL: ALL code changes MUST follow this workflow - NO EXCEPTIONS**

### Parallel Development with Git Worktrees

**When working on multiple tasks in parallel, use git worktrees to avoid conflicts:**

```bash
# Create worktree for Issue #13 (in separate directory)
git worktree add ../arfa-issue-13 -b feature/agent-catalog-page

# Create worktree for Issue #14 (in another separate directory)
git worktree add ../arfa-issue-14 -b feature/org-agent-config

# Now you can work in both directories independently:
# - Terminal 1: cd ../arfa-issue-13 (frontend work)
# - Terminal 2: cd ../arfa-issue-14 (backend work)

# List all worktrees
git worktree list

# Remove worktree after PR is merged
git worktree remove ../arfa-issue-13
```

**Benefits of Worktrees:**
- ✅ Work on multiple features simultaneously without conflicts
- ✅ Each worktree has its own branch and working directory
- ✅ No need to stash/commit incomplete work when switching tasks
- ✅ Perfect for parallel agent work (frontend + backend agents)
- ✅ Cleaner workflow than juggling branches in one directory

**When to Use Worktrees:**
- Multiple agents working in parallel (e.g., frontend + backend)
- Long-running features that need periodic updates from main
- Testing changes against different branches
- Code review while continuing other work

### Standard Workflow for Every Task

**⚠️ STEP 1: Update Project Status FIRST - MANDATORY!**

```bash
# This MUST be the FIRST command you run when starting ANY task
./scripts/update-project-status.sh --issue <issue-number> --status "In Progress"

# Example:
./scripts/update-project-status.sh --issue 13 --status "In Progress"
```

**WHY THIS IS CRITICAL:**
- ✅ Immediate visibility that work has started
- ✅ Prevents duplicate work (others see it's "In Progress")
- ✅ Required for accurate project tracking
- ✅ Shows real-time progress to stakeholders

**If you skip this step, the task is considered NOT STARTED.**

---

2. ✅ Create feature branch (with optional worktree for parallel work)
   ```bash
   # Option A: Traditional branch (single task)
   git checkout main && git pull origin main
   git checkout -b <type>/<short-description>

   # Option B: Worktree (parallel tasks)
   git worktree add ../arfa-<issue-number> -b <type>/<short-description>
   cd ../arfa-<issue-number>
   ```

3. ✅ Verify project status was updated (double-check)
   ```bash
   # Confirm the issue is now "In Progress" on the board
   gh api graphql -f query='
     mutation($projectId: ID!, $itemId: ID!, $fieldId: ID!, $optionId: String!) {
       updateProjectV2ItemFieldValue(input: {
         projectId: $projectId
         itemId: $itemId
         fieldId: $fieldId
         value: {singleSelectOptionId: $optionId}
       }) {
         projectV2Item { id }
       }
     }
   ' -f projectId="<project-id>" -f itemId="<item-id>" -f fieldId="<status-field-id>" -f optionId="<in-progress-option-id>"

   # Or use helper script (if available)
   ./scripts/update-project-status.sh --issue <issue-number> --status "In Progress"
   ```

4. ✅ Implement changes following TDD
   - Write failing tests FIRST
   - Implement code to pass tests
   - Refactor with tests passing

5. ✅ Run tests locally
   ```bash
   # Backend
   make test

   # Frontend
   npm run type-check && npm run lint && npm run build && npm run test:e2e
   ```

6. ✅ Commit with descriptive message
   ```bash
   git add .
   git commit -m "$(cat <<'EOF'
   <type>: <short summary> (#<issue-number>)

   <detailed description>

   ## <section title>
   - Bullet points for details

   Closes #<issue-number>

   🚀 Generated with [Claude Code](https://claude.com/claude-code)

   Co-Authored-By: Claude <noreply@anthropic.com>
   EOF
   )"
   ```

7. ✅ Push to remote
   ```bash
   git push -u origin <branch-name>
   ```

8. ✅ Create Pull Request
   ```bash
   gh pr create \
     --title "<type>: <Title> (#<issue-number>)" \
     --body "$(cat <<'EOF'
   ## Summary
   Brief description of changes

   ## Changes
   - Change 1
   - Change 2

   ## Testing
   - Test 1
   - Test 2

   Closes #<issue>
   EOF
   )" \
     --base main \
     --head <branch-name>
   ```

9. ✅ Update GitHub Project status to "In Review"
   ```bash
   # After PR is created, move to In Review
   ./scripts/update-project-status.sh --issue <issue-number> --status "In Review"
   ```

10. ✅ Wait for CI/CD checks to pass
    ```bash
    gh pr checks <PR-number> --watch --interval 10
    ```

11. ✅ Verify checks passed
    ```bash
    # Check if all checks passed
    gh pr checks <PR-number>
    ```

12. ✅ Report completion to user
   ```
   ✅ Task complete!

   - PR #XX created: <URL>
   - All CI/CD checks: ✅ PASSED
   - Ready for review and merge

   Next steps:
   - Review the PR
   - Merge when approved
   - Delete feature branch after merge
   ```

13. ✅ Cleanup worktree (if used)
    ```bash
    # After PR is merged, cleanup the worktree
    cd /path/to/main/repo
    git worktree remove ../arfa-<issue-number>
    git branch -d <branch-name>  # Delete local branch
    git push origin --delete <branch-name>  # Delete remote (if needed)
    ```

14. ✅ Update GitHub Project status to "Done" (after PR merge)
    ```bash
    # After PR is successfully merged
    ./scripts/update-project-status.sh --issue <issue-number> --status "Done"
    ```

### GitHub Project Status Lifecycle

**CRITICAL: Update project board status at EVERY stage**

| Stage | Status | When to Update |
|-------|--------|----------------|
| Start work | **Backlog → In Progress** | Immediately when starting implementation |
| Create PR | **In Progress → In Review** | After PR is created and pushed |
| PR merged | **In Review → Done** | After PR is successfully merged |

**Status Update Commands:**

```bash
# Get project and field IDs (one-time lookup)
gh api graphql -f query='{ repository(owner: "sergei-rastrigin", name: "arfa") { projectsV2(first: 5) { nodes { id title } } } }'

# Update status (replace IDs with actual values)
./scripts/update-project-status.sh --issue <issue-number> --status "In Progress"
./scripts/update-project-status.sh --issue <issue-number> --status "In Review"
./scripts/update-project-status.sh --issue <issue-number> --status "Done"
```

**Why This Matters:**
- ✅ Provides real-time visibility into what agents are working on
- ✅ Prevents duplicate work (others see issue is "In Progress")
- ✅ Accurate project tracking and velocity metrics
- ✅ Clear handoff points (In Review = ready for merge)

### DO NOT Skip Any Steps

**NEVER:**
- ❌ Push directly to `main`
- ❌ Commit without running tests
- ❌ Create PR with failing tests
- ❌ Skip the PR process
- ❌ Merge before checks pass
- ❌ Use vague commit messages
- ❌ **Forget to update GitHub Project status**

**ALWAYS:**
- ✅ Create feature branch
- ✅ **Update status to "In Progress" when starting**
- ✅ Follow TDD (tests first!)
- ✅ Run tests before committing
- ✅ Write descriptive commit messages
- ✅ Link commits/PRs to issues
- ✅ **Update status to "In Review" after creating PR**
- ✅ Wait for CI/CD checks
- ✅ Report completion with PR details
- ✅ **Update status to "Done" after PR merge**

### Reference Documentation

For complete details, see:
- **[docs/DEV_WORKFLOW.md](../DEV_WORKFLOW.md)** - Complete workflow guide
- **[CLAUDE.md](../CLAUDE.md)** - Development workflow section

### Branch Naming Convention

```
<type>/<short-description>

Types:
- feature/  - New features
- fix/      - Bug fixes
- docs/     - Documentation only
- refactor/ - Code refactoring
- test/     - Test additions/changes
- chore/    - Maintenance tasks

Examples:
- feature/web-ui-foundation
- fix/employee-integration-test
- docs/api-endpoints
```

### Commit Message Types

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `style:` - Formatting, missing semicolons, etc
- `refactor:` - Code restructuring
- `test:` - Adding tests
- `chore:` - Maintenance

### Task Completion Checklist

Before reporting task completion, verify:

- [ ] Feature branch created from `main`
- [ ] All tests written (TDD)
- [ ] All tests passing locally
- [ ] Code committed with descriptive message
- [ ] Changes pushed to remote
- [ ] Pull request created with full description
- [ ] CI/CD checks completed
- [ ] All checks passed (green ✅)
- [ ] GitHub issue linked in PR
- [ ] User notified with PR URL and status

**Only mark task as complete after ALL steps above are verified.**
```

---

## Backend-Specific Configuration

For `~/.claude/agents/go-backend-developer.md`, also add:

```markdown
### Backend Testing Requirements

Before creating PR, verify:

```bash
# Run all tests
make test

# Check coverage (target: 85%+)
make test-coverage

# Run integration tests
make test-integration

# Verify build
make build
```

**All tests MUST pass before creating PR.**
```

---

## Frontend-Specific Configuration

For `~/.claude/agents/frontend-developer.md`, also add:

```markdown
### Frontend Testing Requirements

Before creating PR, verify:

```bash
# Type check
npm run type-check

# Lint
npm run lint

# Build
npm run build

# E2E tests (if applicable)
npm run test:e2e
```

**All checks MUST pass before creating PR.**
```

---

## How to Apply These Changes

### Option 1: Manual Update

1. Open `~/.claude/agents/go-backend-developer.md`
2. Add the "Mandatory PR Workflow" section
3. Add the backend-specific testing requirements
4. Save the file

5. Open `~/.claude/agents/frontend-developer.md`
6. Add the "Mandatory PR Workflow" section
7. Add the frontend-specific testing requirements
8. Save the file

### Option 2: Verify Agent Behavior

After updating the configurations, test by asking each agent to:

1. Implement a small feature
2. Verify the agent automatically:
   - Creates a feature branch
   - Writes tests first (TDD)
   - Commits with proper message format
   - Pushes to remote
   - Creates a PR
   - Waits for CI/CD checks
   - Reports completion with PR URL

---

## Benefits

**Consistency:**
- Every change follows the same process
- No ad-hoc commits to main
- Predictable workflow for all agents

**Quality Gates:**
- All tests must pass before PR
- CI/CD validates changes
- Code review before merge

**Visibility:**
- All changes tracked via PRs
- Clear history in GitHub
- Easy to review and rollback

**Automation:**
- Agents handle entire workflow
- No manual steps required
- Fast feedback on failures

---

## Verification

After updating agent configurations, verify by:

1. Asking `go-backend-developer` to implement a small backend feature
2. Asking `frontend-developer` to implement a small frontend feature
3. Confirming both agents:
   - Created feature branches
   - Followed TDD
   - Created PRs
   - Waited for checks
   - Reported completion properly

If any agent skips steps, review and update their configuration.

---

## See Also

- **[docs/DEV_WORKFLOW.md](./DEV_WORKFLOW.md)** - Complete workflow guide
- **[docs/PR_REVIEWER_AGENT.md](./PR_REVIEWER_AGENT.md)** - PR review, merge, and cleanup agent
- **[CLAUDE.md](../CLAUDE.md)** - Development workflow section
- **[docs/TESTING.md](./TESTING.md)** - TDD methodology

---

## Additional Agent: PR Reviewer

**Purpose:** Automated PR review, conflict resolution, merge, and cleanup

**Configuration File:** `~/.claude/agents/pr-reviewer.md`

**See [docs/PR_REVIEWER_AGENT.md](./PR_REVIEWER_AGENT.md) for complete configuration.**

**Usage:**
```
User: "Review and merge PR #20"
```

**What It Does:**
1. Reviews code changes
2. Resolves merge conflicts (if any)
3. Waits for CI/CD checks to pass
4. Merges PR to main
5. Deletes feature branch
6. Updates GitHub Project status to "Done"
7. Cleans up worktree

**This completes the full development cycle:**
- `go-backend-developer` / `frontend-developer` → Implement & Create PR
- `pr-reviewer` → Review, Merge, Clean up

**Full automation from task start to completion!**

---

🚀 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
