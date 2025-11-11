# GitHub Release Notes Template

Use this template when creating GitHub Releases with `gh release create`.

## Minor Release Template

```markdown
## 🌟 Highlights

🎉 **[Major accomplishment summary]**

- ✅ [Key achievement 1]
- ✅ [Key achievement 2]
- ✅ [Key achievement 3]

## 📊 What's New

### Features
- **[Feature name]**: [Brief description]
- **[Feature name]**: [Brief description]
- **[Feature name]**: [Brief description]

### Improvements
- [Improvement description]
- [Improvement description]

### Bug Fixes
- Fix [issue description] ([#123](https://github.com/rastrigin-org/ubik-enterprise/issues/123))
- Fix [issue description] ([#124](https://github.com/rastrigin-org/ubik-enterprise/issues/124))

## 📈 Statistics

- **Commits:** X
- **New Features:** Y
- **Bug Fixes:** Z
- **Test Coverage:** W%
- **Contributors:** N

## 🔧 Technical Stack

[List key technologies, frameworks, libraries used]

## 🚀 Upgrade Guide

No breaking changes - upgrade by pulling latest code.

```bash
git pull
npm install  # if dependencies changed
```

## 📝 Full Changelog

[Compare view link]

https://github.com/rastrigin-org/ubik-enterprise/compare/vX.Y-1.0...vX.Y.0

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

## Patch Release Template

```markdown
## 🐛 Bug Fixes

Critical bug fixes for vX.Y.0:

- **[Component]**: Fix [issue description] ([#123](https://github.com/rastrigin-org/ubik-enterprise/issues/123))
- **[Component]**: Fix [issue description] ([#124](https://github.com/rastrigin-org/ubik-enterprise/issues/124))

## 📝 Changelog

[Compare view link]

https://github.com/rastrigin-org/ubik-enterprise/compare/vX.Y.0...vX.Y.Z

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

## Major Release Template

```markdown
## 🚀 MAJOR RELEASE

**vX.0.0** introduces breaking changes. Please review the upgrade guide carefully.

## ⚠️ Breaking Changes

- **[Area]**: [What changed and why]
- **[Area]**: [What changed and why]

## 🌟 New Features

- **[Major feature]**: [Description]
- **[Major feature]**: [Description]

## 📖 Upgrade Guide

### Prerequisites

- [Requirement 1]
- [Requirement 2]

### Migration Steps

1. **[Step 1]**: [Instructions]
2. **[Step 2]**: [Instructions]
3. **[Step 3]**: [Instructions]

### API Changes

#### Before (v(X-1).Y.Z)
```javascript
// Old API
```

#### After (vX.0.0)
```javascript
// New API
```

## 🐛 Bug Fixes

- [Bug fix]

## 📈 Statistics

- **Breaking Changes:** X
- **New Features:** Y
- **Bug Fixes:** Z

## 📝 Full Changelog

https://github.com/rastrigin-org/ubik-enterprise/compare/v(X-1).0.0...vX.0.0

---

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

## Usage Examples

### Create Minor Release

```bash
gh release create v0.4.0 \
  --title "Release v0.4.0 - Analytics Dashboard" \
  --notes "$(cat <<'EOF'
## 🌟 Highlights

🎉 **Analytics and cost tracking dashboard**

- ✅ Real-time usage analytics
- ✅ Cost tracking per employee
- ✅ Usage trends visualization

[... rest of template ...]
EOF
)"
```

### Create Patch Release

```bash
gh release create v0.3.1 \
  --title "Release v0.3.1 - Bug Fixes" \
  --notes "Critical bug fixes for v0.3.0. See details for list of fixes."
```

### Create Pre-release

```bash
gh release create v0.4.0-beta.1 \
  --title "Release v0.4.0-beta.1" \
  --notes "Beta release for testing" \
  --prerelease
```

## Tips

1. **Link to issues**: Use `#123` or full URLs
2. **Be specific**: Users want to know exactly what changed
3. **Include screenshots**: For UI changes
4. **Add migration guide**: For breaking changes
5. **Thank contributors**: Acknowledge community contributions
6. **Keep it scannable**: Use headers, lists, and emojis
