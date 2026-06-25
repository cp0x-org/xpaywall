# 12 — control-api CLI

control-api ships with a small CLI for **migrations**, **user management**, **global payment methods**, and **demo data**. It is the only way to create users today — the admin panel does not yet have a Users screen.

> User and project management UI is on the [roadmap](./11-roadmap.md). Until it lands, every login account is created from the command line or the login-page registration.

## Commands at a glance

| Command | What it does |
|---|---|
| `control-api` | Start the HTTP server (default action — no subcommand). |
| `control-api migrate` | Apply pending database migrations (goose up). |
| `control-api install user --username U --password P --email E [--role user\|superadmin]` | Create a single login account. |
| `control-api install payment-methods --superadmin U` | Seed the global x402 + MPP payment methods/assets/facilitator, owned by an existing superadmin. |
| `control-api install demo` | Seed a demo workspace (project, routes, request logs) for a user. Requires `install payment-methods` first. |
| `control-api install demo-mpp` | Same as `install demo` but for an MPP/Tempo charge project. |

Every command accepts `--env-file <path>` and reads the same env vars as the running service (`CONTROL_DB_DSN` is the only one required for CLI use). See [03 — Configuration](./03-configuration.md#control-api) for the full env-var table.

---

## How to run the CLI

You can run the CLI in three places. Pick whichever fits your setup.

### 1. From a running stack (recommended)

When `docker compose up -d` is already running:

```bash
docker compose run --rm control-api migrate
docker compose run --rm control-api install user --username alice --password s3cret --email alice@example.com --role superadmin
docker compose run --rm control-api install payment-methods --superadmin alice
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
go run ./cmd/control-api --env-file .env install user --username alice --password s3cret --email alice@example.com --role superadmin
go run ./cmd/control-api --env-file .env install payment-methods --superadmin alice
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
control-api install user --username <name> --password <plaintext> --email <email> [--role user|superadmin]
```

Inserts a row into `users` with a bcrypt-hashed password. Fails if the username already exists.

| Flag | Required | Default | Notes |
|---|---|---|---|
| `--username` | yes | — | Login username. Must be unique and **slug-safe**: only letters, digits, `_` and `-` (it appears in the proxy URL). |
| `--password` | yes | — | Plaintext password — the CLI hashes it with bcrypt before storing. |
| `--email` | yes | — | Account email. Must be unique. |
| `--role` | no | `user` | `user` or `superadmin`. Superadmins manage global payment methods/assets/facilitators. |

Example — create the first superadmin (needed before seeding global payment methods):

```bash
docker compose run --rm control-api install user \
  --username alice \
  --password 'choose-a-long-passphrase' \
  --email alice@example.com \
  --role superadmin
```

> **Role.** `--role` is the only place to choose a role at creation. Accounts created via login-page
> registration always get the `user` role; promote them with the CLI is not possible after the fact —
> change the role directly in Postgres: `UPDATE users SET role='superadmin' WHERE username='...';`.
> See [01 — Login & Users](./04-admin-panel/01-login-and-users.md).

---

## `install payment-methods` — seed the global payment methods

```
control-api install payment-methods --superadmin <username>
```

Creates the **global** payment building blocks that the demos (and your projects) attach to:

- An **x402 Base Sepolia** payment method with a **USDC** asset and the **`x402 Coinbase`** facilitator.
- An **MPP Tempo Charge** payment method with a **pathUSD** asset (no facilitator).

All of them are marked `is_global` (visible to every user) and **owned by the named superadmin**.

| Flag | Required | Notes |
|---|---|---|
| `--superadmin` | yes | Username of an existing superadmin who will own the global entities. |

If the user does not exist, or is not a superadmin, the command fails with a clear error telling you to create a superadmin first (`install user … --role superadmin`).

The command is idempotent — re-running it upserts the same rows by code/name.

```bash
docker compose run --rm control-api install payment-methods --superadmin alice
```

---

## `install demo` — seed a demo workspace

```
control-api install demo [--username demo] [--password demo] [--email demo@example.com]
                         [--skip-logs] [--log-days 7] [--log-count 75]
```

Seeds a workspace **for a user** so you can poke at the admin panel and gateway without building everything by hand. It creates only the user-owned pieces — the global payment methods must already exist (`install payment-methods`):

- A user (default **non-superadmin** `demo` / `demo` / `demo@example.com`), **created if missing**. The chosen credentials are printed to the console.
- A **Default Project** with slug `default`, owned by that user.
- Routes settings pointing at `http://localhost:4021` (the bundled example-server).
- The project ↔ payment-method link to the global **x402 Base Sepolia** method.
- Six routes: `/health` (paid), `/api/metered/*` (paid), `/weather` (paid), `/free-endpoint`, `/free-multipoint`, `/free-multipoint/*`.
- 75 randomised entries in `request_logs` spread over the last 7 days (so the dashboard charts and the requests table have something to show).

| Flag | Default | Notes |
|---|---|---|
| `--username` | `demo` | Login username for the demo workspace owner (created if missing). |
| `--password` | `demo` | Plaintext password. Hashed before storing. |
| `--email` | `demo@example.com` | Email for the demo user. |
| `--skip-logs` | off | Skip the `request_logs` seed. |
| `--log-days` | `7` | Spread log entries over this many past days. |
| `--log-count` | `75` | Number of `request_logs` rows to insert. |

> **Prerequisite.** `install demo` requires the global payment method to exist. If you have not run
> `install payment-methods` yet, it fails with `payment method "x402_base_sepolia" not found — run
> "install payment-methods" first`.

Once seeded, the demo project's proxy URL is `http://<gateway>/demo/default/<route>` — the path is
`/{username}/{project_slug}/{route}` (here username `demo`, slug `default`).

Example — seed for a specific user with no logs:

```bash
docker compose run --rm control-api install demo \
  --username demo \
  --password 'demo-only-do-not-ship' \
  --email demo@example.com \
  --skip-logs
```

`install demo-mpp` is identical but seeds a `mpp-demo` project wired to the global **MPP Tempo Charge** method.

The seed is **idempotent for structural data** — re-running it does not duplicate the project, link, or routes. The `request_logs` seed runs only when the table is empty for that project, so reruns skip it as well.

---

## Recommended bootstrap order

For a brand-new database:

```bash
docker compose up -d postgres                                            # 1. start Postgres
docker compose run --rm control-api migrate                             # 2. create schema
docker compose run --rm control-api install user \
  --username alice --password '<passphrase>' \
  --email alice@example.com --role superadmin                           # 3. create a superadmin
docker compose run --rm control-api install payment-methods \
  --superadmin alice                                                    # 4. seed global payment methods
docker compose run --rm control-api install demo                        # 5. (optional) seed demo data
docker compose up -d                                                    # 6. start the rest of the stack
```

If you only want the schema, stop after step 2 and create your first user with `install user`. If you
do not need the demo workspace, stop after step 4.

---

## Next steps

- [01 — Login & Users](./04-admin-panel/01-login-and-users.md) — how accounts and roles work, and how to log in.
- [02 — Setup](./02-setup.md) — full-stack quick start with Docker Compose.
- [11 — Roadmap](./11-roadmap.md) — when the admin panel will gain a Users screen.
