-- name: ListPaymentMethods :many
SELECT * FROM payment_methods ORDER BY protocol, name;

-- name: GetPaymentMethod :one
SELECT * FROM payment_methods WHERE id = $1;

-- name: CreatePaymentMethod :one
INSERT INTO payment_methods (id, code, protocol, name, caip2_chain_id, enabled)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdatePaymentMethod :one
UPDATE payment_methods
SET code           = COALESCE(sqlc.narg(code), code),
    protocol       = COALESCE(sqlc.narg(protocol), protocol),
    name           = COALESCE(sqlc.narg(name), name),
    caip2_chain_id = COALESCE(sqlc.narg(caip2_chain_id), caip2_chain_id),
    enabled        = COALESCE(sqlc.narg(enabled), enabled),
    updated_at     = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeletePaymentMethod :exec
DELETE FROM payment_methods WHERE id = $1;

-- name: ListPaymentMethodAssets :many
SELECT
    pma.id, pma.payment_method_id, pma.symbol, pma.contract_address, pma.decimals, pma.created_at, pma.updated_at,
    pm.name AS payment_method_name,
    pm.caip2_chain_id AS payment_method_chain
FROM payment_method_assets pma
JOIN payment_methods pm ON pm.id = pma.payment_method_id
ORDER BY pma.created_at DESC;

-- name: ListPaymentMethodAssetsByMethod :many
SELECT
    pma.id, pma.payment_method_id, pma.symbol, pma.contract_address, pma.decimals, pma.created_at, pma.updated_at,
    pm.name AS payment_method_name,
    pm.caip2_chain_id AS payment_method_chain
FROM payment_method_assets pma
JOIN payment_methods pm ON pm.id = pma.payment_method_id
WHERE pma.payment_method_id = $1
ORDER BY pma.created_at DESC;

-- name: GetPaymentMethodAsset :one
SELECT
    pma.id, pma.payment_method_id, pma.symbol, pma.contract_address, pma.decimals, pma.created_at, pma.updated_at,
    pm.name AS payment_method_name,
    pm.caip2_chain_id AS payment_method_chain
FROM payment_method_assets pma
JOIN payment_methods pm ON pm.id = pma.payment_method_id
WHERE pma.id = $1;

-- name: CreatePaymentMethodAsset :one
INSERT INTO payment_method_assets (id, payment_method_id, symbol, contract_address, decimals)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdatePaymentMethodAsset :one
UPDATE payment_method_assets
SET symbol           = COALESCE(sqlc.narg(symbol), symbol),
    contract_address = COALESCE(sqlc.narg(contract_address), contract_address),
    decimals         = COALESCE(sqlc.narg(decimals), decimals),
    updated_at       = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeletePaymentMethodAsset :exec
DELETE FROM payment_method_assets WHERE id = $1;

-- name: ListFacilitators :many
SELECT * FROM facilitators ORDER BY name;

-- name: GetFacilitator :one
SELECT * FROM facilitators WHERE id = $1;

-- name: CreateFacilitator :one
INSERT INTO facilitators (id, name, url, enabled)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateFacilitator :one
UPDATE facilitators
SET name       = COALESCE(sqlc.narg(name), name),
    url        = COALESCE(sqlc.narg(url), url),
    enabled    = COALESCE(sqlc.narg(enabled), enabled),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteFacilitator :exec
DELETE FROM facilitators WHERE id = $1;

-- name: ListProjectPaymentMethods :many
SELECT * FROM project_payment_methods WHERE project_id = $1 ORDER BY created_at DESC;

-- name: GetProjectPaymentMethod :one
SELECT * FROM project_payment_methods WHERE id = $1;

-- name: CreateProjectPaymentMethod :one
INSERT INTO project_payment_methods (id, project_id, payment_method_id, asset_id, scheme, facilitator_id, payout_address, config, enabled)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateProjectPaymentMethod :one
UPDATE project_payment_methods
SET scheme         = COALESCE(sqlc.narg(scheme), scheme),
    facilitator_id = COALESCE(sqlc.narg(facilitator_id), facilitator_id),
    payout_address = COALESCE(sqlc.narg(payout_address), payout_address),
    config         = COALESCE(sqlc.narg(config), config),
    enabled        = COALESCE(sqlc.narg(enabled), enabled),
    updated_at     = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteProjectPaymentMethod :exec
DELETE FROM project_payment_methods WHERE id = $1;
