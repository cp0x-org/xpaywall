# 11 — Roadmap

What's not in xpaywall today but is planned. Use this page to decide whether to build a workaround now or wait, and to set expectations about which integrations are coming.

This page is updated alongside major releases. The order within each section is "next up" → "later".

## Payment schemes

### x402 `upto`

Currently xpaywall supports x402 with the `exact` scheme — a fixed price known in advance. The `upto` scheme is on the roadmap:

- **Idea:** the client signs an authorisation for *at most* X. The gateway returns the upstream response, then settles for the actual amount based on the upstream's `X-PAYMENT-AMOUNT` reply header (or a similar mechanism).
- **Why it matters:** usage-based billing where the cost depends on input size, tokens consumed, compute time, etc. You don't have to compute the exact price up front.
- **Status:** `upto` is recognised in the codebase but the verification + settlement flow is incomplete. Not safe for production yet.

### x402 batch payments

- **Idea:** one signed authorisation covers many requests up to a cumulative cap. Cheaper for high-frequency callers because you do not pay verification overhead on every request.
- **Why it matters:** turns x402 from "a payment per request" into "a payment per session", which is closer to how most real APIs are billed.
- **Status:** specification stage. No code yet.

### MPP (Multi-Party Payments)

Tempo's MPP protocol allows splitting one payment across multiple recipients atomically — useful for affiliate/marketplace cases. Code paths for MPP are present but commented out in this release.

- **Idea:** define multiple payout addresses with split percentages on the project payment method; signed authorisations cover the bundle in one shot.
- **Why it matters:** if your business model involves passing revenue through to upstream providers, MPP removes the need for off-chain reconciliation.
- **Status:** partial implementation. Disabled in the current release. Will return as a first-class feature.

### OKX Agentic Payments Protocol

OKX has published an agentic payment standard intended for autonomous AI agents that need to pay for resources.

- **Idea:** integrate as a third protocol alongside x402 and MPP. Routes can accept any combination.
- **Why it matters:** broadens the set of clients that can transact with your API.
- **Status:** scoping. Not yet started.

## Integrations

### Stripe

For fiat-priced endpoints where the client is a human, not an agent:

- **Idea:** Stripe Checkout for paid routes — the 402 response includes a Stripe Checkout URL; the upstream is served after webhook confirmation.
- **Why it matters:** bridges xpaywall into traditional consumer billing without forcing wallets on the client.
- **Status:** design phase.

### Bazaar discovery improvements

The Bazaar field on routes is already accepted but underused. Planned improvements:

- Auto-generation of richer Bazaar metadata from OpenAPI specs.
- A public catalog endpoint on control-api that facilitators can poll.
- Admin-panel UI for editing Bazaar metadata (currently raw JSON).

## Gateway

### Configurable rate limits

Today xgateway has no per-route or per-client rate limiting. The expectation is that a reverse proxy in front of xgateway does this. Future:

- Per-route concurrency limits (don't let a single expensive route saturate the gateway).
- Per-client soft limits backed by a counter store.

### WebSocket and SSE proxying

The current proxy is HTTP request/response only. Long-lived connections (WebSockets, server-sent events) are not paywalled or proxied. Future:

- Free WS/SSE proxying when the project allows it.
- Paid WS connections settled per duration via the eventual `upto` scheme.

### Multi-region resilience

Today xgateway is a single process per deployment. There is no clustering, no shared cache, no leader election. Future:

- Optional Redis-backed shared cache to coordinate route resolution and pending-log correlation across replicas.
- Active/active deployments without per-replica state drift.

## Admin panel

### Audit log UX

Activity logs exist but the UI is sparse. Future:

- Diff view between previous and new state of any entity.
- Filtering by user / entity type / time range.
- Export to CSV.

### Multi-tenant operator scoping

Today every operator with a non-superadmin role can see every project. Future:

- Project membership: an operator only sees projects they are explicitly assigned to.
- Per-project roles (read-only, editor, owner).

### Bulk operations

- Import routes from OpenAPI specs.
- Bulk price updates (e.g. apply +10% to all routes on a project).
- CSV import/export for routes.

## Tooling

### Production Docker Compose

The repo currently ships `docker-compose.yml` tuned for development. A separate `docker-compose.prod.yml` with security defaults (bind localhost, real secrets via env file, sensible logging) is planned.

### Helm chart

For Kubernetes operators. xpaywall as a Helm chart with sensible defaults, ingress, and a values schema. Planned but not started.

### CLI

A `xpaywallctl` CLI for scripting common admin operations: create-route, list-projects, rotate-keys, etc. Planned as a thin wrapper over the existing admin API.

## How to influence the roadmap

- File an issue with your use case — the project prioritises features tied to concrete needs over speculative ones.
- If you build a workaround or a custom integration, share the design. It is a useful signal that the feature should be official.
- Backwards-compatibility is taken seriously. Changes to the gateway 402 response shape or the internal API contract are flagged as breaking and will get migration notes.

## Out of scope

A few things xpaywall deliberately does **not** plan to add:

- **An on-chain settlement service.** That is the facilitator's job. xpaywall stays a gateway and dispatcher.
- **A built-in wallet for clients.** Clients use their own wallets / SDKs.
- **A general-purpose API management product.** xpaywall focuses on the payment layer. Schema versioning, request transformation, traffic shaping etc. belong upstream or in a separate API gateway.

## What's next?

- Already covered: [README](./README.md) — full table of contents.
- Ready to deploy? Revisit [09 — Security](./09-security.md) before going public.
