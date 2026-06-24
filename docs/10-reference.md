# 10 — Reference

Lookup tables for the things you forget mid-deploy: environment variables, ports, important URLs, the public API surface, and a glossary of the terms that appear in the admin panel.

## Default ports

These are the host ports the official `docker-compose.yml` binds. If you changed them, substitute accordingly.

| Service | Container port | Host port | Purpose |
|---|---|---|---|
| `xpaywall-landing` | 3100 | 3100 | Marketing landing page |
| `control-api` | 9091 | 3101 | REST control plane |
| `xgateway` | 8081 | 3102 | The reverse proxy clients call |
| `xpaywall-example-server` | 4021 | 3103 | Demo upstream |
| `adminpanel` | 80 | 3104 | React admin panel |
| `postgres` | 5432 | 5482 | Database |

## Environment variables

### xgateway

| Var | Required | Default | Notes |
|---|---|---|---|
| `CONFIG_PROVIDER` | yes | `file` | `file` or `http`. |
| `CONFIG_FILE` | if `CONFIG_PROVIDER=file` | `config.yaml` | Path to YAML/JSON config inside the container. |
| `CONTROL_API_URL` | if `CONFIG_PROVIDER=http` | — | Where to fetch routes from. |
| `INTERNAL_API_KEY` | if `CONFIG_PROVIDER=http` | — | Shared secret with control-api. |
| `PUBLIC_URL` | recommended | — | Public URL to advertise in 402 responses. |
| `PORT` | no | `8081` | HTTP listen port. |
| `GIN_MODE` | no | `release` | `debug` for verbose Gin logs. |
| `LOG_LEVEL` | no | `info` | `debug` shows log dispatch and cache hits. |

### control-api

| Var | Required | Default    | Notes |
|---|---|------------|---|
| `CONTROL_DB_DSN` | yes | —          | PostgreSQL DSN: `postgres://user:pass@host:port/db`. |
| `INTERNAL_API_KEY` | yes | —          | Must match xgateway. |
| `JWT_SECRET` | yes | —          | Signs admin panel sessions. |
| `PROXY_URL` | yes | —          | Public xgateway URL — shown in the admin panel. |
| `PORT` | no | `9091`     | HTTP listen port. |
| `MODE` | no | `release`  | `debug` for verbose Gin logs. |
| `APP_BASE_URL` | no | `http://localhost:3000` | Frontend base URL for password-reset links. |
| `GOOGLE_CLIENT_ID` | no | —          | OAuth client ID; required for Google sign-in. |

> Superadmin is not provisioned via env. Grant it in Postgres: `UPDATE users SET role='superadmin' WHERE username='...';`

### Admin panel

| Var | Required | Default | Notes |
|---|---|---|---|
| `VITE_APP_API_URL` | yes | — | URL of control-api as reachable from the **browser**. |

### Example server

| Var | Required | Default | Notes |
|---|---|---|---|
| `PORT` | no | `4021` | Demo upstream listen port. |

## Admin panel API surface

JWT-authenticated unless noted. All routes are under `/api/v1/`.

| Group | Endpoints |
|---|---|
| Auth | `POST /auth/login`, `GET /auth/me` |
| Users | `GET/POST /users`, `PUT/DELETE /users/:id` |
| Projects | `GET/POST /projects`, `PUT/DELETE /projects/:id`, `GET /projects/:id/payment-methods`, `POST /projects/:id/payment-methods` |
| Routes | `GET/POST /routes`, `PUT/DELETE /routes/:id` |
| Facilitators | `GET/POST /facilitators`, `PUT/DELETE /facilitators/:id` |
| Payment Methods | `GET/POST /payment-methods`, `PUT/DELETE /payment-methods/:id` |
| Payment Assets | `GET/POST /payment-assets`, `PUT/DELETE /payment-assets/:id` |
| Requests | `GET /requests`, `GET /requests/:id`, `GET /requests/:id/events` |
| Dashboard | `GET /dashboard/summary`, `GET /dashboard/top-routes` |
| Activity | `GET /activity-logs` |

## Internal API surface (gateway ↔ control-api)

All require `X-Api-Key: <INTERNAL_API_KEY>`.

