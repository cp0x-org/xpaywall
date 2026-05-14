# xpaywall proxy

HTTP reverse proxy that enforces x402 payment rules before forwarding requests to an upstream service.

## Run

```powershell
$env:PROXY_CONFIG = ".\\config.yaml"
go run ./cmd/xgateway
```

You can also load env vars from a file:

```powershell
go run ./cmd/xgateway --env-file .env
```

## Config

Copy [`config.example.yaml`](./config.example.yaml) to `config.yaml` and set:

- `upstream.url`
- `x402.facilitator_url`
- `x402.network`
- `pricing.rules[].pay_to`

## Test

```powershell
go test ./...
```
