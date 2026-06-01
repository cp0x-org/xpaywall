# Guide 02 — Wildcard routes

A wildcard route is one rule that covers many paths. Useful for REST APIs where many sub-paths should all carry the same price.

This guide assumes you have already followed [Guide 01](./01-first-paid-route.md) and have a working project.

## Two kinds of patterns

- **Exact** — `/forecast`. Matches one path and nothing else.
- **Wildcard** — `/api/v1/*`. Matches any path that has exactly one extra segment after `/api/v1/`.

xpaywall uses Go's `path.Match` for globs. The most important rule:

> `*` matches **one segment**. It does not match `/`.

That means `/api/v1/*`:

| Path | Matches? |
|---|---|
| `/api/v1/users` | yes |
| `/api/v1/users/42` | no — two segments after `/api/v1/` |
| `/api/v1` | no — no trailing segment |

To cover deeper paths you have to add more rules, or use a longer pattern like `/api/v1/users/*`.

## A wildcard route, step by step

Open **Routes** and click **Create Route**.

> **Screenshot placeholder:** ![Route form — wildcard](./../images/routes-form-wildcard.png)

- **Project:** pick yours.
- **Route Name:** `API v1 wildcard`
- **Path Pattern:** `/api/v1/*`
- **Description:** `All v1 endpoints`
- **Free:** off
- **Price (USD):** `0.01`

Save.

Now every request to `/<slug>/api/v1/<anything>` (one segment) costs one cent of USDC.

## Combining exact and wildcard

You can keep an exact rule for one path inside the wildcard's territory. Exact paths beat wildcards.

Two rules side by side:

| Route name | Path | Price |
|---|---|---|
| Free health | `/api/v1/health` | free |
| API v1 paid | `/api/v1/*` | `0.01` |

Result:

| Path called | Rule applied |
|---|---|
| `/api/v1/health` | Free health (exact wins) |
| `/api/v1/users` | API v1 paid (wildcard) |
| `/api/v1/whatever` | API v1 paid |

## Different prices for different sub-paths

When you want one wildcard at `0.01` for most paths but `0.10` for one expensive endpoint, add an exact rule for the expensive endpoint at the higher price. xpaywall picks the more specific match.

| Path | Price |
|---|---|
| `/api/v1/embeddings` | `0.10` (exact) |
| `/api/v1/*` | `0.01` (wildcard) |

## When to stick to exact

Wildcards save typing but they hide intent. If you find yourself with only a handful of paths, write them out one by one. The admin panel sorts and filters them; the cost of an extra row is small and the resulting setup is easier to audit.

Use wildcards when:
- A new sub-path is added often and you do not want to touch the admin panel every time.
- All paths in a section truly share the same price.
- The list of priced paths would be very long.

Otherwise, prefer exact paths.

## A note on the project's `Allow Unmatched` flag

A wildcard route covers a *known* set of paths. Unmatched paths (those outside any wildcard or exact rule) still depend on the project's **Allow Unmatched Routes** flag:

- **off** (default) — unmatched paths return 403. The wildcard does not paper over typos or undocumented endpoints.
- **on** — unmatched paths are proxied for free. This effectively makes the wildcard's price an *opt-in* tax on the matched paths, not a default for the project.

For paid APIs, leave **Allow Unmatched** off and rely on explicit rules.

## What's next?

- Probe a wildcard with curl to confirm it matches what you think it does: [Guide 03 — Testing with curl](./03-testing-with-curl.md).
- Background on how matching decisions are made: [Concepts → Route resolution](./../05-concepts.md#route-resolution).
