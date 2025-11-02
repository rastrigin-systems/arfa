---
name: github-task-manager
description: Manages GitHub issues and project tasks with proper parent-child relationships, status updates, and workflow automation. Use when creating tasks, splitting large tasks into subtasks, updating task status, linking related issues, or querying tasks. Ensures consistent task metadata (labels, milestones) and GitHub Project integration. Critical for maintaining proper task hierarchy and project organization.
---

# GitHub Task Manager Skill

Standardized GitHub task management across all agents with proper parent-child relationships and GitHub Project integration.

## When to Use This Skill

- Creating new GitHub issues with proper metadata
- Breaking down large tasks into linked subtasks
- Updating task status in GitHub Projects
- Establishing task relationships (parent-child, dependencies)
- Querying tasks by status, labels, or milestones
- Ensuring consistent task management across the codebase

## Core Capabilities

### 1. Create Task

Create a new GitHub issue with complete metadata and automatic project integration.

**Usage:**
```bash
gh issue create \
  --title "TITLE" \
  --label "LABELS" \
  --body "DESCRIPTION" \
  --milestone "MILESTONE" \
  --assignee "@me"
```

**Required Metadata:**
- **Title**: Clear, concise description
- **Labels**: Must include at least:
  - Area: `area/api`, `area/cli`, `area/web`, `area/db`, `area/infra`
  - Type: `type/feature`, `type/bug`, `type/chore`, `type/refactor`
  - Priority: `priority/p0`, `priority/p1`, `priority/p2`, `priority/p3`
  - Size: `size/xs`, `size/s`, `size/m`, `size/l`, `size/xl`
- **Description**: Clear problem statement and acceptance criteria
- **Milestone**: Target release version (if applicable)

**Auto-add to Project:**
```bash
ISSUE_URL=$(gh issue create ... | tail -1)
gh project item-add 3 --owner sergei-rastrigin --url "$ISSUE_URL"
```

**Set Initial Status:**
```bash
./scripts/update-project-status.sh --issue ISSUE_NUMBER --status "Todo"
```

**Available Statuses:**
- `Backlog` - Not yet prioritized
- `Todo` - Ready to work on
- `In Progress` - Currently being worked on
- `Blocked` - Waiting on dependencies
- `In Review` - PR created, awaiting review
- `Done` - Completed and merged

### 2. Create Sub-Issue

Create a child task properly linked to a parent issue using GitHub's sub-issue feature.

**Important:** This creates a **proper sub-issue** (not just a checklist item):
- **Sub-issues**: Proper parent-child relationship in GitHub's database (appears in "Sub-issues" section)
- **Subtasks** (checklist): Just text references like `- [ ] #39` (appears in "Subtasks" section)

Always use sub-issues for proper tracking!

**Step 1: Get Parent Issue Node ID**
```bash
PARENT_NODE_ID=$(gh api graphql -f query='
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    issue(number: $number) {
      id
    }
  }
}' -f owner='sergei-rastrigin' -f repo='ubik-enterprise' -F number=PARENT_NUM -q .data.repository.issue.id)
```

**Step 2: Create Subtask with Parent Reference**
```bash
SUB_ISSUE=$(gh issue create \
  --title "Subtask: TASK_TITLE" \
  --label "LABELS,subtask" \
  --body "$(cat <<'EOF'
Part of #PARENT_NUM

## Description
[Subtask description]

## Parent Task
This subtask is part of the larger feature tracked in #PARENT_NUM

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
EOF
)" | grep -oE '#[0-9]+' | cut -c2-)
```

**Step 3: Link to Parent (GitHub Sub-Issue API)**
```bash
# Get subtask node ID
SUB_NODE_ID=$(gh api graphql -f query='
query($owner: String!, $repo: String!, $number: Int!) {
  repository(owner: $owner, name: $repo) {
    issue(number: $number) {
      id
    }
  }
}' -f owner='sergei-rastrigin' -f repo='ubik-enterprise' -F number=$SUB_ISSUE -q .data.repository.issue.id)

# Link via GitHub's addSubIssue mutation (creates proper sub-issue relationship)
gh api graphql -f query='
mutation {
  addSubIssue(input: {
    issueId: "'"$PARENT_NODE_ID"'",
    subIssueId: "'"$SUB_NODE_ID"'"
  }) {
    issue {
      id
      number
    }
  }
}'
```

**Step 4: Add to Project**
```bash
gh project item-add 3 --owner sergei-rastrigin --url "https://github.com/sergei-rastrigin/ubik-enterprise/issues/$SUB_ISSUE"
```

**Step 5: Update Parent Issue**
```bash
gh issue comment PARENT_NUM --body "$(cat <<'EOF'
## Subtasks Created

Breaking this down into smaller tasks:

- [ ] #SUB1 - First subtask
- [ ] #SUB2 - Second subtask
- [ ] #SUB3 - Third subtask

Each subtask can be worked on independently.
EOF
)"
```

