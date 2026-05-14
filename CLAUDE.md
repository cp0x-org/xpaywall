# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**xpaywall** is an x402/MPP payment gateway system. It sits in front of upstream APIs and enforces micropayments before proxying requests. The system has three runnable services:

| Service | Dir | Port | Role |
|---|---|---|---|
| xgateway | `xgateway/` | 8081 | Reverse proxy that enforces payment rules |
| control-api | `control-api/` | 9091 | REST control plane — manages projects, routes, users, logs |
| adminpanel | `frontend/adminpanel/` | 3000 | React admin dashboard for control-api |

An example upstream server lives in `examples/server/` (port 4021).

---

## Commands

### Run all services
```bash
docker-compose up
```

### xgateway
```bash
cd xgateway
go run ./cmd/xgateway --env-file .env
go test ./...
go test -run TestName ./internal/rules/   # single test or package
golangci-lint run
```

### control-api
```bash
cd control-api
go run ./cmd/control-api --env-file .env
go run ./cmd/control-api install          # run DB migrations then exit
go test ./...
golangci-lint run

# Migrations via goose CLI (requires CONTROL_DB_DSN env var)
go tool goose -dir migrations postgres "$CONTROL_DB_DSN" status
go tool goose -dir migrations postgres "$CONTROL_DB_DSN" up
go tool goose -dir migrations postgres "$CONTROL_DB_DSN" down

# Regenerate sqlc (never hand-edit internal/storage/postgres/generated/)
sqlc generate
```

### adminpanel
```bash
cd frontend/adminpanel
yarn start        # dev server
yarn build        # production build
yarn lint
yarn lint:fix
yarn prettier
yarn tsc          # type-check only
```

---

## Environment Variables

### xgateway (required)
| Var | Values | Notes |
|---|---|---|
| `CONFIG_PROVIDER` | `file` / `http` | Selects rule provider mode |
| `CONFIG_FILE` | path | Required when `CONFIG_PROVIDER=file` |
| `CONTROL_API_URL` | URL | Required when `CONFIG_PROVIDER=http` |
| `INTERNAL_API_KEY` | string | Required when `CONFIG_PROVIDER=http`; sent as `X-Api-Key` header |

### control-api (required)
| Var | Notes |
|---|---|
| `CONTROL_DB_DSN` | PostgreSQL DSN |
| `INTERNAL_API_KEY` | Must match xgateway's key |
| `JWT_SECRET` | Signs admin panel tokens |
| `PROXY_URL` | Public URL of xgateway (returned in 402 responses) |

---

## Architecture

### xgateway — Proxy Service

Entry: `cmd/xgateway/main.go` → `internal/app.Run()` → `internal/proxy`

The proxy intercepts every request and resolves a payment rule for it. If no valid payment proof is present, it returns an HTTP 402 response with payment instructions. If payment is verified, it proxies to the configured upstream target.

- `internal/rules/` — two rule providers:
  - **file**: loads rules from a local YAML config at startup
  - **http**: fetches rules from control-api per request (used in production); calls `GET /proxy/resolve/*path` with `X-Api-Key` auth
- `internal/proxy/` — Gin HTTP server, payment verification, upstream proxying
- `internal/logger/` — async client that POSTs request logs to control-api; no-op when `baseURL` is empty

Config mode is set via `CONFIG_PROVIDER` env var or `--config-provider` flag.

#### File-based config format (YAML)
```yaml
x402:
  - name: mychain
    facilitator_url: https://...
    network: eip155:8453          # CAIP-2 format
    scheme: exact                 # exact | upto
    merchant_address: "0x..."

routes:
  - path: /api/resource
    target: http://upstream:4021
    price: "0.01"
    payment_methods:
      - x402:mychain

allow_unmatched: false            # proxy unmatched paths without payment
```

### control-api — Control Plane

Entry: `cmd/control-api/main.go` → `internal/app.Run()` → `internal/http`

Route groups (registered under Gin):
- **`/api/v1/`** — JWT-authenticated routes for the dashboard (projects, routes, users, payment-channels, stats, logs)
- **`/proxy/`** — Unauthenticated payment callback endpoints; `GET /proxy/resolve/*path` used by xgateway
- **`/auth/`** — Login and token refresh

xgateway authenticates to control-api with `X-Api-Key: $INTERNAL_API_KEY` on `/proxy/` and log-ingestion endpoints.

Database access is via **sqlc**-generated code in `internal/storage/postgres/generated/`. Never hand-edit files in that directory — regenerate with `sqlc generate` (config at `sqlc.yaml`).

Migrations are managed with **goose** in `migrations/`. The `install` CLI subcommand runs migrations automatically on startup.

### adminpanel — React Dashboard

Entry: `src/index.tsx` → Redux store → React Router → page views

- `src/api/` — typed API client (axios-based) for control-api
- `src/store/` — Redux Toolkit slices
- `src/views/` — page-level components (dashboard, projects, routes, logs, etc.)
- `src/ui-component/` — shared presentational components
- `src/types/` — TypeScript type definitions mirroring control-api response shapes

UI framework: **MUI v7**. State: **Redux Toolkit**. Build: **Vite 7**.

---

## Key Patterns

### Go backend
- Go 1.26; both services use `urfave/cli/v3` for CLI parsing
- Gin for HTTP routing in both services
- `pgx/v5` + sqlc for all database access — no ORM
- Config is loaded from env vars; `.env` file supported via `--env-file` flag
- `golangci-lint` is enforced — keep linter clean; no `logrus`, no `pkg/errors`; max line length 140

### Frontend
- API calls go through `src/api/` wrappers, not direct `fetch`/axios calls
- Use `yarn` (v4), not npm
- `VITE_APP_API_URL` controls which control-api the frontend points to

### Payment flow
1. Client hits xgateway
2. xgateway resolves rule (file or HTTP provider)
3. No valid payment → 402 + payment requirements returned
4. Client pays (x402 or MPP/Tempo or Stripe)
5. Client retries with proof header
6. xgateway verifies proof → proxies request → logs to control-api
