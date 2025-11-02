---
name: release-manager
description: Standardized release management process for creating versioned releases with proper tagging, GitHub Releases, and changelog generation. Use when preparing releases, creating tags, or documenting version history.
---

# Release Manager Skill

Complete release management workflow for Ubik Enterprise monorepo with semantic versioning, automated checks, and comprehensive documentation.

## When to Use This Skill

- Preparing a new release (minor, major, or patch)
- Creating git tags and GitHub Releases
- Generating changelogs from commit history
- Resolving version conflicts or tag issues
- Documenting release history
- Understanding what goes into each release

## Release Philosophy

**Ubik Enterprise follows milestone-driven releases:**

1. **Milestones define features** - GitHub milestones (v0.3.0, v0.4.0) track feature sets
2. **Releases when complete** - Tag and release when milestone goals are met
3. **Quality over speed** - Only release when CI is green and features are tested
4. **Semantic versioning** - Follow semver for predictable version numbers

## Semantic Versioning Strategy

**Format:** `vMAJOR.MINOR.PATCH` (e.g., `v0.3.0`, `v1.2.1`)

### Version Components

- **MAJOR** (v1.0.0 → v2.0.0): Breaking API changes, major architecture changes
- **MINOR** (v0.3.0 → v0.4.0): New features, milestone completions (backward compatible)
- **PATCH** (v0.3.0 → v0.3.1): Bug fixes, small improvements (backward compatible)

### Pre-1.0 Versioning (Current)

We're in **0.x.y** phase (pre-production):
- **0.x.0** - New milestone features (Web UI, Analytics, etc.)
- **0.x.y** - Bug fixes and polish within a milestone
- Breaking changes allowed in 0.x releases

### Post-1.0 Versioning (Future)

After **v1.0.0** (production-ready):
- **1.x.0** - New features (backward compatible)
- **1.0.x** - Bug fixes only
- **2.0.0** - Breaking changes (API redesign, major refactor)

## Release Criteria

**ALL criteria must be met before tagging a release:**

### ✅ Mandatory Checks

1. **CI/CD Passing**
   ```bash
   gh run list --limit 1 --json conclusion -q '.[0].conclusion'
   # Must output: "success"
   ```

2. **All Tests Passing**
   ```bash
   make test           # Backend tests
   cd services/web && npm test  # Frontend tests
   ```

3. **Milestone Complete**
   ```bash
   gh issue list --milestone "v0.X.0" --state open
   # Should return empty (all issues closed)
   ```

4. **No Uncommitted Changes**
   ```bash
   git status
   # Should be clean (nothing to commit)
   ```

5. **On Main Branch**
   ```bash
   git branch --show-current
   # Must output: "main"
   ```

### 📋 Optional (But Recommended)

- Manual testing completed
- Documentation updated (CLAUDE.md, README.md)
- Breaking changes documented
- Migration guide written (if needed)

## Release Workflow

### 1. Pre-Release Checklist

```bash
# Update issue status
./scripts/update-project-status.sh --issue 49 --status "In Progress"

# Verify CI is green
gh run list --limit 1

# Check milestone completion
gh issue list --milestone "v0.3.0" --state open

# Verify tests pass
make test
cd services/web && npm test && cd ../..

# Ensure on main branch with latest code
git checkout main
git pull origin main

# Verify working tree is clean
git status
```

### 2. Determine Version Number

**Decision Matrix:**

| Change Type | Example | Version Bump |
|------------|---------|--------------|
| New milestone features (Web UI, Analytics) | Multiple new pages | 0.x.0 → 0.(x+1).0 |
| Bug fixes within milestone | Fix login issue | 0.3.0 → 0.3.1 |
| Critical hotfix | Security patch | 0.3.1 → 0.3.2 |
| Breaking API change (pre-1.0) | Change endpoint structure | 0.3.0 → 0.4.0 |

**Current Release:** v0.3.0 (Web UI milestone)
**Next Minor:** v0.4.0 (Analytics/Approvals milestone)
**Next Patch:** v0.3.1 (Web UI bug fixes)

### 3. Generate Changelog

