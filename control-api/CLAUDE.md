# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run
go run ./cmd/control-api --env-file .env

# Test
go test ./...
go test ./internal/http/...          # single package
go test -run TestFooBar ./...        # single test

# Lint
golangci-lint run

# Migrations (requires CONTROL_DB_DSN env var)
go tool goose -dir migrations postgres "$CONTROL_DB_DSN" status
go tool goose -dir migrations postgres "$CONTROL_DB_DSN" up
go tool goose -dir migrations postgres "$CONTROL_DB_DSN" down

# Regenerate sqlc (after editing queries or migrations)
sqlc generate

# Seed demo data
go run ./cmd/control-api --env-file .env install demo
```

## Required env vars

| Var | Notes |
|---|---|
| `CONTROL_DB_DSN` | Postgres DSN |
| `INTERNAL_API_KEY` | Shared secret for xgateway → control-api calls |
| `JWT_SECRET` | Signs admin JWT tokens |
| `PROXY_URL` | Public xgateway URL (returned to dashboard) |
| `SUPERADMIN_USERNAME` / `SUPERADMIN_PASSWORD` | Bootstrap credentials |
| `PORT` | Default `9090` |
| `MODE` | `release` or `debug` (Gin mode) |

## Architecture

### Entry point
`cmd/control-api/main.go` → `internal/app.Run()` → builds DB pool, wires handlers, calls `internalhttp.SetupRouter()`, starts Gin.

### Route groups

| Prefix | Auth | Purpose |
|---|---|---|
| `/auth/...` | none | Login (`POST /auth/login`), me (`GET /auth/me`) |
| `/api/v1/...` | JWT Bearer | Admin CRUD — projects, routes, users, logs, stats, payment channels |
| `/proxy/resolve/*path` | `X-Internal-API-Key` header | xgateway route resolution |
| `/api/v1/request-logs`, `/api/v1/request-events` | `X-Internal-API-Key` | xgateway log ingestion |

Admin routes are registered in `internal/http/routes/admin.go`; internal/proxy routes in `internal/http/routes/internal.go` and `proxy.go`.

### Handler structure
- `internal/http/handlers/handler.go` — `Handler` struct holding `*postgres.Queries` and `DBTX`; all admin handlers are methods on it.
- `internal/http/handlers/auth/` — separate `Handler` for login/JWT; holds superadmin credentials.
- `internal/http/handlers/gateway/` — `ResolveRoute` for xgateway. Path format: `/{projectSlug}/{inboundPath}`. Looks up route + payment channels, returns full proxy rule as JSON.

### Database layer
- All queries live in `internal/storage/postgres/queries/*.sql`.
- Generated Go code is in `internal/storage/postgres/generated/` — **never hand-edit these files**.
- After changing any `.sql` query or migration, run `sqlc generate`.
- `sqlc.yaml` maps `uuid` → `github.com/google/uuid.UUID`; nullable text/varchar → `*string`.

### Data model (key tables)
- `projects` — top-level tenant; identified externally by `slug`.
- `routes` — per-project inbound path patterns with `price_usd`.
- `project_routes_settings` — per-project upstream `base_url`, optional auth header, `allow_unmatched`.
- `payment_channels` — global channel definitions (protocol, method, scheme, JSON metadata).
- `project_payment_configs` — links a project to a payment channel with a `payout_address`.
- `request_logs` / `request_events` — written by xgateway via internal API.

### Authentication
- JWT middleware: `internal/http/middleware/jwt.go` — validates `Authorization: Bearer <token>`.
- API key middleware: `internal/http/middleware/apikey.go` — validates `X-Internal-API-Key` header against `cfg.InternalAPIKey`.

### `pkg/sdk`
Public Go client package for external consumers of control-api; currently scaffolded only.
