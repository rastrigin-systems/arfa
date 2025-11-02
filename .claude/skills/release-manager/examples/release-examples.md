# Release Manager - Real-World Examples

Complete examples of release workflows for different scenarios.

## Example 1: Minor Release (New Milestone Features)

**Scenario:** v0.3.0 Web UI milestone is complete. Ready to release v0.3.0 properly.

```bash
# 1. Verify all checks
echo "=== Pre-Release Checklist ==="

# Check CI status
echo "CI Status:"
gh run list --limit 1 --json conclusion,status -q '.[0] | "\(.status): \(.conclusion)"'

# Check milestone completion
echo -e "\nOpen issues in v0.3.0:"
gh issue list --milestone "v0.3.0" --state open --json number,title

# Run all tests
echo -e "\nRunning tests..."
make test
cd services/web && npm test && cd ../..

# Verify on main and clean
git checkout main
git pull
git status

# 2. Analyze changes since v0.2.0
echo "=== Changes Since v0.2.0 ==="
LAST_TAG="v0.2.0"
NEW_TAG="v0.3.0"

# Count commits
COMMIT_COUNT=$(git log $LAST_TAG..HEAD --oneline --no-merges | wc -l | tr -d ' ')
echo "Total commits: $COMMIT_COUNT"

# Features
echo -e "\n## Features:"
git log $LAST_TAG..HEAD --oneline --no-merges | grep -E "^[a-f0-9]+ feat:"

# Bug fixes
echo -e "\n## Bug Fixes:"
git log $LAST_TAG..HEAD --oneline --no-merges | grep -E "^[a-f0-9]+ fix:"

# 3. Create git tag
git tag -a $NEW_TAG -m "Release v0.3.0 - Complete Web UI

🎉 MAJOR MILESTONE: Production-ready web interface with 11 pages!

## 🌟 Highlights

- ✅ Complete Next.js 14 web UI with authentication
- ✅ Dark/Light mode theming
- ✅ Comprehensive E2E test suite with Playwright
- ✅ MSW API mocking for reliable tests
- ✅ GitHub workflow automation skills

## 📊 New Features

### Web UI (11 Pages)
- Login page with JWT authentication
- Dashboard home page
- Agent catalog page
- Settings/Agent configuration page
- Employee list and detail pages
- Employee creation/editing forms
- Employee agent overrides management
- Team agent assignment UI
- Organization agent configuration pages (3 tabs)

### Testing Infrastructure
- Playwright E2E tests with MSW mocking
- Headless/headed test modes
- CI/CD optimizations (caching, parallelization)
- 24 passing E2E tests

### Developer Experience
- GitHub Task Manager skill
- Development workflow skills
- Release Manager skill
- Comprehensive agent configs

## 🐛 Bug Fixes

- Fixed API client for server-side/client-side env vars
- Fixed MSW integration for E2E tests
- Fixed component loading race conditions
- Fixed employee creation test response unmarshaling
- Fixed health check endpoint

## 📈 Statistics

- $COMMIT_COUNT commits since v0.2.0
- 11 web UI pages implemented
- 24+ E2E tests
- 3 new GitHub workflow skills
- 100% CI/CD pass rate (Build + Test)

## 🔧 Technical Details

- Next.js 14 with App Router
- Tailwind CSS + shadcn/ui components
- Playwright for E2E testing
- MSW for API mocking
- OpenAPI-generated TypeScript types
- JWT authentication with session management

## 🎯 What's Next (v0.4.0)

- Approval workflow UI
- Analytics dashboard
- Cost tracking
- Performance optimizations

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# 4. Push tag
git push origin $NEW_TAG

# 5. Create GitHub Release
gh release create $NEW_TAG \
  --title "Release v0.3.0 - Complete Web UI" \
  --notes "$(cat <<'EOF'
## 🌟 Highlights

🎉 **MAJOR MILESTONE:** Production-ready web interface!

- ✅ Complete Next.js 14 web UI with 11 pages
- ✅ Dark/Light mode theming
- ✅ Comprehensive E2E test suite
- ✅ GitHub workflow automation

## 📊 What's New

### Web UI Pages
1. **Login** - JWT authentication
2. **Dashboard** - Home page
3. **Agent Catalog** - Browse available agents
4. **Settings** - Agent configurations
5-7. **Employees** - List, detail, create/edit
8. **Agent Overrides** - Per-employee customization
9. **Team Assignments** - Team agent management
10-11. **Org Config** - Organization-level settings

### Testing
- 24 Playwright E2E tests
- MSW API mocking
- CI/CD optimized (cached, parallel)

### Developer Tools
- GitHub Task Manager skill
- Release Manager skill
- Development workflow skills

## 🐛 Bug Fixes

- API client environment variable support
- MSW integration
- Component loading issues
- Test response parsing

## 📈 Statistics

- 69 commits since v0.2.0
- 11 web pages
- 24+ E2E tests
- 3 workflow skills

## 🔧 Stack

- Next.js 14 + App Router
- Tailwind CSS + shadcn/ui
- Playwright + MSW
- TypeScript + OpenAPI

## 📝 Full Changelog

https://github.com/sergei-rastrigin/ubik-enterprise/compare/v0.2.0...v0.3.0

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"

# 6. Update documentation
echo "## v0.3.0 - Complete Web UI (2025-11-02)

**Major milestone:** Production-ready web interface with 11 pages

- Complete Next.js 14 web UI
- Dark/Light mode theming
- 24 Playwright E2E tests
- GitHub workflow skills
- 69 commits since v0.2.0

[Release Notes](https://github.com/sergei-rastrigin/ubik-enterprise/releases/tag/v0.3.0)
" >> docs/RELEASES.md

git add docs/RELEASES.md CLAUDE.md
git commit -m "docs: Update for v0.3.0 release

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"
git push

# 7. Announce
gh issue comment 49 --body "🎉 Release management system completed and v0.3.0 released!"

echo "✅ Release v0.3.0 complete!"
```