```bash
# Get commits since last release
LAST_TAG=$(git describe --tags --abbrev=0)
echo "Changes since $LAST_TAG:"
git log $LAST_TAG..HEAD --oneline --no-merges

# Categorize by type (from conventional commits)
echo "## Features:"
git log $LAST_TAG..HEAD --oneline --no-merges --grep="^feat"

echo "## Bug Fixes:"
git log $LAST_TAG..HEAD --oneline --no-merges --grep="^fix"

echo "## Chores:"
git log $LAST_TAG..HEAD --oneline --no-merges --grep="^chore"

# Count commits
git log $LAST_TAG..HEAD --oneline --no-merges | wc -l
```

### 4. Create Git Tag

```bash
VERSION="v0.4.0"  # Change this!

# Create annotated tag with detailed message
git tag -a $VERSION -m "Release $VERSION - [Milestone Name]

🎉 [Brief description of major changes]

## 🌟 Highlights

- ✅ [Major feature 1]
- ✅ [Major feature 2]
- ✅ [Major feature 3]

## 📊 New Features

[List new features with details]

## 🐛 Bug Fixes

[List bug fixes]

## 📈 Stats

- X commits since [last version]
- Y new features
- Z bug fixes

## 🔧 Technical Details

[Architecture changes, library updates, etc.]

## 🎯 What's Next ([next version])

[Preview of next milestone]

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# Verify tag was created
git tag -l -n9 $VERSION
```

### 5. Push Tag to Remote

```bash
# Push tag
git push origin $VERSION

# Or push all tags
git push --tags
```

### 6. Create GitHub Release

```bash
VERSION="v0.4.0"  # Match your tag!

# Create release with changelog
gh release create $VERSION \
  --title "Release $VERSION - [Milestone Name]" \
  --notes "$(cat <<'EOF'
## 🌟 Highlights

- ✅ [Major accomplishment 1]
- ✅ [Major accomplishment 2]

## 📊 What's New

### Features
- [Feature 1 description]
- [Feature 2 description]

### Bug Fixes
- [Fix 1 description]
- [Fix 2 description]

### Technical Improvements
- [Improvement 1]
- [Improvement 2]

## 📈 Statistics

- X commits
- Y new features
- Z bug fixes
- W% test coverage

## 🚀 Upgrade Guide

[If breaking changes, include upgrade instructions]

## 📝 Full Changelog

https://github.com/sergei-rastrigin/ubik-enterprise/compare/[PREV_TAG]...$VERSION

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"

# Verify release was created
gh release view $VERSION
```

### 7. Update Documentation

```bash
# Update CLAUDE.md current version
vim CLAUDE.md
# Change "Current Status" section to reflect new version

# Commit documentation updates
git add CLAUDE.md docs/RELEASES.md
git commit -m "docs: Update documentation for $VERSION release

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
git push
```

### 8. Archive Milestone

```bash
# Archive milestone and remove items from project board
./scripts/archive-milestone.sh --milestone $VERSION

# This will:
# - Add "archived" label to all milestone issues
# - Close any remaining open issues
# - Archive all items from project board (GraphQL API)
# - Close the milestone on GitHub
# - Update docs/MILESTONES_ARCHIVE.md
```

**What Gets Archived:**
- All issues in the milestone are labeled "archived"
- All project board items are archived (removed from active view)
- Milestone is closed on GitHub
- Documentation updated in `docs/MILESTONES_ARCHIVE.md`

**Note:** Archived project items can still be viewed via:
```bash
# View archived items (requires GraphQL)
gh api graphql -f query='
query {
  node(id: "PVT_kwHOAGhClM4BG_A3") {
    ... on ProjectV2 {
      items(first: 100) {
        nodes {
          id
          isArchived
          content {
            ... on Issue {
              number
              title
            }
          }
        }
      }
    }
  }
}'
```

### 9. Announce Release

```bash
# Comment on related issues
gh issue comment [ISSUE_NUM] --body "🎉 Released in $VERSION!"

# Post to discussions (if enabled)
# https://github.com/sergei-rastrigin/ubik-enterprise/discussions
```

## Handling Special Cases

### Moving/Updating an Existing Tag

**⚠️ Use with caution - only for recent tags not yet pulled by others!**

