---
name: github-dev-workflow
description: Complete development workflow from task start to completion. Standardizes git worktrees, branch naming, status updates, PR creation, CI monitoring, and merge process. Use when starting work on a task, creating PRs, or merging completed work. Ensures consistent workflow across all agents with built-in quality gates.
---

# GitHub Development Workflow Skill

Complete, standardized development workflow from task assignment to merge, ensuring consistency across all AI agents.

## When to Use This Skill

- **Starting work** on a GitHub issue
- **Creating a PR** after implementation
- **Merging a PR** when approved and CI passes
- Ensuring consistent workflow across all development tasks
- Enforcing quality gates (CI must pass, proper status updates)

## Overview

This skill defines three complete workflows that agents MUST follow:

1. **Start Task Workflow** - Set up environment, update status
2. **Create PR Workflow** - Commit, push, create PR, wait for CI
3. **Merge PR Workflow** - Resolve conflicts, verify checks, merge, cleanup

## Workflow 1: Start Task

**Trigger:** Agent is asked to work on a GitHub issue

**Steps:**

### 1. Update Task Status to "In Progress"
```bash
ISSUE_NUM=47  # Your issue number

./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Progress"
```

### 2. Self-Assign the Task
```bash
gh issue edit $ISSUE_NUM --add-assignee "@me"
```

### 3. Create Git Worktree with Standard Branch Name
```bash
# Branch naming convention: feature/{num}-{short-description}
# Example: feature/47-dev-workflow-skill

BRANCH_NAME="feature/${ISSUE_NUM}-short-description"

# Create worktree in parallel directory
git worktree add ../$(basename $(pwd))-issue-${ISSUE_NUM} -b $BRANCH_NAME

# Navigate to worktree
cd ../$(basename $(pwd))-issue-${ISSUE_NUM}
```

**Why Worktrees?**
- Work on multiple issues simultaneously without branch switching
- Clean separation of work
- No risk of accidental commits to wrong branch
- Easy cleanup on completion

### 4. Verify Setup
```bash
# Confirm you're on the right branch
git branch --show-current
# Output: feature/47-short-description

# Confirm clean working directory
git status
# Output: On branch feature/47-short-description, nothing to commit

echo "✅ Ready to start work on issue #${ISSUE_NUM}"
```

### 5. Begin Implementation (TDD)
```bash
# Follow Test-Driven Development:
# 1. Write failing tests
# 2. Implement minimal code to pass
# 3. Refactor
# 4. Repeat
```

---

## Workflow 2: Create PR

**Trigger:** Implementation complete, tests passing locally

**Steps:**

### 1. Commit Changes
```bash
ISSUE_NUM=47

# Stage all changes
git add .

# Commit with proper format
git commit -m "feat: Implement feature description

Detailed description of what was implemented.

- Change 1
- Change 2
- Change 3

Closes #${ISSUE_NUM}

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
```

**Commit Message Format:**
- `type: Description` - Title line (no issue number in commit)
- Blank line
- Detailed explanation
- Blank line
- `Closes #issue` - Auto-close on merge
- Attribution

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `refactor:` - Code refactoring
- `chore:` - Maintenance
- `docs:` - Documentation
- `test:` - Tests

### 2. Push to Remote
```bash
# Push and set upstream
git push -u origin feature/${ISSUE_NUM}-short-description
```

### 3. Create Pull Request (with Issue Number in Title - REQUIRED!)
```bash
ISSUE_TITLE=$(gh issue view $ISSUE_NUM --json title -q .title)
ISSUE_LABELS=$(gh issue view $ISSUE_NUM --json labels -q '.labels[].name' | grep -E '^area/' | head -1)

# CRITICAL: Include (#ISSUE_NUM) in title for automatic status updates
gh pr create \
  --title "feat: ${ISSUE_TITLE} (#${ISSUE_NUM})" \
  --label "$ISSUE_LABELS" \
  --body "$(cat <<EOF
## Summary
Implements ${ISSUE_TITLE} as described in #${ISSUE_NUM}.

## Changes
- [List key changes]
- [Be specific and concise]

## Testing
- [ ] Unit tests passing
- [ ] Integration tests passing
- [ ] Manual testing complete
- [ ] No breaking changes

## Additional Notes
[Any important context for reviewers]

---

Closes #${ISSUE_NUM}

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"

PR_NUM=$(gh pr view --json number -q .number)
echo "✅ Created PR #${PR_NUM}"
```

### 4. Monitor CI Checks (CRITICAL!)
```bash
echo "🔍 Waiting for CI checks to complete..."
gh pr checks $PR_NUM --watch --interval 10
```

**This step is MANDATORY. Never skip it!**

