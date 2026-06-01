# Guide 04 — Connecting a real upstream

The example server inside the repo is fine for trying things out, but it does not match what you actually want to monetise. This guide replaces the example with your real API.

There is nothing exotic to do — you change one URL on the project and (optionally) one auth header.

## Where the upstream URL lives

Every project has a **Server Base URL** field. That is the only place the gateway looks when forwarding a request.

The forwarding rule is simple:

```
<Server Base URL> + <route's path>
```

So if the project's base URL is `https://api.example.com` and a route's path is `/forecast`, calls to `https://gateway.example.com/<slug>/forecast` are forwarded to `https://api.example.com/forecast`. Any query string the client sent is appended unchanged.

## Step 1 — Point the project at your upstream

Open **Projects → Project List**, edit your project, and change **Server Base URL** to your real API's URL. Save.

> **Screenshot placeholder:** ![Project form](./../images/projects-form.png)

Common values:

- `https://api.example.com` — public API on a different host.
- `http://my-service:8080` — internal service reachable inside the same Docker network (use the container name, not `localhost`).
- `https://api.example.com/v2` — sub-path of a larger API. The routes' paths are appended on top, so a route `/forecast` becomes `https://api.example.com/v2/forecast`.

Save. The change applies on the next request — no restart needed.

## Step 2 — Add an auth header (if your upstream needs one)

If your upstream rejects calls that do not carry an authentication header, set it on the project. The gateway will inject it on every forwarded request, transparent to clients.

In the project form:

- **Auth Header Name:** `Authorization`
- **Auth Header Value:** `Bearer sk_live_abc123` (or whatever your upstream expects)

Save.

This header is **never sent back to the client** — the client only sees the upstream's response body. Clients have no way to read this token from the gateway.

## Step 3 — Verify with curl

```bash
curl -i http://localhost:3102/<slug>/<some-existing-route>
```

You should see:
- `402` for paid routes (which is fine — payment is unrelated to the upstream URL).
- A real `200` response from your upstream for free routes.

If you see a 502 or 504, the gateway could not reach the upstream. Check that the URL is reachable from inside the gateway container (`docker exec xgateway sh` then `wget -qO- <your-url>`).

## A subtlety: container networking

If your upstream runs on the same host as xpaywall, `localhost` *inside* the gateway container does not mean the same machine — it means the container itself. Use one of:

- The container name, if the upstream is in the same Docker network: `http://my-upstream:8080`.
- The host's DNS name from inside Docker: `http://host.docker.internal:8080` (Docker Desktop on Mac/Windows).
- The host's IP on the Docker bridge.

For production, your upstream is probably reachable by public DNS — use that.

## When the upstream is behind another auth scheme

xpaywall sends exactly one auth header at most: the one you configured on the project. If your upstream needs more headers, you have two options:

1. Put a small adapter service in front of your real upstream, with the simple auth header xpaywall can inject. Have the adapter translate to whatever your real upstream needs.
2. Set up your upstream to accept the gateway's standard auth header.

There is no per-route auth header today — the auth is project-wide.

## What's next?

- Lock down the public surface: [09 — Security](./../09-security.md).
- Move from a YAML file to the admin panel: [Guide 05 — Migrating from file mode to HTTP mode](./05-migrating-file-to-http.md).
