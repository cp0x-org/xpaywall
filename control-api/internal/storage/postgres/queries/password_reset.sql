-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (id, user_id, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetActivePasswordResetToken :one
SELECT * FROM password_reset_tokens
WHERE token_hash = $1
  AND used_at IS NULL
  AND expires_at > CURRENT_TIMESTAMP;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens
SET used_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteUserPasswordResetTokens :exec
DELETE FROM password_reset_tokens
WHERE user_id = $1;
