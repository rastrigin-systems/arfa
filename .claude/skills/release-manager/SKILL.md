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

### 8. Close Milestone

```bash
# Close completed milestone (via web UI)
# https://github.com/sergei-rastrigin/ubik-enterprise/milestones
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

## Automation (Future)

**Potential automations to add:**

- Automated changelog generation from conventional commits
- Release notes template generation
- Version bump automation (read last tag, increment)
- Automated GitHub Release creation in CI/CD
- Release announcement to Slack/Discord

## Examples

See `examples/release-examples.md` for complete real-world release workflows.

## Templates

See `templates/` directory for:
- `release-notes.md` - GitHub Release template
- `tag-message.md` - Git tag message template
- `changelog.md` - Changelog format template

---

**This skill ensures consistent, professional releases across all project phases.**
