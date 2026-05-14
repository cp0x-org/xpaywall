-- name: ResolveOutboundRoute :one
SELECT
    oroute.id,
    oroute.project_id,
    oroute.name,
    oroute.path_pattern,
    oroute.price_amount,
    oroute.price_usd,
    oroute.description,
    oroute.free,
    prs.base_url,
    prs.auth_header_name,
    prs.auth_header_value,
    prs.allow_unmatched
FROM routes oroute
JOIN projects p ON p.id = oroute.project_id
JOIN project_routes_settings prs ON prs.project_id = oroute.project_id
WHERE p.slug = $1
  AND sqlc.arg(inbound_path)::text LIKE REPLACE(oroute.path_pattern, '*', '%')
ORDER BY
    CASE WHEN oroute.path_pattern = sqlc.arg(inbound_path) THEN 0 ELSE 1 END,
    length(oroute.path_pattern) DESC
LIMIT 1;

-- name: GetPaymentChannelsByProjectSlug :many
SELECT
    pc.protocol,
    pc.method,
    pc.scheme,
    pc.metadata,
    ppc.enabled,
    ppc.payout_address,
    pc.id AS channel_id,
    ppc.payment_channel_asset_id
FROM project_payment_configs ppc
JOIN payment_channels pc ON pc.id = ppc.payment_channel_id
JOIN projects p ON p.id = ppc.project_id
WHERE p.slug = $1
  AND ppc.enabled = $2
  AND pc.enabled = $3;
