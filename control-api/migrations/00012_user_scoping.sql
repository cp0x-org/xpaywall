-- +goose Up
-- +goose StatementBegin

-- Role decides who may manage global entities. Superadmin is provisioned manually
-- in Postgres (UPDATE users SET role='superadmin' WHERE ...); there is no API for it.
ALTER TABLE users ADD COLUMN role VARCHAR NOT NULL DEFAULT 'user'
    CHECK (role IN ('user', 'superadmin'));

-- Global-capable entities: is_global makes a row visible to everyone; owner_user_id
-- scopes a personal row to its creator. NULL owner = system/global (no personal owner).
ALTER TABLE payment_methods
    ADD COLUMN is_global BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN owner_user_id UUID REFERENCES users (id);
ALTER TABLE payment_method_assets
    ADD COLUMN is_global BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN owner_user_id UUID REFERENCES users (id);
ALTER TABLE facilitators
    ADD COLUMN is_global BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN owner_user_id UUID REFERENCES users (id);

-- Existing rows were shared by everyone; keep them visible to all.
UPDATE payment_methods       SET is_global = TRUE;
UPDATE payment_method_assets SET is_global = TRUE;
UPDATE facilitators          SET is_global = TRUE;

CREATE INDEX idx_payment_methods_owner       ON payment_methods (owner_user_id);
CREATE INDEX idx_payment_method_assets_owner ON payment_method_assets (owner_user_id);
CREATE INDEX idx_facilitators_owner          ON facilitators (owner_user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_facilitators_owner;
DROP INDEX IF EXISTS idx_payment_method_assets_owner;
DROP INDEX IF EXISTS idx_payment_methods_owner;

ALTER TABLE facilitators          DROP COLUMN IF EXISTS owner_user_id, DROP COLUMN IF EXISTS is_global;
ALTER TABLE payment_method_assets DROP COLUMN IF EXISTS owner_user_id, DROP COLUMN IF EXISTS is_global;
ALTER TABLE payment_methods       DROP COLUMN IF EXISTS owner_user_id, DROP COLUMN IF EXISTS is_global;

ALTER TABLE users DROP COLUMN IF EXISTS role;

-- +goose StatementEnd
