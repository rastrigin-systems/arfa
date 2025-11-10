# Work on Task - Automated Development Workflow

Automate the complete task workflow: pick a task, create worktree, implement, create PR, and update status.

## Usage

```bash
/work                    # Pick next Todo task from project board
/work 123                # Work on specific issue number
/work "feature name"     # Work on task matching title
```

## Workflow

You will execute the following steps:

### 1. Task Selection

**If no argument provided:**
- Query GitHub project for tasks with status "Todo"
- Filter by priority (P0 > P1 > P2 > P3)
- Select highest priority task
- Confirm with user before proceeding

**If issue number provided (e.g., `/work 123`):**
- Fetch issue #123 details
- Verify it exists and is not closed
- Confirm with user before proceeding

**If search term provided (e.g., `/work "wireframes"`):**
- Search open issues for matching title
- Show matches to user
- Let user select which one to work on

### 2. Update Task Status

```bash
# Move task to "In Progress"
./scripts/update-project-status.sh --issue ISSUE_NUM --status "In Progress"
```

### 3. Create Git Worktree

```bash
# Ensure on main branch
git checkout main
git pull

# Create worktree with branch name: feature/{num}-{slug}
git worktree add /Users/sergeirastrigin/Projects/ubik-issue-{NUM} -b feature/{NUM}-{slug}

# Example: feature/158-wireframes
```

**Branch naming:**
- Format: `feature/{num}-{slug}`
- Slug: lowercase, hyphenated, max 3-4 words from title
- Example: Issue #158 "Create wireframes for user stories" → `feature/158-wireframes`

**Worktree path:**
- Path still uses `ubik-issue-{NUM}` for consistency with existing worktrees
- Example: `/Users/sergeirastrigin/Projects/ubik-issue-158`

### 4. Understand the Task

Read the issue description and comments:
```bash
gh issue view ISSUE_NUM
```

**Extract:**
- Acceptance criteria
- Related issues/dependencies
- Area labels (area/api, area/web, area/cli, area/db)
- Type (feature, bug, refactor, etc.)
- Any specific requirements or constraints

### 5. Determine Agent Assignment

Based on area labels, assign to appropriate agent:

**Area → Agent Mapping:**
- `area/api` → **go-backend-developer**
- `area/cli` → **go-backend-developer**
- `area/web` → **frontend-developer**
- `area/db` → **go-backend-developer** (schema changes)
- `area/docs` → **product-designer** (if wireframes/UI) or handle directly
- `area/infra` → **tech-lead** (architecture decisions)

**Special Cases:**
- Multiple areas (e.g., API + Web) → Use **tech-lead** to coordinate
- UI/UX design needed → Use **product-designer** first, then **frontend-developer**
- Large feature (size/xl) → Use **tech-lead** to break down into subtasks

### 6. Execute Work

**If delegating to agent:**
```
Use Task tool with appropriate agent:
- Provide full task context
- Include acceptance criteria
- Specify deliverables
- Request they work in the worktree path
```

**If working directly:**
- Follow TDD workflow (write tests first)
- Implement changes
- Run tests and ensure they pass
- Commit changes with descriptive messages

### 7. Create Pull Request

Once work is complete:

```bash
cd /Users/sergeirastrigin/Projects/ubik-issue-{NUM}

# Push branch to remote
git push -u origin feature/{NUM}-{slug}

# Create PR with proper title format
gh pr create \
  --title "{type}: {title} (#{NUM})" \
  --body "$(cat <<'EOF'
## Summary
[Brief description of changes]

## Changes
- Change 1
- Change 2
- Change 3

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Related
Closes #{NUM}

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

**PR Title Format:**
- `feat: Description (#123)` - New feature
- `fix: Description (#123)` - Bug fix
- `docs: Description (#123)` - Documentation
- `refactor: Description (#123)` - Code refactoring
- `chore: Description (#123)` - Maintenance

### 8. Update Task Status

```bash
# Move task to "In Review"
./scripts/update-project-status.sh --issue ISSUE_NUM --status "In Review"
```

### 9. Report Completion

Provide summary:
```
✅ Task #{NUM} Complete

**Issue:** {title}
**Branch:** feature/{NUM}-{slug}
**PR:** https://github.com/rastrigin-org/ubik-enterprise/pull/{PR_NUM}
**Status:** In Review

**Summary:**
{Brief description of what was implemented}

**Next Steps:**
- Wait for CI checks to pass
- Address review feedback if any
- Merge when approved
```

## Error Handling

**If task is already in progress:**
- Show who is assigned
- Ask if you should take over or pick a different task

**If worktree already exists:**
- Ask if you should:
  1. Delete and recreate (lose uncommitted work)
  2. Resume work in existing worktree
  3. Pick a different task

**If CI fails after PR creation:**
- Keep task in "In Progress"
- Report CI failures
- Fix issues before moving to "In Review"

**If no Todo tasks available:**
- Check Backlog
- Suggest prioritizing a backlog task
- Or ask user what to work on

## Best Practices

1. **Always confirm task selection** - Don't start without user approval
2. **Follow TDD** - Write tests first, then implementation
3. **Commit frequently** - Small, logical commits with clear messages
4. **Test before PR** - Ensure all tests pass locally
5. **Only move to "In Review" if CI passes** - Never create PR with failing tests
6. **Clean commits** - Squash fixup commits before PR
7. **Link issues** - Always use "Closes #123" in PR description

## Examples

### Example 1: Pick Next Task
```
User: /work
Assistant:
🔍 Finding next Todo task...

Found: Issue #160 "Implement admin dashboard API endpoints" (priority/p1, area/api, size/m)

**Description:**
Create REST API endpoints for admin dashboard data...

**Acceptance Criteria:**
- [ ] GET /api/v1/approvals/pending
- [ ] GET /api/v1/activity-logs
- [ ] GET /api/v1/organizations/current/stats

Should I start working on this? (yes/no)

User: yes
Assistant:
✅ Starting work on Issue #160

[Creates worktree with branch feature/160-admin-dashboard-api, delegates to go-backend-developer agent, creates PR, updates status]
```

### Example 2: Work on Specific Issue
```
User: /work 158
Assistant:
📋 Loading Issue #158...

**Title:** Create wireframes for documented user stories
**Status:** Todo
**Labels:** area/docs, type/feature, priority/p1, size/m

This requires product-designer agent for wireframe creation.

Should I proceed? (yes/no)

User: yes
Assistant:
[Creates worktree with branch feature/158-wireframes, delegates to product-designer, creates PR, updates status]
```

### Example 3: Search and Select
```
User: /work "dashboard"
Assistant:
🔍 Searching for tasks matching "dashboard"...

Found 3 matches:
1. #160 - Implement admin dashboard API endpoints (Todo, P1)
2. #161 - Create dashboard loading states (Backlog, P2)
3. #162 - Add dashboard filters (Backlog, P3)

Which task should I work on? (1/2/3)

User: 1
Assistant:
[Proceeds with Issue #160 using branch feature/160-admin-dashboard-api]
```

## Notes

- This command is designed for single-task workflow
- For multi-task work, run `/work` multiple times
- Always review PR before merging
- This automates the mechanical parts, not the thinking
- You can still intervene at any step if needed
- Branch naming uses `feature/{num}-{slug}` but worktree path uses `ubik-issue-{NUM}` for consistency
