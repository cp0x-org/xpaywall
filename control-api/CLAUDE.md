# CLAUDE.md

This file provides rules and guidance Claude Code (claude.ai/code) when working with code in this repository.


# Rules
These rules apply to every task in this project unless explicitly overridden.
Bias: caution over speed on non-trivial work. Use judgment on trivial tasks.

## Rule 1 — Think Before Coding
State assumptions explicitly. If uncertain, ask rather than guess.
Present multiple interpretations when ambiguity exists.
Push back when a simpler approach exists.
Stop when confused. Name what's unclear.

## Rule 2 — Simplicity First
Minimum code that solves the problem. Nothing speculative.
No features beyond what was asked. No abstractions for single-use code.
Test: would a senior engineer say this is overcomplicated? If yes, simplify.

## Rule 3 — Surgical Changes
Touch only what you must. Clean up only your own mess.
Don't "improve" adjacent code, comments, or formatting.
Don't refactor what isn't broken. Match existing style.

## Rule 4 — Goal-Driven Execution
Define success criteria. Loop until verified.
Don't follow steps. Define success and iterate.
Strong success criteria let you loop independently.

## Rule 5 — Use the model only for judgment calls
Use me for: classification, drafting, summarization, extraction.
Do NOT use me for: routing, retries, deterministic transforms.
If code can answer, code answers.

## Rule 6 — Token budgets are not advisory
Per-task: 4,000 tokens. Per-session: 30,000 tokens.
If approaching budget, summarize and start fresh.
Surface the breach. Do not silently overrun.

## Rule 7 — Surface conflicts, don't average them
If two patterns contradict, pick one (more recent / more tested).
Explain why. Flag the other for cleanup.
Don't blend conflicting patterns.

## Rule 8 — Read before you write
Before adding code, read exports, immediate callers, shared utilities.
"Looks orthogonal" is dangerous. If unsure why code is structured a way, ask.

## Rule 9 — Tests verify intent, not just behavior
Tests must encode WHY behavior matters, not just WHAT it does.
A test that can't fail when business logic changes is wrong.

## Rule 10 — Checkpoint after every significant step
Summarize what was done, what's verified, what's left.
Don't continue from a state you can't describe back.
If you lose track, stop and restate.

## Rule 11 — Match the codebase's conventions, even if you disagree
Conformance > taste inside the codebase.
If you genuinely think a convention is harmful, surface it. Don't fork silently.

## Rule 12 — Fail loud
"Completed" is wrong if anything was skipped silently.
"Tests pass" is wrong if any were skipped.
Default to surfacing uncertainty, not hiding it.

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

# Regenerate Swagger docs (after adding/changing handler annotations)
go tool swag init -g swagger_meta.go --dir ./cmd/control-api,./ --output docs --parseInternal

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
| `PORT` | Default `9090` |
| `MODE` | `release` or `debug` (Gin mode) |
| `APP_BASE_URL` | Frontend base URL for password-reset links (default `http://localhost:3000`) |
| `GOOGLE_CLIENT_ID` | OAuth client ID; audience for verifying Google ID tokens. Required for `POST /auth/google` |

### Roles & data scoping
- `users.role` is `user` (default) or `superadmin`. There is no env superadmin and no API to
  change roles — provision one directly in Postgres:
  `UPDATE users SET role='superadmin' WHERE username='...'`. The role is carried in the JWT.
- **User-scoped data** (projects, routes, project settings, project-payment-methods, request
  logs/events, stats) is visible/mutable **only by the owning user** — role grants no access to
  other users' data.
- **Global-capable entities** (`payment_methods`, `payment_method_assets`, `facilitators`) carry
  `is_global` + `owner_user_id`: visible when `is_global` OR owned by the caller (superadmin sees
  all). Only a superadmin may set `is_global`; deleting a global entity is superadmin-only.

## Architecture

### Entry point
`cmd/control-api/main.go` → `internal/app.Run()` → builds DB pool, wires handlers, calls `internalhttp.SetupRouter()`, starts Gin.

### Route groups

| Prefix | Auth | Purpose |
|---|---|---|
| `/auth/...` | none | Login (`POST /auth/login`), register (`POST /auth/register`), forgot/reset password (`POST /auth/forgot-password`, `POST /auth/reset-password`), Google (`POST /auth/google`) |
| `/auth/me`, `/auth/change-password` | JWT Bearer | Current user; change password |
| `/api/v1/...` | JWT Bearer | Admin CRUD — projects, routes, users, logs, stats, payment channels |
| `/proxy/resolve/*path` | `X-Internal-API-Key` header | xgateway route resolution |
| `/api/v1/request-logs`, `/api/v1/request-events` | `X-Internal-API-Key` | xgateway log ingestion |

Admin routes are registered in `internal/http/routes/admin.go`; internal/proxy routes in `internal/http/routes/internal.go` and `proxy.go`.

### Handler structure
- `internal/http/handlers/handler.go` — `Handler` struct holding `*postgres.Queries` and `DBTX`; all admin handlers are methods on it.
- `internal/http/handlers/auth/` — separate `Handler` for login/JWT; holds superadmin credentials.
- `internal/http/handlers/gateway/` — `ResolveRoute` for xgateway. Path format: `/{username}/{projectSlug}/{inboundPath}`. Looks up route + payment channels, returns full proxy rule as JSON.

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
