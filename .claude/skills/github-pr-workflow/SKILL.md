---
name: github-pr-workflow
description: Standardized pull request workflows with CI monitoring, review automation, and status management. Use for creating PRs, monitoring CI checks, handling reviews, and merging. Enforces quality gates (CI must pass) and auto-updates issue status. Critical for maintaining consistent PR workflows across all agents.
---

# GitHub PR Workflow Skill

Standardized pull request management with intelligent CI monitoring, automated status updates, and quality gates.

## When to Use This Skill

- Creating pull requests with proper formatting
- Monitoring CI checks and handling failures
- Managing PR reviews and approvals
- Merging PRs with appropriate strategies
- Auto-updating GitHub Issues and Projects based on PR status
- Enforcing quality gates (CI must pass before review)

## Core Capabilities

### 1. Create Pull Request

Create a well-formatted PR with automatic issue linking and project integration.

**Step 1: Ensure you're on a feature branch**
```bash
# Should be on branch like: issue-123-feature-name
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" = "main" ]; then
  echo "❌ Error: Cannot create PR from main branch"
  exit 1
fi
```

**Step 2: Get issue number from branch name**
```bash
# Extract issue number from branch name (e.g., issue-123-feature → 123)
ISSUE_NUM=$(echo "$CURRENT_BRANCH" | grep -oE '[0-9]+' | head -1)

if [ -z "$ISSUE_NUM" ]; then
  echo "⚠️ Warning: No issue number found in branch name"
fi
```

**Step 3: Get issue details**
```bash
ISSUE_TITLE=$(gh issue view $ISSUE_NUM --json title -q .title)
ISSUE_LABELS=$(gh issue view $ISSUE_NUM --json labels -q '.labels[].name' | grep -E '^area/' | head -1)
```

