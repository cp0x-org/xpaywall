# 08 — Troubleshooting

When something doesn't work, the symptom is usually one of: an unexpected HTTP status, a missing log entry, or a UI screen that won't load. This page maps the common ones to their causes and fixes.

If the section heading describes what you are seeing, jump there.

## "I get 403 Forbidden on a path I configured"

The gateway resolves the request to a path, but no rule matches.

Common causes:

1. **Wrong slug.** Are you calling `/<slug>/<path>` where the slug matches the project's exact slug in the admin panel? Slugs are case-sensitive.
2. **Path mismatch.** A rule for `/api/v1/users` does not match `/api/v1/users/` (trailing slash). Likewise `/api/v1/*` does not match `/api/v1` (no extra segment).
3. **Allow Unmatched is off.** If the path truly has no rule and you expected the upstream to be hit anyway, turn on **Allow Unmatched Routes** on the project — but understand it makes every unmatched path free.

Open the request in the **Requests** page and look at the events. The `route_resolved` event has the matched rule (or absence thereof).

![Requests detail showing 403](./images/requests-detail-403.png)

## "I get 402 every time — even after paying"

If the same path returns 402 to repeated calls and you are sure you sent a valid `X-PAYMENT` header, one of these is happening:

1. **The header was stripped.** A reverse proxy in front of xgateway may filter unknown headers. Confirm by tailing the gateway request log — the event chain will start `route_resolved → payment_required` instead of `... → payment_completed`.
2. **The facilitator rejected the proof.** Check the `lastError` field in the second 402's JSON body. Common values: expired authorisation, wrong amount, wrong network, wrong asset, insufficient balance.
3. **The proof is for the wrong route.** The signed authorisation is tied to a specific `resource` URL. If the client computed the URL differently from what xgateway sees, the facilitator rejects it. The most common cause is `PUBLIC_URL` mismatch — see *PUBLIC_URL pitfalls* below.

## "Every request returns 500"

The gateway can answer a 402 without anything downstream. A 500 means a runtime failure inside the gateway itself. The two common origins:

1. **Cannot reach control-api (HTTP mode).** `CONTROL_API_URL` is wrong or unreachable from the gateway container. Test with `docker exec xgateway wget -qO- $CONTROL_API_URL/healthz`.
2. **Cannot reach the upstream.** `target` is wrong. Test with `docker exec xgateway wget -qO- <target-url>`.

A 502 / 504 specifically points at the upstream; a 500 points at the gateway itself.

## "I see 401 Unauthorized in the control-api logs"

xgateway tried to call control-api with the wrong `X-Api-Key`. Causes:

- The two services have different `INTERNAL_API_KEY` values.
- Whitespace in the env var (yes, a trailing newline in a `.env` file does this).
- The .env file is not being loaded — start the gateway with `--env-file .env`.

Fix the key on whichever side is wrong and restart both. To be sure they match, run:

```bash
docker exec xgateway sh -c 'echo "$INTERNAL_API_KEY"' | shasum
docker exec control-api sh -c 'echo "$INTERNAL_API_KEY"' | shasum
```

The two hashes must be identical.

## "Admin panel won't load / blank page"

Three layers can be wrong:

1. **The admin panel container itself didn't start.** Check `docker ps` and `docker logs adminpanel`.
2. **The frontend is built but can't reach control-api.** Open the browser's network tab — failed calls to control-api will show up. Confirm the `VITE_APP_API_URL` is correct for your environment.
3. **CORS is rejecting the calls.** control-api ships permissive CORS by default; if you have customised it, look at the response headers on the failed call.

If you see a successful login but every subsequent call returns 401, the JWT secret changed between issuing and verification. This happens after a control-api restart with a fresh `JWT_SECRET`. Log out, then log back in.

## "Routes I configured don't apply after I save them"

xgateway caches resolved rules for a short TTL (a few minutes). After an admin-panel edit:

- New paths (never resolved before) use the new rule immediately.
- Existing paths use the **old** rule until their cache entry expires.

For an immediate effect everywhere, restart xgateway (`docker restart xgateway`). For verification without restarting, wait a few minutes.

In file mode this is even simpler: the file is read once at startup, period. Restart after any edit.

## "Payments succeed but I see no row in Requests"

xgateway writes logs to control-api asynchronously. If control-api is slow or unreachable for the duration of the request, the log can be dropped. Single drops happen occasionally under load.

If **every** request is missing from the log:

- `CONTROL_API_URL` is empty or wrong on the gateway. Without a base URL, the log writer is a no-op.
- `INTERNAL_API_KEY` does not match (the gateway gets 401 and drops).

Add `LOG_LEVEL=debug` to xgateway and look for failed log dispatch messages.

## "Two rows appear for one logical interaction"

xgateway correlates the 402 and the paid retry within a ~10-minute window using a fingerprint of `(method, path, client IP)`. If the retry arrives **after** the window, it shows as a standalone row. If the client uses a different IP for the retry (NAT, VPN reconnect), the correlation fails.

This is mostly cosmetic — both rows have the same data; only the grouping in the UI is affected.

## "curl shows a 402 but the body is empty"

You probably used `curl <url>` without `-i`. By default curl prints only the body. 402 responses *do* have a body, but if you are not seeing one, double-check that:

- You are using `-i` to print headers, and
- The gateway returned `Content-Type: application/json` — if your reverse proxy strips bodies on non-2xx codes, the body never reaches curl.

## PUBLIC_URL pitfalls

The 402 response includes a `resource` field — the full URL the client paid for. xgateway constructs it from `PUBLIC_URL + request.URL.Path`.

If `PUBLIC_URL` is wrong:

- Clients sign authorisations for the wrong URL.
- The facilitator rejects them (the `resource` in the proof doesn't match the canonical URL).
- Every retry produces another 402.

Set `PUBLIC_URL` to the URL clients use to call the gateway — **not** the internal Docker URL. If you run behind a reverse proxy, this is the proxy's public URL.

Local Docker default: `PUBLIC_URL=http://localhost:3102`.

## File mode "config not loading"

```
panic: open /etc/xgateway/config.yaml: no such file or directory
```

The volume mount is missing or pointing at the wrong host path. Confirm:

```bash
docker exec xgateway ls -la /etc/xgateway/
```

If the file is there but the gateway still rejects it, check `docker logs xgateway` for the validation message (`x402[0].network is required` etc).

## Database migrations didn't run

After a control-api upgrade, you may see SQL errors on previously working endpoints. Reason: migrations are not auto-applied on `up`; you have to run them.

```bash
docker exec control-api control-api install
```

Or the goose CLI if you have it installed locally.

## Getting more information

Both Go services support a debug mode:

- xgateway: `GIN_MODE=debug` and `LOG_LEVEL=debug`.
- control-api: `MODE=debug`.

These produce verbose logs that are useful for one-off debugging. Turn them off again in production — they include request bodies and are noisy.

## What's next?

- Lock down the open admin panel and gateway: [09 — Security](./09-security.md).
- Look up the meaning of every env var: [10 — Reference](./10-reference.md).
