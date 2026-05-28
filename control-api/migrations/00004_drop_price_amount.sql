-- +goose Up
ALTER TABLE routes DROP COLUMN price_amount;

-- +goose Down
ALTER TABLE routes ADD COLUMN price_amount INTEGER NOT NULL DEFAULT 0;
