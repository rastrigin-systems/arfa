# Standard Pull Request Template

Use this template for feature PRs, refactoring, and general changes.

## Template

```markdown
## Summary
[Brief description of what this PR does and why]

## Changes
- [List key changes]
- [Be specific and concise]
- [Focus on "what" changed, not "how"]

## Testing
- [ ] Unit tests passing
- [ ] Integration tests passing
- [ ] Manual testing complete
- [ ] No breaking changes

## Screenshots
[If UI changes, add before/after screenshots]

## Additional Notes
[Any important context for reviewers]
[Known limitations or follow-up work needed]

---

Closes #[ISSUE_NUMBER]

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Usage Example

```bash
ISSUE_NUM=123
ISSUE_TITLE="Implement JWT authentication"

gh pr create \
  --title "feat: ${ISSUE_TITLE} (#${ISSUE_NUM})" \
  --label "area/api" \
  --body "$(cat <<EOF
## Summary
Implements JWT-based authentication for API endpoints as described in #${ISSUE_NUM}.

## Changes
- Added JWT generation on successful login
- Implemented JWT validation middleware
- Added token expiration handling (24h)
- Updated authentication tests

## Testing
- [x] Unit tests passing (98% coverage)
- [x] Integration tests passing
- [x] Manual testing complete
- [x] No breaking changes

## Additional Notes
- Token refresh functionality will be added in #124
- Database schema unchanged

---

Closes #${ISSUE_NUM}

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

## Tips

1. **Be Specific in Changes** - List actual code changes, not vague descriptions
2. **Check All Boxes** - Only check boxes that are truly complete
3. **Add Screenshots** - For any UI changes, always include screenshots
4. **Link Issues** - Use `Closes #123` to auto-close related issues
5. **Add Context** - Help reviewers understand "why" not just "what"