| Endpoint | Used by xgateway for |
|---|---|
| `GET /proxy/resolve/{slug}/{path}` | Looking up the rule for an inbound request. |
| `POST /api/v1/request-logs` | Writing one row per finished request. |
| `POST /api/v1/request-events` | Writing granular per-stage events. |

See [07-xgateway/02 — HTTP mode](./07-xgateway/02-http-mode.md) for the full contract.

## Public files inside the repo

| Path | What's in it |
|---|---|
| `docker-compose.yml` | The default development stack. |
| `docker-compose.prod.yml` | The production override (may not exist yet — see Roadmap). |
| `xgateway/config.dev.yaml` | Example file-mode config with x402 and MPP channels. |
| `control-api/migrations/` | Database schema (goose migrations). |
| `examples/server/` | The demo upstream you used in Guide 01. |

## Useful CAIP-2 chain IDs

| Network | CAIP-2 |
|---|---|
| Base Mainnet | `eip155:8453` |
| Base Sepolia | `eip155:84532` |
| Ethereum Mainnet | `eip155:1` |
| Optimism | `eip155:10` |
| Polygon | `eip155:137` |
| Arbitrum One | `eip155:42161` |

## Useful USDC contracts

| Network | USDC contract | Decimals |
|---|---|---|
| Base Mainnet | `0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913` | 6 |
| Base Sepolia | `0x036CbD53842c5426634e7929541eC2318f3dCF7e` | 6 |
| Ethereum Mainnet | `0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48` | 6 |
| Polygon | `0x3c499c542cEF5E3811e1192ce70d8cC03d5c3359` | 6 |

Always verify these against the official chain explorer before pointing real money at them.

## Glossary

**402** — HTTP status code for "Payment Required". The whole product is built around making this code useful.

**Asset** — A specific token on a specific chain. Identified by its contract address and decimals.

**Bazaar** — The x402 ecosystem's discovery extension — a catalog of paid endpoints that facilitators can advertise to clients.

**CAIP-2** — A standard format for identifying blockchains. Format: `<namespace>:<reference>`, e.g. `eip155:8453`.

**CAIP-19** — A standard format for identifying assets. xpaywall stores just the address portion; the chain comes from the parent payment method.

**Facilitator** — An external service that verifies x402 payment proofs and settles them on-chain. Public ones exist for testing; you can run your own.

**File mode** — xgateway reading its configuration from a YAML file at startup. Static, single-upstream.

**HTTP mode** — xgateway fetching its configuration from control-api per request. Dynamic, multi-project.

**MPP** — Machine Payments Protocol. A payment rail that settles an on-chain charge directly against a blockchain RPC, with no facilitator. xpaywall supports the Tempo `charge` scheme.

**Payment Asset** — Admin-panel entity: a token (symbol, contract, decimals) attached to a payment method.

**Payment Channel** — In gateway-internal API responses, the combination that a route accepts: for x402 (payment method, asset, scheme, facilitator, payout address), for MPP (method, asset, scheme, rpc_url, payout address). One route can accept many.

**Payment Method** — Admin-panel entity: for x402 a (protocol, network) pair, e.g. x402-on-Base; for MPP a (protocol, method, scheme) triple, e.g. MPP-Tempo-charge.

**Payout Address** — The wallet address that receives the money for paid requests.

**Project** — Admin-panel entity: an upstream API plus a slug. Routes belong to a project.

**Project Payment Method** — Admin-panel entity: the link between a project and a payment method, with the payout address attached.

**Project Slug** — A short identifier (e.g. `demo`) used in the proxy URL: `/{slug}/{path}`.

**Resource** — The full URL a client paid for. Appears in the 402 response and in the signed authorisation.

**Route** — Admin-panel entity: a path pattern plus a price.

**Scheme** — The kind of authorisation a client signs. For x402: `exact` (a fixed price). For MPP: `charge` (a one-time on-chain charge).

**Tempo** — The blockchain network MPP charges are verified and settled against, addressed by an `rpc_url`.

**X-PAYMENT** — The HTTP header a client sends on the retry, carrying the signed proof of payment.

**x402** — The protocol that makes HTTP 402 usable for actual micropayments.

## What's next?

- See what's planned but not yet shipped: [11 — Roadmap](./11-roadmap.md).
- Back to the table of contents: [README](./README.md).
