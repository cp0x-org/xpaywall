# Admin Panel — Payment Methods

A **payment method** is a template for "this protocol on this network". It says, for example, "x402 on Base Mainnet". It does not yet say which asset, which facilitator or which wallet receives the money — those come from assets and from the per-project link.

You define payment methods once, then reuse them across many projects.

> **Screenshot placeholder:** ![Payment methods list](./../images/payment-methods-list.png)

## Fields

Open **Payments → Payment Methods** and click **Create Payment Method**.

> **Screenshot placeholder:** ![Payment method form — x402](./../images/payment-methods-form-x402.png)

| Field | What to put |
|---|---|
| **Code** | A unique short identifier. Lowercase letters, digits and hyphens. Example: `x402-base-usdc`. You will see it everywhere in the UI alongside the human name, so make it descriptive. |
| **Protocol** | `x402`. (MPP is reserved for the future — see [11 — Roadmap](./../11-roadmap.md).) |
| **Network** | Pick a network from the dropdown, or switch to **Custom** to type a CAIP-2 chain ID directly. |
| **Name** | Human-readable label, e.g. `Base Mainnet`. The select-network mode fills this in for you. |
| **CAIP-2 Chain ID** | Auto-filled when you pick a known network. For custom entries, format is `eip155:<chainId>`. Examples: `eip155:8453` (Base Mainnet), `eip155:84532` (Base Sepolia). |
| **Enabled** | Leave on. Switching off keeps the row but stops the method appearing in selections. |

### Network: select vs custom

The form has a toggle between **Select network** and **Custom**.

- **Select network** — pick from a curated list of CAIP-2 chains known to control-api. This is the safer option; the name and ID stay in sync.
- **Custom** — type any `eip155:<chainId>` you like, plus your own human name. Use this only when the network you need is not in the list.

## What "payment method" means in practice

A payment method on its own does not let anyone pay. You always combine three things:

1. **Payment Method** — protocol + network (this page).
2. **Payment Asset** — what coin (USDC, etc.) on this method's network.
3. **Project Payment Method** — links the method+asset to a specific project, with a payout address and a facilitator.

Think of payment method as the *protocol/network slot* and asset as the *currency slot* — they slot together when you attach them to a project.

## Edit / delete

You can rename, change the protocol/network or disable an existing method. You **cannot delete** a method that has assets or active project links pointing at it — clean those up first.

## What's next?

- Add at least one asset for this method: [Payment Assets](./04-payment-assets.md).
- Then attach the method+asset combo to a project: [Project Payment Methods](./07-project-payment-methods.md).
