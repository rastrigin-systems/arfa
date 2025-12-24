# Development Workflow - PR & Git Best Practices

**Last Updated:** 2025-11-07
**Applies to:** All developers (human and AI agents)

**🎉 NEW: Automated PR Workflow!**
- ✅ Branch protection enforces PR-based workflow (no direct commits to `main`)
- ✅ Issue status auto-updates when PR created (`status/in-review`)
- ✅ Issues auto-close when PR merges (`status/done`)
- ✅ Branches auto-delete after merge
- ✅ **PR title MUST include issue number:** `feat: Description (#123)`

---

## 📋 Standard Workflow for All Changes

**ALL code changes must follow this workflow:**

1. ✅ Create feature branch
2. ✅ Implement changes (following TDD)
3. ✅ Run tests locally
4. ✅ Commit with descriptive message
5. ✅ Push to remote
6. ✅ Create Pull Request
7. ✅ Wait for CI/CD checks
8. ✅ Review and merge
9. ✅ Delete feature branch

**No exceptions.** This applies to:
- Backend API changes (go-backend-developer agent)
- Frontend web changes (frontend-developer agent)
- CLI changes
- Documentation changes
- Database migrations

---

## 🔀 Git Branching Strategy

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
- refactor/auth-middleware
- test/e2e-dashboard
- chore/cleanup-unused-files
```

### Base Branch

- **Always branch from:** `main`
- **Always merge to:** `main`
- **Never push directly to:** `main`

---

## 💾 Commit Message Format

### Template

```
<type>: <short summary> (#<issue-number>)

<detailed description>

## <section title>
- Bullet points for details

Closes #<issue-number>

🚀 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

### Types

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation
- `style:` - Formatting, missing semicolons, etc
- `refactor:` - Code restructuring
- `test:` - Adding tests
- `chore:` - Maintenance

### Examples

**Good:**
```
feat: Add Web UI foundation with Next.js 14 and authentication (#12)

Implements Issue #12 - Web UI Foundation & Authentication

## Features
- Next.js 14 App Router with TypeScript strict mode
- Tailwind CSS + shadcn/ui component library
- Auto-generated API client from OpenAPI spec

Closes #12

🚀 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

**Bad:**
```
updated files
```

---

## 🔄 Pull Request Workflow

### 1. Create Feature Branch

**Option A: Traditional Branch (Single Task)**

```bash
# Start from main
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/your-feature-name
```

**Option B: Git Worktree (Parallel Tasks) ⭐ RECOMMENDED for parallel work**

```bash
# Create a new worktree in a separate directory
git worktree add ../ubik-issue-<number> -b feature/your-feature-name

# Move to the new worktree directory
cd ../ubik-issue-<number>

# Now you can work independently from the main repo
# Multiple worktrees = Multiple agents working in parallel!
```

**Why Use Worktrees for Parallel Development?**

- ✅ **No Conflicts**: Each worktree has its own working directory
- ✅ **Parallel Agents**: Frontend + backend agents can work simultaneously
- ✅ **No Stashing**: Switch between tasks without committing incomplete work
- ✅ **Independent State**: Each worktree has its own branch checkout
- ✅ **Easy Cleanup**: Remove worktree after PR merge

**Example: Running 2 Agents in Parallel**

```bash
# Agent 1: Frontend work on Issue #13
git worktree add ../ubik-issue-13 -b feature/agent-catalog-page
# Frontend agent works in ../ubik-issue-13

# Agent 2: Backend work on Issue #3
git worktree add ../ubik-issue-3 -b fix/employee-integration-test
# Backend agent works in ../ubik-issue-3

# Both agents can commit, push, and create PRs independently!
```

**Cleanup After PR Merge:**

```bash
# Return to main repo
cd /path/to/ubik-enterprise

# Remove worktree
git worktree remove ../ubik-issue-13

# Delete branch (if needed)
git branch -d feature/agent-catalog-page
```

### 2. Make Changes & Commit

```bash
# Make your changes following TDD

# Stage changes
git add <files>

# Commit with detailed message
git commit -m "$(cat <<'EOF'
feat: Your feature title (#<issue>)

Detailed description here...

Closes #<issue>
EOF
)"
```

### 3. Push to Remote

```bash
# Push branch
git push -u origin feature/your-feature-name
```

### 4. Create Pull Request

```bash
# Create PR with gh CLI
gh pr create \
  --title "feat: Your Feature Title (#<issue>)" \
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
  --head feature/your-feature-name
```

**PR Title Format:**
```
<type>: <Title> (#<issue-number>)

Examples:
- feat: Web UI Foundation with Next.js 14 (#12)
- fix: Employee integration test nil pointer (#3)
- docs: Add API endpoint documentation (#45)
```

### 5. Wait for CI/CD Checks

```bash
# Check PR status
gh pr checks <PR-number>

# Wait for all checks to pass (green ✅)
```

**Do NOT merge until:**
- ✅ All CI/CD checks pass
- ✅ Code review approved (if applicable)
- ✅ No merge conflicts

### 6. Merge Pull Request

```bash
# Merge PR (after checks pass)
gh pr merge <PR-number> --merge

# Or merge via GitHub UI
# Click "Merge pull request" → "Confirm merge"
```

### 7. Cleanup

```bash
# Switch back to main
git checkout main

# Pull latest
git pull origin main

# Delete local branch
git branch -d feature/your-feature-name

# Delete remote branch (usually auto-deleted)
git push origin --delete feature/your-feature-name
```

---

## 🔍 CI-Aware Development Workflow

**IMPORTANT:** Both `go-backend-developer` and `frontend-developer` agents now automatically wait for CI checks before completing tasks.

### Complete Workflow (Automated by Agents)

```bash
# 1. Pick task from GitHub Projects
gh issue list --label="backend,status/ready"

# 2. Create branch + workspace
git checkout -b feature/123-feature-description
git worktree add ../ubik-issue-123 feature/123-feature-description
cd ../ubik-issue-123

# 3. Implement feature (TDD)
# - Write failing tests
# - Implement code
# - All tests pass locally

# 4. Create PR with issue number in title (REQUIRED for automation!)
gh pr create --title "feat: Feature description (#123)" --body "..." --label "backend"
PR_NUM=$(gh pr view --json number -q .number)

# 5. Wait for CI checks (CRITICAL!)
gh pr checks $PR_NUM --watch --interval 10

# 6. Verify CI passed
CI_STATUS=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE" or .state == "CANCELLED")) | length')

if [ "$CI_STATUS" -eq 0 ]; then
  # All checks passed!
  # ✅ GitHub Actions automatically:
  #    - Updates issue status to "In Review"
  #    - Adds status/in-review label
  #    - Posts comment linking PR to issue

  echo "✅ PR #${PR_NUM} created and all CI checks passing."
  echo "✅ GitHub automation updated issue #123 status to 'In Review'."
else
  # Checks failed - investigate and fix
  gh pr checks $PR_NUM
  echo "❌ CI checks failed for PR #${PR_NUM}. Fix failures and push again."
  # Fix failures and push again - CI will re-run automatically
fi
```

### Why This Matters

- ✅ **Quality Gate**: All tests must pass in CI before moving to review
- ✅ **Automated**: Agents handle the entire workflow without manual intervention
- ✅ **Visibility**: GitHub Project status auto-updates to "In Review" when ready
- ✅ **Fast Feedback**: Failures caught immediately, not during code review
- ✅ **Clean Pipeline**: No broken PRs waiting for review

### Status Updates (Now Automated!)

**✨ NEW:** Status updates are now automated! Just include issue number in PR title:

```bash
# PR title format (automation triggers on this!)
gh pr create --title "feat: Your feature (#123)" ...

# When PR is created with passing CI:
#   → Issue status automatically updates to "In Review"
#   → status/in-review label added
#   → Comment posted linking PR to issue

# When PR is merged:
#   → Issue automatically closes
#   → status/done label added
#   → Branch automatically deleted
```

**Manual status updates only needed for:**
- Moving issues to "In Progress" when starting work
- Marking issues as "Blocked"
- Non-PR status changes (e.g., research tasks)

```bash
# Manual status update (rarely needed now)
./scripts/update-project-status.sh --issue 123 --status "In Progress"
./scripts/update-project-status.sh --issue 123 --status "Blocked"
```

**See agent configurations:**
- `.claude/agents/go-backend-developer.md` - Backend workflow (versioned in project)
- `.claude/agents/frontend-developer.md` - Frontend workflow (versioned in project)

---

## 🤖 Agent-Specific Instructions

### For `go-backend-developer` Agent

**Always follow this workflow when implementing backend features:**

1. Create feature branch: `feature/<feature-name>`
2. Write failing tests (TDD)
3. Implement feature to pass tests
4. Run full test suite: `make test`
5. Verify coverage: `make test-coverage`
6. Commit changes
7. Push to remote
8. Create PR
9. Wait for CI checks
10. Report completion to user

**Example Task Completion:**
```
Task: Implement GET /employees endpoint

Steps:
1. git checkout -b feature/list-employees-endpoint
2. Write failing test in tests/integration/employees_test.go
3. Implement handler in internal/handlers/employees.go
4. Run: make test
5. git commit -m "feat: Add GET /employees endpoint (#5)"
6. git push -u origin feature/list-employees-endpoint
7. gh pr create --title "feat: List Employees Endpoint (#5)" ...
8. Wait for checks
9. Report: "PR #XX created, all checks passing, ready for review"
```

### For `frontend-developer` Agent

**Always follow this workflow when implementing frontend features:**

1. Create feature branch: `feature/<feature-name>`
2. Write E2E tests first (TDD)
3. Implement UI components
4. Run type-check: `npm run type-check`
5. Run lint: `npm run lint`
6. Run build: `npm run build`
7. Run E2E tests: `npm run test:e2e`
8. Commit changes
9. Push to remote
10. Create PR
11. Wait for CI checks
12. Report completion to user

**Example Task Completion:**
```
Task: Implement Agent Catalog Page

Steps:
1. git checkout -b feature/agent-catalog-page
2. Write E2E test in tests/e2e/agent-catalog.spec.ts
3. Implement components in app/(dashboard)/agents/
4. npm run type-check && npm run lint && npm run build
5. git commit -m "feat: Add Agent Catalog Page (#13)"
6. git push -u origin feature/agent-catalog-page
7. gh pr create --title "feat: Agent Catalog Page (#13)" ...
8. Wait for checks
9. Report: "PR #XX created, all checks passing, ready for review"
```

---

## ✅ CI/CD Checks

### Current Checks (if configured)

- **TypeScript:** No compilation errors
- **ESLint:** No linting errors
- **Tests:** All tests passing
- **Build:** Production build successful

### If Checks Fail

1. Review error logs
2. Fix issues locally
3. Push fixes to same branch
4. Checks will re-run automatically

---

## 📊 Code Review Guidelines

### Before Requesting Review

- [ ] All tests pass locally
- [ ] Code follows project style guide
- [ ] No console.log or debug statements
- [ ] Documentation updated (if needed)
- [ ] CHANGELOG.md updated (for major changes)

### Review Checklist

- [ ] Code is clean and maintainable
- [ ] Tests cover new functionality
- [ ] No security vulnerabilities
- [ ] Performance considerations addressed
- [ ] Accessibility requirements met (for UI)

---

## 🚫 What NOT to Do

### ❌ DON'T:

1. Push directly to `main` branch
2. Commit without running tests
3. Create PR with failing tests
4. Merge before checks pass
5. Use vague commit messages ("fixed stuff")
6. Skip pull request process
7. Commit without linking to issue
8. Leave commented-out code
9. Commit secrets or credentials
10. Ignore CI/CD failures

### ✅ DO:

1. Always create feature branch
2. Run tests before committing
3. Write descriptive commit messages
4. Link commits/PRs to issues
5. Wait for CI/CD checks
6. Clean up branches after merge
7. Follow TDD methodology
8. Document complex changes
9. Keep PRs focused and small
10. Respond to review feedback

---

## 🎯 Quick Reference

### Full Workflow Commands

```bash
# 1. Create branch
git checkout main && git pull
git checkout -b feature/my-feature

# 2. Make changes
# ... write code, tests ...

# 3. Test
make test  # Backend
npm run build && npm run test:e2e  # Frontend

# 4. Commit
git add .
git commit -m "feat: My feature (#<issue>)

Details here...

Closes #<issue>

🚀 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# 5. Push
git push -u origin feature/my-feature

# 6. Create PR
gh pr create \
  --title "feat: My Feature (#<issue>)" \
  --body "Summary...

Closes #<issue>"

# 7. Check status
gh pr checks <PR-number>

# 8. Merge (after checks pass)
gh pr merge <PR-number> --merge

# 9. Cleanup
git checkout main
git pull
git branch -d feature/my-feature
```

---

## 📝 Issue Linking

### In Commits

```
Closes #123
Fixes #123
Resolves #123
Part of #123
Ref #123
```

### In PRs

```
Closes #123
Fixes #123
Resolves #123
```

**Auto-closes issue when PR is merged.**

---

## 🔄 Monorepo Considerations

### Service-Specific Changes

```bash
# API changes
cd services/api
make test

# CLI changes
cd services/cli
go test ./...

# Web changes
cd services/web
npm run build
```

### Cross-Service Changes

If changes affect multiple services:
1. Test each service individually
2. Test integration between services
3. Note dependencies in PR description

---

## 📚 Additional Resources

- **TDD Guide:** [docs/TESTING.md](./TESTING.md)
- **Development Guide:** [docs/DEVELOPMENT.md](./DEVELOPMENT.md)
- **Project Structure:** [CLAUDE.md](../CLAUDE.md)
- **Roadmap:** [IMPLEMENTATION_ROADMAP.md](../IMPLEMENTATION_ROADMAP.md)

---

## ✅ Summary

**For EVERY code change:**

1. ✅ Feature branch
2. ✅ TDD (tests first)
3. ✅ Commit
4. ✅ Push
5. ✅ PR
6. ✅ Wait for checks
7. ✅ Merge
8. ✅ Cleanup

**No shortcuts. No exceptions.**

---

🚀 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