## Example 2: Patch Release (Bug Fixes)

**Scenario:** Critical bug found in v0.3.0, need to release v0.3.1 quickly.

```bash
VERSION="v0.3.1"
PREV_VERSION="v0.3.0"

# 1. Quick checks (skip milestone check for patches)
git checkout main
git pull
make test

# 2. Generate bug fix changelog
echo "Bug fixes:"
git log $PREV_VERSION..HEAD --oneline --no-merges | grep -E "^[a-f0-9]+ fix:"

# 3. Create patch tag (shorter message for patches)
git tag -a $VERSION -m "Release v0.3.1 - Bug Fixes

## 🐛 Bug Fixes

- Fix login redirect loop in Safari (#123)
- Fix agent config validation error (#124)
- Fix employee list pagination (#125)

Patch release with critical bug fixes for v0.3.0.

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>"

# 4. Push and release
git push origin $VERSION

gh release create $VERSION \
  --title "Release v0.3.1 - Bug Fixes" \
  --notes "Critical bug fixes for v0.3.0.

## 🐛 Fixes

- Login redirect loop in Safari
- Agent config validation
- Employee list pagination

[Full Changelog](https://github.com/sergei-rastrigin/ubik-enterprise/compare/v0.3.0...v0.3.1)"

echo "✅ Patch release v0.3.1 complete!"
```

## Example 3: Moving an Existing Tag

**Scenario:** v0.3.0 tag points to wrong commit (monorepo migration), need to move it to Web UI completion.

```bash
OLD_TAG="v0.3.0"
TARGET_COMMIT="8e7f02e"  # Current HEAD with Web UI

# 1. Verify current state
echo "Current tag points to:"
git show $OLD_TAG --no-patch --format="%H %s"

echo -e "\nTarget commit:"
git show $TARGET_COMMIT --no-patch --format="%H %s"

# 2. Delete local tag
git tag -d $OLD_TAG
echo "✓ Deleted local tag"

# 3. Delete remote tag
git push origin :refs/tags/$OLD_TAG
echo "✓ Deleted remote tag"

# 4. Create new tag at target commit
git checkout $TARGET_COMMIT
git tag -a $OLD_TAG -m "Release v0.3.0 - Complete Web UI

(Tag moved from monorepo migration to Web UI completion)

🎉 MAJOR MILESTONE: Production-ready web interface!

[Full release notes - see examples above]"

# 5. Force push new tag
git push origin $OLD_TAG --force
echo "✓ Pushed new tag"

# 6. Update GitHub Release (if exists)
gh release delete $OLD_TAG --yes
gh release create $OLD_TAG \
  --title "Release v0.3.0 - Complete Web UI" \
  --notes "[See Example 1 for full notes]"

echo "✅ Tag successfully moved to Web UI completion!"
```

## Example 4: Querying Release Information

**Scenario:** Need to understand what's in the current release and what's coming next.

