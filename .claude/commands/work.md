# Start Work on Task

Automate the complete development workflow from task selection to PR creation.

## Usage

```
/work                    # Pick next Todo task from project
/work 123                # Work on specific issue #123
/work "feature name"     # Create new issue and start work
```

## Workflow

### If Issue Number Provided (`/work 123`)

1. Fetch issue details: `gh issue view 123`
2. Confirm with user before proceeding
3. Continue to **Start Work** below

### If Search Term Provided (`/work "description"`)

1. Create new issue using **github-task-manager** skill
2. Add to project, set labels, assign to @me
3. Continue to **Start Work** below

### If No Argument (`/work`)

1. Query project for Todo tasks: `gh issue list --state open --label "status/todo"`
2. Filter by priority (P0 > P1 > P2)
3. Show top task, confirm with user
4. Continue to **Start Work** below

## Start Work

Use **github-dev-workflow** skill (Workflow 1: Start Task):

```bash
ISSUE_NUM=<selected issue>

# 1. Update status
./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Progress"

# 2. Self-assign
gh issue edit $ISSUE_NUM --add-assignee "@me"

# 3. Create worktree
git worktree add ../$(basename $(pwd))-${ISSUE_NUM} -b feature/${ISSUE_NUM}-description
cd ../$(basename $(pwd))-${ISSUE_NUM}
```

## Execute Work

Based on area labels, either:
- **area/api, area/cli, area/db** → Implement directly or use go-backend-developer agent
- **area/web** → Use frontend-developer agent
- **Multiple areas** → Use tech-lead agent to coordinate

Follow TDD: write tests first, then implementation.

## Create PR

Use **github-pr-workflow** skill:

```bash
# 1. Commit and push
git add . && git commit -m "feat: Description

Closes #${ISSUE_NUM}

🤖 Generated with [Claude Code](https://claude.com/claude-code)
Co-Authored-By: Claude <noreply@anthropic.com>"

git push -u origin feature/${ISSUE_NUM}-description

# 2. Create PR (MUST include #ISSUE_NUM in title)
gh pr create --title "feat: Title (#${ISSUE_NUM})" --body "Closes #${ISSUE_NUM}"

# 3. Wait for CI (MANDATORY)
gh pr checks $(gh pr view --json number -q .number) --watch

# 4. Update status only if CI passes
./scripts/update-project-status.sh --issue $ISSUE_NUM --status "In Review"
```

## Completion Report

```
✅ Task #${ISSUE_NUM} Complete

Issue: {title}
Branch: feature/${ISSUE_NUM}-{slug}
PR: {url}
Status: In Review

Next: Wait for CI, address reviews, merge when approved
```

## Skills Used

- **github-task-manager** - Issue creation, status updates, labels
- **github-dev-workflow** - Worktree setup, branch naming
- **github-pr-workflow** - PR creation, CI monitoring, merge
