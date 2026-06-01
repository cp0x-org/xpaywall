# Admin Panel — Dashboard

The Dashboard is the first page you see after logging in. It is a read-only summary of the activity going through the gateway.

![Dashboard](./../images/dashboard-main.png)

## Summary cards

The top of the page shows aggregate counters. Typical cards:

- **Revenue** — how much the gateway has earned from successful payments.
- **Total requests** — every request the gateway has handled.
- **Success Rate** — the percentage of requests that got a successful payment proof and were proxied to the upstream. A low success rate means many clients are starting the payment flow but not finishing it.
- **Active Routes** — how many routes are enabled. 

## Top routes

A list of the most-hit routes in the selected period, with revenue and call counts. Useful for spotting your most profitable endpoints and your highest-traffic ones.

![Top routes card](./../images/dashboard-top-routes.png "medium")

## Recent requests

A timeline of the latest individual requests. Click through to open the **Requests** page filtered to that request for full details.

![Recent requests card](./../images/dashboard-recent-requests.png "medium")

## What's next?

- Drill into individual requests: [Requests](./09-requests.md).
- Add a payment configuration so the counters start moving: [Facilitators](./03-facilitators.md).
