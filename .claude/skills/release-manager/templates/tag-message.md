# Git Tag Message Template

Use this template when creating annotated git tags.

## Minor Release (New Features)

```
Release vX.Y.0 - [Milestone Name]

🎉 [One-line summary of major accomplishment]

## 🌟 Highlights

- ✅ [Major feature/change 1]
- ✅ [Major feature/change 2]
- ✅ [Major feature/change 3]

## 📊 New Features

### [Category 1]
- [Feature description]
- [Feature description]

### [Category 2]
- [Feature description]
- [Feature description]

## 🐛 Bug Fixes

- [Bug fix description] (#issue-number)
- [Bug fix description] (#issue-number)

## 📈 Statistics

- X commits since [last version]
- Y new features
- Z bug fixes
- W% test coverage

## 🔧 Technical Details

[Architecture changes, library updates, dependencies, etc.]

## 🎯 What's Next (vX.Y+1.0)

[Preview of next milestone features]

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Patch Release (Bug Fixes Only)

```
Release vX.Y.Z - Bug Fixes

## 🐛 Bug Fixes

- [Bug fix description] (#issue-number)
- [Bug fix description] (#issue-number)
- [Bug fix description] (#issue-number)

Patch release with critical bug fixes for vX.Y.0.

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Major Release (Breaking Changes)

```
Release vX.0.0 - [Major Milestone Name]

🚀 MAJOR RELEASE with breaking changes!

## ⚠️ Breaking Changes

- [Breaking change description - what changed]
- [Breaking change description - migration needed]

See upgrade guide below.

## 🌟 New Features

- [Major new feature]
- [Major new feature]

## 🐛 Bug Fixes

- [Bug fix]

## 📖 Upgrade Guide

### Migrating from v(X-1).Y.Z

1. [Step 1 - what to change]
2. [Step 2 - what to update]
3. [Step 3 - how to test]

### API Changes

**Before:**
```
[Old API example]
```

**After:**
```
[New API example]
```

## 📈 Statistics

- X commits since v(X-1).0.0
- Y breaking changes
- Z new features

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Usage

```bash
# Minor release
git tag -a v0.4.0 -m "$(cat .claude/skills/release-manager/templates/tag-message.md | sed -n '/## Minor Release/,/## Patch Release/p' | sed '1d;$d')"

# Patch release
git tag -a v0.3.1 -m "$(cat .claude/skills/release-manager/templates/tag-message.md | sed -n '/## Patch Release/,/## Major Release/p' | sed '1d;$d')"
```

Or copy and fill in manually:

```bash
git tag -a v0.4.0 -m "Release v0.4.0 - [Your Content Here]"
```
