-- +goose Up
ALTER TABLE request_logs DROP COLUMN amount_paid;

-- +goose Down
ALTER TABLE request_logs ADD COLUMN amount_paid BIGINT;
