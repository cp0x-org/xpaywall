# 05 — Concepts

This page explains the ideas behind xpaywall in plain language: how payment actually flows, what x402 and MPP are, how networks and assets are identified, and how the gateway decides which rule to apply to an incoming request. Read this once and the admin panel will make much more sense.

## The payment flow

HTTP has a status code called `402 Payment Required`. It was reserved in the original HTTP spec but never widely used — until x402 turned it into a real protocol.

The flow looks like this:

1. **Unauthorised request.** A client calls the gateway with no payment proof.
2. **402 response.** The gateway looks up the rule for that path and returns `402 Payment Required` together with everything the client needs to pay: the price, the asset, the network, the wallet address that should receive the money, and an expiration.
3. **Client pays.** Off-band, the client signs a transfer with its wallet. With x402, the signature is over an `EIP-712`-style typed message — the client does not have to broadcast the transaction itself; the facilitator does that.
4. **Retry with proof.** The client repeats the original request, this time with an `X-PAYMENT` header containing the signed proof.
5. **Gateway verifies.** The gateway forwards the proof to the configured facilitator. The facilitator confirms it on-chain (or rejects it).
6. **Forward to upstream.** If the proof is valid, the gateway forwards the original request to your upstream API and streams the response back to the client.

From the client's point of view, the magic is: one request → 402 with instructions → pay → retry → success. From the gateway's point of view: every paid request is a "verify then forward" pipeline.

