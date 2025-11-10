# Task: Implement Password Reset Backend (Issue #161)

## Objective
Implement the backend foundation for the password reset (forgot password) flow, including database schema, OpenAPI spec updates, and 3 API endpoints with comprehensive tests.

## Work Location
**All changes must be made in:** `/Users/sergeirastrigin/Projects/ubik-issue-161`

## Phase 1: Database & OpenAPI Spec (DO THIS FIRST)

### 1. Add `password_reset_tokens` table to schema

**File:** `shared/schema/schema.sql`

Add this table after the `sessions` table (around line 82):

```sql
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '1 hour'),
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_employee_id ON password_reset_tokens(employee_id);
```

### 2. Add 3 endpoints to OpenAPI spec

**File:** `shared/openapi/spec.yaml`

Add these 3 endpoints under the `auth` tag:

#### Endpoint 1: POST /auth/forgot-password

```yaml
  /auth/forgot-password:
    post:
      tags: [auth]
      summary: Request password reset link
      description: Sends a password reset email to the provided address (if it exists). Always returns success to prevent email enumeration.
      operationId: forgotPassword
      security: []  # Public endpoint
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ForgotPasswordRequest'
      responses:
        '200':
          description: Success (generic message for security)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ForgotPasswordResponse'
        '429':
          description: Too many requests (rate limited)
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
```

#### Endpoint 2: GET /auth/verify-reset-token

```yaml
  /auth/verify-reset-token:
    get:
      tags: [auth]
      summary: Verify password reset token
      description: Checks if a password reset token is valid (not expired, not used)
      operationId: verifyResetToken
      security: []  # Public endpoint
      parameters:
        - name: token
          in: query
          required: true
          schema:
            type: string
          description: The password reset token to verify
      responses:
        '200':
          description: Token is valid
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/VerifyResetTokenResponse'
        '400':
          description: Token invalid, expired, or already used
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
```

#### Endpoint 3: POST /auth/reset-password

```yaml
  /auth/reset-password:
    post:
      tags: [auth]
      summary: Reset password using token
      description: Updates the employee's password using a valid reset token. Marks token as used.
      operationId: resetPassword
      security: []  # Public endpoint
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ResetPasswordRequest'
      responses:
        '200':
          description: Password reset successful
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ResetPasswordResponse'
        '400':
          description: Invalid token or validation error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
```

#### Add schemas to `components.schemas`

```yaml
    ForgotPasswordRequest:
      type: object
      required:
        - email
      properties:
        email:
          type: string
          format: email
          example: "alice@acme.com"

    ForgotPasswordResponse:
      type: object
      required:
        - message
      properties:
        message:
          type: string
          example: "If an account exists with this email, you will receive a password reset link within a few minutes."

    VerifyResetTokenResponse:
      type: object
      required:
        - valid
      properties:
        valid:
          type: boolean
          example: true

    ResetPasswordRequest:
      type: object
      required:
        - token
        - new_password
      properties:
        token:
          type: string
          example: "abc123xyz456..."
        new_password:
          type: string
          format: password
          minLength: 8
          example: "NewSecurePass123!"

    ResetPasswordResponse:
      type: object
      required:
        - message
      properties:
        message:
          type: string
          example: "Password reset successful"
```

### 3. Run code generation

After updating schema and spec:

```bash
cd /Users/sergeirastrigin/Projects/ubik-issue-161
make generate
```

This will:
- Generate database types from schema
- Generate API types from OpenAPI spec
- Update ERD documentation

## Phase 2: SQL Queries (AFTER code generation)

**File:** `sqlc/queries/password_reset_tokens.sql`

Create new file with these queries:

```sql
-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (
    employee_id,
    token,
    expires_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetPasswordResetToken :one
SELECT * FROM password_reset_tokens
WHERE token = $1
  AND expires_at > NOW()
  AND used_at IS NULL
LIMIT 1;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens
SET used_at = NOW()
WHERE token = $1;

-- name: GetEmployeeByEmail :one
SELECT * FROM employees
WHERE email = $1
LIMIT 1;

-- name: UpdateEmployeePassword :exec
UPDATE employees
SET password_hash = $1,
    updated_at = NOW()
WHERE id = $2;

-- name: CountRecentPasswordResetRequests :one
SELECT COUNT(*) FROM password_reset_tokens
WHERE employee_id = $1
  AND created_at > NOW() - INTERVAL '1 hour';
```

