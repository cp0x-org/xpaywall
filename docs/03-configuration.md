# 03 — Configuration

This page lists every environment variable used by xpaywall, how to change the host ports, and what you must change before exposing xpaywall to the internet.

If you have not booted xpaywall yet, start with [02 — Setup](./02-setup.md).

## Where configuration lives

All service configuration is set in `docker-compose.yml`. After editing it, apply changes with:

```bash
docker compose down && docker compose up -d
```

Docker Compose restarts only the services whose configuration changed.

---

## Environment variables

### xgateway

| Variable | Required | Default | What it does |
|---|---|---|---|
| `CONFIG_PROVIDER` | yes | `file` | `http` to fetch rules from control-api, `file` to load them from a YAML file. |
| `CONTROL_API_URL` | http mode | — | URL of control-api as seen by the gateway, e.g. `http://xpaywall-control-api:9091`. |
| `INTERNAL_API_KEY` | http mode | — | Shared secret with control-api. **Must match** the value in control-api. |
| `CONFIG_FILE` | file mode | `config.yaml` | Path to the YAML rules file inside the container. |
| `PORT` | no | `8080` | Listen port inside the container. The host port is set in `docker-compose.yml`. |
| `LOG_LEVEL` | no | `INFO` | `DEBUG`, `INFO`, `WARN`, `ERROR`. |
| `GIN_MODE` | no | `release` | `debug` adds verbose request logs; `release` is quiet. |
| `DEBUG` | no | `false` | Enables extra request/response tracing. Disable in production. |
| `PUBLIC_URL` | no | — | Override the public-facing URL returned inside 402 responses. Needed when the gateway sits behind a reverse proxy. |

### control-api

| Variable | Required | Default | What it does |
|---|---|---|---|
| `CONTROL_DB_DSN` | yes | — | PostgreSQL connection string, e.g. `postgres://user:password@xpaywall-postgres:5432/xpaywalldb?sslmode=disable`. |
| `INTERNAL_API_KEY` | yes | — | Shared secret with xgateway. **Must match** the value in xgateway. |
| `JWT_SECRET` | yes | — | Signs admin-panel session tokens. Treat it like a password. |
| `PROXY_URL` | yes | — | Public URL of xgateway. Shown in the admin panel and used to build example requests. |
| `PORT` | no | `9090` | Listen port inside the container. |
| `MODE` | no | `release` | Gin mode: `debug` or `release`. |
| `SUPERADMIN_USERNAME` | no | — | Username for the bootstrap superadmin account. |
| `SUPERADMIN_PASSWORD` | no | — | Password for the bootstrap superadmin account. |
| `DEBUG` | no | `false` | Verbose logs (request bodies included). Disable in production. |

### admin panel

The admin panel is a static React app served by nginx. It reads two values at container start to build a runtime `config.js`:

| Variable | Required | Default | What it does |
|---|---|---|---|
| `API_URL` | yes | — | URL of control-api **as seen from the user's browser**. Not from the server. |
| `PROXY_URL` | yes | — | URL of xgateway **as seen from the user's browser**. Used for the "test this route" hints. |

Both can be changed without rebuilding the image — restart the container and the new values take effect.

### PostgreSQL

The bundled `postgres:16` container uses these settings:

| Variable | Default | What it does |
|---|---|---|
| `POSTGRES_DB` | `xpaywalldb` | Database name. Must match the DB name in `CONTROL_DB_DSN`. |
| `POSTGRES_USER` | `user` | Username. Must match. |
| `POSTGRES_PASSWORD` | `password` | Password. **Must be changed** before production. |

---

## Changing ports

If `3100`, `3101`, `3102`, `3103`, `3104` or `5482` are already in use on your machine, change them in `docker-compose.yml`. Each service has a `ports:` block like this:

```yaml
adminpanel:
  ports:
    - "3104:80"      # ← left side is the host port; change it
```

After editing, run `docker compose up -d` to apply the new mapping.

You do **not** need to change the container-side port (the right side) — services talk to each other over the internal `xpaywall-net` network using their container names, not host ports.

---

## Production checklist

Before exposing xpaywall to a real network, change every default. The list is short but every item is important.

### 1. Replace shared secrets

In `docker-compose.yml`, replace both occurrences (control-api and xgateway):

```yaml
INTERNAL_API_KEY: <generate a long random string>
```

Use the same value for both services. Generate it with, for example:

```bash
openssl rand -hex 32
```

Then change `JWT_SECRET` in control-api the same way.

### 2. Change the superadmin password

Either change `SUPERADMIN_USERNAME` and `SUPERADMIN_PASSWORD` in `docker-compose.yml` before the first boot, or log into the running admin panel and update the password from the user form.

### 3. Set `PROXY_URL` correctly

In control-api, set `PROXY_URL` to the real public URL of the gateway, including scheme and port if non-standard. Example:

```yaml
PROXY_URL: https://api.example.com
```

This value is shown to clients inside the 402 response, so it must be reachable from outside.

### 4. Set `API_URL` and `PROXY_URL` for the admin panel

These are URLs the **browser** uses, so they must be reachable from wherever your team accesses the admin panel — not just from inside the Docker network.

### 5. Change the PostgreSQL password

Replace `POSTGRES_PASSWORD` and the matching password in `CONTROL_DB_DSN`. Mount the data volume outside the container so an accidental `docker compose down -v` does not wipe your database.

### 6. Turn off debug

For both control-api and xgateway, set:

```yaml
MODE: release        # control-api
GIN_MODE: release    # xgateway
DEBUG: false         # both
LOG_LEVEL: INFO      # xgateway
```

### 7. Lock down what is exposed

Do not publish the PostgreSQL port (`5482`) on the public internet. Restrict control-api's port (`3101`) to your office network or VPN if you do not need it to be public. Only `xgateway` (the proxy) needs to be reachable from clients.

For more on what to expose and what to keep internal, see [09 — Security](./09-security.md).

---

## Next steps

- Understand the architecture before customising further: [01 — Overview](./01-overview.md).
- Start configuring projects, payments and routes: [04 — Admin panel](./04-admin-panel/01-login-and-users.md).

