-- name: ListProjects :many
SELECT * FROM projects
WHERE archived_at IS NULL
Order BY name;

-- name: ListProjectsWithConfig :many
SELECT
    p.*,
    prs.base_url,
    COALESCE(
        ARRAY_AGG(DISTINCT pm.protocol) FILTER (WHERE pm.protocol IS NOT NULL),
        ARRAY[]::VARCHAR[]
    ) AS payment_methods
FROM projects p
         LEFT JOIN project_routes_settings prs
                   ON prs.project_id = p.id
         LEFT JOIN project_payment_methods ppm
                   ON ppm.project_id = p.id
                       AND ppm.enabled = TRUE
         LEFT JOIN payment_methods pm
                   ON pm.id = ppm.payment_method_id
                       AND pm.enabled = TRUE
WHERE p.archived_at IS NULL
GROUP BY
    p.id,
    prs.base_url
ORDER BY p.name;

-- name: ListProjectsWithConfigByOwner :many
SELECT
    p.*,
    prs.base_url,
    COALESCE(
        ARRAY_AGG(DISTINCT pm.protocol) FILTER (WHERE pm.protocol IS NOT NULL),
        ARRAY[]::VARCHAR[]
    ) AS payment_methods
FROM projects p
         LEFT JOIN project_routes_settings prs
                   ON prs.project_id = p.id
         LEFT JOIN project_payment_methods ppm
                   ON ppm.project_id = p.id
                       AND ppm.enabled = TRUE
         LEFT JOIN payment_methods pm
                   ON pm.id = ppm.payment_method_id
                       AND pm.enabled = TRUE
WHERE p.archived_at IS NULL AND p.owner_user_id = $1
GROUP BY
    p.id,
    prs.base_url
ORDER BY p.name;

-- name: GetProject :one
SELECT * FROM projects
WHERE id = $1 AND archived_at IS NULL;

-- name: GetProjectBySlug :one
SELECT * FROM projects
WHERE slug = $1 AND archived_at IS NULL;

-- name: GetActiveProjectByOwnerAndSlug :one
SELECT * FROM projects
WHERE owner_user_id = $1 AND slug = $2 AND archived_at IS NULL;

-- name: ListProjectsByOwner :many
SELECT * FROM projects
WHERE owner_user_id = $1 AND archived_at IS NULL
ORDER BY name;

-- name: CreateProject :one
INSERT INTO projects (id, owner_user_id, name, slug)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateProject :one
UPDATE projects
SET name       = COALESCE(sqlc.narg(name), name),
    slug       = COALESCE(sqlc.narg(slug), slug),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND archived_at IS NULL
RETURNING *;

-- name: ArchiveProject :exec
UPDATE projects
SET archived_at = CURRENT_TIMESTAMP,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = $1 AND archived_at IS NULL;
