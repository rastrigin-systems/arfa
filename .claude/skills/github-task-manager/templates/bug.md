# Bug Report Template

Use this template when reporting bugs or defects.

## Usage

```bash
gh issue create \
  --title "Bug: Clear description of the issue" \
  --label "type/bug,area/AREANAME,priority/PRIORITYLEVEL" \
  --body "$(cat <<'EOF'
## Bug Description

[Clear, concise description of what's wrong]

## Steps to Reproduce

1. [First step]
2. [Second step]
3. [Third step]

## Expected Behavior

[What should happen]

## Actual Behavior

[What actually happens]

## Environment

- **Version**: [e.g., v0.2.0, commit sha, or "main branch"]
- **OS**: [e.g., macOS 13.4, Ubuntu 22.04, Windows 11]
- **Component**: [API Server, CLI Client, Web UI]
- **Database**: [PostgreSQL version if relevant]
- **Browser**: [If web issue - Chrome 115, Firefox 116, etc.]

## Logs/Screenshots

[Paste relevant logs or attach screenshots]

```
[error logs here]
```

## Impact

- **Severity**: [Critical/High/Medium/Low]
- **Affected Users**: [All users/Specific org/Single user]
- **Workaround**: [Is there a workaround? If yes, describe it]

## Additional Context

[Any other relevant information]

## Possible Root Cause

[If known or suspected]

## Related Issues

[Link to related bugs or features]
- Related: #XXX
- Duplicate of: #YYY (if duplicate, close this one)
EOF
)"
```

## Example: API Bug

```bash
gh issue create \
  --title "Bug: JWT token expiration not being validated" \
  --label "type/bug,area/api,priority/p0" \
  --body "$(cat <<'EOF'
## Bug Description

Expired JWT tokens are being accepted as valid, allowing unauthorized access to protected endpoints.

## Steps to Reproduce

1. Obtain a valid JWT token
2. Wait for token to expire (24 hours + 1 second)
3. Make request to protected endpoint with expired token
4. Request succeeds when it should return 401

```bash
# Get token
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token')

# Wait 24+ hours or manually set token exp to past

# This should fail but succeeds:
curl http://localhost:8080/api/v1/employees \
  -H "Authorization: Bearer $TOKEN"
# Returns 200 OK with employee list
```

## Expected Behavior

Expired tokens should be rejected with:
- HTTP Status: 401 Unauthorized
- Error message: "Token has expired"
- Headers: `WWW-Authenticate: Bearer error="invalid_token", error_description="Token has expired"`

## Actual Behavior

Expired tokens are accepted as valid and requests succeed with 200 OK.

## Environment

- **Version**: v0.2.0 (commit 6f57ba5)
- **OS**: macOS 13.4
- **Component**: API Server
- **Database**: PostgreSQL 15.3
- **Browser**: N/A (API only)

## Logs/Screenshots

Server logs show no expiration validation:

```
[INFO] 2025-11-02T12:00:00Z Incoming request: GET /api/v1/employees
[INFO] 2025-11-02T12:00:00Z JWT validation passed
[INFO] 2025-11-02T12:00:00Z Response: 200 OK
```

## Impact

- **Severity**: Critical (Security vulnerability)
- **Affected Users**: All users with expired tokens
- **Workaround**: None - security issue requires immediate fix

## Additional Context

Discovered during security audit. The JWT middleware in `services/api/internal/middleware/auth.go` validates the signature but doesn't check the `exp` claim.

## Possible Root Cause

Missing expiration check in `validateJWT()` function:

```go
// services/api/internal/middleware/auth.go:45
func (m *JWTMiddleware) validateJWT(tokenString string) (*jwt.Token, error) {
    token, err := jwt.Parse(tokenString, m.keyFunc)
    if err != nil {
        return nil, err
    }

    // BUG: Missing expiration validation here
    // Should check token.Claims.(jwt.MapClaims)["exp"]

    return token, nil
}
```

## Related Issues

- Related: #23 (Authentication system implementation)
- Blocks: #45 (v0.2.1 security release)
EOF
)"
```

## Example: CLI Bug

```bash
gh issue create \
  --title "Bug: `ubik sync` fails with Docker permission error on Linux" \
  --label "type/bug,area/cli,priority/p1" \
  --body "$(cat <<'EOF'
## Bug Description

The `ubik sync --start-containers` command fails on Linux systems when Docker daemon socket permissions are restrictive.

## Steps to Reproduce

1. Install ubik CLI on Linux (Ubuntu 22.04)
2. Ensure Docker is running
3. Run `ubik sync --start-containers`
4. Command fails with permission error

```bash
$ ubik sync --start-containers
Fetching agent configurations...
✓ Retrieved 2 agent configs
Starting Docker containers...
✗ Error: failed to connect to Docker daemon: permission denied

