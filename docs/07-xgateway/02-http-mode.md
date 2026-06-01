# 07-xgateway — HTTP mode

In HTTP mode, xgateway has no idea what your routes look like at startup. Every time a request lands on a path it has not seen before, it asks control-api: *what is the rule for this path?* The answer is cached in memory and reused for subsequent requests.

This page documents the contract between xgateway and control-api. You only need to read it carefully if you are:

- Replacing control-api with your own service that speaks the same HTTP contract.
- Debugging unexpected gateway behaviour (404, 403, wrong price).
- Writing a custom monitoring tool.

If you are just using the official admin panel, you can stop after the next section.

## How HTTP mode is enabled

On xgateway:

```bash
CONFIG_PROVIDER=http
CONTROL_API_URL=http://control-api:9091
INTERNAL_API_KEY=<some-shared-secret>
```

On control-api:

```bash
INTERNAL_API_KEY=<the-same-shared-secret>
```

The two services must share the same `INTERNAL_API_KEY`. xgateway sends it on every internal call; control-api rejects calls without it.

Both can run on the same Docker network, so the URL is typically a container name — `http://control-api:9091`. From outside Docker it would be `http://localhost:3101` (the default mapped port).

## The route resolution endpoint

xgateway calls one endpoint to resolve a route:

```
GET /proxy/resolve/{projectSlug}/{inboundPath}
X-Api-Key: <INTERNAL_API_KEY>
```

For example, a client request to `http://gateway:8081/demo/weather` causes:

```
GET /proxy/resolve/demo/weather
```

The path after the slug is preserved exactly, including any sub-path segments. So `/demo/api/v1/users/42` becomes `GET /proxy/resolve/demo/api/v1/users/42`.

### Response — route found

`200 OK` with a JSON body. Below is every field, in the order control-api emits them:

```json
{
  "project_id": "9e0e8d12-...-...",
  "outbound_route_id": "5b9c...-...",
  "name": "Weather",
  "inbound_path": "/weather",
  "target": "http://xpaywall-example-server:4021",
  "auth_header_name": "Authorization",
  "auth_header_value": "Bearer sk_live_...",
  "allow_unmatched": false,
  "price": "0.001",
  "free": false,
  "mime_type": "",
  "description": "Returns the current weather",
  "bazaar": null,
  "payment_channels": [
    {
      "protocol": "x402",
      "code": "x402-base-sepolia",
      "scheme": "exact",
      "caip2_chain_id": "eip155:84532",
      "facilitator_url": "https://x402.org/facilitator",
      "payout_address": "0xYourPayoutAddress",
      "asset_symbol": "USDC",
      "contract_address": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
      "amount": "1000",
      "decimals": 6,
      "enabled": true,
      "payment_method_id": "...",
      "asset_id": "..."
    }
  ]
}
```

What each field is for:

| Field | What xgateway does with it |
|---|---|
| `project_id`, `outbound_route_id` | Logging only — sent back in request logs so you can correlate to the admin panel. |
| `name`, `description` | Logging and 402 metadata. |
| `inbound_path` | Echoed back for verification — should match what xgateway sent. |
| `target` | The upstream's base URL. Forwarding URL = `target + inbound_path`. |
| `auth_header_name` / `auth_header_value` | If non-empty, xgateway injects this header on the forwarded request. |
| `allow_unmatched` | Project-wide flag; controls whether unknown paths are proxied for free. |
| `price` | USD price as a decimal string. xgateway echoes this in the 402 body but does **not** use it for the on-chain amount — control-api has already done that conversion. |
| `free` | If `true`, payment middleware is skipped. |
| `mime_type` | Hint for clients in the 402 body. Optional. |
| `bazaar` | Opaque JSON object passed through to the 402 body. xgateway does not interpret it. |
| `payment_channels[]` | One entry per `(method, asset)` pair the route accepts. Each becomes one entry in the 402 `accepts` array. |

For each payment channel:

| Field | Use |
|---|---|
| `protocol`, `code`, `scheme` | Stamped on the 402 entry so the client knows which protocol/scheme to sign for. |
| `caip2_chain_id` | The `network` field in the 402 body. |
| `facilitator_url` | Where xgateway sends the proof on the retry. |
| `payout_address` | The `payTo` field in the 402 body. |
| `asset_symbol`, `contract_address`, `decimals` | Stamped on the 402 entry. `contract_address` is the `asset` field. |
| `amount` | The pre-computed on-chain amount as a decimal string. **control-api computes this**: `round(price * 10^decimals)`. xgateway uses this value as-is. |
| `enabled` | If `false`, the channel is skipped. (control-api already filters; this is belt-and-suspenders.) |

### Response — no route

If no rule matches the path:

```
404 Not Found
Content-Type: application/json

{"error": "route not found"}
```

xgateway treats this as: project's `Allow Unmatched` decides what happens. If it is off (the default), xgateway returns 403 to the client. If on, xgateway proxies the request unauthenticated to the project's base URL.

To distinguish these two cases, the 404 response uses the most recent `allow_unmatched` value the gateway has cached for that project, or falls back to "off" if there is no cache hit.

### Response — auth failure

If `INTERNAL_API_KEY` is missing or wrong:

```
401 Unauthorized
{"error": "unauthorized"}
```

This will fail every request — xgateway returns a 500 to the client. If you see this happening, the two services have a key mismatch.

## Request log ingestion

xgateway also writes request logs back to control-api via two endpoints:

- `POST /api/v1/request-logs` — one record per request (final state).

Both require the same `X-Api-Key` header. They are write-only from xgateway's perspective — it never reads back.

If control-api is down or slow, xgateway buffers logs in memory and retries with backoff. Sustained outages result in dropped logs (xgateway prefers serving traffic over preserving telemetry).

## Caching behaviour

xgateway caches resolve responses by `(projectSlug, inboundPath)` for a short TTL — measured in minutes, not hours. Within that window, repeated requests skip the call to control-api entirely.

Consequences:

- An admin-panel edit takes effect on the **next request to that path after the TTL expires**. For an immediate effect, restart the gateway container.
- A new path (one xgateway has never seen) is always resolved fresh — there is no negative cache for misses.
- The cache is per-process. Multiple gateway replicas each have their own cache.

There is no API to explicitly invalidate the cache. If you need to be 100% sure of behaviour after a change, restart xgateway.

## What can go wrong

| Symptom | Likely cause |
|---|---|
| Every request returns 500. | xgateway cannot reach control-api. Check `CONTROL_API_URL` and the network. |
| Every request returns 403, even valid paths. | `INTERNAL_API_KEY` mismatch. control-api returns 401, xgateway treats it as "no route". |
| Some requests still use old prices. | Cache TTL hasn't expired yet. Wait or restart. |
| 502 / 504 on paid routes after correct payment. | The 402 + verification succeeded but the upstream is unreachable. Check `target` in the response and verify reachability from the gateway container. |

## Replacing control-api

The above is the entire contract. Anything that returns this JSON shape from `GET /proxy/resolve/{slug}/{path}` with `X-Api-Key` auth is a valid drop-in for control-api, as far as xgateway is concerned. The same applies to the two log-ingestion endpoints.

If you want a custom backend (e.g. one that pulls routes from your own database), implement those three endpoints and point `CONTROL_API_URL` at it.

## What's next?

- The simpler way to configure xgateway: [03 — File mode](./03-file-mode.md).
- Overview of how the gateway handles a request: [01 — Overview](./01-overview.md).
