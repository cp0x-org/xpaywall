# Guide 05 — Migrating from file mode to HTTP mode

You started with a YAML config because it was the fastest way to get going. Now you want the admin panel: live edits, request logs, multiple operators, a dashboard. This guide walks the move, rule by rule.

The migration is mechanical — every concept in the YAML maps to one or two entities in the admin panel.

## What changes and what doesn't

The **gateway** keeps doing the same job. What changes:

- **Where rules live.** YAML → PostgreSQL behind control-api.
- **How the gateway gets them.** Reads at startup → fetches per request from control-api (cached in memory).
- **Who can edit.** A file on disk → anyone with an admin panel login.

What does *not* change:

- The proxy URLs your clients call.
- The 402 response shape.
- Pricing, networks, assets, facilitators — same values, just stored differently.

If you do the mapping carefully, clients see no difference.

## The mapping at a glance

| YAML field | Admin panel entity |
|---|---|
| `x402[].facilitator_url` | Facilitator |
| `x402[].network`, `scheme` | Payment Method (+ scheme on the project payment method) |
| `x402[].merchant_address` | Payout Address on the project payment method |
| `outbound[].url` + `auth_header` | Project (Server Base URL + Auth Header) |
| `routes[].path` | Route — Path Pattern |
| `routes[].target` | Project Server Base URL (one project per upstream) |
| `routes[].price` | Route — Price (USD) |
| `routes[].free` | Route — Free flag |
| `routes[].mime_type`, `description`, `bazaar` | Route — same fields |
| `allow_unmatched` | Project — Allow Unmatched Routes |

The asset (USDC contract + decimals) is implicit in the YAML — x402 facilitators usually know which asset is configured on their side. In the admin panel you spell it out as a **Payment Asset** entity. You will need the contract address and decimals for each token you accept.

## Step 1 — Write down what your YAML currently does

Before you touch the admin panel, jot down the answers to:

- Which networks do you accept payments on? For each: chain ID (CAIP-2), asset symbol, asset contract, decimals, facilitator URL.
- For each upstream you forward to: its base URL, any auth header it needs, its slug.
- For each route: path pattern, price, free or paid, which payment method(s) it accepts.

If your YAML has only one upstream and one facilitator, this is a five-minute exercise.

> Tip: keep a copy of `config.yaml` open in another window. You will copy values out of it.

## Step 2 — Bring up control-api and the admin panel

If you started with file mode, your `docker-compose.yml` probably only runs xgateway. You need to enable the full stack — see [02 — Setup](./../02-setup.md). The relevant additions:

- PostgreSQL service.
- control-api service.
- Admin panel service.

Set `CONFIG_PROVIDER=http` on xgateway. Keep `CONFIG_FILE` empty (the file provider will not run).

Make sure `INTERNAL_API_KEY` is the same on both xgateway and control-api.

Restart the stack. xgateway will start trying to fetch rules from control-api. Until you create rules, every request gets a 403 — that is expected.

## Step 3 — Recreate the payment entities

In this order (it is the same order as [Guide 01](./01-first-paid-route.md)):

1. **Facilitator.** One row per `facilitator_url` you used in YAML.
2. **Payment Method.** One row per `(protocol, network)` pair. Code can be anything memorable — `x402-base-sepolia`, `x402-mainnet`, etc.
3. **Payment Asset.** One row per `(payment method, token)` pair. Fill in contract address and decimals.

If your YAML accepted only one combination (the common case), you are creating three rows total.

## Step 4 — Create a project per upstream

For each `outbound[]` entry in YAML, create a project:

- **Server Base URL:** the YAML `url`.
- **Slug:** pick a short identifier. **This becomes part of the URL** your clients call (`/<slug>/...`), so once you publish it, you cannot rename it without breaking clients. If you want to keep the same client-facing URLs you had with file mode, pick a slug that matches the prefix you were using — or use an empty slug if your file-mode setup did not have one (the admin panel may not allow this; in that case set a default slug and accept that URLs change).
- **Auth Header Name / Value:** the YAML `auth_header` fields.
- **Allow Unmatched Routes:** mirror the YAML `allow_unmatched` flag.

> **URLs may change.** In file mode the proxy URL was usually `http://gateway/<path>`. In HTTP mode it becomes `http://gateway/<slug>/<path>`. If you cannot afford to change client URLs, see *Keeping URLs identical* at the end of this guide.

## Step 5 — Attach a payment method to the project

For each project, open it and go to the **Payment Methods** tab. Add:

- **Payment Method:** the one from Step 3.
- **Asset:** the asset from Step 3.
- **Scheme:** `exact` (this is what file mode used; if you had something else, mirror it).
- **Facilitator:** the one from Step 3.
- **Payout Address:** copy from YAML `merchant_address`.

If your YAML had multiple `x402[]` entries on the same upstream (multi-network support), add one project payment method per entry.

## Step 6 — Recreate the routes

For each `routes[]` entry, create one route in the admin panel:

- **Project:** the project from Step 4 (look up by upstream URL if you forgot the mapping).
- **Path Pattern:** copy verbatim. Glob syntax is the same.
- **Free:** mirror the YAML.
- **Price (USD):** copy from YAML `price` — but in **dollar format**. If your YAML had `price: "100000"` (on-chain integer for USDC), convert: 100000 / 10^6 = `0.10`. The admin panel always stores USD; the on-chain conversion happens automatically.
- **Description / MIME / Bazaar:** copy from YAML if set.

This is the longest step. If you have many routes, do it one project at a time and test as you go.

## Step 7 — Compare before flipping clients

Both modes can run side by side (different ports, different containers). A good check before you tell clients to switch:

1. Pick a known-priced route. Call the old file-mode proxy with curl. Note the 402 response.
2. Call the new HTTP-mode proxy with curl. Note the 402 response.
3. The bodies should match in: `network`, `asset`, `payTo`, `maxAmountRequired`, `resource` (modulo the URL prefix).

If they differ, the most common cause is the price field: HTTP mode does the USD→on-chain conversion based on the asset's `decimals`. Verify the asset row.

## Step 8 — Cut over

When you are happy:

- Stop the file-mode xgateway container.
- Point DNS / load balancer at the HTTP-mode xgateway.
- Watch the Requests page of the admin panel for the first real traffic.

If something goes wrong, you can flip `CONFIG_PROVIDER` back to `file` and restart — the YAML is still there.

## Keeping URLs identical

The biggest behavioural change is the project slug appearing in the path. There are two ways to avoid it:

1. **Reverse proxy rewrite.** Put a small nginx in front of xgateway that strips the leading client-facing prefix or adds the slug. Clients keep their old URL; xgateway sees the new one.
2. **Run xgateway with a single project and treat its slug as silent.** If you have exactly one project and operate a separate xgateway instance per upstream, the slug is purely cosmetic. Choose a short one and document the URL change once.

In practice, most teams accept a one-time URL change and use it as a chance to version their public surface.

## What's next?

- After cutover, lock down the open admin panel: [09 — Security](./../09-security.md).
- xgateway internals to understand what the HTTP provider is doing under the hood: [07 — xgateway](./../07-xgateway/01-overview.md).