Run `make generate-db` after creating this file.

## Phase 3: Backend Implementation (TDD - Tests First!)

### Required Components

1. **Token Generator** (`pkg/auth/token_generator.go`)
   - Function: `GenerateSecureToken() (string, error)`
   - Use `crypto/rand` for 256-bit random tokens
   - Base64 encode the result
   - Write tests FIRST

2. **Rate Limiter** (`pkg/auth/rate_limiter.go`)
   - Function: `CheckPasswordResetRateLimit(employeeID uuid.UUID, db *database.Queries) error`
   - Max 3 requests per employee per hour
   - Write tests FIRST

3. **Mock Email Service** (`pkg/email/mock_service.go`)
   - Interface: `EmailService` with `SendPasswordResetEmail(email, token string) error`
   - Mock implementation logs to console
   - Easy to swap with real service later
   - Write tests FIRST

4. **API Handlers** (`services/api/handlers/auth_handlers.go`)
   - `ForgotPassword(w http.ResponseWriter, r *http.Request)`
   - `VerifyResetToken(w http.ResponseWriter, r *http.Request)`
   - `ResetPassword(w http.ResponseWriter, r *http.Request)`
   - Write integration tests FIRST

### TDD Workflow (MANDATORY)

For EACH component:
1. ✅ Write failing tests first
2. ✅ Implement minimal code to pass tests
3. ✅ Refactor with tests passing
4. ✅ Verify 85%+ coverage

### Security Requirements

**Must Implement:**
- ✅ 256-bit cryptographically secure tokens (crypto/rand)
- ✅ Generic success message (no email enumeration)
- ✅ 1-hour expiration enforced
- ✅ Single-use tokens (check used_at)
- ✅ Rate limiting: 3 requests per email per hour
- ✅ Bcrypt password hashing (cost factor 12)

### Test Coverage Requirements

**Unit Tests:**
- Token generation (cryptographic randomness, uniqueness)
- Rate limiter (3 requests limit, time window)
- Password validation (bcrypt, strength)

**Integration Tests:**
- POST /auth/forgot-password (success, non-existent email, rate limited)
- GET /auth/verify-reset-token (valid, expired, used, invalid)
- POST /auth/reset-password (success, invalid token, weak password)

**Target: 85%+ coverage**

## Reference Documentation

- **Wireframes:**
  - `docs/wireframes/epic-1-authentication/1.4-forgot-password.md`
  - `docs/wireframes/epic-1-authentication/1.4-reset-password.md`
- **User Story:** `docs/user-stories/epic-1-authentication/1.4-password-reset.md`
- **Testing Guide:** `docs/TESTING.md`
- **Database Guide:** `docs/DATABASE.md`

## Success Criteria

- ✅ All tests passing (85%+ coverage)
- ✅ All security requirements enforced
- ✅ Schema and spec updated
- ✅ Code generated and committed
- ✅ Rate limiting working
- ✅ Mock email service logging correctly
- ✅ Ready for frontend integration

## Deliverables Checklist

- [ ] `password_reset_tokens` table added to schema.sql
- [ ] 3 endpoints added to OpenAPI spec.yaml
- [ ] SQL queries in `sqlc/queries/password_reset_tokens.sql`
- [ ] Code generation run successfully
- [ ] Token generator implemented with tests
- [ ] Rate limiter implemented with tests
- [ ] Mock email service implemented with tests
- [ ] 3 API handlers implemented with integration tests
- [ ] All tests passing (85%+ coverage)
- [ ] Security requirements verified

## Notes

- Work ONLY in `/Users/sergeirastrigin/Projects/ubik-issue-161`
- Follow strict TDD: tests first, then implementation
- Use existing auth patterns from login/register handlers
- Reference existing tests for integration test structure
- Mock email service for MVP (structure for easy swap later)

Please implement the backend foundation following this plan, and report back when all backend tests are passing.
