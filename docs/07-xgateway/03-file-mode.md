# 07-xgateway — File mode

In file mode, xgateway reads a single YAML file at startup. There is no database, no control-api, no admin panel. All routes, prices, and payment configuration live in one file you maintain in version control.

File mode is great for:
- Quick local experiments.
- Single-upstream setups that rarely change.
- Pinning a known-good configuration in CI.

It is less convenient when:
- You want to change rules without restarting.
- You want request logs in a UI.
- Multiple operators need to manage routes without touching files.

## Enabling file mode

```bash
CONFIG_PROVIDER=file
CONFIG_FILE=/etc/xgateway/config.yaml
```

`CONFIG_FILE` may be a path to a YAML (`.yaml`/`.yml`) or JSON file. YAML is the recommended format.

On startup xgateway reads the file, validates it, and holds the parsed structure in memory for the life of the process. **There is no auto-reload.** To pick up edits, restart the container.

## Full file structure

```yaml
x402:
  - name: base-sepolia
    facilitator_url: https://x402.org/facilitator
    network: eip155:84532
    scheme: exact
    merchant: "0xYourPayoutAddress"
    asset: "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
    decimals: 6
    timeout_seconds: 30

mpp:
  - name: tempo-charge
    method: tempo
    scheme: charge
    rpc_url: https://rpc.moderato.tempo.xyz
    merchant: "0xYourPayoutAddress"
    asset: "0x20c0000000000000000000000000000000000000"
    decimals: 6
    secret_key: "your-hmac-secret"
    timeout_seconds: 30

outbound:
  target: http://my-upstream:4021
  auth_header:
    enable: true
    name: Authorization
    value: "Bearer sk_live_abc123"
  allow_unmatched: false
  rules:
    - name: health
      path: /health
      free: true

    - name: weather
      path: /weather
      price: "$0.001"
      description: "Returns the current weather"
      payment_methods: [base-sepolia]

    - name: paid-rpc          # MPP-only route (Tempo charge)
      path: /rpc
      price: "10000"          # base units — MPP needs the integer form, not "$0.01"
      payment_methods: [tempo-charge]

    - name: api-v1
      path_glob: /api/v1/*
      price: "0.01"
```

Three top-level blocks: `x402`, `mpp`, and `outbound`. Let's walk each.

## The `x402` block

Each entry defines one accepted payment channel.

| Field | Required | Notes |
|---|---|---|
| `name` | yes | A unique identifier you use inside `rules[].payment_methods`. |
| `facilitator_url` | yes | The verifier endpoint. e.g. `https://x402.org/facilitator`. |
| `network` | yes | CAIP-2 chain ID. `eip155:8453` = Base Mainnet, `eip155:84532` = Base Sepolia. |
| `scheme` | no | `exact` (default) or `upto`. xgateway currently supports `exact` end-to-end. |
| `merchant` | yes | The wallet address that receives payment. |
| `asset` | no | Token contract address (the `asset` field in 402). |
| `decimals` | no | Token decimals. USDC = 6. Used to convert `$x.xx` prices to on-chain amounts. |
| `timeout_seconds` | no | HTTP timeout when calling the facilitator. Default 30. |
| `sync_facilitator_on_start` | no | If true, contact the facilitator at startup to validate config. |

You can define multiple `x402` entries — for example one on Base Mainnet for production and one on Base Sepolia for testing. Each rule then picks the entries it accepts via `payment_methods`.

## The `mpp` block

Each entry defines one accepted MPP (Machine Payments Protocol) payment channel. MPP settles an on-chain charge directly against a blockchain RPC endpoint — there is no facilitator. Only the **Tempo** method and the **`charge`** scheme work end-to-end today.

| Field | Required | Notes |
|---|---|---|
| `name` | yes | A unique identifier you use inside `rules[].payment_methods`. |
| `method` | yes | Only `tempo` works end-to-end. `stripe` is parsed but rejected during validation. |
| `scheme` | yes | Only `charge` works end-to-end. `session` is parsed but rejected during validation. |
| `rpc_url` | yes | The Tempo JSON-RPC endpoint the charge is verified and settled against. |
| `merchant` | yes | The wallet address that receives payment. |
| `asset` | yes | Token contract address — e.g. pathUSD, the Tempo stablecoin. |
| `decimals` | yes | Base-unit precision of the route `price`. pathUSD = 6. |
| `secret_key` | yes | HMAC key used to sign the 402 challenge. |
| `timeout_seconds` | no | RPC call timeout. Default 30. |

A rule can list both x402 and MPP channels in `payment_methods`. The gateway uses MPP when the client sends an MPP `Authorization` header, and x402 otherwise; a rule that lists only MPP channels always uses MPP.

> **MPP price must be in base units.** Unlike x402, any rule that accepts an MPP channel must express `price` as the raw integer in the asset's base units — `"100000"` for $0.10 at 6 decimals — not the `"$0.10"` dollar form. See *Price formats* below.

## The `outbound` block

`outbound` describes the single upstream and the routes that point at it. File mode supports exactly one upstream — that is the main structural difference from HTTP mode.

### `outbound.target`

The upstream's base URL. Every rule's path is appended to this when forwarding.

If your rule is `path: /weather` and `target: http://api:4021`, requests to `/weather` are forwarded to `http://api:4021/weather`.

### `outbound.auth_header`

Injects a single header on every forwarded request:

```yaml
auth_header:
  enable: true
  name: Authorization
  value: "Bearer sk_live_..."
```

If `enable` is `false` (or the block is missing), no header is injected. If `name` is omitted but `enable` is true, the header defaults to `Authorization`.

