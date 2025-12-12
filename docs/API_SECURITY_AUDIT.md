# API Security Audit Report

**Date:** December 11, 2025
**Auditor:** Claude Code (Automated Security Audit)
**Scope:** All 39 REST API endpoints

---

## Executive Summary

A comprehensive security audit was performed on the Ubik Enterprise API. The audit identified **1 critical vulnerability** and **1 medium-severity issue**. The critical vulnerability has been fixed via PR #260.

| Severity | Count | Status |
|----------|-------|--------|
| 🔴 Critical | 1 | Fixed (PR #260) |
| 🟡 Medium | 1 | Open (#259) |
| 🟢 Low | 0 | - |

---

## Findings

### 🔴 CRITICAL: Teams API Missing Org Scoping (#258)

**Status:** ✅ Fixed in PR #260

**Affected Endpoints:**
- `PATCH /api/v1/teams/{team_id}` - UpdateTeam
- `DELETE /api/v1/teams/{team_id}` - DeleteTeam

**Description:**
The `UpdateTeam` and `DeleteTeam` handlers did not verify that the team belongs to the authenticated user's organization before performing operations. This allowed any authenticated user to modify or delete teams from any organization.

**Impact:**
- Data integrity compromise across organizations
- Multi-tenant isolation breach
- Potential compliance violations (GDPR, SOC2)

**Fix Applied:**
1. SQL queries updated to include `org_id` in WHERE clause
2. Handlers updated to extract and pass `org_id` from JWT context
3. Tests updated to verify org scoping is enforced

---

### 🟡 MEDIUM: Roles API Missing Admin Authorization (#259)

**Status:** 🔓 Open

**Affected Endpoints:**
- `POST /api/v1/roles` - CreateRole
- `PATCH /api/v1/roles/{role_id}` - UpdateRole
- `DELETE /api/v1/roles/{role_id}` - DeleteRole

**Description:**
Roles handlers have comments indicating "admin only" access but there is no actual authorization check enforcing this. Any authenticated user can create, modify, or delete system-wide roles.

**Impact:**
- Privilege escalation risk
- System-wide role modifications by non-admins

**Recommended Fix:**
Add role-based authorization middleware or checks to restrict these operations to administrators.

---

## Endpoints Audited

### Authentication (`/auth/*`) - ✅ Secure

| Endpoint | Auth | Org Scoping | Notes |
|----------|------|-------------|-------|
| POST /auth/login | Public | N/A | Generic error messages (prevents enumeration) |
| POST /auth/register | Public | N/A | Creates org + admin atomically |
| GET /auth/check-slug | Public | N/A | Read-only check |
| POST /auth/logout | JWT | ✅ | Session invalidation working |
| GET /auth/me | JWT | ✅ | Returns only authenticated user data |
| POST /auth/forgot-password | Public | N/A | Rate limited, generic responses |
| GET /auth/verify-reset-token | Public | N/A | Token-based validation |
| POST /auth/reset-password | Public | N/A | Token required |

### Employees (`/employees/*`) - ✅ Secure

| Endpoint | Auth | Org Scoping | Notes |
|----------|------|-------------|-------|
| GET /employees | JWT | ✅ | Properly scoped to org |
| POST /employees | JWT | ✅ | Creates in authenticated org |
| GET /employees/{id} | JWT | ✅ | 404 for other orgs (secure) |
| PATCH /employees/{id} | JWT | ✅ | Verifies org ownership |
| DELETE /employees/{id} | JWT | ✅ | Verifies org ownership |

### Teams (`/teams/*`) - ⚠️ Fixed

| Endpoint | Auth | Org Scoping | Notes |
|----------|------|-------------|-------|
| GET /teams | JWT | ✅ | Properly scoped |
| POST /teams | JWT | ✅ | Creates in authenticated org |
| GET /teams/{id} | JWT | ✅ | Properly scoped |
| PATCH /teams/{id} | JWT | ✅ **FIXED** | Was missing org check |
| DELETE /teams/{id} | JWT | ✅ **FIXED** | Was missing org check |

### Organizations (`/organizations/*`) - ✅ Secure

| Endpoint | Auth | Org Scoping | Notes |
|----------|------|-------------|-------|
| GET /organizations/current | JWT | ✅ | Returns only authenticated org |
| PATCH /organizations/current | JWT | ✅ | Updates only authenticated org |

### Agent Configs - ✅ Secure

All agent configuration endpoints properly verify organization membership through context.

### Invitations (`/invitations/*`) - ✅ Secure

| Endpoint | Auth | Org Scoping | Notes |
|----------|------|-------------|-------|
| GET /invitations | JWT | ✅ | Lists org invitations only |
| POST /invitations | JWT | ✅ | Rate limited (20/day) |
| GET /invitations/{token} | Public | Token-based | Secure token validation |
| POST /invitations/{token}/accept | Public | Token-based | Creates employee in correct org |
| DELETE /invitations/{id} | JWT | ✅ | Verifies org ownership |

### Roles (`/roles/*`) - ⚠️ Authorization Issue

| Endpoint | Auth | Org Scoping | Notes |
|----------|------|-------------|-------|
| GET /roles | JWT | System-wide | Read-only, acceptable |
| GET /roles/{id} | JWT | System-wide | Read-only, acceptable |
| POST /roles | JWT | System-wide | ⚠️ No admin check |
| PATCH /roles/{id} | JWT | System-wide | ⚠️ No admin check |
| DELETE /roles/{id} | JWT | System-wide | ⚠️ No admin check |

### Logs & Activity - ✅ Secure

All logging endpoints properly scope to the authenticated organization.

### Sync (`/sync/*`) - ✅ Secure

Employee-scoped sync endpoint properly extracts employee ID from JWT.

---

## Security Best Practices Observed

1. **JWT Token Management:** Proper token generation, validation, and session management
2. **Password Hashing:** Using bcrypt with appropriate cost factor
3. **Rate Limiting:** Password reset limited to 3 requests/hour
4. **Email Enumeration Prevention:** Generic messages for login/reset flows
5. **Secure Token Generation:** 256-bit cryptographically random tokens for invitations
6. **HTTPS Only:** CORS configured for production domains

---

## Recommendations

### Immediate (P0)
- [x] Fix teams org scoping vulnerability (PR #260)

### Short-term (P1)
- [ ] Add admin role authorization for roles endpoints (#259)
- [ ] Implement audit logging for sensitive operations
- [ ] Add request validation middleware (input sanitization)

### Long-term (P2)
- [ ] Implement API rate limiting across all endpoints
- [ ] Add CSRF protection for cookie-based authentication
- [ ] Implement IP-based blocking for suspicious activity
- [ ] Add security headers middleware (CSP, X-Frame-Options, etc.)

---

## Files Modified

| File | Change |
|------|--------|
| `platform/database/sqlc/queries/organizations.sql` | Added org_id to UpdateTeam/DeleteTeam queries |
| `services/api/internal/handlers/teams.go` | Added org verification to handlers |
| `services/api/internal/handlers/teams_test.go` | Updated tests for org scoping |

---

## Conclusion

The Ubik Enterprise API demonstrates strong security foundations with proper JWT authentication, multi-tenant isolation (with the fixed teams endpoints), and secure password handling. The identified vulnerabilities have been addressed or documented for remediation.

---

*This report was generated automatically as part of the security audit process.*
