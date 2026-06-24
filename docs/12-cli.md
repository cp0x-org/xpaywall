# 12 — control-api CLI

control-api ships with a small CLI for **migrations**, **user management**, and **demo data**. It is the only way to create users today — the admin panel does not yet have a Users screen.

> User and project management UI is on the [roadmap](./11-roadmap.md). Until it lands, every login account is created from the command line.

## Commands at a glance

| Command | What it does |
|---|---|
| `control-api` | Start the HTTP server (default action — no subcommand). |
| `control-api migrate` | Apply pending database migrations (goose up). |
| `control-api install user --username U --password P` | Create a single login account. |
| `control-api install demo` | Seed a complete demo workspace: admin user, project, payment method, routes, and sample request logs. |

Every command accepts `--env-file <path>` and reads the same env vars as the running service (`CONTROL_DB_DSN` is the only one required for CLI use). See [03 — Configuration](./03-configuration.md#control-api) for the full env-var table.

---

## How to run the CLI

You can run the CLI in three places. Pick whichever fits your setup.

### 1. From a running stack (recommended)

When `docker compose up -d` is already running:

```bash
docker compose run --rm control-api migrate
docker compose run --rm control-api install user --username alice --password s3cret
docker compose run --rm control-api install demo
```

`docker compose run` reuses the `control-api` image and env vars, but starts a fresh, ephemeral container that exits when the command finishes. The `--rm` flag deletes it after.

### 2. With the dedicated CLI service

`docker-compose.yml` defines a profile-gated `control-api-cli` service that uses the same image but never starts automatically. Use it when you want to script init steps:

```bash
docker compose --profile cli run --rm control-api-cli migrate
docker compose --profile cli run --rm control-api-cli install demo
```

Because it lives behind the `cli` profile, `docker compose up` ignores it.

### 3. From source (Go installed)

For local development:

```bash
cd control-api
go run ./cmd/control-api --env-file .env migrate
go run ./cmd/control-api --env-file .env install user --username alice --password s3cret
go run ./cmd/control-api --env-file .env install demo
```

---

## `migrate` — apply database migrations

```
control-api migrate [--dir migrations]
```

Runs `goose up` against `CONTROL_DB_DSN`. The bundled Docker image has migrations at `/app/migrations`, so the default `--dir` value works without further configuration.

| Flag | Default | Notes |
|---|---|---|
| `--dir` | `migrations` | Path to the goose migrations directory. |

Example — re-run migrations after pulling a new release:

```bash
docker compose run --rm control-api migrate
```

> The `control-api` HTTP server does **not** run migrations on boot. If you skip this step after a release that adds migrations, the API will likely fail with a missing-column or missing-table error.

---

## `install user` — create a login account

```
control-api install user --username <name> --password <plaintext>
```

Inserts a row into `users` with a bcrypt-hashed password. Fails if the username already exists.

| Flag | Required | Notes |
|---|---|---|
| `--username` | yes | Login username. Must be unique. |
| `--password` | yes | Plaintext password — the CLI hashes it with bcrypt before storing. |

Example — create your first non-bootstrap admin:

```bash
docker compose run --rm control-api install user --username alice --password 'choose-a-long-passphrase'
```

> **Role.** Accounts created this way (and via the login-page registration) have the `user` role.
> To make one a superadmin — required to manage global payment methods/assets/facilitators — set it
> in Postgres: `UPDATE users SET role='superadmin' WHERE username='...';`. There is no env var or
> CLI flag for it. See [01 — Login & Users](./04-admin-panel/01-login-and-users.md).

---

## `install demo` — seed a demo workspace

```
control-api install demo [--admin-username admin] [--admin-password admin]
                          [--skip-logs] [--log-days 7] [--log-count 75]
```

Seeds a complete sandbox you can poke at without configuring anything from scratch:

- An admin user (default `admin` / `admin`).
- A **Default Project** with slug `default`.
- Routes settings pointing at `http://localhost:4021` (the bundled example-server).
- A `Base Sepolia` payment method with USDC asset, attached to the project via the `x402 Coinbase` facilitator.
- Six routes: `/health` (paid), `/api/metered/*` (paid), `/weather` (paid), `/free-endpoint`, `/free-multipoint`, `/free-multipoint/*`.
- 75 randomised entries in `request_logs` spread over the last 7 days (so the dashboard charts and the requests table have something to show).

| Flag | Default | Notes |
|---|---|---|
| `--admin-username` | `admin` | Username for the seeded admin. |
| `--admin-password` | `admin` | Plaintext password. Hashed before storing. |
| `--skip-logs` | off | Skip the `request_logs` seed. |
| `--log-days` | `7` | Spread log entries over this many past days. |
| `--log-count` | `75` | Number of `request_logs` rows to insert. |

Example — seed an empty stack with custom credentials and no logs:

```bash
docker compose run --rm control-api install demo \
  --admin-username demo \
  --admin-password 'demo-only-do-not-ship' \
  --skip-logs
```

The seed is **idempotent for structural data** — re-running it does not duplicate the project, payment method, asset, or routes. The `request_logs` seed runs only when the table is empty for the default project, so reruns will skip it as well.

---

## Recommended bootstrap order

For a brand-new database:

```bash
docker compose up -d postgres                      # 1. start Postgres
docker compose run --rm control-api migrate        # 2. create schema
docker compose run --rm control-api install demo   # 3. (optional) seed demo data
docker compose up -d                               # 4. start the rest of the stack
```

If you only want the schema, stop after step 2 and create your first user with `install user` instead of `install demo`.

---

## Next steps

- [01 — Login & Users](./04-admin-panel/01-login-and-users.md) — how the bootstrap account works and how to log in.
- [02 — Setup](./02-setup.md) — full-stack quick start with Docker Compose.
- [11 — Roadmap](./11-roadmap.md) — when the admin panel will gain a Users screen.