The header is added on the outbound side — clients never see it.

### `outbound.allow_unmatched`

When a request arrives for a path that no rule matches:

- `false` (default) — return **403 Forbidden** without touching the upstream.
- `true` — proxy the request to `target + path` for **free**, no 402, no payment.

Default to `false` for paid APIs. Use `true` only if you intend the whole upstream to be public except for the paths that have explicit paid rules.

### `outbound.rules[]`

The interesting part. Each entry is one path rule.

| Field | Required | Notes |
|---|---|---|
| `name` | yes | A human label; appears in logs. |
| `path` | one of | Exact path: `/weather`. |
| `path_glob` | one of | Glob pattern: `/api/v1/*`. Uses Go's `path.Match` — `*` matches one segment. |
| `price` | required if not `free` | USD amount. See *Price formats* below. |
| `free` | no | If true, skip payment entirely. `price` and `payment_methods` are ignored. |
| `payment_methods` | conditional | List of `x402[].name` or `mpp[].name` entries this rule accepts. If omitted, defaults to all supported entries. |
| `description` | no | Human-readable; appears in logs and the 402 metadata. |
| `bazaar` | no | Discovery extension metadata; see *Bazaar* below. |

Use **either** `path` or `path_glob`, not both. If both are set, `path` wins.

#### Price formats

Two forms are accepted on the same field:

- `price: "$0.10"` — human form, parsed back to a number using the asset's decimals.
- `price: "100000"` — raw on-chain integer, used as-is.

Both produce the same on-chain amount when the asset is USDC with 6 decimals. The dollar form is easier to maintain.

**MPP routes require the integer form.** The dollar form is x402-only; for any rule that accepts an MPP channel, use the base-units integer (e.g. `"100000"`).

Free routes omit `price` entirely:

```yaml
- name: health
  path: /health
  free: true
```

#### Bazaar

The x402 Bazaar discovery extension is an optional way to advertise the endpoint to facilitator-side catalogs. If you do not need this, leave it out — xgateway auto-generates a minimal `GET` declaration for every paid route.

To override:

```yaml
bazaar:
  method: POST
  body_type: application/json
  input_schema:
    type: object
    properties:
      city: { type: string }
  output_example:
    temp: 21
    units: C
```

Set `disabled: true` to opt this route out of the catalog entirely.

## Matching order

Within a single rule list, xgateway returns the first match. Order your rules from most specific to least specific:

```yaml
rules:
  - { name: health,   path: /health,        free: true }      # exact
  - { name: v1-users, path: /api/v1/users,  price: "$0.10" } # exact
  - { name: v1,       path_glob: /api/v1/*, price: "$0.01" } # glob
  - { name: default,  path_glob: /*,        price: "$0.05" } # catch-all
```

A request to `/api/v1/users` matches the second rule (exact wins over glob because it is listed first).

If you swap the order so `/api/v1/*` comes before `/api/v1/users`, the wildcard wins because xgateway picks the **first match** during iteration — it does not optimise for specificity in file mode. The lesson: keep exact rules above the globs they overlap with.

## Validation at startup

xgateway refuses to start if any of:

- `CONFIG_FILE` is empty or unreadable.
- The file is not valid YAML or JSON.
- `outbound.target` is empty.
- Any `x402[]` entry is missing `name`, `facilitator_url`, `network`, `merchant`, or has an invalid `scheme`.
- Any `mpp[]` entry is missing `name`, `rpc_url`, `merchant`, `asset`, or `secret_key`, uses a `method` other than `tempo`, or a `scheme` other than `charge`.
- Any rule has neither `path` nor `path_glob`.
- A paid rule references a `payment_methods[]` entry that is not defined.

The error message tells you the path of the bad field (e.g. `x402[0].network is required`).

## Complete example

A realistic small config — one free endpoint, two paid routes, two networks accepted:

```yaml
x402:
  - name: mainnet
    facilitator_url: https://x402.org/facilitator
    network: eip155:8453
    scheme: exact
    merchant: "0xMainnetPayoutAddress"
    asset: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
    decimals: 6

  - name: sepolia
    facilitator_url: https://x402.org/facilitator
    network: eip155:84532
    scheme: exact
    merchant: "0xTestnetPayoutAddress"
    asset: "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
    decimals: 6

outbound:
  target: http://my-api:4021
  auth_header:
    enable: true
    name: Authorization
    value: "Bearer prod-key-abc"
  allow_unmatched: false
  rules:
    - name: health
      path: /health
      free: true

    - name: weather
      path: /weather
      price: "$0.001"
      description: "Returns the current weather"
      payment_methods: [mainnet, sepolia]

    - name: forecast
      path_glob: /forecast/*
      price: "$0.005"
      payment_methods: [mainnet, sepolia]
```

Clients calling `/weather` will see a 402 with **two** entries in `accepts[]` — one for mainnet and one for sepolia. They pick whichever they have funds on.

## Running it

```bash
docker run \
  -e CONFIG_PROVIDER=file \
  -e CONFIG_FILE=/etc/xgateway/config.yaml \
  -e PUBLIC_URL=http://localhost:3102 \
  -v ./config.yaml:/etc/xgateway/config.yaml:ro \
  -p 3102:8081 \
  xpaywall/xgateway:latest
```

There is no project slug in file mode — paths map directly. A client request to `http://localhost:3102/weather` matches `rules[].path: /weather`.

## What's next?

- Background on how matching decisions are made: [05 — Concepts](./../05-concepts.md#route-resolution).
