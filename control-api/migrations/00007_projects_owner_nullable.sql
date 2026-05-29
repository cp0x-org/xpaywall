-- +goose Up
ALTER TABLE projects
    DROP CONSTRAINT projects_owner_user_id_fkey,
    ALTER COLUMN owner_user_id DROP NOT NULL;

-- +goose Down
ALTER TABLE projects
    ALTER COLUMN owner_user_id SET NOT NULL,
    ADD CONSTRAINT projects_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES users (id);
