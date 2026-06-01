-- +goose Up
-- bazaar holds the x402 Bazaar Discovery Extension declaration for the route
-- (method, body_type, input/output examples and JSON schemas). NULL means the
-- gateway will auto-generate a minimal GET declaration; explicit JSON gives
-- discovery clients richer metadata. JSONB so the schema can evolve without
-- migrations as Bazaar grows.
ALTER TABLE routes ADD COLUMN bazaar JSONB;

-- +goose Down
ALTER TABLE routes DROP COLUMN bazaar;