```bash
# Delete local tag
git tag -d v0.3.0

# Delete remote tag
git push origin :refs/tags/v0.3.0

# Create new tag at current HEAD
git tag -a v0.3.0 -m "Release v0.3.0 - [New description]"

# Force push new tag
git push origin v0.3.0 --force
```

### Creating a Patch Release

```bash
# For bug fix releases (0.3.0 → 0.3.1)
VERSION="v0.3.1"

git tag -a $VERSION -m "Release $VERSION - Bug Fixes

## 🐛 Bug Fixes

- Fix [issue description] (#123)
- Fix [issue description] (#124)

Patch release with critical bug fixes.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

git push origin $VERSION
gh release create $VERSION --title "Release $VERSION - Bug Fixes" \
  --notes "Critical bug fixes for v0.3.0. See tag for details."
```

### Rolling Back a Release

**If you need to unpublish a broken release:**

```bash
VERSION="v0.4.0"

# Delete GitHub Release (keeps tag)
gh release delete $VERSION --yes

# Or delete both release and tag
gh release delete $VERSION --cleanup-tag --yes
```

## Monorepo Versioning

**Single Version for Entire Monorepo:**

- API Server, CLI, and Web UI share the same version
- Version applies to the entire platform release
- Individual services don't have separate versions

**Why Unified Versioning?**
- Simpler for users (one version to track)
- Services are tightly coupled (API + Web)
- Easier to test and release together

**If we ever need independent service versions:**
- Tag format: `api-v1.0.0`, `cli-v1.0.0`, `web-v1.0.0`
- Only use if services become truly independent

## Release History

### Existing Releases

| Version | Date | Milestone | Description |
|---------|------|-----------|-------------|
| v0.1.0 | 2025-10-29 | API Foundation | Authentication + Employee CRUD |
| v0.2.0 | 2025-10-29 | CLI Client | Employee CLI with Docker integration |
| v0.3.0 | 2025-10-29 | Monorepo Migration | Go workspace restructure |

**v0.3.0 Tag Confusion:**
- Current v0.3.0 tag points to monorepo migration commit
- v0.3.0 milestone (Web UI) is actually at current HEAD
- Decision needed: Move tag or create v0.4.0?

## Version Querying

```bash
# Current version (latest tag)
git describe --tags --abbrev=0

# All releases
gh release list

# Commits since last release
git log $(git describe --tags --abbrev=0)..HEAD --oneline

# Compare two versions
git log v0.2.0..v0.3.0 --oneline
gh release compare v0.2.0..v0.3.0
```

## Best Practices

1. **Always use annotated tags** (`-a` flag) - includes metadata
2. **Write detailed release notes** - help users understand changes
3. **Follow conventional commits** - makes changelog generation easier
4. **Test before tagging** - releases should be stable
5. **Update docs first** - commit doc updates before tagging
6. **Keep changelog updated** - don't rely on git log alone
7. **Link to issues/PRs** - use #123 syntax in release notes

## Milestone Transition Workflow

**Complete workflow for transitioning between milestones after a release.**

### Overview

After successfully releasing a milestone (e.g., v0.3.0), follow this process to:
1. Archive completed milestone issues
2. Plan and start the next milestone (e.g., v0.4.0)
3. Prioritize backlog and split large tasks

### Phase 1: Archive Completed Milestone

**Purpose:** Clean up completed work and preserve history

```bash
# Archive all issues from completed milestone
./scripts/archive-milestone.sh --milestone v0.3.0
```

**What this does:**
- Labels all milestone issues as "archived"
- Closes any remaining open issues
- Closes the milestone
- Updates `docs/MILESTONES_ARCHIVE.md` with completion record

**Example output:**
```
📦 Archiving Milestone: v0.3.0
==================================

📋 Fetching issues in milestone...
✓ Found 15 issues:
  - Closed: 14
  - Open: 1

⚠️  WARNING: Milestone has 1 open issues:
  #42: Polish E2E test output

Continue archiving? (y/N): y

📦 Archiving issues...
  #32: ✓ Labeled as archived
  #33: ✓ Labeled as archived
  ...

🎯 Closing milestone...
✓ Milestone closed

📝 Updating documentation...
✓ Updated docs/MILESTONES_ARCHIVE.md

✅ Milestone v0.3.0 archived successfully!
```

