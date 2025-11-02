# Bug Fix Pull Request Template

Use this template for bug fixes.

## Template

```markdown
## Bug Description
[Brief description of the bug that was fixed]

## Root Cause
[Explain what was causing the bug]

## Fix
[Explain how the fix resolves the issue]

## Changes
- [List specific code changes]
- [Be precise about what was modified]

## Testing
- [ ] Bug reproduction test added
- [ ] Existing tests still pass
- [ ] Manual verification complete
- [ ] Edge cases tested

## Regression Risk
[Low/Medium/High] - [Explain why]

## Additional Notes
[Any important context about the fix]

---

Fixes #[ISSUE_NUMBER]

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Usage Example

```bash
ISSUE_NUM=87
BUG_TITLE="Login fails with valid credentials"

gh pr create \
  --title "fix: ${BUG_TITLE} (#${ISSUE_NUM})" \
  --label "area/api,type/bug" \
  --body "$(cat <<EOF
## Bug Description
Login endpoint returned 401 Unauthorized even with valid credentials.

## Root Cause
Password comparison was using incorrect bcrypt salt rounds due to configuration mismatch between seed data (10 rounds) and authentication code (12 rounds expected).

## Fix
- Updated seed data to use consistent bcrypt rounds (12)
- Added validation to ensure bcrypt configuration matches across app
- Added test to verify password hashing consistency

## Changes
- Modified \`schema/seed.sql\` to use bcrypt with 12 rounds
- Updated \`internal/auth/password.go\` validation
- Added \`TestPasswordHashingConsistency\` integration test

## Testing
- [x] Bug reproduction test added
- [x] Existing tests still pass
- [x] Manual verification complete (login works!)
- [x] Edge cases tested (wrong password still fails)

## Regression Risk
Low - Fix only affects authentication, well-tested area

## Additional Notes
This issue only affected development environment with seed data. Production users were not impacted as they have properly hashed passwords from registration flow.

---

Fixes #${ISSUE_NUM}

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

## Tips for Bug Fix PRs

1. **Add Reproduction Test** - Always add a test that would fail before the fix
2. **Explain Root Cause** - Help reviewers understand what went wrong
3. **Assess Regression Risk** - Be honest about potential side effects
4. **Test Edge Cases** - Bug fixes often expose related edge cases
5. **Document Impact** - Note if bug affected production or just dev/staging