Error: exit status 1
```

## Expected Behavior

CLI should:
1. Detect Docker socket permissions issue
2. Provide helpful error message suggesting solutions:
   - Add user to docker group
   - Use sudo
   - Check Docker daemon is running

## Actual Behavior

CLI shows generic "permission denied" error without guidance.

## Environment

- **Version**: v0.2.0
- **OS**: Ubuntu 22.04 LTS
- **Component**: CLI Client
- **Docker**: Docker 24.0.5
- **User**: Non-root user, not in docker group

## Logs/Screenshots

Full error trace:

```
$ ubik sync --start-containers -v
[DEBUG] Config loaded from: /home/user/.ubik/config.json
[DEBUG] API endpoint: https://api.ubik.dev
[DEBUG] Fetching configs for org: 550e8400-e29b-41d4-a716-446655440000
[INFO] Retrieved 2 agent configurations
[DEBUG] Initializing Docker client at: unix:///var/run/docker.sock
[ERROR] Docker client error: permission denied while trying to connect to Docker daemon socket

Error: failed to connect to Docker daemon: permission denied
```

## Impact

- **Severity**: High (Blocks CLI usage on Linux)
- **Affected Users**: Linux users not in docker group (~60% of CLI users)
- **Workaround**: Add user to docker group: `sudo usermod -aG docker $USER` (requires logout)

## Additional Context

This is a common issue on Linux where Docker socket is owned by root:docker with 660 permissions. macOS and Windows Docker Desktop handle this differently.

## Possible Root Cause

Docker client initialization doesn't check socket permissions before attempting connection:

```go
// services/cli/internal/docker/client.go:32
func NewDockerClient() (*DockerClient, error) {
    cli, err := client.NewClientWithOpts(client.FromEnv)
    if err != nil {
        // Generic error - doesn't distinguish permission vs daemon not running
        return nil, fmt.Errorf("failed to connect to Docker daemon: %w", err)
    }
    return &DockerClient{cli: cli}, nil
}
```

Should add pre-flight check for socket permissions and provide actionable error.

## Related Issues

- Related: #34 (CLI Docker integration)
- Similar: #56 (Docker networking issues)
EOF
)"
```

## Example: Web UI Bug

```bash
gh issue create \
  --title "Bug: Employee form validation allows invalid email format" \
  --label "type/bug,area/web,priority/p2" \
  --body "$(cat <<'EOF'
## Bug Description

The employee creation form accepts invalid email addresses that don't match RFC 5322 format, leading to validation errors on the backend.

## Steps to Reproduce

1. Navigate to /employees/new
2. Fill in form with invalid email: `user@` (missing domain)
3. Click "Create Employee"
4. Form submits to API
5. API returns 400 Bad Request
6. User sees generic error message

## Expected Behavior

1. Frontend validation should catch invalid email before submission
2. Show inline validation error: "Please enter a valid email address"
3. Prevent form submission until valid
4. Match backend validation rules

## Actual Behavior

1. Form accepts invalid email
2. Submits to API
3. API rejects with "invalid email format"
4. User sees generic error: "Failed to create employee"

## Environment

- **Version**: v0.2.0
- **OS**: macOS 13.4
- **Component**: Web UI
- **Browser**: Chrome 115.0.5790.110

## Logs/Screenshots

![Invalid email accepted](screenshot-invalid-email.png)

Browser console shows API error:

```
POST /api/v1/employees 400 Bad Request
{
  "error": "validation_error",
  "message": "invalid email format",
  "field": "email"
}
```

## Impact

- **Severity**: Medium (Poor UX but not blocking)
- **Affected Users**: All users creating employees
- **Workaround**: Users learn to enter valid emails after first error

## Additional Context

Current frontend validation uses simple regex that doesn't match backend validation. Backend uses Go's `mail.ParseAddress()` which is RFC 5322 compliant.

Examples that should be invalid but are accepted:
- `user@` (missing domain)
- `@example.com` (missing local part)
- `user @example.com` (space in local part)

## Possible Root Cause

Frontend email validation is too permissive:

```typescript
// services/web/src/components/EmployeeForm.tsx:45
const emailRegex = /^[^\s@]+@[^\s@]+$/;  // Too simple!

if (!emailRegex.test(email)) {
  setError('email', { message: 'Invalid email' });
}
```

Should use a more robust regex or HTML5 input type="email" + custom validation.

## Related Issues

- Related: #12 (Employee management UI)
- Similar: #67 (Form validation consistency)
EOF
)"
```

## Priority Guidelines

### P0 - Critical
- Security vulnerabilities
- Data loss or corruption
- Complete system outage
- Revenue-blocking bugs

### P1 - High
- Feature completely broken
- Affects many users
- No workaround available
- Significant UX degradation

### P2 - Medium
- Feature partially broken
- Affects some users
- Workaround exists
- Minor UX issues

### P3 - Low
- Cosmetic issues
- Rare edge cases
- Easy workarounds available
- Nice-to-have fixes

## Best Practices

1. **Clear Title**: Start with "Bug:" and be specific
2. **Reproducible Steps**: Make it easy to reproduce
3. **Evidence**: Include logs, screenshots, or error traces
4. **Environment**: Specify exact versions and configuration
5. **Impact Assessment**: Help prioritize with severity and affected users
6. **Root Cause**: If you can identify it, share your analysis
7. **Workarounds**: Document any temporary solutions
8. **Link Related**: Connect to related bugs or the original feature

## After Reporting

1. **Add to Project**: `gh project item-add 3 --owner sergei-rastrigin --url ISSUE_URL`
2. **Set Priority**: Based on impact and severity
3. **Assign**: If you know who should fix it
4. **Notify**: If critical, notify team immediately