The steps above describe the **x402** rail. **MPP** follows the same shape — 402 → pay → retry → forward — but the proof travels in an `Authorization` header and is settled against a blockchain RPC endpoint instead of a facilitator. See [MPP](#mpp-machine-payments-protocol) below.

## x402

x402 is the protocol that defines how the 402 response is structured and how the proof header is signed. Think of it as "HTTP + a payment instruction format".

xpaywall today supports x402 with the **`exact`** scheme.

- **exact** — the price is fixed and known in advance. The client signs an authorisation for exactly that amount. The facilitator settles it. The request is then served.

Two more schemes are on the roadmap:

- **upto** — the client signs an authorisation for *up to* a maximum amount. The real amount is determined by the upstream response (e.g. usage-based billing) and settled afterwards.
- **batch-payment** — one signature covers many requests up to a cap. Cheaper for high-frequency callers.

See [11 — Roadmap](./11-roadmap.md) for the schedule.

## MPP (Machine Payments Protocol)

MPP is a second payment rail xpaywall supports alongside x402. Where x402 delegates verification to an external facilitator, MPP settles an on-chain **charge** directly against a blockchain RPC endpoint — there is no facilitator in the loop.

xpaywall today supports MPP with the **Tempo** method and the **`charge`** scheme:

- **Tempo `charge`** — a one-time on-chain charge. The client signs an authorisation and sends it in an `Authorization` header (not the `X-PAYMENT` header x402 uses). The gateway verifies and settles the charge against the configured Tempo RPC endpoint, then forwards the request.

In **file mode**, a single route can accept **both** rails at once. When a request carries an MPP `Authorization` header the gateway uses MPP; otherwise it falls back to x402. A route that lists only MPP channels always uses MPP. (In HTTP mode the admin panel commits each project to a single protocol — see [Project Payment Methods](./04-admin-panel/07-project-payment-methods.md).)

The Tempo `session` scheme and the `stripe` method are recognised in configuration but rejected during validation — they are not available yet. See [11 — Roadmap](./11-roadmap.md).

## Networks and assets

xpaywall does not assume any particular blockchain. Networks and assets are identified by **CAIP** standards, which give every chain and every token a globally unique short string.

- **Network — CAIP-2.** Format: `<namespace>:<reference>`. For EVM chains, the namespace is `eip155` and the reference is the chain ID.
  - `eip155:8453` = Base Mainnet
  - `eip155:84532` = Base Sepolia (test network)
- **Asset — CAIP-19.** Format: `<chain>/<assetType>:<address>`. In xpaywall you only store the *address* part of the asset — the chain is implicit from the parent payment method.
  - `0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913` = USDC contract on Base Mainnet
  - `0x036CbD53842c5426634e7929541eC2318f3dCF7e` = USDC contract on Base Sepolia

For most setups you will use **USDC on Base** and that is the only pair you need to remember.

## Price formats

xpaywall stores prices in **USD as a decimal string** — for example `0.10`. The gateway converts USD to the on-chain amount using the asset's `decimals` field:

```
on_chain_amount = round( usd_price * 10 ** decimals )
```

For USDC with `decimals = 6`:

| USD price | On-chain amount |
|---|---|
| `0.001` | `1000` (= 0.001 USDC) |
| `0.10` | `100000` (= 0.10 USDC) |
| `1.00` | `1000000` (= 1.00 USDC) |

You only type the USD value in the route form. The on-chain conversion happens automatically every time a 402 is returned.

In **file mode**, the YAML file accepts two formats for the same field:

- `price: "$0.10"` — human-readable, parsed back to a number.
- `price: "100000"` — raw on-chain integer string. Useful when you want to be explicit, but you have to compute it yourself.

The two are equivalent. The dollar-sign form is easier to read; the integer form is unambiguous.

## Free vs paid

A route is **paid** by default. Add the **Free** flag (file mode: `free: true`) to skip payment entirely.

Free routes:
- Never return 402.
- Never call a facilitator.
- Do not need a payout address.
- Still go through the gateway, so they still get logged and proxied to upstream.

Use free for health checks, schema endpoints, public metadata. Use paid for everything else.

A route cannot be partially free. If you want "first 100 requests free, then paid", do it in your client or upstream — xpaywall has no quota notion.

## Route resolution

When a request arrives, the gateway has to pick which route's rule applies. The lookup happens once per path and the result is cached in memory, so the cost is paid only on the first hit.

### How the path is matched

In **HTTP mode**, the gateway sends the path `/<username>/<projectSlug>/<rest>` to control-api, which finds the route in PostgreSQL. The `username` scopes the lookup to that owner's project — a slug is unique **per user**, so two users may each own a `default` project. PostgreSQL evaluates exact paths first, then patterns.

In **file mode**, the gateway iterates through the rules in the YAML, in order, and returns the first one whose pattern matches. Glob matching is done with Go's `path.Match` — so `*` matches a single segment.

### Specificity

When two patterns could match the same path, the more specific one wins.

| Rule A | Rule B | Path | Winner |
|---|---|---|---|
| `/health` | `/*` | `/health` | A (exact path beats wildcard) |
| `/api/v1/users` | `/api/v1/*` | `/api/v1/users` | A |
| `/api/v1/*` | `/api/*` | `/api/v1/foo` | A (longer pattern) |

If no rule matches and the project has **Allow Unmatched** off, the gateway returns `403 Forbidden`. If it is on, the request is proxied without any payment check.

### Caching

After the first request to a path, the resolved rule is cached. Changes you make in the admin panel apply to **new** paths immediately, and to **existing** paths after the cache entry expires or the gateway restarts. In practice this is short — the cache exists for performance, not for permanence — but it is the reason a test request right after a config change might briefly use the old rule.

In file mode there is no cache invalidation: the gateway holds the rules in memory for its entire lifetime. To pick up a YAML change you must restart the gateway.

## Putting it together

A complete configuration is built from five layers:

```
Facilitator        Payment Method            Payment Asset
(verifier)         (protocol + network)      (currency)
    \                   |                        /
     \                  |                       /
      \                 |                      /
       Project Payment Method (payout address)
              |
              ├── attached to ───── Project (slug, upstream URL, owner)
                                          |
                                          ├── Route ── Route ── Route
                                          └── Route ── Route ── Route
```


The next page — [Guide 01](./06-guides/01-first-paid-route.md) — walks through building these layers in order.
