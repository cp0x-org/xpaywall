-- name: GetProjectRouteSettings :one
SELECT id, project_id, base_url, auth_header_name, auth_header_value, allow_unmatched, created_at, updated_at
FROM project_routes_settings
WHERE project_id = $1;

-- name: UpsertProjectRouteSettings :one
INSERT INTO project_routes_settings (id, project_id, base_url, auth_header_name, auth_header_value, allow_unmatched)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (project_id) DO UPDATE
SET base_url = EXCLUDED.base_url,
    auth_header_name = EXCLUDED.auth_header_name,
    auth_header_value = EXCLUDED.auth_header_value,
    allow_unmatched = EXCLUDED.allow_unmatched,
    updated_at = CURRENT_TIMESTAMP
RETURNING id, project_id, base_url, auth_header_name, auth_header_value, allow_unmatched, created_at, updated_at;

-- name: ListOutboundRoutes :many
SELECT r.id, r.project_id, p.slug AS project_slug, r.name, r.path_pattern, r.price_usd, r.description, r.free, r.bazaar, r.created_at, r.updated_at
FROM routes r
JOIN projects p ON p.id = r.project_id
ORDER BY r.name;

-- name: ListOutboundRoutesByOwner :many
SELECT r.id, r.project_id, p.slug AS project_slug, r.name, r.path_pattern, r.price_usd, r.description, r.free, r.bazaar, r.created_at, r.updated_at
FROM routes r
JOIN projects p ON p.id = r.project_id
WHERE p.owner_user_id = $1
ORDER BY r.name;

-- name: ListOutboundRoutesByProject :many
SELECT r.id, r.project_id, p.slug AS project_slug, r.name, r.path_pattern, r.price_usd, r.description, r.free, r.bazaar, r.created_at, r.updated_at
FROM routes r
JOIN projects p ON p.id = r.project_id
WHERE r.project_id = $1
ORDER BY r.name;

-- name: GetOutboundRoute :one
SELECT id, project_id, name, path_pattern, price_usd, description, free, bazaar, created_at, updated_at
FROM routes
WHERE id = $1;

-- name: CreateOutboundRoute :one
INSERT INTO routes (id, project_id, name, path_pattern, price_usd, description, free, bazaar)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, project_id, name, path_pattern, price_usd, description, free, bazaar, created_at, updated_at;

-- name: UpdateOutboundRoute :one
UPDATE routes
SET project_id   = COALESCE(sqlc.narg(project_id), project_id),
    name         = COALESCE(sqlc.narg(name), name),
    path_pattern = COALESCE(sqlc.narg(path_pattern), path_pattern),
    price_usd    = COALESCE(sqlc.narg(price_usd), price_usd),
    description  = COALESCE(sqlc.narg(description), description),
    free         = COALESCE(sqlc.narg(free), free),
    bazaar       = COALESCE(sqlc.narg(bazaar), bazaar),
    updated_at   = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, project_id, name, path_pattern, price_usd, description, free, bazaar, created_at, updated_at;

-- name: NullifyRouteInRequestLogs :exec
UPDATE request_logs SET outbound_route_id = NULL WHERE outbound_route_id = $1;

-- name: DeleteRouteDailyStats :exec
DELETE FROM route_daily_stats WHERE outbound_route_id = $1;

-- name: DeleteOutboundRoute :exec
DELETE FROM routes WHERE id = $1;