```bash
# Current version
CURRENT=$(git describe --tags --abbrev=0)
echo "Current release: $CURRENT"

# Latest commits not yet released
echo -e "\n=== Unreleased Changes ==="
git log $CURRENT..HEAD --oneline --no-merges

# Statistics
echo -e "\n=== Stats ==="
echo "Commits since $CURRENT: $(git log $CURRENT..HEAD --oneline --no-merges | wc -l)"
echo "Features: $(git log $CURRENT..HEAD --oneline --no-merges | grep 'feat:' | wc -l)"
echo "Bug fixes: $(git log $CURRENT..HEAD --oneline --no-merges | grep 'fix:' | wc -l)"

# Compare two versions
echo -e "\n=== Compare v0.2.0 and v0.3.0 ==="
git log v0.2.0..v0.3.0 --oneline | head -10

# View release details
echo -e "\n=== GitHub Releases ==="
gh release list --limit 5

# View specific release
gh release view v0.3.0
```

## Example 5: Pre-Release Checklist

**Scenario:** About to release v0.4.0, want to verify everything is ready.

```bash
VERSION="v0.4.0"
MILESTONE="v0.4.0"

echo "=== Release Readiness Checklist for $VERSION ==="

# 1. CI Status
echo -n "✓ CI Status: "
gh run list --limit 1 --json conclusion -q '.[0].conclusion'

# 2. Test Status
echo -n "✓ Running tests... "
if make test > /dev/null 2>&1 && cd services/web && npm test > /dev/null 2>&1; then
  echo "PASSED"
else
  echo "FAILED ❌"
  exit 1
fi
cd ../..

# 3. Milestone Completion
OPEN_ISSUES=$(gh issue list --milestone "$MILESTONE" --state open --json number -q 'length')
echo "✓ Open issues in $MILESTONE: $OPEN_ISSUES"
if [ "$OPEN_ISSUES" -gt 0 ]; then
  echo "  ⚠️  Warning: Milestone not complete!"
  gh issue list --milestone "$MILESTONE" --state open
fi

# 4. Branch Status
BRANCH=$(git branch --show-current)
echo "✓ Current branch: $BRANCH"
if [ "$BRANCH" != "main" ]; then
  echo "  ❌ Not on main branch!"
  exit 1
fi

# 5. Working Tree
if git diff-index --quiet HEAD --; then
  echo "✓ Working tree: Clean"
else
  echo "  ❌ Uncommitted changes!"
  git status --short
  exit 1
fi

# 6. Changelog Preview
echo -e "\n=== Changelog Preview ==="
LAST_TAG=$(git describe --tags --abbrev=0)
git log $LAST_TAG..HEAD --oneline --no-merges | head -10
echo "..."
echo "Total: $(git log $LAST_TAG..HEAD --oneline --no-merges | wc -l) commits"

echo -e "\n✅ Ready to release $VERSION!"
echo "Next steps:"
echo "  1. git tag -a $VERSION -m '...'"
echo "  2. git push origin $VERSION"
echo "  3. gh release create $VERSION ..."
```

## Example 6: Automated Changelog Generation

**Scenario:** Generate structured changelog from commit history using conventional commits.

```bash
LAST_TAG=$(git describe --tags --abbrev=0)
NEW_TAG="v0.4.0"

echo "# Changelog - $NEW_TAG"
echo ""
echo "Release Date: $(date +%Y-%m-%d)"
echo ""

# Features
FEATURES=$(git log $LAST_TAG..HEAD --oneline --no-merges | grep "^[a-f0-9]* feat")
if [ ! -z "$FEATURES" ]; then
  echo "## 🚀 Features"
  echo ""
  echo "$FEATURES" | sed 's/^[a-f0-9]* feat: /- /'
  echo ""
fi

# Bug Fixes
FIXES=$(git log $LAST_TAG..HEAD --oneline --no-merges | grep "^[a-f0-9]* fix")
if [ ! -z "$FIXES" ]; then
  echo "## 🐛 Bug Fixes"
  echo ""
  echo "$FIXES" | sed 's/^[a-f0-9]* fix: /- /'
  echo ""
fi

# Chores
CHORES=$(git log $LAST_TAG..HEAD --oneline --no-merges | grep "^[a-f0-9]* chore")
if [ ! -z "$CHORES" ]; then
  echo "## 🔧 Chores"
  echo ""
  echo "$CHORES" | sed 's/^[a-f0-9]* chore: /- /'
  echo ""
fi

# Statistics
echo "## 📊 Statistics"
echo ""
echo "- **Commits:** $(git log $LAST_TAG..HEAD --oneline --no-merges | wc -l | tr -d ' ')"
echo "- **Contributors:** $(git log $LAST_TAG..HEAD --format='%an' | sort -u | wc -l | tr -d ' ')"
echo "- **Files Changed:** $(git diff --stat $LAST_TAG..HEAD | tail -1 | awk '{print $1}')"
echo ""

echo "---"
echo ""
echo "**Full Changelog:** https://github.com/sergei-rastrigin/ubik-enterprise/compare/$LAST_TAG...$NEW_TAG"
```

