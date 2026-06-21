-- +goose Up
-- +goose StatementBegin

-- MPP payment methods carry a `method` (e.g. "tempo") and a `scheme` (e.g.
-- "charge") at the definition level; x402 methods have neither (they use a
-- facilitator and the per-project scheme in project_payment_methods.scheme).
-- Nullable so x402 rows leave them unset rather than holding a misleading default.
ALTER TABLE payment_methods ADD COLUMN method VARCHAR;
ALTER TABLE payment_methods ADD COLUMN scheme VARCHAR;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE payment_methods DROP COLUMN IF EXISTS method;
ALTER TABLE payment_methods DROP COLUMN IF EXISTS scheme;

-- +goose StatementEnd
