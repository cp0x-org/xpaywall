# Admin Panel — Login & Users

The admin panel is a small web app at `http://<your-host>:3104` (port `3104` by default). Projects, payment methods and routes are all managed here.

> **Accounts** can be self-registered from the login page (with password reset and Google sign-in)
> or created from the control-api CLI. A full user-management screen is on the [roadmap](./../11-roadmap.md).
> See [12 — control-api CLI](./../12-cli.md) for the CLI commands.

## First login

There is no bootstrap-admin env var. Create your first account one of two ways:

- **Register on the login page** — the panel has Sign-up, "Forgot password", and Google sign-in.
- **Seed the demo** — `docker compose run --rm control-api install demo` creates `admin / admin123`.

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
  --password 'choose-a-long-passphrase'
```

The CLI hashes the password with bcrypt before storing it. Run it once per account you want to create. The new account can log in immediately — no restart required.

For local development without Docker:

```bash
cd control-api
go run ./cmd/control-api --env-file .env install user --username alice --password 'choose-a-long-passphrase'
```

See [12 — control-api CLI](./../12-cli.md) for full flag reference and the dedicated `control-api-cli` Docker Compose profile.

## Seeding a complete demo workspace

If you want a ready-to-explore stack (admin user, sample project, payment method, routes, and ~75 fake request logs), use the demo seeder:

```bash
docker compose run --rm control-api install demo
```

That gives you an `admin` / `admin` account plus a **Default Project** with realistic data behind it. Full details in [12 — control-api CLI](./../12-cli.md#install-demo--seed-a-demo-workspace).

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

If you lose the password, create a fresh account directly with `install user`, then promote it
if you need superadmin rights:

```bash
docker compose run --rm control-api install user --username recovery --password '<long-passphrase>'
```

```sql
UPDATE users SET role='superadmin' WHERE username='recovery';
```

You can also reset an existing account's password by updating its bcrypt `password_hash` in
Postgres, or restore the database from a backup.

## What's next?

- See what the home page shows: [Dashboard](./02-dashboard.md).
- Start setting up payments: [Facilitators](./03-facilitators.md).
- Full CLI reference: [12 — control-api CLI](./../12-cli.md).
