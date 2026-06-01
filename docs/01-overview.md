# 01 — Overview

## What is xpaywall?

**xpaywall** is a self-hosted HTTP payment gateway. You put it in front of any web API you own, and it makes that API pay-per-request.

A client calls your API → xpaywall checks whether the call has been paid → if not, it returns an HTTP `402 Payment Required` response that tells the client how to pay → the client pays on-chain → the client retries with proof of payment → xpaywall verifies the proof and finally forwards the request to your real API ("upstream").

You decide the price per route. You decide which networks and assets you accept. You decide which routes are free and which are paid.

### What problem does it solve?

Traditional API monetisation needs all of:
- API keys
- A billing system
- A customer signup flow
- Monthly invoicing

xpaywall replaces all of that with a single primitive: **pay for this exact call, right now**. No account, no key, no subscription. The protocol it speaks (x402) was designed for AI agents and machines, but it works for human-built clients too.

Common use-cases:
- Monetising AI inference (image generation, transcription, embeddings).
- Charging per call for weather, news, scraping, enrichment APIs.
- Pay-per-download for files or datasets.
- A "no signup needed" tier for an existing paid API.

---

## How it works

```
┌────────┐      1. GET /weather                              ┌──────────┐
│ Client │ ───────────────────────────────────────────────►  │ xgateway │
└────────┘                                                   └────┬─────┘
                                                                  │
                                                                  │ 2. lookup rule for /weather
                                                                  │    (file mode: read YAML;
                                                                  │     HTTP mode: ask control-api)
                                                                  │
┌────────┐      3. HTTP 402 + payment instructions           ┌────▼─────┐
│ Client │ ◄───────────────────────────────────────────────  │ xgateway │
└───┬────┘                                                   └──────────┘
    │ 4. signs payment on-chain
    │
    │ 5. GET /weather  (with X-PAYMENT header)               ┌──────────┐
    └────────────────────────────────────────────────────►   │ xgateway │
                                                             └────┬─────┘
                                                                  │
                                                                  │ 6. verify payment with facilitator
                                                                  │
                                                                  │ 7. forward to your upstream API
                                                                  ▼
                                                             ┌──────────┐
                                                             │ Upstream │
                                                             │   API    │
                                                             └────┬─────┘
                                                                  │
┌────────┐      8. real response                             ┌────▼─────┐
│ Client │ ◄───────────────────────────────────────────────  │ xgateway │
└────────┘                                                   └──────────┘
```

Steps:
1. The client calls a path on the gateway.
2. The gateway looks up which route this path matches and what it costs.
3. If a paid route has no valid payment proof, the gateway answers `402` with everything the client needs to pay: which network, which asset, which address, how much.
4. The client signs a payment transaction with its wallet.
5. The client retries the original call with the proof attached.
6. The gateway asks a *facilitator* (a verifier service) to confirm the proof is valid.
7. The gateway forwards the call to your real upstream API.
8. The upstream's response is streamed back to the client.

If a route is marked **free**, steps 3–6 are skipped and the gateway proxies the call directly.

> **Screenshot placeholder:** ![Architecture overview](./images/arch-overview.png)

---

## Components

A full xpaywall installation has four parts. In file mode, only the first one runs.

| Component | What it does |
|---|---|
| **xgateway** | The proxy itself. Sits in front of your API and enforces payment. |
| **control-api** | REST backend that stores projects, routes, prices and request logs. Used by xgateway in HTTP mode and by the admin panel. |
| **admin panel** | Web UI for managing routes, payments and users. Talks only to control-api. |
| **PostgreSQL** | Database for control-api. |

Plus, optionally:

| Component | What it does |
|---|---|
| **example-server** | A toy upstream API that ships with the repo so you have something to point xpaywall at while you experiment. Port 4021. |

---

## Two operation modes

xgateway can run in two modes. They produce the same behaviour for the client — the only difference is where the rules come from.

### HTTP mode (recommended)

xgateway asks control-api for a rule on every incoming request, then caches the result. You manage everything through the admin panel. This is what `docker compose up` gives you out of the box.

Use HTTP mode when:
- You have more than one project on the same gateway.
- You want to add, change or remove routes from a UI without restarting the gateway.
- You want a log of every request stored and searchable.
- You want to give other people (with their own logins) the ability to manage their own projects.

> **Screenshot placeholder:** ![HTTP mode architecture](./images/arch-http-mode.png)

### File mode

xgateway loads a `config.yaml` file at startup and uses it for the rest of its lifetime. control-api, the admin panel and PostgreSQL are not needed at all.

Use file mode when:
- You have a single project and a small number of routes.
- You are happy to change configuration by editing a file and restarting the gateway.
- You do not need a UI and do not care about persistent request logs.
- You are deploying to a constrained environment (one container, no database).

> **Screenshot placeholder:** ![File mode architecture](./images/arch-file-mode.png)

### Quick comparison

| Feature | HTTP mode | File mode |
|---|---|---|
| Components needed | Gateway + control-api + admin panel + Postgres | Gateway only |
| Configuration source | Web UI → database → API call | YAML file on disk |
| Hot reload | Yes (on next request) | No (restart required) |
| Multiple projects | Yes | One project per gateway |
| Per-request log storage | Yes | No |
| Multi-user management | Yes (JWT auth) | No |
| Setup complexity | Medium | Low |

You can also **start in file mode and migrate to HTTP mode** later — see [Guide 05](./06-guides/05-migrating-file-to-http.md).

---

## What's next?

- If you want to get xpaywall running on your machine: [02 — Setup](./02-setup.md).
- If you want to understand the moving parts before you install anything: [05 — Concepts](./05-concepts.md).
- If you only need the gateway and a YAML file: jump straight to [07 — xgateway / File mode](./07-xgateway/03-file-mode.md).