**Key Principles:**
- Sub-issues inherit area labels from parent
- Add `subtask` label to all child issues (optional)
- Reference parent in body: `Part of #PARENT_NUM`
- Use `addSubIssue` mutation to create proper parent-child relationship
- This creates a real sub-issue (not just a checklist reference)

### 3. Split Large Task

Analyze a task and break it into logical subtasks.

**When to Split:**
- Task is sized `size/l` or `size/xl`
- Task description has multiple acceptance criteria
- Task spans multiple areas (API + CLI + Web)
- Estimated time > 1 week

**Splitting Strategy:**
1. **Vertical Slices**: Each subtask delivers end-to-end value
2. **Dependencies First**: Order subtasks by dependencies
3. **Logical Grouping**: Group related work together
4. **Balanced Size**: Aim for `size/s` or `size/m` subtasks

**Example Split:**

**Parent:** "Implement Agent Configuration Management" (size/xl)

**Subtasks:**
1. "API: Create agent_configs endpoints" (size/m, area/api)
2. "Database: Add agent_configs table" (size/s, area/db)
3. "CLI: Add `ubik agents` command" (size/m, area/cli)
4. "Web: Agent configuration UI" (size/l, area/web)
5. "Tests: E2E agent config workflow" (size/m, area/testing)

**Implementation:**
```bash
# Get parent details
PARENT_NUM=36
PARENT_NODE_ID=$(gh api graphql ... -F number=$PARENT_NUM ...)

# Create subtasks
for task in "${SUBTASKS[@]}"; do
  SUB_NUM=$(gh issue create --title "$task" --body "Part of #$PARENT_NUM" ...)
  gh project item-add 3 --owner sergei-rastrigin --url "https://github.com/sergei-rastrigin/ubik-enterprise/issues/$SUB_NUM"
done

# Update parent
gh issue comment $PARENT_NUM --body "Split into subtasks: #SUB1, #SUB2, ..."
```

### 4. Update Task Status

Move tasks through the workflow in GitHub Projects.

**Usage:**
```bash
./scripts/update-project-status.sh --issue ISSUE_NUM --status "STATUS"
```

**Standard Workflow:**
```
Backlog → Todo → In Progress → In Review → Done
                      ↓
                   Blocked
```

**Status Transitions:**
- `Backlog → Todo`: Task is prioritized
- `Todo → In Progress`: Work starts
- `In Progress → In Review`: PR created, CI passing
- `In Review → Done`: PR merged
- `In Progress → Blocked`: Waiting on dependencies
- `Blocked → In Progress`: Dependencies resolved

**When to Update:**
- **Start work**: `Todo → In Progress`
- **Create PR**: `In Progress → In Review` (only after CI passes!)
- **Blocked**: `In Progress → Blocked` (add comment explaining blocker)
- **Merge PR**: `In Review → Done`

### 5. Link Related Tasks

Create relationships between issues (not parent-child).

**Relationship Types:**
- **Blocks**: This issue blocks another
- **Blocked by**: This issue is blocked by another
- **Depends on**: This issue depends on another
- **Related to**: General relationship

**Implementation:**
```bash
# Add relationship in issue body or comments
gh issue comment ISSUE_NUM --body "$(cat <<'EOF'
## Dependencies

**Blocks:** #OTHER_ISSUE
**Depends on:** #DEPENDENCY_ISSUE

[Explanation of relationship]
EOF
)"

# Or use GitHub's task list feature
gh issue edit ISSUE_NUM --body "$(cat <<'EOF'
## Description
...

## Dependencies
- Depends on #123
- Blocks #456
EOF
)"
```

### 6. Query Tasks

Find tasks by criteria using `gh` CLI.

**By Status:**
```bash
# Open issues
gh issue list --state open

# Closed issues
gh issue list --state closed

# All issues
gh issue list --state all
```

**By Label:**
```bash
# Backend tasks
gh issue list --label "area/api"

# High priority bugs
gh issue list --label "priority/p0,type/bug"

# Ready to work on
gh issue list --label "priority/p1" --state open
```

**By Milestone:**
```bash
gh issue list --milestone "v0.3.0"
```

**By Assignee:**
```bash
# My tasks
gh issue list --assignee "@me"

# Unassigned
gh issue list --assignee ""
```

**Project-Specific:**
```bash
# View project board
gh project view 3 --owner sergei-rastrigin

# List items in project
gh project item-list 3 --owner sergei-rastrigin --format json

# Filter by status
gh project item-list 3 --owner sergei-rastrigin | jq '.items[] | select(.status=="In Progress")'
```

## Label Standards

### Area Labels (Required - Pick One)
- `area/api` - Backend API changes
- `area/cli` - CLI client changes
- `area/web` - Web UI changes
- `area/db` - Database/schema changes
- `area/infra` - Infrastructure/DevOps
- `area/testing` - Test infrastructure
- `area/docs` - Documentation

### Type Labels (Required - Pick One)
- `type/feature` - New feature
- `type/bug` - Bug fix
- `type/chore` - Maintenance/tooling
- `type/refactor` - Code improvement
- `type/research` - Research/spike
- `type/epic` - Large multi-issue feature

