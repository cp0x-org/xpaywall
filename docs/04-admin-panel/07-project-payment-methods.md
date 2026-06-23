# Admin Panel — Project Payment Methods

This is the page where the global pieces (facilitator, payment method, asset) get attached to a specific project, with a **payout address** — the wallet that actually receives the money.

Until you create at least one Project Payment Method, paid routes on that project cannot work: the gateway has no idea where to send the funds. Free routes still work without it.

![Project payment methods list](./../images/project-payment-methods-list.png)

## Two ways to open the form

- Sidebar **Projects → Payment Methods** shows links across all projects.
- Or open a project and switch to the **Payment Methods** tab inside it.

Click **Add Payment Method** to open the dialog.

![Add project payment method](./../images/project-payment-methods-form.png "medium")

## Fields

| Field | What to put |
|---|---|
| **Payment Method** | Pick one of the global Payment Methods (e.g. `x402 Base Mainnet` or `MPP Tempo`). The fields below adapt to its protocol. |
| **Asset** | Filtered to assets that belong to the chosen payment method. Pick the currency you accept, e.g. `USDC` (x402) or `pathUSD` (MPP/Tempo). |
| **Scheme** | For an x402 method: `exact` (`upto` and `batch-payment` are listed for forward compatibility but not yet supported — see [11 — Roadmap](./../11-roadmap.md)). For an MPP method: `charge`. |
| **Facilitator** *(x402)* | Pick a facilitator that supports the network of the chosen method. This is the verifier service that will check incoming payment proofs. MPP methods have no facilitator — leave it blank. |
| **Method / RPC URL / Secret Key** *(MPP)* | MPP links carry these instead of a facilitator: the `method` (`tempo`), the Tempo JSON-RPC `rpc_url` the charge is settled against, and the HMAC `secret_key` that signs the 402 challenge. |
| **Payout Address** | The wallet address that receives the payments on this network. Required for paid routes. Format depends on the chain — for EVM chains this is a `0x...` address. |
| **Enabled** | Leave on. Off keeps the row but stops it from being offered as a payment option to clients. |

> **Why the form filters assets.** Each asset belongs to exactly one payment method. Once you choose the method, only its assets appear in the next dropdown.

## What the schemes mean

For an **x402** method, the `exact` scheme means the client signs and pays the exact USD price you set on the route — nothing more, nothing less, before the request is forwarded. It is the simplest mode.

For an **MPP** method, the `charge` scheme means a one-time on-chain charge, settled against the Tempo RPC before the request is forwarded.

x402's `upto` (pay up to a maximum, settle the real consumption afterwards) and `batch-payment` (pay several requests in one signature) are planned. Selecting them in the form has no effect until they ship — see the roadmap.

## Why payout address matters

The gateway tells every paying client where to send the money. That address is the **payout address**. If you leave it blank or paste the wrong one:

- Best case: clients see a malformed 402 response and cannot pay.
- Worst case: you accept payments to a wallet you do not control.


## Multiple methods per project

You can attach several Project Payment Methods to the same project. The gateway will offer all of them in the 402 response and the client picks which one to pay with. Use this when you want to accept the same asset on different networks, or different assets.

A project commits to **one protocol**. Once you attach an x402 method, the form locks further additions to x402 (and likewise for MPP) — you can mix multiple networks or assets within a protocol, but not x402 and MPP on the same project. (Offering both rails on a single route is a file-mode feature; see [07-xgateway/03 — File mode](./../07-xgateway/03-file-mode.md).)

## Edit / delete

The form lets you edit scheme, facilitator, payout address and the enabled flag. Payment Method and Asset are locked after creation — to change those, delete the row and add a new one. Disable a row instead of deleting it if you want to keep it as a temporary off-switch.

## What's next?

- Now that money has a place to land, define which paths are paid: [Routes](./08-routes.md).
- See it end-to-end with screenshots: [Guide 01 — Add your first paid route](./../06-guides/01-first-paid-route.md).
