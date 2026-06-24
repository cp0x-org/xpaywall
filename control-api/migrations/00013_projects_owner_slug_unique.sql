-- +goose Up
-- +goose StatementBegin
-- Slug is now unique per owner instead of globally: routes are resolved by
-- {username}/{slug}, so two different users may each own a 'default' project,
-- but a single user cannot have two active projects with the same slug.
DROP INDEX IF EXISTS projects_slug_active_uidx;
CREATE UNIQUE INDEX projects_owner_slug_active_uidx
    ON projects (owner_user_id, slug) WHERE archived_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS projects_owner_slug_active_uidx;
CREATE UNIQUE INDEX projects_slug_active_uidx ON projects (slug) WHERE archived_at IS NULL;
-- +goose StatementEnd
