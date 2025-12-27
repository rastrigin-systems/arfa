# Development Workflows

**Last Updated:** 2025-11-05

This document covers specialized workflows for milestone planning, release management, and project coordination.

---

## Table of Contents

- [Milestone Planning](#milestone-planning)
- [Milestone Transitions](#milestone-transitions)
- [Task Splitting](#task-splitting)
- [Release Workflow](#release-workflow)
- [Best Practices](#best-practices)

---

## Milestone Planning

### After Releasing a Milestone

**Complete workflow for transitioning between milestones.**

#### 1. Archive Completed Milestone

```bash
# Archive all issues from completed milestone
./scripts/archive-milestone.sh --milestone v0.3.0
```

**This will:**
- Label all milestone issues as "archived"
- Close any remaining open issues
- Close the milestone
- Update `docs/MILESTONES_ARCHIVE.md` with completion record

---

#### 2. Start New Milestone

```bash
# Create new milestone and populate from backlog
./scripts/start-milestone.sh \
  --version v0.4.0 \
  --description "Analytics Dashboard & Approval Workflows" \
  --due-date "2026-01-31" \
  --auto-split
```

**This will:**
1. Create GitHub milestone with description and due date
2. Query backlog for `priority/p0` and `priority/p1` issues
3. Display issues for review and confirmation
4. Add confirmed issues to milestone
5. Move issues to "Todo" status on project board
6. Flag large tasks (size/l, size/xl) for splitting
7. Create milestone kickoff issue

---

#### 3. Split Large Tasks

```bash
# Find tasks flagged for splitting
gh issue list --label "needs-splitting" --milestone "v0.4.0"

# Split a large task
./scripts/split-large-tasks.sh --issue 51

# Or use auto-split with github-task-manager skill
./scripts/split-large-tasks.sh --issue 51 --auto
```

**Purpose:** Break down size/xl and size/l tasks into manageable subtasks (size/s or size/m).

---

### Milestone Planning Best Practices

#### Before Starting New Milestone

**Review and Prepare:**
- ✅ Review backlog and update priorities
- ✅ Ensure issue descriptions are clear
- ✅ Verify all issues have size labels
- ✅ Check for dependencies between issues
- ✅ Set realistic due date (4-6 weeks typical)

---

#### When Populating Milestone

**Selection Criteria:**
- ✅ Focus on p0/p1 priority issues
- ✅ Aim for mix of sizes (not all large tasks)
- ✅ Balance features vs bug fixes vs tech debt
- ✅ Include testing and documentation tasks
- ✅ Leave buffer for unexpected work (70-80% capacity)

---

#### Task Splitting Guidelines

**Split When:**
- Task is size/l or size/xl
- Task has multiple distinct deliverables
- Task spans multiple systems or modules
- Estimation uncertainty is high

**How to Split:**
- ✅ Each subtask should be independently testable
- ✅ Subtasks should be size/s or size/m (1-3 days each)
- ✅ Use parent-child relationship (blockedBy in GitHub)
- ✅ Update parent with checklist of subtasks
- ✅ Close parent only when all subtasks complete

**Example Split:**

```
Parent: Implement Analytics Dashboard (size/xl)
├─ Subtask 1: Design dashboard data model (size/s)
├─ Subtask 2: Implement backend API endpoints (size/m)
├─ Subtask 3: Create dashboard UI components (size/m)
├─ Subtask 4: Add real-time updates via WebSocket (size/s)
└─ Subtask 5: Write tests and documentation (size/s)
```

---

## Milestone Transitions

### Complete Workflow Example

```bash
# After releasing v0.3.0, transition to v0.4.0

# 1. Archive completed milestone
./scripts/archive-milestone.sh --milestone v0.3.0

# 2. Start new milestone
./scripts/start-milestone.sh \
  --version v0.4.0 \
  --description "Analytics & Approvals" \
  --due-date "2026-01-31" \
  --auto-split

# 3. Split flagged large tasks
for issue in $(gh issue list --label "needs-splitting" --milestone "v0.4.0" --json number -q '.[].number'); do
  ./scripts/split-large-tasks.sh --issue $issue
done

# 4. Verify milestone ready
gh issue list --milestone "v0.4.0" --json number,title,labels

# 5. Start working on first task
FIRST_TASK=$(gh issue list --milestone "v0.4.0" --label "priority/p0" --assignee "" --limit 1 --json number -q '.[0].number')
git checkout -b issue-$FIRST_TASK-feature
./scripts/update-project-status.sh --issue $FIRST_TASK --status "In Progress"
```

---

## Task Splitting

### When to Split Tasks

**Size-Based Triggers:**
- size/xl (>5 days) - Always split
- size/l (3-5 days) - Usually split
- size/m (1-3 days) - Rarely split
- size/s (<1 day) - Never split

**Complexity Triggers:**
- Multiple components/modules affected
- Requires changes across API + CLI + Web
- High estimation uncertainty
- Spans multiple skill domains

---

### Splitting Strategies

#### 1. Vertical Slicing (Recommended)

**Split by end-to-end features:**

```
Task: User Authentication
├─ Login flow (API + UI)
├─ Registration flow (API + UI)
├─ Password reset (API + UI)
└─ Session management (API + UI)
```

**Benefits:**
- Each subtask delivers user value
- Can test end-to-end
- Can release incrementally

---

#### 2. Horizontal Slicing

**Split by technical layer:**

```
Task: User Dashboard
├─ Database schema and migrations
├─ API endpoints
├─ UI components
└─ Integration tests
```

**Benefits:**
- Clear technical boundaries
- Easier to parallelize
- Good for specialized skills

**Drawbacks:**
- No user value until all complete
- Integration risk at end

---

#### 3. Hybrid Approach

**Combine vertical and horizontal:**

```
Task: Analytics Dashboard (size/xl)
├─ User metrics API + UI (vertical)
├─ Agent usage API + UI (vertical)
├─ Real-time updates (horizontal)
└─ Export functionality (vertical)
```

---

### Using GitHub Task Manager Skill

**Automatic Task Splitting:**

```bash
# Use github-task-manager skill
./scripts/split-large-tasks.sh --issue 51 --auto

# This will:
# 1. Analyze parent task description
# 2. Suggest subtask breakdown
# 3. Create subtasks with proper labels
# 4. Link to parent issue
# 5. Update parent with checklist
```

**See:** [github-task-manager skill](../../.claude/skills/github-task-manager/SKILL.md)

---

## Release Workflow

### Quick Reference

**See [Release Manager Skill](../../.claude/skills/release-manager/SKILL.md) for complete workflow.**

**Quick Release Checklist:**
```
✅ 1. All CI/CD checks passing
✅ 2. All tests passing (make test)
✅ 3. Milestone issues closed
✅ 4. On main branch, clean working tree
✅ 5. Documentation updated
✅ 6. Create annotated git tag
✅ 7. Push tag to remote
✅ 8. Create GitHub Release
✅ 9. Update CLAUDE.md and docs/RELEASES.md
```

**Versioning Strategy:**
- **v0.x.0** - New milestone features (Web UI, Analytics, etc.)
- **v0.x.y** - Bug fixes and polish within a milestone
- **v1.0.0+** - Production releases (post-launch)

**Key Commands:**

```bash
# Check release readiness
gh run list --limit 1  # Verify CI green
make test              # Run all tests
gh issue list --milestone "v0.X.0" --state open  # Check milestone

# Create release
git tag -a v0.X.0 -m "Release v0.X.0 - [Description]"
git push origin v0.X.0
gh release create v0.X.0 --title "..." --notes "..."
```

**Release History:** See [releases/RELEASES.md](../releases/RELEASES.md)

---

## Best Practices

### General Principles

1. **Plan Incrementally**
   - Start with high-priority items
   - Add stretch goals if time permits
   - Leave buffer for unexpected work

2. **Maintain Flexibility**
   - Re-prioritize as needed
   - Don't hesitate to move tasks between milestones
   - Focus on delivering value over completing all tasks

3. **Communicate Progress**
   - Update issue statuses regularly
   - Comment on blockers immediately
   - Share wins with the team

4. **Learn and Improve**
   - Review velocity after each milestone
   - Adjust estimation based on actuals
   - Document lessons learned

---

### Milestone Sizing

**Typical Milestone Capacity:**

| Team Size | Weeks | Story Points | Features |
|-----------|-------|--------------|----------|
| 1 developer | 4-6 | 20-30 | 3-5 major |
| 2 developers | 4-6 | 40-60 | 6-10 major |
| 3+ developers | 4-6 | 60-90 | 10-15 major |

**Adjust based on:**
- Team experience with codebase
- Complexity of features
- Amount of technical debt
- External dependencies

---

### Issue Labeling

**Priority Labels:**
- `priority/p0` - Critical, blocks release
- `priority/p1` - High, should be in milestone
- `priority/p2` - Medium, nice to have
- `priority/p3` - Low, backlog

**Size Labels:**
- `size/xs` - < 4 hours
- `size/s` - < 1 day
- `size/m` - 1-3 days
- `size/l` - 3-5 days
- `size/xl` - > 5 days (should be split)

**Status Labels:**
- `status/todo` - Ready to start
- `status/in-progress` - Currently working
- `status/blocked` - Waiting on dependency
- `status/waiting-for-review` - PR submitted
- `status/done` - Completed and merged

---

## See Also

- [Release Manager Skill](../../.claude/skills/release-manager/SKILL.md) - Complete release workflow
- [PR Workflow](./workflows.md) - Standard PR and Git workflow
- [Testing](./testing.md) - Testing best practices
- [Release History](../releases/RELEASES.md) - Release history
