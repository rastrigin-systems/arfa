# Task Template

Use this template when creating standard GitHub issues.

## Usage

```bash
gh issue create \
  --title "ACTION: Clear description of what needs to be done" \
  --label "type/TYPENAME,area/AREANAME,priority/PRIORITYLEVEL,size/SIZE" \
  --milestone "vX.Y.Z" \
  --body "$(cat <<'EOF'
## Overview

[Brief summary of what needs to be done and why]

## Problem Statement

[What problem does this solve? What pain point does it address?]

## Proposed Solution

[High-level approach to solving the problem]

### Implementation Details
- [ ] Detail 1
- [ ] Detail 2
- [ ] Detail 3

## Acceptance Criteria

- [ ] Criterion 1 - [Specific, measurable outcome]
- [ ] Criterion 2 - [Specific, measurable outcome]
- [ ] Criterion 3 - [Specific, measurable outcome]

## Technical Notes

[Any technical considerations, constraints, or dependencies]

## Dependencies

[List any issues this depends on or blocks]
- Depends on: #XXX
- Blocks: #YYY

## References

[Links to relevant docs, discussions, or related issues]
- [Doc/URL]
- Related: #ZZZ
EOF
)"
```

## Example: Feature Task

```bash
gh issue create \
  --title "Implement GET /api/v1/employees endpoint" \
  --label "type/feature,area/api,priority/p1,size/m" \
  --milestone "v0.3.0" \
  --body "$(cat <<'EOF'
## Overview

Add API endpoint for listing employees with filtering and pagination support.

## Problem Statement

The web UI needs to display a list of employees for the organization with search and filter capabilities.

## Proposed Solution

Create a new GET endpoint following the existing API patterns:
- URL: `/api/v1/employees`
- Method: GET
- Auth: Required (JWT)
- Multi-tenancy: Scoped to org_id

### Implementation Details
- [ ] Add OpenAPI spec definition
- [ ] Create SQL query with filters (status, role, team)
- [ ] Implement handler with pagination
- [ ] Add unit tests (TDD)
- [ ] Add integration tests
- [ ] Update API documentation

## Acceptance Criteria

- [ ] Endpoint returns paginated list of employees
- [ ] Filtering works for status, role, and team
- [ ] Responses are properly scoped to organization
- [ ] Test coverage > 85%
- [ ] API docs updated

## Technical Notes

- Use existing pagination pattern from agents endpoint
- Ensure RLS policies enforce org scoping
- Return employee details but not sensitive fields (password hash, etc.)

## Dependencies

- Depends on: #45 (Employee schema migration)
- Blocks: #67 (Employee management UI)

## References

- [API Design Guide](../../docs/API_DESIGN.md)
- Related: #34 (Organization endpoints)
EOF
)"
```

## Example: Bug Task

```bash
gh issue create \
  --title "Fix: JWT token expiration not being validated" \
  --label "type/bug,area/api,priority/p0" \
  --body "$(cat <<'EOF'
## Overview

JWT tokens are not being validated for expiration, allowing expired tokens to access protected endpoints.

## Problem Statement

Security vulnerability - expired JWT tokens are accepted as valid, allowing unauthorized access.

## Proposed Solution

Add expiration validation to JWT middleware:
- Check `exp` claim in token
- Return 401 if expired
- Add test coverage for expired tokens

### Implementation Details
- [ ] Add expiration check to auth/jwt.go
- [ ] Write unit test for expired token
- [ ] Write integration test for expired token access
- [ ] Update error messages
- [ ] Document token lifetime in API docs

## Acceptance Criteria

- [ ] Expired tokens return 401 Unauthorized
- [ ] Error message is clear and helpful
- [ ] Test coverage includes expiration cases
- [ ] No regression in valid token handling

## Technical Notes

- Current token lifetime: 24 hours
- Need to test edge cases around clock skew
- Consider adding refresh token flow (future enhancement)

## Dependencies

None - critical security fix

## References

- JWT RFC: https://tools.ietf.org/html/rfc7519
- Related: #23 (Authentication system)
EOF
)"
```

## Label Reference

### Required Labels (Pick One from Each Category)

**Type:**
- `type/feature` - New functionality
- `type/bug` - Something isn't working
- `type/chore` - Maintenance/tooling
- `type/refactor` - Code improvement
- `type/research` - Investigation/spike
- `type/epic` - Large multi-issue feature

**Area:**
- `area/api` - Backend API
- `area/cli` - CLI client
- `area/web` - Web dashboard
- `area/db` - Database/schema
- `area/infra` - Infrastructure
- `area/testing` - Test infrastructure
- `area/docs` - Documentation

**Priority:**
- `priority/p0` - Critical (revenue/security)
- `priority/p1` - High (significant impact)
- `priority/p2` - Medium (nice to have)
- `priority/p3` - Low (future/speculative)

**Size (Recommended):**
- `size/xs` - < 2 hours
- `size/s` - 2-4 hours
- `size/m` - 1-2 days
- `size/l` - 3-5 days
- `size/xl` - > 1 week (should split!)

## Best Practices

1. **Clear Titles**: Use action verbs (Implement, Fix, Add, Update)
2. **Complete Description**: Include all sections
3. **Specific Criteria**: Make acceptance criteria measurable
4. **Link Dependencies**: Always link related issues
5. **Assign Milestone**: If target release is known
6. **Add to Project**: Use `gh project item-add`
7. **Set Initial Status**: Use update-project-status.sh
