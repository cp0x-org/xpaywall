# xpaywall control-api

REST control plane for the [xpaywall](../README.md) stack. control-api owns the database schema and exposes:

- the admin REST API consumed by `frontend/adminpanel`
- the internal `/proxy/resolve/*path` endpoint that `xgateway` calls to fetch payment rules per request
- the log-ingestion endpoints that `xgateway` posts request logs and events to

Built with **Go 1.26**, **Gin**, **pgx/v5**, **sqlc**, and **goose**.

## Run

```bash
export CONTROL_DB_DSN="postgres://user:password@localhost:5432/xpaywalldb?sslmode=disable"
export INTERNAL_API_KEY="change-me-internal-secret-key"
export JWT_SECRET="change-me-jwt-secret-key"
export PROXY_URL="http://127.0.0.1:8081"

go run ./cmd/control-api --env-file .env
```

Or use the bundled `install` subcommand to run migrations once and exit:

```bash
go run ./cmd/control-api --env-file .env install
```

`install demo` additionally seeds demo data.

## Environment variables

| Variable | Required | Default | Notes |
|---|---|---|---|
| `CONTROL_DB_DSN` | yes | — | PostgreSQL DSN |
| `INTERNAL_API_KEY` | yes | — | Shared secret with xgateway (sent as `X-Api-Key`) |
| `JWT_SECRET` | yes | — | Signs admin JWT tokens |
| `PROXY_URL` | yes | — | Public URL of xgateway (returned to the dashboard and in 402 responses) |
| `PORT` | no | `9090` | HTTP listen port |
| `MODE` | no | `release` | Gin mode (`release` or `debug`) |
| `DEBUG` | no | `false` | Verbose logging |
| `CORS_ORIGINS` | no | `*` | Comma-separated allowed origins |
| `APP_BASE_URL` | no | `http://localhost:3000` | Frontend base URL for password-reset links |
| `GOOGLE_CLIENT_ID` | no | — | OAuth client ID; required for Google sign-in (`POST /auth/google`) |

> **Superadmin:** there is no bootstrap-admin env var. Provision the role directly in
> Postgres: `UPDATE users SET role='superadmin' WHERE username='...';`

`--env-file <path>` loads variables from a dotenv file before parsing. `.env` in the working directory is loaded automatically when present.

## Migrations

Migrations live in `migrations/` and are managed by goose. The `install` subcommand applies them automatically. To run them manually:

```bash
go tool goose -dir migrations postgres "$CONTROL_DB_DSN" status
go tool goose -dir migrations postgres "$CONTROL_DB_DSN" up
go tool goose -dir migrations postgres "$CONTROL_DB_DSN" down
```

## Database access

All queries live in `internal/storage/postgres/queries/*.sql`. Generated Go code is in `internal/storage/postgres/generated/` — **never hand-edit those files**. After changing any `.sql` query or migration, regenerate:

```bash
sqlc generate
```

`sqlc.yaml` maps `uuid` → `github.com/google/uuid.UUID` and nullable text/varchar → `*string`.

## Route groups

| Prefix | Auth | Purpose |
|---|---|---|
| `/auth/...` | none | Login (`POST /auth/login`), current user (`GET /auth/me`) |
| `/api/v1/...` | JWT Bearer | Admin CRUD — projects, routes, users, logs, stats, payment channels |
| `/proxy/resolve/*path` | `X-Api-Key` header | xgateway route resolution |
| `/api/v1/request-logs`, `/api/v1/request-events` | `X-Api-Key` header | xgateway log ingestion |

Handlers live in `internal/http/handlers/`. JWT middleware is in `internal/http/middleware/jwt.go`; API-key middleware in `internal/http/middleware/apikey.go`.

## Test and lint

```bash
go test ./...
go test ./internal/http/...      # single package
go test -run TestFooBar ./...    # single test
golangci-lint run
```

## Swagger

After changing handler annotations, regenerate the OpenAPI spec:

```bash
go tool swag init -g swagger_meta.go --dir ./cmd/control-api,./ --output docs --parseInternal
```

## Contributing

Contributions follow the rules in the top-level [`CONTRIBUTING.md`](../CONTRIBUTING.md).

## License

Released under the [MIT License](../LICENSE).
