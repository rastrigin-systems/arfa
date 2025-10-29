-- name: CreateAgentRequest :one
INSERT INTO agent_requests (
    employee_id,
    request_type,
    request_data,
    status,
    reason
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetAgentRequest :one
SELECT * FROM agent_requests
WHERE id = $1;

-- name: ListAgentRequests :many
SELECT * FROM agent_requests
WHERE ($1::varchar IS NULL OR status = $1)
  AND ($2::uuid IS NULL OR employee_id = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountAgentRequests :one
SELECT COUNT(*) FROM agent_requests
WHERE ($1::varchar IS NULL OR status = $1)
  AND ($2::uuid IS NULL OR employee_id = $2);

-- name: UpdateAgentRequestStatus :one
UPDATE agent_requests
SET
    status = $2,
    resolved_at = CASE WHEN $2 IN ('approved', 'rejected', 'cancelled') THEN NOW() ELSE NULL END
WHERE id = $1
RETURNING *;

-- name: CountPendingRequestsByOrg :one
SELECT COUNT(*) as pending_count
FROM agent_requests ar
JOIN employees e ON ar.employee_id = e.id
WHERE e.org_id = $1
  AND ar.status = 'pending';