### Phase 2: Start New Milestone

**Purpose:** Create new milestone and populate with prioritized backlog

```bash
# Start new milestone
./scripts/start-milestone.sh \
  --version v0.4.0 \
  --description "Analytics & Approvals" \
  --due-date "2026-01-31" \
  --auto-split
```

**What this does:**
1. Creates GitHub milestone with description and due date
2. Queries backlog for `priority/p0` and `priority/p1` issues
3. Displays issues for review
4. Adds confirmed issues to milestone
5. Moves issues to "Todo" status on project board
6. Flags large tasks (size/l, size/xl) for splitting
7. Creates milestone kickoff issue

**Example output:**
```
🚀 Starting Milestone: v0.4.0
=================================

📅 Creating milestone...
✓ Milestone 'v0.4.0' created

📋 Querying backlog...
✓ Found 8 backlog issues

🎯 Backlog Issues (priority/p0, p1):
-----------------------------------
  #50 [size/m]: Add approval workflow UI
  #51 [size/l]: Analytics dashboard
  #52 [size/s]: Cost tracking per employee
  #53 [size/xl]: Usage trends visualization
  ...

⚠️  Found 2 large tasks (size/l or size/xl)
   Will automatically split after adding to milestone

Add these 8 issues to milestone v0.4.0? (Y/n): y

📌 Adding issues to milestone...
  ✓ Added #50
  ✓ Added #51
  ...

📊 Updating project board...
  ✓ Moved #50 to Todo
  ✓ Moved #51 to Todo
  ...

✂️  Splitting large tasks...
  Splitting #51... ✓ Flagged for splitting
  Splitting #53... ✓ Flagged for splitting

✓ Large tasks flagged (manual splitting required)
  Tip: Use .claude/skills/github-task-manager to split tasks

📝 Creating milestone kickoff issue...
✓ Kickoff issue created

✅ Milestone v0.4.0 ready!

Summary:
  - Milestone created: v0.4.0
  - Issues added: 8
  - Todo status: 8 issues
  - Large tasks: 2 (flagged for splitting)

🎯 Start working on issues:
   gh issue list --milestone 'v0.4.0' --assignee ''

📊 View on project board:
   https://github.com/users/sergei-rastrigin/projects/3
```

### Phase 3: Split Large Tasks

**Purpose:** Break down size/xl and size/l tasks into manageable subtasks

```bash
# Split a large task manually
./scripts/split-large-tasks.sh --issue 51

# Or use auto-split with github-task-manager skill
./scripts/split-large-tasks.sh --issue 51 --auto
```

**What this does:**
1. Analyzes the large task
2. Provides guidance on task breakdown
3. Creates subtasks with parent-child relationship
4. Updates parent with subtask checklist
5. Removes `needs-splitting` label when complete

**Example workflow:**
```
✂️  Splitting Large Task: #51
==================================

📋 Fetching issue details...
✓ Issue: Analytics dashboard
  Size: size/l
  Milestone: v0.4.0

🧠 Task Breakdown Workflow:
-----------------------------------

1. Analyze the task and identify logical components
2. Break into subtasks (size/s or size/m each)
3. Ensure subtasks are independent and testable
4. Create sub-issues with parent-child relationship
5. Update parent issue with subtask checklist

📝 Manual Workflow:

Step 1: Identify subtasks (example breakdown)
  - Subtask 1: [Component A] - size/s
  - Subtask 2: [Component B] - size/m
  - Subtask 3: [Component C] - size/s

Step 2: Create sub-issues
  Run: ./scripts/create-sub-issue.sh --parent 51 --title "Subtask title" --size s

Step 3: Update parent with checklist
  Add to issue body:
  ## Subtasks
  - [ ] #XX - Subtask 1
  - [ ] #YY - Subtask 2
  - [ ] #ZZ - Subtask 3

Or use github-task-manager skill for automated workflow:
  .claude/skills/github-task-manager/SKILL.md

Create first subtask now? (y/N):
```

### Complete Milestone Transition Example

**Scenario:** Just released v0.3.0 (Web UI), starting v0.4.0 (Analytics)

