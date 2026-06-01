# Guide 03 — Testing with curl

curl is the fastest way to confirm that a route works the way you think. You can see the 402 response, inspect what the gateway tells clients to pay, and check that free routes are actually free.

This guide is curl-only. To complete a paid flow with a real on-chain payment, use an x402 client SDK — see [Guide 01](./01-first-paid-route.md).

## Setup

Assume:
- Gateway is at `http://localhost:3102`.
- Your project slug is `demo`.
- You have at least one route configured at `/weather`.

If your URLs differ, substitute as needed.

## See the 402 response for a paid route

```bash
curl -i http://localhost:3102/demo/weather
```

You should get something like:

```
HTTP/1.1 402 Payment Required
Content-Type: application/json
...

{
  "x402Version": 1,
  "accepts": [
    {
      "scheme": "exact",
      "network": "eip155:84532",
      "asset": "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
      "payTo": "0xYourPayoutAddress",
      "maxAmountRequired": "1000",
      "resource": "http://localhost:3102/demo/weather",
      "description": "Returns the current weather (sample upstream)",
      "mimeType": "",
      "maxTimeoutSeconds": 30
    }
  ]
}
```

The interesting fields:

- `network` — the CAIP-2 chain ID.
- `asset` — the token contract on that chain.
- `payTo` — your payout address. Confirm this is correct.
- `maxAmountRequired` — the on-chain amount in the asset's base unit. For USDC at 6 decimals, `1000` = `0.001` USDC.
- `resource` — the full URL the client is paying for.

> **Screenshot placeholder:** ![curl 402 response](./../images/curl-402-response.png)

## See a free route succeed

If you have a route marked **Free** at `/health`:

```bash
curl -i http://localhost:3102/demo/health
```

You should get a `200 OK` (or whatever your upstream returns) with no 402 stage at all.

## Hit a path that has no rule

If the project has **Allow Unmatched** off:

```bash
curl -i http://localhost:3102/demo/no-such-path
```

You should see a `404 Not Found` (route not found) or `403 Forbidden`. The exact code depends on how the gateway is configured; both mean "no rule matched".

If you instead get a `200` from your upstream, the project has **Allow Unmatched** turned on. Consider whether that is what you want — see [Projects](./../04-admin-panel/06-projects.md#allow-unmatched--what-it-does).

## Inspect what the gateway sees

Pass `-H 'X-Forwarded-For: 1.2.3.4'` to spoof a client IP, or `-H 'User-Agent: my-test'` to set the user agent. Both show up in the request log and are useful when you want to find a specific test in a noisy environment.

## Repeated 402 means the client did not pay

If you call the same path twice without an `X-PAYMENT` header, you get two 402 responses. xpaywall correlates them in the request log so you can see the retry pattern, but each call is its own row at the database level.

To complete the flow, you need to:
1. Sign an x402 authorisation with a wallet that has at least the requested asset on the requested network.
2. Send the proof in the `X-PAYMENT` header.

curl can do step 2 (it is just a header), but you need an x402 SDK or wallet plugin to produce the signed proof. See [Guide 01](./01-first-paid-route.md) for SDK pointers.

## Common curl flags

| Flag | What it does |
|---|---|
| `-i` | Show response headers — needed to see the 402 code. |
| `-v` | Verbose; show the request the client sends. |
| `-s` | Quiet, no progress bar. |
| `-o /dev/null -w "%{http_code}\n"` | Print only the status code. Handy for quick smoke tests. |
| `-H "Header: value"` | Set a custom request header. |

## Smoke-test a deployment

A one-liner you can run after every config change:

```bash
curl -s -o /dev/null -w "402 expected: %{http_code}\n" http://localhost:3102/demo/weather
```

You can wrap a handful of these into a shell script and run it on a schedule to catch broken routes.

## What's next?

- Replace the example upstream with your real API: [Guide 04 — Connecting a real upstream](./04-connecting-real-upstream.md).
- Read 402 errors and request events in detail: [Requests](./../04-admin-panel/09-requests.md).
