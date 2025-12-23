-- name: ListWebhookDestinations :many
-- List all webhook destinations for an organization
SELECT
    id,
    org_id,
    name,
    url,
    auth_type,
    -- Note: auth_config contains secrets, handled in application layer
    event_types,
    event_filter,
    enabled,
    batch_size,
    timeout_ms,
    retry_max,
    retry_backoff_ms,
    created_by,
    created_at,
    updated_at
FROM webhook_destinations
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: GetWebhookDestination :one
-- Get a specific webhook destination by ID
SELECT
    id,
    org_id,
    name,
    url,
    auth_type,
    auth_config,
    event_types,
    event_filter,
    enabled,
    batch_size,
    timeout_ms,
    retry_max,
    retry_backoff_ms,
    signing_secret,
    created_by,
    created_at,
    updated_at
FROM webhook_destinations
WHERE id = $1 AND org_id = $2;

-- name: GetWebhookDestinationByName :one
-- Get a webhook destination by name within an organization
SELECT
    id,
    org_id,
    name,
    url,
    auth_type,
    auth_config,
    event_types,
    event_filter,
    enabled,
    batch_size,
    timeout_ms,
    retry_max,
    retry_backoff_ms,
    signing_secret,
    created_by,
    created_at,
    updated_at
FROM webhook_destinations
WHERE org_id = $1 AND name = $2;

-- name: CreateWebhookDestination :one
-- Create a new webhook destination
INSERT INTO webhook_destinations (
    org_id,
    name,
    url,
    auth_type,
    auth_config,
    event_types,
    event_filter,
    enabled,
    batch_size,
    timeout_ms,
    retry_max,
    retry_backoff_ms,
    signing_secret,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
) RETURNING *;

-- name: UpdateWebhookDestination :one
-- Update a webhook destination
UPDATE webhook_destinations SET
    name = COALESCE(sqlc.narg(name), name),
    url = COALESCE(sqlc.narg(url), url),
    auth_type = COALESCE(sqlc.narg(auth_type), auth_type),
    auth_config = COALESCE(sqlc.narg(auth_config), auth_config),
    event_types = COALESCE(sqlc.narg(event_types), event_types),
    event_filter = COALESCE(sqlc.narg(event_filter), event_filter),
    enabled = COALESCE(sqlc.narg(enabled), enabled),
    batch_size = COALESCE(sqlc.narg(batch_size), batch_size),
    timeout_ms = COALESCE(sqlc.narg(timeout_ms), timeout_ms),
    retry_max = COALESCE(sqlc.narg(retry_max), retry_max),
    retry_backoff_ms = COALESCE(sqlc.narg(retry_backoff_ms), retry_backoff_ms),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND org_id = sqlc.arg(org_id)
RETURNING *;

-- name: DeleteWebhookDestination :exec
-- Delete a webhook destination
DELETE FROM webhook_destinations
WHERE id = $1 AND org_id = $2;

-- name: EnableWebhookDestination :exec
-- Enable a webhook destination
UPDATE webhook_destinations
SET enabled = true, updated_at = NOW()
WHERE id = $1 AND org_id = $2;

-- name: DisableWebhookDestination :exec
-- Disable a webhook destination
UPDATE webhook_destinations
SET enabled = false, updated_at = NOW()
WHERE id = $1 AND org_id = $2;

-- name: RotateSigningSecret :one
-- Rotate the signing secret for a webhook destination
UPDATE webhook_destinations
SET signing_secret = $3, updated_at = NOW()
WHERE id = $1 AND org_id = $2
RETURNING id, signing_secret;

-- name: ListEnabledDestinations :many
-- List all enabled webhook destinations (for forwarder)
SELECT
    id,
    org_id,
    name,
    url,
    auth_type,
    auth_config,
    event_types,
    event_filter,
    batch_size,
    timeout_ms,
    retry_max,
    retry_backoff_ms,
    signing_secret
FROM webhook_destinations
WHERE enabled = true;

-- ============================================================================
-- WEBHOOK DELIVERIES
-- ============================================================================

-- name: CreateWebhookDelivery :one
-- Create a new delivery record
INSERT INTO webhook_deliveries (
    destination_id,
    log_id,
    status,
    next_retry_at
) VALUES (
    $1, $2, 'pending', NOW()
) RETURNING *;

-- name: GetPendingDeliveries :many
-- Get pending deliveries ready for processing
SELECT
    d.id,
    d.destination_id,
    d.log_id,
    d.status,
    d.attempts,
    d.next_retry_at
FROM webhook_deliveries d
WHERE d.status IN ('pending', 'failed')
    AND d.next_retry_at <= NOW()
ORDER BY d.next_retry_at ASC
LIMIT $1;

-- name: MarkDeliverySuccess :exec
-- Mark a delivery as successful
UPDATE webhook_deliveries SET
    status = 'delivered',
    attempts = attempts + 1,
    last_attempt_at = NOW(),
    delivered_at = NOW(),
    response_status = $2,
    response_body = $3,
    error_message = NULL
WHERE id = $1;

-- name: MarkDeliveryFailed :exec
-- Mark a delivery as failed (will retry)
UPDATE webhook_deliveries SET
    status = CASE WHEN attempts + 1 >= $4 THEN 'dead' ELSE 'failed' END,
    attempts = attempts + 1,
    last_attempt_at = NOW(),
    next_retry_at = NOW() + ($5 * POWER(2, attempts))::interval,
    response_status = $2,
    response_body = $3,
    error_message = $6
WHERE id = $1;

-- name: ListDeliveriesByDestination :many
-- List delivery history for a destination
SELECT
    id,
    destination_id,
    log_id,
    status,
    attempts,
    last_attempt_at,
    next_retry_at,
    response_status,
    response_body,
    error_message,
    created_at,
    delivered_at
FROM webhook_deliveries
WHERE destination_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountDeliveriesByStatus :many
-- Count deliveries by status for a destination
SELECT status, COUNT(*) as count
FROM webhook_deliveries
WHERE destination_id = $1
GROUP BY status;

-- name: GetUndeliveredLogs :many
-- Get logs that haven't been delivered to a destination yet
SELECT l.id
FROM activity_logs l
LEFT JOIN webhook_deliveries d ON d.log_id = l.id AND d.destination_id = $1
WHERE l.org_id = $2
    AND d.id IS NULL
    AND l.created_at > $3
ORDER BY l.created_at ASC
LIMIT $4;

-- name: DeleteOldDeliveries :exec
-- Delete old completed delivery records
DELETE FROM webhook_deliveries
WHERE status = 'delivered' AND delivered_at < $1;
