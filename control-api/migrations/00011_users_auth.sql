-- +goose Up
-- +goose StatementBegin

-- Expand users into a full account: email + Google OAuth identity.
-- email/google_id are nullable so pre-existing rows (and the env superadmin)
-- stay valid; uniqueness is enforced by partial indexes only when set.
-- password_hash becomes nullable because Google-only users have no password.
ALTER TABLE users ADD COLUMN email VARCHAR;
ALTER TABLE users ADD COLUMN google_id VARCHAR;
ALTER TABLE users ADD COLUMN auth_provider VARCHAR NOT NULL DEFAULT 'local';
ALTER TABLE users ADD COLUMN avatar_url VARCHAR;
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

CREATE UNIQUE INDEX users_email_key     ON users (email)     WHERE email IS NOT NULL;
CREATE UNIQUE INDEX users_google_id_key ON users (google_id) WHERE google_id IS NOT NULL;

-- Password reset: store only sha256(raw token); the raw token lives in the email link.
CREATE TABLE password_reset_tokens (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token_hash VARCHAR NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used_at    TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_password_reset_tokens_user ON password_reset_tokens (user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS password_reset_tokens;

DROP INDEX IF EXISTS users_google_id_key;
DROP INDEX IF EXISTS users_email_key;

ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider;
ALTER TABLE users DROP COLUMN IF EXISTS google_id;
ALTER TABLE users DROP COLUMN IF EXISTS email;

-- +goose StatementEnd
