# Admin Panel — Requests

The **Requests** page is a log of every request that has gone through the gateway. Use it to debug payment problems, audit traffic, and confirm that a paid endpoint is actually being paid for.

![Requests list](./../images/requests-list.png)

## The table

Each row is one client request. Columns typically include:

- **Time** — when the request hit the gateway.
- **Method** — HTTP method (GET, POST, ...).
- **Path** — the inbound path the client called.
- **Project / Route** — which configured route matched.
- **Status** — the final outcome:
  - `paid` — payment verified, upstream returned 2xx.
  - `free` — free route, proxied without payment.
  - `payment_required` — a 402 was returned (the client has not paid yet, or chose not to).
  - `upstream_error` — payment was fine but the upstream returned 4xx/5xx.
  - `error` — the gateway itself failed.
- **HTTP code** — the response the client got.
- **Latency** — total time from request to response, in milliseconds.
- **Amount** — when paid, the USD value charged.
- **Client IP** — the client's IP as seen by the gateway.

Sort, filter and paginate from the column headers. Use filters to narrow by status, project, route or time window when looking for a specific event.

## What's next?

- Recipes to fix specific errors: [08 — Troubleshooting](./../08-troubleshooting.md).
- Glossary of every term used in the events: [10 — Reference](./../10-reference.md#glossary).
