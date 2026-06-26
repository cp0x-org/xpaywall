-- +goose Up
-- +goose StatementBegin

-- x402 payment methods were created without a definition-level scheme, leaving the
-- column NULL. x402 settles with the 'exact' scheme by default, so backfill existing
-- rows; new rows now set it explicitly (form + seed).
UPDATE payment_methods SET scheme = 'exact' WHERE protocol = 'x402' AND scheme IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

UPDATE payment_methods SET scheme = NULL WHERE protocol = 'x402' AND scheme = 'exact';

-- +goose StatementEnd
