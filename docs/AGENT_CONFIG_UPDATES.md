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

### Standard Workflow for Every Task

1. ✅ Create feature branch from `main`
   ```bash
   git checkout main && git pull origin main
   git checkout -b <type>/<short-description>
   ```

2. ✅ Implement changes following TDD
   - Write failing tests FIRST
   - Implement code to pass tests
   - Refactor with tests passing

3. ✅ Run tests locally
   ```bash
   # Backend
   make test

   # Frontend
   npm run type-check && npm run lint && npm run build && npm run test:e2e
   ```

4. ✅ Commit with descriptive message
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

5. ✅ Push to remote
   ```bash
   git push -u origin <branch-name>
   ```

6. ✅ Create Pull Request
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

7. ✅ Wait for CI/CD checks to pass
   ```bash
   gh pr checks <PR-number> --watch --interval 10
   ```

8. ✅ Verify checks passed
   ```bash
   # Check if all checks passed
   gh pr checks <PR-number>
   ```

9. ✅ Report completion to user
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

### DO NOT Skip Any Steps

**NEVER:**
- ❌ Push directly to `main`
- ❌ Commit without running tests
- ❌ Create PR with failing tests
- ❌ Skip the PR process
- ❌ Merge before checks pass
- ❌ Use vague commit messages

**ALWAYS:**
- ✅ Create feature branch
- ✅ Follow TDD (tests first!)
- ✅ Run tests before committing
- ✅ Write descriptive commit messages
- ✅ Link commits/PRs to issues
- ✅ Wait for CI/CD checks
- ✅ Report completion with PR details

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
- **[CLAUDE.md](../CLAUDE.md)** - Development workflow section
- **[docs/TESTING.md](./TESTING.md)** - TDD methodology

---

🚀 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
