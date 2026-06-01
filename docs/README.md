# xpaywall — User Documentation

Welcome to the xpaywall documentation. xpaywall is a self-hosted payment gateway that lets you charge per request for any HTTP API — no API keys, no subscriptions, no billing accounts. Clients pay on-chain (or via Stripe in the future) and the gateway proxies the request only after the payment is verified.

This documentation is organised so you can read it top-to-bottom on your first visit, or jump to a specific topic later.

---

## Table of contents

### Getting started
- [01 — Overview](./01-overview.md) — what xpaywall is, how it works, the two operation modes, common use-cases.
- [02 — Setup](./02-setup.md) — install with Docker Compose and run xpaywall in minutes.
- [03 — Configuration](./03-configuration.md) — environment variables, ports, production checklist.

### Admin panel
- [01 — Login & Users](./04-admin-panel/01-login-and-users.md)
- [02 — Dashboard](./04-admin-panel/02-dashboard.md)
- [03 — Facilitators](./04-admin-panel/03-facilitators.md)
- [04 — Payment Assets](./04-admin-panel/04-payment-assets.md)
- [05 — Payment Methods](./04-admin-panel/05-payment-methods.md)
- [06 — Projects](./04-admin-panel/06-projects.md)
- [07 — Project Payment Methods](./04-admin-panel/07-project-payment-methods.md)
- [08 — Routes](./04-admin-panel/08-routes.md)
- [09 — Requests](./04-admin-panel/09-requests.md)

### Understanding the system
- [05 — Concepts](./05-concepts.md) — payment flow, x402, networks, price formats, route resolution.

### Step-by-step guides
- [01 — Add your first paid route](./06-guides/01-first-paid-route.md)
- [02 — Wildcard routes](./06-guides/02-wildcard-routes.md)

### xgateway (the proxy)
- [01 — Overview](./07-xgateway/01-overview.md) — architecture, middleware chain, caching, providers.
- [02 — HTTP mode](./07-xgateway/02-http-mode.md) — using xgateway with control-api or your own backend.
- [03 — File mode](./07-xgateway/03-file-mode.md) — running xgateway with only a YAML config.

### Operations
- [08 — Troubleshooting](./08-troubleshooting.md)
- [09 — Security](./09-security.md)
- [10 — Reference](./10-reference.md)
- [11 — Roadmap](./11-roadmap.md)
- [12 — control-api CLI](./12-cli.md) — migrations, user management, demo seed

---

## Two ways to run xpaywall

xpaywall can run in two modes. Pick the one that fits your case:

| | **HTTP mode** (full stack) | **File mode** (gateway only) |
|---|---|---|
| Components | `xgateway` + `control-api` + admin panel + PostgreSQL | `xgateway` + a YAML file |
| Configuration | Web UI | YAML file on disk |
| Best for | Production, multiple projects, several routes that change often | Single project, dev/staging, simple monetisation |
| Live changes | Yes — UI updates apply on the next request | No — gateway restart required to reload |
| Request logs | Stored in PostgreSQL, browsable in admin panel | Not stored |

If you are just trying things out, **start with HTTP mode** — the [02 — Setup](./02-setup.md) page boots the whole stack with one `docker compose up` command.

---

## Need help?

- Read [08 — Troubleshooting](./08-troubleshooting.md) for common errors.
- Check [10 — Reference](./10-reference.md) for env variables, API endpoints, and the glossary.
- For future capabilities (MPP, batch payments, OKX APP, Stripe), see [11 — Roadmap](./11-roadmap.md).