**Step 4: Create PR with template**
```bash
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

## Screenshots
[If UI changes, add screenshots]

## Additional Notes
[Any important context for reviewers]

---

Closes #${ISSUE_NUM}

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

**Step 5: Capture PR number**
```bash
PR_NUM=$(gh pr view --json number -q .number)
echo "✅ Created PR #${PR_NUM}"
```

**Automatic Behaviors:**
- PR title follows format: `type: Description (#issue)`
- Closes issue automatically via `Closes #123`
- Inherits area labels from issue
- Includes Claude Code attribution

**PR Title Conventions:**
- `feat:` - New feature
- `fix:` - Bug fix
- `refactor:` - Code refactoring
- `chore:` - Maintenance tasks
- `docs:` - Documentation
- `test:` - Test additions/changes

### 2. Monitor CI Checks

Watch CI checks and only proceed when all checks pass.

**Usage:**
```bash
PR_NUM=45  # Your PR number

# Watch CI checks with polling
echo "🔍 Watching CI checks for PR #${PR_NUM}..."
gh pr checks $PR_NUM --watch --interval 10
```

**Check CI Status:**
```bash
# Count failed/cancelled checks
CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE" or .state == "CANCELLED")) | length')

if [ "$CI_FAILED" -eq 0 ]; then
  echo "✅ All CI checks passed!"
  EXIT_CODE=0
else
  echo "❌ Some CI checks failed:"
  gh pr checks $PR_NUM
  EXIT_CODE=1
fi
```

**Quality Gate Enforcement:**
```bash
if [ $EXIT_CODE -eq 0 ]; then
  # Update issue status to "In Review"
  ./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Review"

  # Add success comment to issue
  gh issue comment $ISSUE_NUM --body "✅ PR #${PR_NUM} created and all CI checks passing. Ready for review."
else
  # Keep status as "In Progress", add failure comment
  gh issue comment $ISSUE_NUM --body "⚠️ PR #${PR_NUM} created but CI checks failed. Investigating..."

  # DO NOT update status to "In Review"
  echo "❌ Fix CI failures before moving to review"
fi
```

**Critical Rule:**
> **NEVER update issue status to "In Review" until ALL CI checks pass!**

### 3. Analyze CI Failures

Intelligently diagnose CI failures and suggest fixes.

**Fetch Failure Details:**
```bash
# Get failed check details
gh pr checks $PR_NUM --json name,state,detailsUrl \
  -q '.[] | select(.state == "FAILURE") | "\(.name): \(.detailsUrl)"'
```

**View Workflow Logs:**
```bash
# Get latest workflow run ID
RUN_ID=$(gh run list --limit 1 --json databaseId -q '.[0].databaseId')

# View logs
gh run view $RUN_ID --log
```

**Common Failure Patterns:**

| Failure Type | Pattern | Likely Cause | Fix |
|-------------|---------|--------------|-----|
| Build failure | `go build` error | Compilation error | Check syntax, imports |
| Test failure | `FAIL:` in output | Test assertion failed | Fix test or implementation |
| Lint failure | `golangci-lint` error | Code style violation | Run `golangci-lint run --fix` |
| Integration test | `testcontainers` error | Docker/DB issue | Check container setup |
| E2E test | Playwright error | UI interaction failed | Check selectors, timing |

**Auto-Retry Failed Checks:**
```bash
# Re-run failed jobs
gh run rerun $RUN_ID --failed
```

### 4. Update PR After Feedback

Make changes based on review feedback while maintaining CI quality gate.

**Workflow:**
```bash
# 1. Make changes based on review
git add .
git commit -m "fix: Address review feedback"
git push

# 2. Wait for CI again
gh pr checks $PR_NUM --watch --interval 10

# 3. Comment on PR
gh pr comment $PR_NUM --body "✅ Addressed review feedback. CI checks passing."
```

**Critical:** After every push, CI must pass again before merge!

### 5. Merge Pull Request

Merge PR only after CI passes and approvals received.

**Pre-Merge Checklist:**
```bash
# 1. Verify CI passed
CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE")) | length')
if [ "$CI_FAILED" -gt 0 ]; then
  echo "❌ Cannot merge: CI checks failed"
  exit 1
fi

# 2. Verify approvals (optional - depends on repo settings)
APPROVALS=$(gh pr view $PR_NUM --json reviewDecision -q .reviewDecision)
if [ "$APPROVALS" != "APPROVED" ]; then
  echo "⚠️ Warning: PR not yet approved"
fi

# 3. Check if PR is up to date with base branch
gh pr view $PR_NUM --json mergeable -q .mergeable
```

**Merge Strategies:**

**Squash Merge (Recommended for feature branches):**
```bash
gh pr merge $PR_NUM --squash --delete-branch
```
- Combines all commits into one
- Cleaner git history
- Preserves PR number in commit message

**Merge Commit (For larger features with meaningful commit history):**
```bash
gh pr merge $PR_NUM --merge --delete-branch
```
- Preserves all commits
- Creates merge commit
- Use when commit history tells a story

**Rebase (For linear history, use cautiously):**
```bash
gh pr merge $PR_NUM --rebase --delete-branch
```
- Rewrites commit history
- Linear history
- Avoid if PR has been reviewed (changes SHAs)

**Auto-Update After Merge:**
```bash
# Issue status auto-updates to "Done" when PR closes issue
# No manual action needed if PR description has "Closes #123"

# Verify issue closed
gh issue view $ISSUE_NUM --json state -q .state
# Should output: CLOSED
```

### 6. Handle Merge Conflicts

Resolve conflicts when PR branch is behind base branch.

**Update PR Branch:**
```bash
# Option 1: Merge base into feature (preserves history)
git checkout issue-123-feature
git fetch origin
git merge origin/main
# Resolve conflicts manually
git add .
git commit -m "chore: Merge main into feature branch"
git push

# Option 2: Rebase onto base (linear history, rewrites)
git checkout issue-123-feature
git fetch origin
git rebase origin/main
# Resolve conflicts manually
git add .
git rebase --continue
git push --force-with-lease

# Option 3: Use GitHub's "Update branch" button (safest)
gh pr update-branch $PR_NUM
```

**After Conflict Resolution:**
```bash
# Wait for CI again
gh pr checks $PR_NUM --watch --interval 10

# Verify no conflicts
gh pr view $PR_NUM --json mergeable -q .mergeable
```

### 7. Cancel/Close PR

Cancel a PR if approach changes or issue is resolved differently.

**Close PR without merging:**
```bash
gh pr close $PR_NUM --comment "Closing: [reason for closing]"

# Update issue status back to "In Progress" or "Blocked"
./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Progress"

# Add comment to issue
gh issue comment $ISSUE_NUM --body "PR #${PR_NUM} closed. [Explanation]"
```

**Delete branch:**
```bash
# Delete remote branch
git push origin --delete issue-123-feature

# Delete local branch
git checkout main
git branch -D issue-123-feature
```

## Complete Workflow Example

### End-to-End: From Feature to Merged PR

```bash
# ============================================
# STEP 1: Start Working on Issue
# ============================================
ISSUE_NUM=89
git checkout main
git pull
git checkout -b issue-${ISSUE_NUM}-new-feature

./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Progress"
gh issue edit $ISSUE_NUM --add-assignee "@me"

# ============================================
# STEP 2: Implement Feature (TDD)
# ============================================
# Write tests, implement code, ensure tests pass locally
make test

# ============================================
# STEP 3: Create Pull Request
# ============================================
ISSUE_TITLE=$(gh issue view $ISSUE_NUM --json title -q .title)
ISSUE_LABELS=$(gh issue view $ISSUE_NUM --json labels -q '.labels[].name' | grep -E '^area/' | head -1)

gh pr create \
  --title "feat: ${ISSUE_TITLE} (#${ISSUE_NUM})" \
  --label "$ISSUE_LABELS" \
  --body "$(cat <<EOF
## Summary
Implements ${ISSUE_TITLE}.

## Changes
- Added new feature X
- Updated service layer
- Added comprehensive tests

## Testing
- [x] Unit tests passing (95% coverage)
- [x] Integration tests passing
- [x] Manual testing complete

Closes #${ISSUE_NUM}
EOF
)"

PR_NUM=$(gh pr view --json number -q .number)
echo "✅ Created PR #${PR_NUM}"

# ============================================
# STEP 4: Wait for CI Checks
# ============================================
echo "🔍 Waiting for CI checks..."
gh pr checks $PR_NUM --watch --interval 10

# ============================================
# STEP 5: Check CI Status
# ============================================
CI_FAILED=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE" or .state == "CANCELLED")) | length')

if [ "$CI_FAILED" -eq 0 ]; then
  echo "✅ All CI checks passed!"

  # Update issue status to In Review
  ./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Review"

  # Add comment to issue
  gh issue comment $ISSUE_NUM --body "✅ PR #${PR_NUM} created and all CI checks passing. Ready for review."

  echo "✅ PR ready for review!"
else
  echo "❌ Some CI checks failed. Please investigate:"
  gh pr checks $PR_NUM

  # Add comment but DON'T update status
  gh issue comment $ISSUE_NUM --body "⚠️ PR #${PR_NUM} created but CI checks failed. Investigating..."

  echo "❌ Fix failures and push again"
  exit 1
fi

# ============================================
# STEP 6: Handle Review Feedback (if needed)
# ============================================
# Make changes based on review
git add .
git commit -m "fix: Address review feedback"
git push

# Wait for CI again
gh pr checks $PR_NUM --watch --interval 10

# ============================================
# STEP 7: Merge PR
# ============================================
# After approval and CI passing
gh pr merge $PR_NUM --squash --delete-branch

# Verify issue closed and status updated
gh issue view $ISSUE_NUM --json state -q .state
# Should output: CLOSED

echo "✅ Feature complete! Issue #${ISSUE_NUM} closed."
```

## Best Practices

### 1. PR Creation
- ✅ Always create PR from feature branch, never from `main`
- ✅ Use descriptive PR titles following convention (feat/fix/etc)
- ✅ Include `Closes #123` to auto-close issue
- ✅ Inherit area labels from linked issue
- ✅ Fill out PR template completely

### 2. CI Monitoring
- ✅ **ALWAYS** wait for CI checks before marking "In Review"
- ✅ Use `--watch` flag to poll CI status automatically
- ✅ Investigate failures immediately, don't merge with failures
- ✅ Re-run flaky tests if needed
- ✅ Comment on PR/issue with CI status

### 3. Review Process
- ✅ Request reviews from appropriate team members
- ✅ Address all review feedback promptly
- ✅ Run CI again after every change
- ✅ Comment when ready for re-review

### 4. Merging
- ✅ Verify CI passed one final time before merge
- ✅ Use squash merge for most feature PRs
- ✅ Delete branch after merge (`--delete-branch`)
- ✅ Verify issue auto-closed after merge

### 5. Status Management
- ✅ `In Progress` → Create PR → Wait for CI
- ✅ CI Pass → Update to `In Review`
- ✅ CI Fail → Stay `In Progress`, fix issues
- ✅ PR Merged → Auto-update to `Done`

## Quality Gates

### Mandatory Gates (MUST Pass)
1. **All CI checks pass** - No exceptions
2. **PR links to issue** - Use `Closes #123`
3. **Area label applied** - For project tracking
4. **Tests included** - For all code changes

### Recommended Gates
1. **Code review approval** - At least 1 reviewer
2. **Coverage maintained** - No coverage drops
3. **Documentation updated** - For user-facing changes
4. **Changelog updated** - For notable changes

## Common Scenarios

### Scenario 1: CI Fails After PR Creation
```bash
# Don't panic! This is normal.
# 1. View the failure
gh pr checks $PR_NUM

# 2. Fix the issue locally
git add .
git commit -m "fix: Resolve CI failure"
git push

# 3. Wait for CI again
gh pr checks $PR_NUM --watch

# 4. Only update status when CI passes
```

### Scenario 2: Review Requests Changes
```bash
# 1. Make requested changes
git add .
git commit -m "fix: Address review feedback"
git push

# 2. Wait for CI
gh pr checks $PR_NUM --watch

# 3. Comment on PR
gh pr comment $PR_NUM --body "✅ Changes made. Please review again."

# 4. Re-request review
gh pr review $PR_NUM --comment --body "@reviewer ready for re-review"
```

### Scenario 3: PR Has Merge Conflicts
```bash
# 1. Update branch with main
gh pr update-branch $PR_NUM
# OR manually:
# git fetch origin && git merge origin/main

# 2. Resolve conflicts
git add .
git commit -m "chore: Resolve merge conflicts"
git push

# 3. Wait for CI
gh pr checks $PR_NUM --watch
```

### Scenario 4: Need to Make Emergency Fix
```bash
# For critical P0 bugs, still follow CI workflow but expedite

# 1. Create PR as normal
gh pr create --title "fix: Critical bug (#123)" ...

# 2. Monitor CI closely
gh pr checks $PR_NUM --watch --interval 5  # Poll every 5s

# 3. Request immediate review
gh pr review $PR_NUM --request-reviewer @tech-lead

# 4. Merge as soon as CI passes and approved
gh pr merge $PR_NUM --squash --delete-branch
```

## Integration with Other Skills

### Works With:
- **github-task-manager** - PR creation updates issue status
- **github-ci-monitor** - Detailed CI failure analysis
- **github-project-manager** - Auto-update project fields

### Workflow Example:
```bash
# 1. Use github-task-manager to start work
./scripts/update-project-status.sh --issue 45 --status "In Progress"

# 2. Implement feature...

# 3. Use github-pr-workflow to create PR
# (This skill)

# 4. Use github-ci-monitor to diagnose failures
# (Future skill)

# 5. Merge PR
# github-task-manager auto-updates to "Done"
```

## Troubleshooting

### Issue: PR doesn't auto-close issue
**Solution:** Check PR description has exact text `Closes #123` (case-insensitive)

### Issue: CI checks never finish
**Solution:** Check workflow file syntax, view logs with `gh run view`

### Issue: Can't merge due to branch protection
**Solution:** Ensure required approvals received, all checks passed

### Issue: Deleted branch still shows in git
**Solution:** Run `git fetch --prune` to clean up remote tracking branches

## Scripts Used

### update-project-status.sh
```bash
./scripts/update-project-status.sh --issue ISSUE_NUM --status "STATUS"
```
- Used to update GitHub Projects status
- Called automatically during PR workflow
- Ensures status stays in sync

## Templates

See `templates/` directory:
- `pr-template.md` - Standard PR template
- `pr-bugfix.md` - Bug fix PR template
- `pr-hotfix.md` - Emergency hotfix template

## Examples

See `examples/pr-workflows.md` for complete real-world examples.

---

**This skill ensures consistent, high-quality PR workflows with mandatory CI quality gates and automatic status management.**
