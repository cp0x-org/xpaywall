-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: GetUserByUsernameOrEmail :one
SELECT * FROM users
WHERE username = $1 OR email = $1
LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByGoogleID :one
SELECT * FROM users
WHERE google_id = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (id, username, email, password_hash, auth_provider)
VALUES ($1, $2, $3, sqlc.narg(password_hash), $4)
RETURNING *;

-- name: CreateGoogleUser :one
INSERT INTO users (id, username, email, google_id, avatar_url, auth_provider)
VALUES ($1, $2, $3, $4, $5, 'google')
RETURNING *;

-- name: LinkGoogleAccount :one
UPDATE users
SET google_id  = $2,
    avatar_url = COALESCE(sqlc.narg(avatar_url), avatar_url),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET username   = COALESCE(sqlc.narg(username), username),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
