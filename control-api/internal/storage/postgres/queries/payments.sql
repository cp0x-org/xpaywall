-- name: ListPaymentChannels :many
SELECT * FROM payment_channels ORDER BY created_at DESC;

-- name: GetPaymentChannel :one
SELECT * FROM payment_channels WHERE id = $1;

-- name: CreatePaymentChannel :one
INSERT INTO payment_channels (id, protocol, method, scheme, enabled)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdatePaymentChannel :one
UPDATE payment_channels
SET protocol   = COALESCE(sqlc.narg(protocol), protocol),
    method     = COALESCE(sqlc.narg(method), method),
    scheme     = COALESCE(sqlc.narg(scheme), scheme),
    enabled    = COALESCE(sqlc.narg(enabled), enabled),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeletePaymentChannel :exec
DELETE FROM payment_channels WHERE id = $1;

-- name: ListPaymentChannelAssets :many
SELECT * FROM payment_channel_assets ORDER BY created_at DESC;

-- name: ListPaymentChannelAssetsByChannel :many
SELECT * FROM payment_channel_assets WHERE payment_channel_id = $1 ORDER BY created_at DESC;

-- name: GetPaymentChannelAsset :one
SELECT * FROM payment_channel_assets WHERE id = $1;

-- name: CreatePaymentChannelAsset :one
INSERT INTO payment_channel_assets (id, payment_channel_id, asset_symbol, asset_address, decimals)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdatePaymentChannelAsset :one
UPDATE payment_channel_assets
SET asset_symbol  = COALESCE(sqlc.narg(asset_symbol), asset_symbol),
    asset_address = COALESCE(sqlc.narg(asset_address), asset_address),
    decimals      = COALESCE(sqlc.narg(decimals), decimals),
    updated_at    = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeletePaymentChannelAsset :exec
DELETE FROM payment_channel_assets WHERE id = $1;