## Best Practices Demonstrated

1. **Always verify before releasing** - Check CI, tests, milestone
2. **Use annotated tags** - Include detailed messages
3. **Follow semver strictly** - Users expect predictable versioning
4. **Link to issues** - Use #123 syntax for traceability
5. **Generate changelogs** - Don't write manually
6. **Document everything** - Update RELEASES.md and CLAUDE.md
7. **Announce releases** - Comment on related issues

## Common Pitfalls

❌ **Don't:**
- Release with failing tests
- Skip milestone completion check
- Forget to update documentation
- Use lightweight tags (git tag without -a)
- Release from feature branches

✅ **Do:**
- Always test before tagging
- Write detailed release notes
- Follow conventional commit format
- Keep RELEASES.md updated
- Release from main branch only

## Example 7: Complete Milestone Transition (Release + Archive + Start New)

**Scenario:** v0.3.0 is released and complete. Now archive it and start v0.4.0 milestone.

```bash
# ========================================
# PART 1: Release v0.3.0 (if not done yet)
# ========================================

VERSION="v0.3.0"
PREV_VERSION="v0.2.0"

# Pre-release checks
gh run list --limit 1
gh issue list --milestone "$VERSION" --state open
make test && cd services/web && npm test && cd ../..

# Create tag and release
git tag -a $VERSION -m "Release v0.3.0 - Complete Web UI

🎉 Production-ready web interface with 11 pages!

[... full release notes ...]"

git push origin $VERSION

gh release create $VERSION \
  --title "Release v0.3.0 - Complete Web UI" \
  --notes "..."

echo "✅ v0.3.0 released!"

# ========================================
# PART 2: Archive Completed Milestone
# ========================================

echo -e "\n=== Archiving v0.3.0 Milestone ==="

# Run archive script
./scripts/archive-milestone.sh --milestone v0.3.0

# What this does:
# - Labels all 15 issues as "archived"
# - Closes any remaining open issues
# - Closes the milestone
# - Updates docs/MILESTONES_ARCHIVE.md

echo "✅ v0.3.0 milestone archived!"

# Verify archival
gh issue list --label "archived" --milestone "v0.3.0" | wc -l
# Should show 15 issues

# Check milestone is closed
gh api /repos/sergei-rastrigin/ubik-enterprise/milestones \
  --jq '.[] | select(.title=="v0.3.0") | .state'
# Should output: "closed"

# ========================================
# PART 3: Start New Milestone (v0.4.0)
# ========================================

echo -e "\n=== Starting v0.4.0 Milestone ==="

# Review and update backlog priorities first
gh issue list --label "priority/p0,priority/p1" --state open --json number,title

# Start new milestone
./scripts/start-milestone.sh \
  --version v0.4.0 \
  --description "Analytics Dashboard & Approval Workflows" \
  --due-date "2026-01-31" \
  --auto-split

# What this does:
# 1. Creates milestone v0.4.0 with due date
# 2. Queries backlog for p0/p1 issues (finds 8 issues)
# 3. Shows issues for review
# 4. Adds confirmed issues to milestone
# 5. Moves issues to "Todo" status
# 6. Flags large tasks (#51, #53) for splitting
# 7. Creates kickoff issue

echo "✅ v0.4.0 milestone created with 8 issues!"

# ========================================
# PART 4: Split Large Tasks
# ========================================

echo -e "\n=== Splitting Large Tasks ==="

# Find tasks flagged for splitting
LARGE_TASKS=$(gh issue list \
  --label "needs-splitting" \
  --milestone "v0.4.0" \
  --json number -q '.[].number')

echo "Large tasks to split: $LARGE_TASKS"

# Split each large task
for issue in $LARGE_TASKS; do
  echo -e "\nSplitting issue #$issue..."

  # Option 1: Manual split with guidance
  ./scripts/split-large-tasks.sh --issue $issue

  # Option 2: Auto-split with github-task-manager
  # ./scripts/split-large-tasks.sh --issue $issue --auto
done

# Example: Manually split issue #51 (Analytics Dashboard)
# This creates 3 subtasks:
#   - #54: Analytics data API endpoint (size/m)
#   - #55: Dashboard UI components (size/s)
#   - #56: Charts and visualizations (size/m)

# Update parent issue #51 with subtask checklist
gh issue edit 51 --body "# Analytics Dashboard

Complete analytics dashboard for tracking agent usage.

## Subtasks

- [ ] #54 - Analytics data API endpoint
- [ ] #55 - Dashboard UI components
- [ ] #56 - Charts and visualizations

## Acceptance Criteria

- [ ] Real-time usage metrics
- [ ] Cost tracking per employee
- [ ] Filterable date ranges
- [ ] Exportable reports

---

Part of milestone v0.4.0

🤖 Generated with [Claude Code](https://claude.com/claude-code)"

# Remove needs-splitting label
gh issue edit 51 --remove-label "needs-splitting"

echo "✅ Large tasks split!"

# ========================================
# PART 5: Verify New Milestone Ready
# ========================================

echo -e "\n=== Verifying v0.4.0 Milestone ==="

# Count total issues (parent + subtasks)
TOTAL_ISSUES=$(gh issue list --milestone "v0.4.0" --json number -q 'length')
echo "Total issues in v0.4.0: $TOTAL_ISSUES"

# Check issue breakdown
echo -e "\nIssue breakdown:"
gh issue list --milestone "v0.4.0" --json number,title,labels \
  --jq 'group_by(.labels[] | select(.name | startswith("size/")) | .name) |
        map({size: .[0].labels[] | select(.name | startswith("size/")) | .name,
             count: length}) | .[]'

# Show issues by priority
echo -e "\nPriority breakdown:"
echo "P0 issues:"
gh issue list --milestone "v0.4.0" --label "priority/p0" --json number,title -q '.[] | "#\(.number): \(.title)"'
echo -e "\nP1 issues:"
gh issue list --milestone "v0.4.0" --label "priority/p1" --json number,title -q '.[] | "#\(.number): \(.title)"'

# View on project board
echo -e "\n📊 View on project board:"
echo "https://github.com/users/sergei-rastrigin/projects/3"

# ========================================
# PART 6: Start Working on v0.4.0
# ========================================

echo -e "\n=== Ready to Start v0.4.0! ==="

# Pick first task (highest priority, unassigned)
FIRST_TASK=$(gh issue list \
  --milestone "v0.4.0" \
  --label "priority/p0" \
  --assignee "" \
  --limit 1 \
  --json number -q '.[0].number')

echo "Recommended first task: #$FIRST_TASK"
gh issue view $FIRST_TASK

# Create feature branch and start work
git checkout -b issue-$FIRST_TASK-feature
./scripts/update-project-status.sh --issue $FIRST_TASK --status "In Progress"

echo "✅ Started work on #$FIRST_TASK!"
echo ""
echo "📋 Milestone transition complete!"
echo "   v0.3.0: Released and archived ✓"
echo "   v0.4.0: Created with $TOTAL_ISSUES issues ✓"
echo "   Ready to build! 🚀"
```

