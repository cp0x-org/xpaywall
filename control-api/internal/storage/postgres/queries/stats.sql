-- name: GetTopRoutesForDashboard :many
SELECT
    r.path_pattern,
    r.price_usd,
    COUNT(rl.id)::BIGINT                                                                        AS total_requests,
    COALESCE(SUM(rl.amount_usd) FILTER (WHERE rl.payment_completed = TRUE), 0)::FLOAT8         AS revenue_usd
FROM routes r
JOIN request_logs rl ON rl.outbound_route_id = r.id
GROUP BY r.id, r.path_pattern, r.price_usd
ORDER BY total_requests DESC
LIMIT 5;

-- name: GetTopRoutesForDashboardByProject :many
SELECT
    r.path_pattern,
    r.price_usd,
    COUNT(rl.id)::BIGINT                                                                        AS total_requests,
    COALESCE(SUM(rl.amount_usd) FILTER (WHERE rl.payment_completed = TRUE), 0)::FLOAT8         AS revenue_usd
FROM routes r
JOIN request_logs rl ON rl.outbound_route_id = r.id
WHERE r.project_id = $1
GROUP BY r.id, r.path_pattern, r.price_usd
ORDER BY total_requests DESC
LIMIT 5;

-- name: GetDailyStats :many
-- Note: for sub-48h periods the handler runs GetHourlyStats (raw query) instead.
SELECT
    DATE(created_at)                                                        AS day,
    COUNT(*)::BIGINT                                                        AS total_requests,
    COALESCE(SUM(amount_usd) FILTER (WHERE payment_completed = TRUE), 0)::FLOAT8 AS total_earnings_usd
FROM request_logs
WHERE created_at >= $1
  AND created_at <  $2
GROUP BY DATE(created_at)
ORDER BY DATE(created_at);

-- name: GetDashboardStats :one
SELECT
    -- current period
    (SELECT COUNT(*) FROM projects WHERE projects.created_at >= @period_start AND projects.created_at < @period_end)::BIGINT AS total_projects,
    (SELECT COUNT(*) FROM routes   WHERE routes.created_at   >= @period_start AND routes.created_at   < @period_end)::BIGINT AS total_routes,
    (COUNT(*) FILTER (WHERE request_logs.created_at >= @period_start AND request_logs.created_at < @period_end))::BIGINT     AS total_requests,
    COALESCE(
        SUM(request_logs.amount_usd) FILTER (
            WHERE request_logs.payment_required = TRUE AND request_logs.payment_completed = TRUE
              AND request_logs.created_at >= @period_start AND request_logs.created_at < @period_end
        ),
        0
    )::FLOAT8                                                                                                                 AS total_earnings_usd,
    CASE
        WHEN (COUNT(*) FILTER (WHERE request_logs.payment_required = TRUE AND request_logs.created_at >= @period_start AND request_logs.created_at < @period_end)) = 0
        THEN 0.0
        ELSE (COUNT(*) FILTER (WHERE request_logs.payment_required = TRUE AND request_logs.payment_completed = TRUE
                AND request_logs.created_at >= @period_start AND request_logs.created_at < @period_end))::FLOAT8
            * 100.0
            / (COUNT(*) FILTER (WHERE request_logs.payment_required = TRUE AND request_logs.created_at >= @period_start AND request_logs.created_at < @period_end))::FLOAT8
    END::FLOAT8                                                                                                               AS success_rate,
    -- previous period
    (SELECT COUNT(*) FROM projects WHERE projects.created_at >= @prev_start AND projects.created_at < @period_start)::BIGINT AS prev_total_projects,
    (SELECT COUNT(*) FROM routes   WHERE routes.created_at   >= @prev_start AND routes.created_at   < @period_start)::BIGINT AS prev_total_routes,
    (COUNT(*) FILTER (WHERE request_logs.created_at >= @prev_start AND request_logs.created_at < @period_start))::BIGINT     AS prev_total_requests,
    COALESCE(
        SUM(request_logs.amount_usd) FILTER (
            WHERE request_logs.payment_required = TRUE AND request_logs.payment_completed = TRUE
              AND request_logs.created_at >= @prev_start AND request_logs.created_at < @period_start
        ),
        0
    )::FLOAT8                                                                                                                 AS prev_total_earnings_usd,
    CASE
        WHEN (COUNT(*) FILTER (WHERE request_logs.payment_required = TRUE AND request_logs.created_at >= @prev_start AND request_logs.created_at < @period_start)) = 0
        THEN 0.0
        ELSE (COUNT(*) FILTER (WHERE request_logs.payment_required = TRUE AND request_logs.payment_completed = TRUE
                AND request_logs.created_at >= @prev_start AND request_logs.created_at < @period_start))::FLOAT8
            * 100.0
            / (COUNT(*) FILTER (WHERE request_logs.payment_required = TRUE AND request_logs.created_at >= @prev_start AND request_logs.created_at < @period_start))::FLOAT8
    END::FLOAT8                                                                                                               AS prev_success_rate
FROM request_logs
WHERE request_logs.created_at >= @prev_start AND request_logs.created_at < @period_end;
