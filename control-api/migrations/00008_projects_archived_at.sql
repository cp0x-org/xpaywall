-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects ADD COLUMN archived_at TIMESTAMP;
ALTER TABLE projects DROP CONSTRAINT projects_slug_key;
CREATE UNIQUE INDEX projects_slug_active_uidx ON projects (slug) WHERE archived_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS projects_slug_active_uidx;
ALTER TABLE projects ADD CONSTRAINT projects_slug_key UNIQUE (slug);
ALTER TABLE projects DROP COLUMN archived_at;
-- +goose StatementEnd