### 5. Check CI Results (Automation Handles Status Updates)
```bash
CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE" or .state == "CANCELLED")) | length')

if [ "$CI_FAILED" -eq 0 ]; then
  echo "✅ All CI checks passed!"
  echo ""
  echo "📋 GitHub Actions will automatically:"
  echo "  - Update issue #${ISSUE_NUM} status to 'In Review'"
  echo "  - Add comment linking PR #${PR_NUM} to issue"
  echo "  - Close issue when PR is merged"
  echo "  - Delete branch after merge"
  echo ""
  echo "✅ Task ready for review. PR: #${PR_NUM}"
else
  echo "❌ Some CI checks failed. Please investigate:"
  gh pr checks $PR_NUM

  echo "❌ Fix CI failures and push again"
  exit 1
fi
```

**Quality Gate: CI must pass before PR can be merged!**
**Automation: GitHub Actions handles all status updates based on PR title (#ISSUE_NUM)**

### 6. Return to Main Workspace (Optional)
```bash
# Navigate back to main workspace
cd ../$(basename $(pwd) | sed 's/-issue-[0-9]*//')

# Worktree remains for potential fixes or future work
# No manual status updates needed - automation handles it!
```

---

## Workflow 3: Merge PR

**Trigger:** PR approved, CI passing, ready to merge

**Steps:**

### 1. Verify All Pre-Merge Conditions
```bash
ISSUE_NUM=47
PR_NUM=50  # Your PR number

# Check CI status
CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')
if [ "$CI_FAILED" -gt 0 ]; then
  echo "❌ Cannot merge: CI checks failed"
  gh pr checks $PR_NUM
  exit 1
fi

# Check if mergeable (no conflicts)
MERGEABLE=$(gh pr view $PR_NUM --json mergeable -q .mergeable)
if [ "$MERGEABLE" != "MERGEABLE" ]; then
  echo "⚠️ Warning: PR has merge conflicts"
  # Continue to conflict resolution
fi

# Check approvals (optional, depends on repo settings)
REVIEW_DECISION=$(gh pr view $PR_NUM --json reviewDecision -q .reviewDecision)
if [ "$REVIEW_DECISION" != "APPROVED" ] && [ "$REVIEW_DECISION" != "" ]; then
  echo "⚠️ Warning: PR not yet approved (status: $REVIEW_DECISION)"
fi

echo "✅ Pre-merge checks complete"
```

### 2. Resolve Merge Conflicts (if any)
```bash
if [ "$MERGEABLE" = "CONFLICTING" ]; then
  echo "🔧 Resolving merge conflicts..."

  # Navigate to worktree
  cd ../$(basename $(pwd))-issue-${ISSUE_NUM}

  # Fetch latest changes
  git fetch origin

  # Option 1: Merge main into feature (preserves history)
  git merge origin/main
  # Resolve conflicts manually
  git add .
  git commit -m "chore: Merge main into feature branch"
  git push

  # Option 2: Use GitHub's update branch feature
  # gh pr update-branch $PR_NUM

  # Wait for CI again after conflict resolution
  echo "🔍 Waiting for CI after conflict resolution..."
  gh pr checks $PR_NUM --watch --interval 10

  # Verify no more conflicts
  MERGEABLE=$(gh pr view $PR_NUM --json mergeable -q .mergeable)
  if [ "$MERGEABLE" != "MERGEABLE" ]; then
    echo "❌ Still has conflicts. Please resolve manually."
    exit 1
  fi

  echo "✅ Conflicts resolved, CI passing"
fi
```

### 3. Merge Pull Request
```bash
# Choose merge strategy based on PR
# - squash: Most feature PRs (clean history)
# - merge: Preserve commit history
# - rebase: Linear history (use cautiously)

gh pr merge $PR_NUM --squash --delete-branch

echo "✅ PR #${PR_NUM} merged!"
echo "📋 GitHub Actions will automatically:"
echo "  - Close issue #${ISSUE_NUM}"
echo "  - Update status to 'Done'"
echo "  - Delete feature branch"
```

**Merge Strategies:**
- **`--squash`** (Recommended): Combines all commits into one, clean history
- **`--merge`**: Creates merge commit, preserves all commits
- **`--rebase`**: Rewrites history, linear timeline

### 4. Verify Issue Closed (Automatic)
```bash
# Issue auto-closes due to "Closes #123" in PR + GitHub Actions
ISSUE_STATE=$(gh issue view $ISSUE_NUM --json state -q .state)

if [ "$ISSUE_STATE" = "CLOSED" ]; then
  echo "✅ Issue #${ISSUE_NUM} automatically closed by GitHub Actions"
else
  echo "⚠️ Issue not yet closed. GitHub Actions may be processing..."
  echo "   Check workflow status: gh run list --limit 5"
fi
```

### 5. Clean Up Worktree
```bash
# Navigate back to main workspace
cd ../$(basename $(pwd) | sed 's/-issue-[0-9]*//')

# Remove worktree
WORKTREE_PATH="../$(basename $(pwd))-issue-${ISSUE_NUM}"
git worktree remove $WORKTREE_PATH

# Verify worktree removed
git worktree list

echo "✅ Worktree cleaned up"
```

### 6. Confirm Status is "Done" (Automatic)
```bash
# GitHub Actions automatically updates status to "Done"
echo "✅ Task complete! Issue #${ISSUE_NUM} closed and merged."
echo "📋 GitHub Actions handled:"
echo "  - Issue closure"
echo "  - Status update to 'Done'"
echo "  - Branch deletion"
```

---

## Complete End-to-End Example

### Scenario: Implement a New Feature

```bash
# ==========================================
# WORKFLOW 1: START TASK
# ==========================================
ISSUE_NUM=47

# 1. Update status
./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Progress"

# 2. Self-assign
gh issue edit $ISSUE_NUM --add-assignee "@me"

# 3. Create worktree
git worktree add ../ubik-issue-47 -b feature/47-dev-workflow-skill
cd ../ubik-issue-47

# 4. Verify setup
git branch --show-current  # feature/47-dev-workflow-skill
pwd  # /Users/you/Projects/ubik-issue-47

# 5. Implement feature (TDD)
# Write tests, implement code, run tests locally
make test

# ==========================================
# WORKFLOW 2: CREATE PR
# ==========================================

# 1. Commit
git add .
git commit -m "feat: Create development workflow skill

Implements comprehensive workflow for all development tasks.

- Start task workflow (status, worktree, setup)
- Create PR workflow (commit, push, CI)
- Merge PR workflow (conflicts, merge, cleanup)

Closes #47

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# 2. Push
git push -u origin feature/47-dev-workflow-skill

# 3. Create PR
ISSUE_TITLE=$(gh issue view 47 --json title -q .title)
gh pr create \
  --title "feat: ${ISSUE_TITLE} (#47)" \
  --label "area/infra" \
  --body "..."

PR_NUM=$(gh pr view --json number -q .number)

# 4. Wait for CI (MANDATORY!)
gh pr checks $PR_NUM --watch --interval 10

# 5. Check CI (automation handles status updates)
CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')
if [ "$CI_FAILED" -eq 0 ]; then
  echo "✅ All CI checks passed!"
  echo "📋 GitHub Actions will automatically:"
  echo "  - Update issue status to 'In Review'"
  echo "  - Close issue when PR is merged"
  echo "  - Delete branch after merge"
else
  echo "❌ Fix CI failures first"
  exit 1
fi

# 6. Return to main workspace
cd ../ubik-enterprise

# ==========================================
# WORKFLOW 3: MERGE PR (after approval)
# ==========================================

# 1. Verify pre-merge conditions
CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')
MERGEABLE=$(gh pr view $PR_NUM --json mergeable -q .mergeable)

if [ "$CI_FAILED" -eq 0 ] && [ "$MERGEABLE" = "MERGEABLE" ]; then
  echo "✅ Ready to merge"
fi

# 2. Merge PR
gh pr merge $PR_NUM --squash --delete-branch

# 3. Verify issue closed (automatic via GitHub Actions)
gh issue view 47 --json state -q .state  # CLOSED

# 4. Clean up worktree
cd ../ubik-enterprise
git worktree remove ../ubik-issue-47

# 5. Status automatically updated to "Done" by GitHub Actions
echo "✅ Task complete! Issue #47 closed and merged."
echo "📋 GitHub Actions handled:"
echo "  - Issue closure"
echo "  - Status update to 'Done'"
echo "  - Branch deletion"
```

---

## Quality Gates (MANDATORY)

### Gate 1: Status Updates (Automated)
- ✅ **MUST** manually update to "In Progress" when starting work
- ✅ **AUTOMATIC** update to "In Review" when CI passes (via GitHub Actions)
- ✅ **AUTOMATIC** update to "Done" after merge (via GitHub Actions)

### Gate 2: CI Checks (Critical!)
- ✅ **MUST** wait for ALL CI checks to complete
- ✅ **MUST NOT** proceed if CI fails (automation won't update status)
- ✅ **MUST** re-run CI after fixing failures
- ✅ **MUST** wait for CI again after resolving conflicts

### Gate 3: Git Workflow
- ✅ **MUST** use git worktrees for parallel work
- ✅ **MUST** follow branch naming: `feature/{num}-{description}`
- ✅ **MUST** include `(#ISSUE_NUM)` in PR title for automation
- ✅ **MUST** include `Closes #issue` in PR description
- ✅ **MUST** clean up worktree after merge

### Gate 4: PR Quality
- ✅ **MUST** use proper commit message format
- ✅ **MUST** include issue number in PR title: `type: Title (#123)`
- ✅ **MUST** fill out PR template completely
- ✅ **MUST** inherit area labels from issue
- ✅ **MUST** include Claude Code attribution

---

## Branch Naming Convention

**Format:** `feature/{number}-{short-description}`

**Examples:**
- `feature/47-dev-workflow-skill`
- `feature/89-list-agents-endpoint`
- `feature/123-fix-auth-bug`

**Rules:**
- Always start with `feature/{num}`
- Use lowercase with dashes
- Keep description short (3-5 words max)
- Be descriptive but concise

---

## Worktree Management

### Why Worktrees?
- Work on multiple issues without branch switching
- Clean separation of work
- No accidental commits to wrong branch
- Parallel development

### Worktree Location
```bash
# Main workspace
/Users/you/Projects/ubik-enterprise

# Worktree for issue 47
/Users/you/Projects/ubik-issue-47
```

### List All Worktrees
```bash
git worktree list
```

### Remove Worktree
```bash
git worktree remove ../ubik-issue-47
```

### Prune Stale Worktrees
```bash
git worktree prune
```

---

## Troubleshooting

### Issue: Worktree already exists
```bash
# Remove existing worktree
git worktree remove ../ubik-issue-47

# Or force remove
git worktree remove --force ../ubik-issue-47
```

### Issue: CI checks never complete
```bash
# Check workflow status
gh run list --limit 5

# View logs for specific run
RUN_ID=$(gh run list --limit 1 --json databaseId -q '.[0].databaseId')
gh run view $RUN_ID --log
```

### Issue: Cannot update status
```bash
# Verify issue is in project
gh project item-list 3 --owner sergei-rastrigin | grep "#47"

# If not found, add to project
gh project item-add 3 --owner sergei-rastrigin --url "https://github.com/sergei-rastrigin/ubik-enterprise/issues/47"
```

### Issue: Merge conflicts
```bash
# Fetch latest
git fetch origin

# Merge main into feature
git merge origin/main

# Resolve conflicts manually
git add .
git commit -m "chore: Resolve merge conflicts"
git push

# Wait for CI again
gh pr checks $PR_NUM --watch
```

---

## Integration with Other Skills

### Works With:
- **github-task-manager** - Task creation, status updates
- **github-pr-workflow** - PR creation, CI monitoring
- **github-ci-monitor** (future) - Detailed CI analysis

### Workflow Chain:
```
github-task-manager → github-dev-workflow → github-pr-workflow → merge
      (create)           (implement)          (PR + CI)         (done)
```

---

## Agent Instructions

**When asked to start work on an issue:**
1. Use **Workflow 1: Start Task**
2. Follow ALL steps in order
3. Never skip status update
4. Always create worktree

**When implementation is complete:**
1. Use **Workflow 2: Create PR**
2. Follow ALL steps in order
3. **WAIT for CI** before updating status
4. Never skip CI checks

**When PR is approved:**
1. Use **Workflow 3: Merge PR**
2. Verify all conditions first
3. Resolve conflicts if needed
4. Clean up worktree after merge

---

## Success Metrics

- ✅ 100% of tasks follow standardized workflow
- ✅ Zero manual status updates (except "In Progress")
- ✅ Zero "In Review" with failing CI
- ✅ All PRs use worktrees
- ✅ All PRs have issue number in title for automation
- ✅ Clean git history (no leftover branches)
- ✅ Complete audit trail

---

## What Changed with Automation

**Before (Manual):**
- Manual status update scripts
- Manual issue comments
- Manual branch deletion
- Manual status label management

**After (Automated via GitHub Actions):**
- ✅ PR created with `(#123)` in title → Auto-updates status to "In Review"
- ✅ PR merged → Auto-closes issue and updates status to "Done"
- ✅ Branch auto-deleted after merge
- ✅ Automatic comment on issue with PR link

**Agent Responsibilities Now:**
- ✅ Create PR with issue number in title: `feat: Title (#123)`
- ✅ Wait for CI checks (still critical!)
- ✅ Merge when approved
- ❌ No manual status updates after PR creation
- ❌ No manual issue closure
- ❌ No manual branch deletion

---

**This skill ensures every agent follows the exact same development workflow from start to finish, with mandatory quality gates and automated status management.**