### Milestone Transition Timeline

```
Day 1 (Release Day):
  09:00 - Final pre-release checks
  10:00 - Create tag and GitHub Release
  11:00 - Announce release
  14:00 - Archive completed milestone (v0.3.0)

Day 2 (Planning Day):
  09:00 - Review backlog and update priorities
  10:00 - Start new milestone (v0.4.0)
  11:00 - Review and confirm issue selection
  14:00 - Split large tasks
  16:00 - Verify milestone ready

Day 3 (Sprint Start):
  09:00 - Team picks first tasks
  10:00 - Development begins on v0.4.0
```

### Key Metrics to Track

```bash
# After milestone transition, track these metrics:

# 1. Milestone velocity (issues per day)
START_DATE=$(gh api /repos/sergei-rastrigin/ubik-enterprise/milestones \
  --jq '.[] | select(.title=="v0.4.0") | .created_at')
DAYS_ELAPSED=$(( ($(date +%s) - $(date -j -f "%Y-%m-%dT%H:%M:%SZ" "$START_DATE" +%s)) / 86400 ))
COMPLETED=$(gh issue list --milestone "v0.4.0" --state closed --json number -q 'length')
echo "Velocity: $(echo "scale=2; $COMPLETED / $DAYS_ELAPSED" | bc) issues/day"

# 2. Burndown (remaining work)
REMAINING=$(gh issue list --milestone "v0.4.0" --state open --json number -q 'length')
echo "Remaining: $REMAINING issues"

# 3. Size distribution
echo "Size distribution:"
gh issue list --milestone "v0.4.0" --state open --json labels \
  --jq '[.[] | .labels[] | select(.name | startswith("size/")) | .name] |
        group_by(.) | map({size: .[0], count: length}) | .[]'
```

---

**These examples cover 95% of release scenarios you'll encounter.**