### Priority Labels (Required - Pick One)
- `priority/p0` - Critical - Revenue blocker / Security issue
- `priority/p1` - High - Significant business impact
- `priority/p2` - Medium - Nice to have
- `priority/p3` - Low - Speculative / Future

### Size Labels (Recommended - Pick One)
- `size/xs` - < 2 hours
- `size/s` - 2-4 hours
- `size/m` - 1-2 days
- `size/l` - 3-5 days
- `size/xl` - > 1 week (should be split!)

### Impact Labels (Optional)
- `impact/revenue` - Directly impacts revenue
- `impact/acquisition` - Helps win customers
- `impact/retention` - Reduces churn
- `impact/efficiency` - Developer productivity

### Special Labels
- `subtask` - Child task of a larger issue
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention needed
- `blocked` - Waiting on dependencies

## Best Practices

### Task Creation
1. **Descriptive Titles**: Use action verbs ("Implement", "Fix", "Add", "Update")
2. **Clear Descriptions**: Include problem statement and acceptance criteria
3. **Proper Labels**: Always include area, type, and priority
4. **Milestones**: Assign to target release if known
5. **Assignees**: Assign to yourself when starting work

### Parent-Child Relationships
1. **Reference Parent**: Always include "Part of #PARENT" in subtask body
2. **Use GraphQL**: Link via GitHub's issue tracking API
3. **Update Parent**: Add checklist of subtasks to parent issue
4. **Inherit Labels**: Subtasks inherit area labels from parent
5. **Add Subtask Label**: Mark all child issues with `subtask` label

### Status Management
1. **Update Frequently**: Keep status current to reflect actual state
2. **Only One In Progress**: Limit WIP to maintain focus
3. **Wait for CI**: Only move to "In Review" after CI passes
4. **Explain Blocks**: Always comment when marking as "Blocked"
5. **Close on Merge**: Auto-close via PR description ("Closes #123")

### Task Splitting
1. **Vertical Slices**: Each subtask should deliver value
2. **Size Appropriately**: Aim for size/s or size/m subtasks
3. **Dependencies Clear**: Order subtasks by dependencies
4. **Test Last**: Create testing subtask after implementation tasks
5. **Document First**: Create documentation subtask alongside features

## Common Workflows

### Starting a New Feature
```bash
# 1. Create main task
ISSUE=$(gh issue create \
  --title "Implement Feature X" \
  --label "type/feature,area/api,priority/p1,size/l" \
  --body "..." | grep -oE '#[0-9]+' | cut -c2-)

# 2. Add to project
gh project item-add 3 --owner sergei-rastrigin --url "https://github.com/sergei-rastrigin/ubik-enterprise/issues/$ISSUE"

# 3. Set status
./scripts/update-project-status.sh --issue $ISSUE --status "Todo"

# 4. If large, split into subtasks
# (Follow "Create Subtask" pattern above)
```

### Reporting a Bug
```bash
gh issue create \
  --title "Bug: Specific issue description" \
  --label "type/bug,area/api,priority/p1" \
  --body "$(cat <<'EOF'
## Bug Description
[What's wrong]

## Steps to Reproduce
1. Step 1
2. Step 2

## Expected Behavior
[What should happen]

## Actual Behavior
[What actually happens]

## Environment
- Version: v0.2.0
- OS: macOS
- Browser: N/A (CLI)

## Logs/Screenshots
[If applicable]
EOF
)"
```

### Working on a Task
```bash
# 1. Check out branch
git checkout -b issue-123-feature-name

# 2. Update status
./scripts/update-project-status.sh --issue 123 --status "In Progress"

# 3. Do the work (TDD!)

# 4. Create PR
gh pr create --title "feat: Feature name (#123)" --body "Closes #123"

# 5. Wait for CI
gh pr checks --watch

# 6. Update status
./scripts/update-project-status.sh --issue 123 --status "In Review"
```

## GitHub Projects

**Engineering Roadmap (Project #3):**
- Owner: `sergei-rastrigin`
- URL: https://github.com/users/sergei-rastrigin/projects/3

**Marketing Board (Project #4):**
- Owner: `sergei-rastrigin`
- URL: https://github.com/users/sergei-rastrigin/projects/4

**Default Project:** Engineering Roadmap (#3)

## Scripts

### update-project-status.sh
```bash
./scripts/update-project-status.sh --issue ISSUE_NUM --status "STATUS" [--project PROJECT_NAME]
```

**Parameters:**
- `--issue`: Issue number
- `--status`: Target status (Backlog, Todo, In Progress, Blocked, In Review, Done)
- `--project`: Optional project name (default: "engineering")

## Examples

See `examples/workflow-examples.md` for complete real-world examples.

## Templates

See `templates/` directory for issue templates:
- `task.md` - Standard task template
- `subtask.md` - Subtask template
- `bug.md` - Bug report template

---

**This skill ensures consistent, high-quality task management across all AI agents and human developers.**
