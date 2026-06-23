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

### MPP extensions

MPP (Machine Payments Protocol) ships today with the Tempo `charge` scheme — see [05 — Concepts](./05-concepts.md#mpp-machine-payments-protocol). Two extensions are recognised in configuration but not yet available:

- **Tempo `session` scheme.** Instead of a one-time charge, open a session that authorises a series of charges up to a cap, settled as they happen. Recognised in config today but rejected during validation.
- **`stripe` method.** Settle MPP charges through Stripe rather than an on-chain Tempo RPC, for operators who want a fiat rail. Recognised in config today but rejected during validation.

### OKX Agentic Payments Protocol

OKX has published an agentic payment standard intended for autonomous AI agents that need to pay for resources.

- **Idea:** integrate as a third protocol alongside x402 and MPP. Routes can accept any combination.
- **Why it matters:** broadens the set of clients that can transact with your API.
- **Status:** scoping. Not yet started.

## Gateway

### Configurable rate limits

Today xgateway has no per-route or per-client rate limiting. The expectation is that a reverse proxy in front of xgateway does this. Future:

- Per-route concurrency limits (don't let a single expensive route saturate the gateway).
- Per-client soft limits backed by a counter store.

## What's next?

- Already covered: [README](./README.md) — full table of contents.
- Ready to deploy? Revisit [09 — Security](./09-security.md) before going public.
