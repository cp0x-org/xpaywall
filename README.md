# xpaywall

**xpaywall** is a self-hosted HTTP 402 payment gateway that enforces micropayments in front of any API. It sits between clients and your upstream services — no API keys, no billing accounts, no subscriptions. Clients pay per request using the [x402](https://www.x402.org/) protocol or the [Machine Payments Protocol](https://mpp.dev/) (MPP / Tempo), and the gateway proxies the request only after verifying payment proof.

---

## Why xpaywall

**Get paid per request. No accounts. No keys. No invoices.**

- ⚡ **Drop-in monetisation.** Point xpaywall at your existing API. No code changes upstream.
- 💸 **On-chain settlement.** Every call is its own payment. You get paid the moment the request is made.
- 🔓 **No client signup.** Callers hit the URL, get `HTTP 402`, pay, retry. That's the entire onboarding.
- 🎛️ **Web UI for everything.** Projects, routes, prices, payment methods, request logs — all in the admin panel.
- 🏠 **Self-hosted and open source.** MIT licensed, one `docker compose up` away.

📖 **Full documentation:** <http://xpaywall.cp0x.com/docs>

---

## How It Works

```
Client Request
     │
     ▼
┌─────────────┐    no payment     ┌─────────────────────────────┐
│  xgateway   │ ────────────────► │  HTTP 402 + payment details │
│  :8081      │                   └─────────────────────────────┘
│             │    proof verified
│             │ ────────────────► Upstream API → response → client
└──────┬──────┘
       │ resolves rules
       ▼
┌─────────────┐
│ control-api │  ◄── + Admin Panel (React)
│  :9091      │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ PostgreSQL  │
└─────────────┘
```

1. Client hits xgateway at any path
2. Gateway resolves the payment rule for that path
3. No valid proof → returns `HTTP 402` with payment requirements
4. Client pays on-chain and retries with proof header
5. Gateway verifies proof → proxies request to upstream → logs result

---

## Admin Dashboard

![Dashboard](docs/images/dashboard.png)

---

## Services

| Service | Directory | Container Port | Role |
|---|---|---|---|
| **xgateway** | `xgateway/` (submodule) | 8081 | Reverse proxy that enforces payment rules |
| **control-api** | `control-api/` | 9091 | REST control plane — projects, routes, users, logs |
| **adminpanel** | `frontend/adminpanel/` | 80 (3000 in dev) | React dashboard for managing everything |
| **example-server** | `examples/server/` | 4021 | Sample upstream API for testing |

---

## Quick Start

### With Docker Compose

```bash
git clone https://github.com/cp0x-org/xpaywall.git
cd xpaywall
git submodule update --init --recursive
docker compose up -d
```

> `xgateway/` is a Git submodule from [`cp0x-org/xgateway`](https://github.com/cp0x-org/xgateway) — don't skip the `submodule update` step or the build will fail.

| Service | URL |
|---|---|
| Landing | http://localhost:3100 |
| Control API | http://localhost:3101 |
| Gateway | http://localhost:3102 |
| Example upstream | http://localhost:3103 |
| Admin Panel | http://localhost:3104 |
| PostgreSQL | localhost:5482 |

There is no default admin account — register one on the login page, or create the first **superadmin** with the CLI: `docker compose run --rm control-api install user --username alice --password '<pass>' --email alice@example.com --role superadmin`. Host ports are mapped in `docker-compose.yml` — adjust them there if you need different externals.

### Local Development

**Prerequisites:** Go 1.26+, Node.js 22+, Yarn 4, PostgreSQL 16, Docker

**control-api:**
```bash
cd control-api
go run ./cmd/control-api --env-file .env migrate                       # run migrations
go run ./cmd/control-api --env-file .env install user \
  --username alice --password p --email alice@example.com --role superadmin
go run ./cmd/control-api --env-file .env install payment-methods --superadmin alice
go run ./cmd/control-api --env-file .env                               # start server
```

**xgateway:**
```bash
cd xgateway
go run ./cmd/xgateway --env-file .env
```

**adminpanel:**
```bash
cd frontend/adminpanel
yarn install
yarn start
```

---

## Setting Up Your First Paywall

Once the services are running, log into the **Admin Panel** and configure your gateway in three steps. Everything below is done through the UI.

### 1. Create a project

**Projects → Create Project**

A project maps a public slug to an upstream server.

| Field | Description                                                                                                                          |
|---|--------------------------------------------------------------------------------------------------------------------------------------|
| **Project Name** | Human-readable name. The **Slug** is auto-filled from it (lowercase, hyphenated) — edit if needed.                                   |
| **Slug** | URL identifier. Requests are served at `<PROXY_URL>/<username>/<slug>/...`.                                                          |
| **Server Base URL** | Upstream target where verified requests are proxied, e.g. `https://api.example.com`.                                                 |
| **Auth Header Name / Value** | *Optional.* Injected into every upstream request — use it to attach your own API key to the origin (e.g. `Authorization: Bearer …`). |
| **Allow Unmatched Routes** | If checked, paths with no matching route are proxied **without** payment. Leave off to block them.                                   |

Save the project, then open it and switch to the **Payment Methods** tab.

### 2. Attach a payment method

**Project → Payment Methods tab → Add Payment Method**

A project link combines a **payment method** (protocol + network), an **asset** (token), and — for x402 — a **facilitator**. A project uses a single protocol; once one method is added, only methods of the same protocol can be added.

If these building blocks don't exist yet, create them first under the **Payments** menu (global ones may already be seeded):

1. **Payments → Facilitators (x402) → Create** *(x402 only)* — name + facilitator `URL` (settles the on-chain payment), e.g. `https://x402.dexter.cash`.
2. **Payments → Payment Methods → Create** — set the **Protocol** and its options: `x402` takes a **Network** (or a custom CAIP-2 chain ID like `eip155:8453`) and a **Scheme** (`exact`); `mpp` takes a **Method** (`tempo`) and a **Scheme** (`charge`). Give it a unique **Code** (e.g. `x402-base-usdc` or `mpp-tempo-charge`).
3. **Payments → Payment Assets → Create** — choose the payment method above, then enter the token **Symbol** (e.g. `USDC`), **Contract Address**, and **Decimals** (`6` for USDC).

Back on the project's **Payment Methods** tab, click **Add Payment Method** and select:

| Field | Description |
|---|---|
| **Payment Method** | The method created above. |
| **Asset** | A token belonging to that method. |
| **Scheme** | `exact` for x402, `charge` for MPP. |
| **Facilitator** | *x402 only.* The facilitator that settles payment. MPP settles directly on Tempo (no facilitator). |
| **Payout Address** | Wallet that receives payments on this network. |
| **Enabled** | Must be on for the method to be offered. |

### 3. Add routes

**Routes → Create Route**

A route decides which upstream paths require payment and how much.

| Field | Description |
|---|---|
| **Project** | The project this route belongs to. |
| **Route Name** | Label for the route. |
| **Path Pattern** | Path to match, e.g. `/weather` or `/api/v1/*`. The live **Proxy URL** and **Target URL** preview updates as you type. |
| **Free** | If checked, the path is proxied with no payment. |
| **Price (USD)** | Per-request price when not free, e.g. `0.10`. |
| **Bazaar Discovery** | *Optional (x402).* JSON describing the endpoint for discovery — leave empty for auto-mode. |

That's it. Callers now hit `<PROXY_URL>/<slug>/<path>`, receive `HTTP 402` with payment requirements, pay, and retry with the proof header to reach your upstream.

---

## Configuration

### docker-compose.yml

All service configuration lives in `docker-compose.yml`. The key values to change before deploying:

```yaml
control-api:
  environment:
    INTERNAL_API_KEY: change-me-internal-secret-key   # shared with xgateway
    JWT_SECRET: change-me-jwt-secret-key
    PROXY_URL: http://your-server-ip:3102             # public gateway URL
    MODE: release                                     # Gin mode: release | debug

xgateway:
  environment:
    INTERNAL_API_KEY: change-me-internal-secret-key   # must match control-api

adminpanel:
  environment:
    API_URL: http://your-server-ip:3101/              # browser-accessible URL
    PROXY_URL: http://your-server-ip:3102/
```

> **Note:** `API_URL` and `PROXY_URL` for the admin panel are runtime env vars — you can change them without rebuilding the image.

---

### xgateway Environment Variables

| Variable | Required | Description |
|---|---|---|
| `CONFIG_PROVIDER` | Yes | `file` (YAML) or `http` (control-api) |
| `CONTROL_API_URL` | http mode | URL of control-api, e.g. `http://control-api:9091` |
| `INTERNAL_API_KEY` | http mode | Shared secret — must match control-api |
| `CONFIG_FILE` | file mode | Path to YAML rules file |
| `PORT` | No | Listen port (default: `8080`) |
| `LOG_LEVEL` | No | `DEBUG`, `INFO`, `WARN`, `ERROR` |
| `GIN_MODE` | No | `debug` or `release` |
| `PUBLIC_URL` | No | Override the public-facing URL injected into 402 responses |

### control-api Environment Variables

| Variable | Required | Description |
|---|---|---|
| `CONTROL_DB_DSN` | Yes | PostgreSQL DSN |
| `INTERNAL_API_KEY` | Yes | Shared secret with xgateway |
| `JWT_SECRET` | Yes | Signs admin JWT tokens |
| `PROXY_URL` | Yes | Public URL of xgateway (returned in 402 responses) |
| `PORT` | No | Listen port (default: `9090`) |
| `MODE` | No | Gin mode: `release` or `debug` |
| `APP_BASE_URL` | No | Frontend base URL for password-reset links (default `http://localhost:3000`) |
| `GOOGLE_CLIENT_ID` | No | OAuth client ID; required for Google sign-in (`POST /auth/google`) |
| `SMTP_HOST` | No | SMTP server for welcome + password-reset email. Empty ⇒ email disabled (links logged and returned in API responses) |
| `SMTP_PORT` | No | SMTP port, STARTTLS (default `587`) |
| `SMTP_USERNAME` / `SMTP_PASSWORD` | No | SMTP credentials (PLAIN auth) |
| `SMTP_FROM` / `SMTP_FROM_NAME` | No | From address (defaults to `SMTP_USERNAME`) and display name (default `xpaywall`) |

> **Superadmin:** there is no bootstrap-admin env var. Create one with
> `install user … --role superadmin`, or grant the role in Postgres:
> `UPDATE users SET role='superadmin' WHERE username='...';`

---

## File-Based Config (xgateway)

When `CONFIG_PROVIDER=file`, rules are loaded from a YAML file at startup:

```yaml
x402:                           # required when any rule uses an x402 method
  - name: base-exact
    facilitator_url: https://x402.dexter.cash
    network: eip155:8453        # Base mainnet (CAIP-2)
    scheme: exact               # exact
    merchant: "0xYourAddress"        # pay_to address
    asset: "0xUSDCAddress"           # CAIP-19 asset

mpp:                            # required when any rule uses an MPP method
  - name: tempo-charge
    method: tempo               # only "tempo" is supported
    scheme: charge              # one-time charge only (no session)
    rpc_url: https://rpc.moderato.tempo.xyz   # Tempo JSON-RPC
    asset: "0x20c0000000000000000000000000000000000000"   # pathUSD (Tempo stablecoin)

outbound:
  target: http://your-upstream:4021
  allow_unmatched: false        # 403 for paths with no rule
  rules:
    - name: weather
      path: /weather
      price: "$0.10"
      payment_methods: [base-exact]

    - name: usage-api           # paid via MPP Tempo charge
      path: /api/usage
      price: "$0.10"
      payment_methods: [tempo-charge]

    - name: health-check
      path: /health
      free: true                # no payment required
```

See [`xgateway/config.dev.yaml`](xgateway/config.dev.yaml) for the full schema.

---

## Payment Methods

| Protocol | Status | Description | Networks |
|---|---|---|---|
| **x402** | Shipped | EVM-based micropayments (`exact` scheme), settled via a facilitator | Base, Base Sepolia, any EVM chain with an x402 facilitator |
| **MPP / Tempo** | Shipped | Machine Payments Protocol — one-time Tempo `charge` payments (pathUSD), no facilitator | Tempo |
| **Stripe** | Roadmap | Traditional card payments | — |

---

## Tech Stack

**xgateway & control-api** — Go 1.26, Gin, pgx/v5, sqlc, goose, urfave/cli/v3

**adminpanel** — React 19, TypeScript, MUI v7, Redux Toolkit, Vite 7, Axios

**Infrastructure** — PostgreSQL 16, nginx, Docker

---

## API Overview

### control-api

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/auth/login` | — | Get JWT token |
| POST | `/auth/register` | — | Self-register a local account |
| POST | `/auth/google` | — | Sign in with a Google ID token |
| POST | `/auth/forgot-password` / `/auth/reset-password` | — | Request a reset link / set a new password |
| GET | `/auth/me` | JWT | Current user info |
| GET/POST | `/api/v1/projects` | JWT | List / create projects |
| GET/PUT/DELETE | `/api/v1/projects/:id` | JWT | Get / update / delete project |
| GET/POST | `/api/v1/outbound-routes` | JWT | Manage payment routes |
| `*` | `/api/v1/payment-methods`, `/payment-method-assets`, `/facilitators` | JWT | Manage payment building blocks (global ones are superadmin-only to mutate) |
| `*` | `/api/v1/project-payment-methods` | JWT | Link a project to a payment method + asset |
| `*` | `/api/v1/users` | JWT (superadmin) | User management |
| GET | `/api/v1/stats/dashboard` | JWT | Dashboard stats |
| GET | `/api/v1/request-logs` | JWT | Paginated request logs |
| GET | `/proxy/resolve/{username}/{slug}/{path}` | API Key | Rule resolution (used by xgateway) |
| POST | `/api/v1/request-logs` | API Key | Log ingestion (used by xgateway) |
| POST | `/api/v1/request-events` | API Key | Per-step event ingestion (used by xgateway) |

---

## Development

### control-api

```bash
# Run migrations manually
go tool goose -dir migrations postgres "$CONTROL_DB_DSN" up

# Regenerate sqlc (never hand-edit internal/storage/postgres/generated/)
sqlc generate

# Tests & lint
go test ./...
golangci-lint run
```

### xgateway

```bash
go test ./...
go test -run TestName ./internal/rules/
golangci-lint run
```

### adminpanel

```bash
yarn tsc          # type-check
yarn lint         # eslint
yarn lint:fix
yarn prettier
yarn build        # production build
```

---

## Project Structure

```
xpaywall/
├── xgateway/               # Payment proxy service
│   ├── cmd/xgateway/       # Entry point
│   └── internal/
│       ├── proxy/          # Gin server, payment verification, upstream proxy
│       ├── rules/          # Rule providers (file, http)
│       └── logger/         # Async log client
│
├── control-api/            # Control plane
│   ├── cmd/control-api/    # Entry point + install subcommand
│   ├── internal/
│   │   ├── http/           # Gin routes, handlers, middleware
│   │   └── storage/postgres/
│   │       ├── generated/  # sqlc output — do not edit
│   │       └── queries/    # SQL query definitions
│   └── migrations/         # goose SQL migrations
│
├── frontend/adminpanel/    # React admin dashboard
│   └── src/
│       ├── api/            # Typed axios wrappers
│       ├── views/          # Page components
│       ├── store/          # Redux slices
│       └── utils/axios.ts  # HTTP client (reads window.__CONFIG__)
│
├── examples/server/        # Sample upstream API
└── docker-compose.yml
```

---

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the dev setup, PR checklist, and contribution rules. Changes scoped to the gateway go through the [xgateway](https://github.com/cp0x-org/xgateway) repository — see [`xgateway/CONTRIBUTING.md`](xgateway/CONTRIBUTING.md).

## License

Released under the [MIT License](LICENSE).
