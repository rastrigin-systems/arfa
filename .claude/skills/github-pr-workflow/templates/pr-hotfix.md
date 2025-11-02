# Hotfix Pull Request Template

Use this template for emergency production fixes (P0 bugs).

## Template

```markdown
## 🚨 HOTFIX - [Brief Description]

**Severity:** Critical (P0)
**Impact:** [Describe production impact]
**Affected Users:** [Number or percentage of users affected]

## Issue
[Describe the critical issue]

## Fix
[Explain the immediate fix applied]

## Root Cause
[If known, explain root cause. If not, note "Under investigation"]

## Changes
- [List minimal changes made]
- [Keep hotfixes as small as possible]

## Testing
- [ ] Bug reproduction confirmed
- [ ] Fix verified in staging
- [ ] Manual testing complete
- [ ] Rollback plan prepared

## Rollback Plan
[Describe how to rollback if issues arise]

## Follow-up Tasks
- [ ] Create issue for root cause analysis
- [ ] Create issue for additional tests
- [ ] Create issue for monitoring improvements

## Post-Deploy Verification
- [ ] Monitor error rates
- [ ] Check user reports
- [ ] Verify metrics returned to normal

---

Fixes #[ISSUE_NUMBER]

🚨 HOTFIX - Expedited merge required

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Usage Example

```bash
ISSUE_NUM=101
HOTFIX_TITLE="API returns 500 for all requests"

gh pr create \
  --title "fix: 🚨 HOTFIX - ${HOTFIX_TITLE} (#${ISSUE_NUM})" \
  --label "area/api,type/bug,priority/p0" \
  --body "$(cat <<EOF
## 🚨 HOTFIX - API Returns 500 for All Requests

**Severity:** Critical (P0)
**Impact:** Complete API outage - all endpoints returning 500
**Affected Users:** 100% of active users
**Duration:** Started at 14:23 UTC (15 minutes ago)

## Issue
After deploying v0.3.1, all API requests return 500 Internal Server Error due to database connection pool exhaustion.

## Fix
- Increased database connection pool max size from 10 to 50
- Added connection timeout of 5s (was infinite)
- Added connection health check before serving requests

## Root Cause
Recent traffic spike (3x normal) combined with slow queries caused connection pool to exhaust. Connections were not timing out, causing cascading failure.

## Changes
- Modified \`internal/db/connection.go\` pool configuration
- Added \`MaxOpenConns: 50\` (was 10)
- Added \`ConnMaxIdleTime: 5 * time.Minute\`
- Added startup health check in \`cmd/server/main.go\`

## Testing
- [x] Bug reproduction confirmed (exhausted pool locally)
- [x] Fix verified in staging environment
- [x] Manual testing complete (load tested with 100 concurrent requests)
- [x] Rollback plan prepared

## Rollback Plan
1. Revert to v0.3.0 commit: \`git revert HEAD && git push\`
2. Redeploy previous version: \`make deploy-production\`
3. Estimated rollback time: 2 minutes

## Follow-up Tasks
- [ ] Issue #102: Investigate slow queries causing connection buildup
- [ ] Issue #103: Add connection pool metrics to monitoring
- [ ] Issue #104: Add load testing to CI pipeline
- [ ] Issue #105: Review and optimize database query performance

## Post-Deploy Verification
- [ ] Monitor error rates (should drop to <0.1%)
- [ ] Check user reports (no new 500 error reports)
- [ ] Verify response times (<200ms p95)
- [ ] Monitor connection pool usage (<80%)

---

Fixes #${ISSUE_NUM}

🚨 HOTFIX - Expedited merge required

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"

# Request immediate review
gh pr review --request-reviewer @tech-lead

# Monitor CI closely
gh pr checks $(gh pr view --json number -q .number) --watch --interval 5
```

## Hotfix Best Practices

1. **Minimize Changes** - Only fix the immediate issue, nothing else
2. **Test in Staging First** - Always verify fix before production deploy
3. **Prepare Rollback** - Know exactly how to rollback if needed
4. **Create Follow-ups** - Log all follow-up work as separate issues
5. **Communicate** - Notify team immediately about hotfix and status
6. **Monitor Closely** - Watch production metrics after deploy
7. **Document** - Write thorough incident report after resolution

## Hotfix Workflow

```bash
# 1. Create hotfix branch from production
git checkout main
git pull
git checkout -b hotfix-critical-issue

# 2. Make MINIMAL fix
# (Only fix the immediate issue)

# 3. Test locally
make test

# 4. Create hotfix PR
gh pr create --title "fix: 🚨 HOTFIX - ..." --label "priority/p0"

# 5. Request immediate review
gh pr review --request-reviewer @tech-lead

# 6. Monitor CI (poll every 5s, not 10s)
gh pr checks $(gh pr view --json number -q .number) --watch --interval 5

# 7. Merge immediately when CI passes
gh pr merge --squash --delete-branch

# 8. Verify in production
# Check metrics, logs, user reports

# 9. Create follow-up issues
# Root cause analysis, additional tests, monitoring
```

## Communication Template

When deploying hotfix, notify team:

```
🚨 HOTFIX DEPLOYED 🚨

Issue: API returning 500 for all requests
PR: #123
Deploy Time: 14:45 UTC
Status: ✅ Fix deployed, monitoring
Impact: All API endpoints affected (15 min downtime)

Fix: Increased DB connection pool size

Post-Deploy:
✅ Error rate dropped to 0%
✅ Response times normal (<200ms)
✅ No new user reports

Follow-up: Issues #102-#105 created for root cause analysis

Questions? See PR #123 or ping @engineer
```
