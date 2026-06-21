-- +goose Up
-- +goose StatementBegin

-- x402 payment methods settle through a facilitator; MPP (Tempo charge) has no
-- facilitator (its rpc_url / secret_key live in project_payment_methods.config).
-- Make facilitator_id optional so MPP project links can be created without one.
ALTER TABLE project_payment_methods ALTER COLUMN facilitator_id DROP NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Reverting requires every row to have a facilitator_id; rows with NULL (MPP)
-- would violate the restored NOT NULL constraint and must be removed first.
ALTER TABLE project_payment_methods ALTER COLUMN facilitator_id SET NOT NULL;

-- +goose StatementEnd
