# 09 — Security

xpaywall sits in the request path of paid APIs and holds the secrets that let your gateway talk to your upstreams. The defaults are safe-ish for local development but **not** safe for the public internet. This page lists the things to harden before you put the stack behind a public DNS name.

The advice is grouped by what you are protecting: secrets, the admin panel, the gateway, and your operating posture.

## Secrets you must change before production

The Docker Compose defaults exist for fast local setup. Every one of these is a security boundary in production.

| Secret | Default | Why it matters |
|---|---|---|
| `INTERNAL_API_KEY` | `please-change-me-...` | Anyone who knows it can call control-api as if they were xgateway — i.e. read every route, change resolution behaviour, ingest fake request logs. |
| `JWT_SECRET` | `please-change-me-...` | Signs admin-panel session tokens. If leaked, an attacker can forge admin sessions and run anything in the admin API. |
| Superadmin account | — | The superadmin role (set in Postgres) can manage global payment methods, assets, and facilitators. Guard the credentials of any user you promote. |
| Database credentials (`CONTROL_DB_DSN`) | `postgres:postgres@...` | Full read-write to your route, project, and payment-channel data — including granting the superadmin role. |
| `SMTP_PASSWORD` | — | If set, lets control-api send welcome / password-reset email. A leaked SMTP credential lets an attacker send mail as you. Keep it in a secret store, rotate on the provider if leaked. |

Practical advice:

- Generate `INTERNAL_API_KEY` and `JWT_SECRET` with `openssl rand -hex 32`. They should look like 64-character hex strings, not memorable words.
- Don't reuse the same value across `INTERNAL_API_KEY` and `JWT_SECRET` — they protect different surfaces.
- Keep secrets out of your `docker-compose.yml`. Move them to a `.env` file that is in `.gitignore`, or use your platform's secret manager (Docker secrets, Kubernetes Secrets, Vault, AWS Secrets Manager, etc.).

## The admin panel

The admin panel is a React app that calls control-api. It has no authentication of its own — *all* protection comes from how you expose it.

### Restrict network exposure

By default Docker Compose binds the admin panel to a host port (e.g. `3104`). That port is reachable from anywhere on the host's network. Production options:

1. **Bind to localhost only** — the panel is only reachable via SSH tunnel or through a reverse proxy. In compose:
   ```yaml
   ports:
     - "127.0.0.1:3104:3104"
   ```
2. **Hide it behind a VPN** — only operators on the VPN can reach it.
3. **Put it behind an authenticating reverse proxy** — Cloudflare Access, an AWS ALB with OIDC, a Tailscale Serve. The proxy enforces identity before the React app even loads.

Avoid putting the admin panel on a raw public IP. Even with strong superadmin credentials, login pages get scanned constantly.

### Guard the superadmin credentials

There is no default admin account — you create the first superadmin yourself with
`install user … --role superadmin` (and if you seeded the demo, a non-superadmin `demo` / `demo`
account also exists — change or remove it before going public). After your first deployment:

1. Create your real superadmin with a strong password (`install user … --role superadmin`).
2. Remove or change the password of any seeded/demo accounts (`demo`).
3. Give regular operators the default `user` role, not `superadmin`.

Treat the superadmin role like root — give it only to people who actually need it. The role is what gates managing **global** payment methods, assets, and facilitators.

### JWT lifetime

JWTs are valid for a fixed period (default a few hours). A leaked JWT remains valid for that period — there is no revocation list. If you suspect a leak, rotate `JWT_SECRET`; all existing tokens become invalid immediately. (Every operator will need to log in again.)

## The gateway

xgateway is the only piece that *must* be publicly reachable — that is its job.

### TLS termination

xgateway speaks plain HTTP. Put it behind a TLS-terminating reverse proxy (nginx, Caddy, an ALB, Cloudflare). Never expose port 8081 directly.

The proxy should:
- Terminate TLS with a valid certificate.
- Forward the original `Host` header.
- Set `X-Forwarded-For` so the gateway sees real client IPs.

### PUBLIC_URL must match reality

`PUBLIC_URL` is what the gateway tells clients to pay for. If a client signs an x402 authorisation for the URL the gateway advertised, the facilitator verifies it against that URL. A mismatch (e.g. you advertise `http://localhost:3102` but clients call `https://api.example.com`) causes every payment to fail validation.

After putting the gateway behind a public hostname, set `PUBLIC_URL` to the public URL.

### Rate limiting

xpaywall does not enforce request rate limits. A client can hammer your gateway with paths it knows return 402, or with valid free routes. Rate limit:

- At the reverse proxy layer (nginx `limit_req`, Cloudflare rules, etc.).
- Or with a sidecar like a small WAF.

Free routes are particularly vulnerable — they cost you upstream resources but produce no revenue.

### Upstream isolation

The gateway forwards client requests with limited filtering. Defensively:

- Run the upstream on an internal-only network — the gateway is its only public entry point.
- Use the project's auth header to ensure the upstream rejects any request not bearing the gateway's secret.
- Treat the gateway as a privileged client of the upstream — its `auth_header_value` is a long-lived token.

If the upstream has its own auth, ensure the token xgateway injects is scoped (read-only if possible, no admin privileges).

## Database

PostgreSQL holds:
- Project routes and prices.
- Payout addresses.
- Hashed admin credentials.
- Request logs (which include client IPs, paths, and response codes).

Do not expose port 5432 publicly. If the database lives outside Docker, keep its security group / firewall locked down to the control-api host only.

Back up the database regularly. The loss of payment-channel configuration costs you revenue; the loss of `request_logs` costs you observability.

## Operating posture

A few practices that aren't features but matter:

- **Patch promptly.** xpaywall depends on Go libraries and Node packages. Subscribe to release notes.
- **Audit logs in the admin panel.** Look at the Activity Log page periodically — unfamiliar edits to payment channels or routes are a yellow flag.
- **Use separate environments.** Don't share an admin panel between staging and production. Slug typos at the API level are recoverable; payout-address typos are not.
- **Drill the rotation.** Practice rotating `INTERNAL_API_KEY` — it requires a coordinated update of both xgateway and control-api. Doing it once in calm conditions saves you when you need to do it under pressure.
- **Treat 402 payment requirements as public.** The 402 body includes payout addresses and asset contracts. That is fine — it is the protocol. Just don't treat the payout address as a secret.

## What to do if a secret leaks

| Leaked secret | What to do |
|---|---|
| `INTERNAL_API_KEY` | Rotate on both control-api and xgateway. Restart both. |
| `JWT_SECRET` | Rotate on control-api. Every operator re-logs in. Audit recent admin actions in the activity log. |
| Superadmin password | Reset it. Then look at recent admin activity for changes you did not make. |
| Payment channel's `auth_header_value` (upstream token) | Rotate the token on the upstream. Update the project's Auth Header Value. |
| Database password | Rotate. Update `CONTROL_DB_DSN` and restart control-api. |
| Payout wallet's private key (out of band) | Move funds out, abandon the address, update all projects to use a new payout address. |

The last one is the worst case. The payout address can be changed in seconds in the admin panel; the rotation cost is operational, not technical.

## What's next?

- After hardening, look up the meaning of any env var: [10 — Reference](./10-reference.md).
- See what's coming that may change your security posture: [11 — Roadmap](./11-roadmap.md).
