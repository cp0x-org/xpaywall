-- name: CreateRequestLog :one
INSERT INTO request_logs (
    id, project_id, outbound_route_id, request_id, method, path,
    client_ip, user_agent, status,
    payment_required, payment_requested_at,
    payment_completed, payment_completed_at,
    payment_channel_id, payment_channel_asset_id,
    amount_paid, amount_usd,
    upstream_url, upstream_status_code, upstream_response_time_ms,
    final_status_code, error_type, error_message
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9,
    $10, $11,
    $12, $13,
    $14, $15,
    $16, $17,
    $18, $19, $20,
    $21, $22, $23
)
RETURNING *;

-- name: UpdateRequestLog :one
UPDATE request_logs
SET status                    = $2,
    outbound_route_id         = COALESCE($3, outbound_route_id),
    payment_required          = $4,
    payment_requested_at      = COALESCE($5, payment_requested_at),
    payment_completed         = $6,
    payment_completed_at      = COALESCE($7, payment_completed_at),
    payment_channel_id        = COALESCE($8, payment_channel_id),
    payment_channel_asset_id  = COALESCE($9, payment_channel_asset_id),
    amount_paid               = COALESCE($10, amount_paid),
    amount_usd                = COALESCE($11, amount_usd),
    upstream_url              = COALESCE($12, upstream_url),
    upstream_status_code      = COALESCE($13, upstream_status_code),
    upstream_response_time_ms = COALESCE($14, upstream_response_time_ms),
    final_status_code         = COALESCE($15, final_status_code),
    error_type                = COALESCE($16, error_type),
    error_message             = COALESCE($17, error_message),
    updated_at                = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetRequestLog :one
SELECT * FROM request_logs WHERE id = $1;

-- name: GetRequestLogByRequestID :one
SELECT * FROM request_logs WHERE request_id = $1;

-- name: ListRequestLogs :many
SELECT * FROM request_logs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListRequestLogsByProject :many
SELECT * FROM request_logs
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetRecentRequestsForDashboard :many
SELECT
    rl.id,
    rl.path,
    rl.method,
    rl.created_at,
    COALESCE(rl.final_status_code,
        CASE WHEN rl.payment_required = TRUE AND rl.payment_completed = FALSE THEN 402 ELSE 200 END
    )::INTEGER                AS status_code,
    pc.protocol               AS payment_channel,
    rl.amount_usd
FROM request_logs rl
LEFT JOIN payment_channels pc ON pc.id = rl.payment_channel_id
ORDER BY rl.created_at DESC
LIMIT 5;

-- name: GetRecentRequestsForDashboardByProject :many
SELECT
    rl.id,
    rl.path,
    rl.method,
    rl.created_at,
    COALESCE(rl.final_status_code,
        CASE WHEN rl.payment_required = TRUE AND rl.payment_completed = FALSE THEN 402 ELSE 200 END
    )::INTEGER                AS status_code,
    pc.protocol               AS payment_channel,
    rl.amount_usd
FROM request_logs rl
LEFT JOIN payment_channels pc ON pc.id = rl.payment_channel_id
WHERE rl.project_id = $1
ORDER BY rl.created_at DESC
LIMIT 5;

-- name: CreateRequestEvent :one
INSERT INTO request_events (id, request_log_id, event_type, metadata)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListRequestEventsByLog :many
SELECT * FROM request_events
WHERE request_log_id = $1
ORDER BY created_at ASC;
