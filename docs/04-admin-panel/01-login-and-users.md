# Admin Panel — Login & Users

The admin panel is a small web app at `http://<your-host>:3104` (port `3104` by default). Projects, payment methods and routes are all managed here.

> **Accounts** can be self-registered from the login page (with password reset and Google sign-in)
> or created from the control-api CLI. A full user-management screen is on the [roadmap](./../11-roadmap.md).
> See [12 — control-api CLI](./../12-cli.md) for the CLI commands.

## First login

There is no bootstrap-admin env var. Create your first account one of two ways:

- **Register on the login page** — the panel has Sign-up, "Forgot password", and Google sign-in. Registration and password reset deliver email when SMTP is configured (see [03 — Configuration](./../03-configuration.md#control-api)); without SMTP the reset link is logged and returned in the API response.
- **Create one from the CLI** — `install user … --role superadmin` (see below). To manage **global** payment methods/assets/facilitators you need a superadmin.

Open the login page, enter the credentials (username **or** email + password), and submit.

![Login page](./../images/login.png "small")

After login you land on the Dashboard. The left sidebar gives you access to every other section.

![Sidebar menu](./../images/sidebar-menu.png "small")

> **Superadmin.** Managing *global* payment methods, assets, and facilitators requires the
> `superadmin` role, which is granted directly in Postgres:
> `UPDATE users SET role='superadmin' WHERE username='...';`

## Creating additional users

There is no Users screen in the admin panel. Add accounts from the CLI:

```bash
docker compose run --rm control-api install user \
  --username alice \
  --password 'choose-a-long-passphrase' \
  --email alice@example.com \
  --role superadmin     # omit for a regular user
```

`--email` is required and `--role` defaults to `user`. The username must be **slug-safe** (letters,
digits, `_`, `-`) because it appears in the proxy URL. The CLI hashes the password with bcrypt before
storing it. Run it once per account. The new account can log in immediately — no restart required.

For local development without Docker:

```bash
cd control-api
go run ./cmd/control-api --env-file .env install user \
  --username alice --password 'choose-a-long-passphrase' --email alice@example.com --role superadmin
```

See [12 — control-api CLI](./../12-cli.md) for full flag reference and the dedicated `control-api-cli` Docker Compose profile.

## Seeding a complete demo workspace

If you want a ready-to-explore stack (a project, routes, and ~75 fake request logs), use the demo seeder. It attaches to the **global** payment methods, so seed those first (owned by a superadmin):

```bash
docker compose run --rm control-api install payment-methods --superadmin alice
docker compose run --rm control-api install demo
```

That gives you a non-superadmin **`demo` / `demo`** account plus a **Default Project** (slug `default`) with realistic data behind it. Full details and the bootstrap order are in [12 — control-api CLI](./../12-cli.md).

## Roles

Accounts have a `role` of `user` (default) or `superadmin`. **User-scoped data — projects,
routes, project settings, request logs, and stats — is visible and editable only by its owner;
this applies to superadmins too (they do not see other users' projects).** The superadmin role
only adds rights over *global* shared entities:

| Capability | Who can do it |
|---|---|
| Log in to the admin panel | Any account in the `users` table. |
| Create a project | Any logged-in account. The creator becomes the project owner. |
| View / edit / archive a project | The owner of that project only. |
| Create a **personal** payment method / facilitator / asset | Any logged-in account (visible only to them). |
| Mark an entity **global** (visible to all) | Superadmin only. |
| Delete a **global** entity | Superadmin only. |

Promote a user with `UPDATE users SET role='superadmin' WHERE username='...';` in Postgres.

## Losing access to the only account

If you lose the password, create a fresh superadmin directly with `install user`:

```bash
docker compose run --rm control-api install user \
  --username recovery --password '<long-passphrase>' --email recovery@example.com --role superadmin
```

You can also reset an existing account's password by updating its bcrypt `password_hash` in
Postgres, or restore the database from a backup. If SMTP is configured, the account owner can also
use the login page's **Forgot password** flow.

## What's next?

- See what the home page shows: [Dashboard](./02-dashboard.md).
- Start setting up payments: [Facilitators](./03-facilitators.md).
- Full CLI reference: [12 — control-api CLI](./../12-cli.md).
