-- name: CreateInvitation :one
-- Create a new invitation with a secure token
-- Used by POST /invitations (admin only)
INSERT INTO invitations (
    org_id,
    inviter_id,
    email,
    role_id,
    team_id,
    token,
    status,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, 'pending', NOW() + INTERVAL '7 days'
) RETURNING *;

-- name: ListInvitations :many
-- List all invitations for an organization with pagination
-- Used by GET /invitations (admin only)
SELECT * FROM invitations
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountInvitations :one
-- Count total invitations for an organization (for pagination)
-- Used by GET /invitations (admin only)
SELECT COUNT(*) FROM invitations
WHERE org_id = $1;

-- name: GetInvitationByToken :one
-- Get invitation details by token (for validation)
-- Used by GET /invitations/{token} (public)
SELECT
    i.*,
    o.name as org_name,
    r.name as role_name,
    t.name as team_name
FROM invitations i
JOIN organizations o ON i.org_id = o.id
JOIN roles r ON i.role_id = r.id
LEFT JOIN teams t ON i.team_id = t.id
WHERE i.token = $1;

-- name: GetInvitationByID :one
-- Get invitation by ID (for cancellation)
-- Used by DELETE /invitations/{id} (admin only)
SELECT * FROM invitations
WHERE id = $1 AND org_id = $2;

-- name: AcceptInvitation :one
-- Accept an invitation (updates status and records acceptance)
-- Used by POST /invitations/{token}/accept
-- Note: Must be called within a transaction that also creates the employee
UPDATE invitations
SET
    status = 'accepted',
    accepted_at = NOW(),
    accepted_by = $2,
    updated_at = NOW()
WHERE token = $1 AND status = 'pending' AND expires_at > NOW()
RETURNING *;

-- name: CancelInvitation :exec
-- Cancel a pending invitation
-- Used by DELETE /invitations/{id} (admin only)
UPDATE invitations
SET
    status = 'cancelled',
    updated_at = NOW()
WHERE id = $1 AND org_id = $2 AND status = 'pending';

-- name: CountInvitationsByOrgToday :one
-- Count invitations created today for rate limiting (20/day)
-- Used by POST /invitations for rate limiting validation
SELECT COUNT(*) FROM invitations
WHERE org_id = $1
  AND created_at >= CURRENT_DATE
  AND created_at < CURRENT_DATE + INTERVAL '1 day';

-- name: ExpireOldInvitations :exec
-- Mark expired invitations (background job)
-- Used by scheduled cleanup task
UPDATE invitations
SET status = 'expired', updated_at = NOW()
WHERE status = 'pending' AND expires_at <= NOW();
