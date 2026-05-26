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

-- name: GetPaymentMethodsByProjectSlug :many
SELECT
    pm.protocol,
    pm.code,
    pm.caip2_chain_id,
    pma.symbol,
    pma.contract_address,
    pma.decimals,
    ppm.scheme,
    ppm.payout_address,
    ppm.config,
    ppm.enabled,
    f.url AS facilitator_url,
    pm.id AS payment_method_id,
    pma.id AS asset_id
FROM project_payment_methods ppm
JOIN payment_methods pm ON pm.id = ppm.payment_method_id
JOIN payment_method_assets pma ON pma.id = ppm.asset_id
JOIN facilitators f ON f.id = ppm.facilitator_id
JOIN projects p ON p.id = ppm.project_id
WHERE p.slug = $1
  AND ppm.enabled = $2
  AND pm.enabled = $3;
