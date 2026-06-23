# 07 — xgateway: Overview

xgateway is the proxy that sits in front of your API. It receives every request, decides whether payment is required, returns a 402 with payment instructions if needed, verifies the proof when the client retries, and forwards verified requests to your upstream.

This page describes what xgateway does internally — middleware, providers, caches. The next two pages describe the two ways it gets its rules: from a YAML file or from control-api over HTTP.

## What xgateway is

xgateway **is**:

- A reverse proxy with payment logic.
- A standalone Go service that runs in a container.
- The thing your clients actually call.

## Where it sits

```
Client ──HTTP──▶ xgateway ──HTTP──▶ Upstream API
                    │
                    ├── (HTTP mode) ──▶ control-api ──▶ PostgreSQL
                    │
                    └── (file mode) ──▶ config.yaml on disk
```

Every request follows the top row. The bottom row is how xgateway gets its rules — that part happens at different cadences depending on the mode.

## Request lifecycle

When a request lands on xgateway, four middleware layers run in order:

```
Client request
      │
      ▼
┌─────────────────┐
│  Logging        │ ── stamps a request ID, starts timing
└────────┬────────┘
         ▼
┌─────────────────┐
│  Resolve        │ ── looks up the rule for this path
└────────┬────────┘    (cache → provider → cache fill)
         ▼
┌─────────────────┐
│  Payment        │ ── inspects X-PAYMENT header
└────────┬────────┘    paid + valid? proceed; missing? 402
         ▼
┌─────────────────┐
│  Proxy upstream │ ── forwards to the rule's target
└────────┬────────┘
         ▼
Response streamed back to client
```

Each layer has one job. None of them are aware of the others — the rule resolver does not know whether payment will happen; the payment layer does not know what the upstream looks like.

### Logging middleware

Stamps the incoming request with an ID, records the start time, attaches a logger that includes request metadata. After the request finishes (success or error), it sends a structured log entry to control-api asynchronously (HTTP mode only). In file mode, log lines go to stdout only.

### Route resolution

Reads the URL path. Asks the configured provider for the rule that matches. Stores the rule in a request-scoped context for the next middleware to read.

If no rule matches and the project does not allow unmatched paths, this middleware short-circuits with **403 Forbidden** — the request never reaches the payment layer.

### Payment middleware

Looks for an `X-PAYMENT` header on the request:

- **Missing** and the route is paid → return **402 Payment Required** with full payment requirements (one per accepted channel: scheme, network, asset, amount, payout address, facilitator). No call to the facilitator is needed for this — the route already knows what it costs.
- **Missing** and the route is free → skip straight to proxying.
- **Present** → POST the proof to the facilitator. If the facilitator says valid, proceed. If invalid, return 402 again with a `lastError` field describing the rejection.

A route can accept two payment rails. If the client sends an MPP `Authorization` header the gateway runs the **MPP** path — verifying and settling a Tempo `charge` against the configured RPC endpoint instead of calling a facilitator. Otherwise it runs the **x402** path described above. A route that lists only MPP channels always uses MPP.

### Proxy

Builds the upstream URL from the project's Server Base URL + the route's path. Copies the client's headers (minus a small denylist), injects the project's configured auth header if one is set, streams the request body, and streams the response back. The client sees the upstream response unchanged.

## Providers

xgateway is provider-agnostic — the rest of the code does not care where rules come from. The provider is selected by the `CONFIG_PROVIDER` env var:

- **`file`** — rules are loaded from a YAML file at startup and held in memory for the lifetime of the process. Detailed in [07-xgateway/03 — File mode](./03-file-mode.md).
- **`http`** — rules are fetched from control-api per request, with caching. Detailed in [07-xgateway/02 — HTTP mode](./02-http-mode.md).

Switching providers does not change any application behaviour. The 402 response is identical, the proxy behaviour is identical, the cache behaviour is identical — only the source of rules differs.

## Caches

xgateway keeps small in-memory caches. None of them touch disk. All are best-effort and reset on restart.

### Route cache

The first time a path is requested, the provider does its lookup (either reading the in-memory YAML or making an HTTP call to control-api). The resolved rule is cached, keyed by `<projectSlug>/<inboundPath>`. Subsequent requests skip the provider.

The cache lives for a short TTL (minutes), not hours. This is intentional: in HTTP mode, you want admin-panel edits to apply quickly without bouncing the gateway. In file mode the cache is redundant — the rules in memory never change — but the same code path is used.

A consequence: an admin-panel edit takes effect on the next request to that path **after the TTL expires**. For an immediate effect, restart xgateway.

## What runs at startup

1. Read environment variables (`CONFIG_PROVIDER`, `PORT`, `PUBLIC_URL`, facilitator settings, etc.).
2. Build the configured provider. **File mode:** read the YAML, validate, hold rules in memory. **HTTP mode:** verify control-api is reachable, store the base URL and API key.
3. Build the proxy middleware chain.
4. Start the HTTP server.

If any of the above fails (missing file, unreachable control-api, invalid YAML), the process exits with an error. There is no degraded mode.

## PUBLIC_URL

The 402 response includes a `resource` field — the full URL the client paid for. xgateway builds it from `PUBLIC_URL + request.URL.Path`.

If you run xgateway behind a reverse proxy or a load balancer, the `Host` header it sees is not the URL clients use. Set `PUBLIC_URL` to the public address so the 402 instructions point clients back at the right place.

For local Docker setups: `PUBLIC_URL=http://localhost:3102` is the standard.

## What's next?

- If you run with control-api: [07-xgateway/02 — HTTP mode](./02-http-mode.md) — the contract between xgateway and control-api.
- If you run from YAML: [07-xgateway/03 — File mode](./03-file-mode.md) — the full file format reference.