```bash
# 1. Archive completed milestone
./scripts/archive-milestone.sh --milestone v0.3.0
# → Closes milestone, archives 15 issues

# 2. Start new milestone
./scripts/start-milestone.sh \
  --version v0.4.0 \
  --description "Analytics Dashboard & Approvals" \
  --due-date "2026-01-31" \
  --auto-split
# → Creates milestone, adds 8 backlog issues, flags 2 large tasks

# 3. Split large tasks (if flagged)
gh issue list --label "needs-splitting" --milestone "v0.4.0"
# → Shows #51 and #53 need splitting

./scripts/split-large-tasks.sh --issue 51 --auto
# → Uses github-task-manager to split into 3 subtasks

./scripts/split-large-tasks.sh --issue 53 --auto
# → Uses github-task-manager to split into 4 subtasks

# 4. Verify milestone ready
gh issue list --milestone "v0.4.0" --label "status/todo"
# → Shows 8 parent issues + 7 subtasks = 15 total tasks

# 5. Start working!
gh issue list --milestone "v0.4.0" --label "priority/p0" --assignee ""
# → Pick highest priority unassigned task
```

### Milestone Planning Best Practices

**Before Starting New Milestone:**
- ✅ Review backlog and update priorities
- ✅ Ensure issue descriptions are clear
- ✅ Verify all issues have size labels
- ✅ Check for dependencies between issues
- ✅ Set realistic due date (4-6 weeks typical)

**When Populating Milestone:**
- ✅ Focus on p0/p1 priority issues
- ✅ Aim for mix of sizes (not all large tasks)
- ✅ Balance features vs bug fixes vs tech debt
- ✅ Include testing and documentation tasks
- ✅ Leave buffer for unexpected work (70-80% capacity)

**Task Splitting Guidelines:**
- ✅ Each subtask should be independently testable
- ✅ Subtasks should be size/s or size/m (1-3 days each)
- ✅ Use parent-child relationship (blockedBy in GitHub)
- ✅ Update parent with checklist of subtasks
- ✅ Close parent only when all subtasks complete

### Integration with GitHub Project

The milestone transition scripts automatically update the GitHub Project board:

- **Archive script**: Leaves issues in "Done" status with "archived" label
- **Start script**: Moves new milestone issues to "Todo" status
- **Status updates**: Use `./scripts/update-project-status.sh` for manual updates

```bash
# Move issue to different status
./scripts/update-project-status.sh --issue 51 --status "In Progress"
./scripts/update-project-status.sh --issue 51 --status "Done"
```

### Troubleshooting

**Issue: Milestone has open issues when archiving**
- Review each open issue
- Close or move to next milestone as appropriate
- Script will warn and ask for confirmation

**Issue: Backlog query returns no results**
- Ensure issues have `priority/p0` or `priority/p1` labels
- Check that issues don't already have milestones
- Verify issues are in "open" state

**Issue: Large tasks not splitting automatically**
- Use `--auto` flag with split-large-tasks.sh
- Or use github-task-manager skill for AI-assisted splitting
- Manual splitting always available as fallback

### Related Scripts

- `archive-milestone.sh` - Archive completed milestone
- `start-milestone.sh` - Start new milestone
- `split-large-tasks.sh` - Split large tasks into subtasks
- `create-sub-issue.sh` - Create subtask with parent link
- `update-project-status.sh` - Update issue status on project board

### Related Skills

- `.claude/skills/github-task-manager/SKILL.md` - Task management and splitting
- `.claude/skills/development-workflow/SKILL.md` - Complete dev workflow

---

## Automation (Future)

**Potential automations to add:**

- Automated changelog generation from conventional commits
- Release notes template generation
- Version bump automation (read last tag, increment)
- Automated GitHub Release creation in CI/CD
- Release announcement to Slack/Discord
- Full milestone transition automation (archive → start → split)

## Examples

See `examples/release-examples.md` for complete real-world release workflows.

## Templates

See `templates/` directory for:
- `release-notes.md` - GitHub Release template
- `tag-message.md` - Git tag message template
- `changelog.md` - Changelog format template

---

**This skill ensures consistent, professional releases across all project phases.**
