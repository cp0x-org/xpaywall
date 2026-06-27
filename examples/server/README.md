# example-server

A minimal sample upstream API used to exercise xgateway end to end. It listens
on port `4021` and returns canned JSON. xgateway sits in front of it and
enforces payment (and, optionally, injects an upstream auth header) before
forwarding requests here.

```bash
go run ./cmd        # listens on :4021
```

## Endpoints

### x402 demo set (seeded by `control-api install demo`)

| Method | Path | Notes |
|---|---|---|
| GET | `/health` | Health probe |
| ANY | `/api/metered/*path` | Mock metered response |
| GET | `/weather` | Weather for `?city=` |
| GET | `/free-endpoint` | Free, no payment |
| GET | `/free-multipoint/*path` | Free, no payment |
| GET | `/http-endpoint` | Static payload |

### MPP demo set (seeded by `control-api install demo-mpp`)

| Method | Path | Notes |
|---|---|---|
| GET | `/time` | Server time |
| ANY | `/api/usage/*path` | Mock usage response |
| GET | `/quote` | Price quote for `?symbol=` |
| GET | `/ping` | `pong` |
| GET | `/echo/*path` | Echoes the path |

### Auth-protected set

These endpoints require an `Authorization: Bearer <token>` header and return
`401` without a valid one. They demonstrate xgateway's upstream auth-header
injection: configure a route's auth header (project `auth_header_name` /
`auth_header_value`, or the file config's `outbound.auth_header`) so the gateway
adds the credential after verifying payment.

| Method | Path | Notes |
|---|---|---|
| GET | `/protected` | Requires a valid bearer token |
| GET | `/protected/*path` | Requires a valid bearer token |

The expected token defaults to `demo-secret-token`. Override it with the
`DEMO_BEARER_TOKEN` environment variable.

```bash
# Rejected — no credential:
curl -i http://localhost:4021/protected/data            # 401

# Accepted — correct bearer token (what xgateway injects for you):
curl -i -H "Authorization: Bearer demo-secret-token" \
  http://localhost:4021/protected/data                  # 200
```

See [Guide 04 — Connecting a real upstream](../../docs/06-guides/04-connecting-real-upstream.md)
for the full walkthrough of putting the gateway in front of an auth-protected
upstream.
