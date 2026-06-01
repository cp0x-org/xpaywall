# xpaywall adminpanel

React dashboard for the [xpaywall](../../README.md) control plane. It talks to `control-api` over HTTP and lets you manage projects, payment routes, payment channels, users, and view request logs and dashboard stats.

Built on the **Berry MUI** template — **React 19**, **TypeScript**, **MUI v7**, **Redux Toolkit**, **Vite 7**, **Yarn 4**.

## Run

```bash
yarn install
yarn start        # dev server on port 3000
```

The dev server proxies API calls to `VITE_APP_API_URL` (defaults to `http://localhost:3010/`). Point it at your local control-api by exporting the env var or editing `.env.local`.

## Commands

| Command | What it does |
|---|---|
| `yarn start` | Vite dev server on port 3000 |
| `yarn build` | `tsc` + `vite build` (production) |
| `yarn lint` | ESLint check |
| `yarn lint:fix` | ESLint auto-fix |
| `yarn prettier` | Format `src/**` |
| `yarn tsc` | Type-check only (no emit) |

No test suite is configured — `yarn tsc` and `yarn lint` are the primary correctness checks.

## Environment variables

| Variable | Required | Default | Notes |
|---|---|---|---|
| `VITE_APP_API_URL` | dev | `http://localhost:3010/` | control-api base URL — used by `src/utils/axios.ts` |
| `VITE_APP_BASE_NAME` | build | `/` | React Router base path |
| `API_URL` | runtime | — | Browser-accessible control-api URL (injected via `public/config.js`) |
| `PROXY_URL` | runtime | — | Browser-accessible xgateway URL (injected via `public/config.js`) |

`API_URL` and `PROXY_URL` are runtime values read from `window.__CONFIG__`, so they can be changed without rebuilding the Docker image.

## Architecture

### API layer

All HTTP calls go through `src/utils/axios.ts`, which is an axios instance pointing at `VITE_APP_API_URL`. It injects the JWT from `localStorage.serviceToken` and redirects to `/login` on 401.

> Use `axiosServices` (the default export) for every control-api call — never raw `fetch` or bare `axios`.

### Auth

JWT-based. `utils/route-guard/AuthGuard` wraps all main routes. The active auth context is `src/contexts/JWTContext.tsx`. The other auth contexts in the template (Firebase, Auth0, AWS, Supabase) are unused leftovers.

### Routing

Three route groups in `src/routes/`:

- `LoginRoutes` — `/login`
- `MainRoutes` — authenticated app routes wrapped in `AuthGuard + MainLayout`
- `AuthenticationRoutes` — misc auth pages

All main pages are lazy-loaded via `ui-component/Loadable`.

### Pages (`src/views/`)

| Route | View | Notes |
|---|---|---|
| `/dashboard` | `views/dashboard` | Stats cards + chart |
| `/projects` | `views/projects` | Full CRUD; archive (soft-delete) preserves request logs |
| `/routes` | `views/routes-page` | Route CRUD with Formik + Yup validation |
| `/payment-channels` | `views/payment-channels` | List from `/api/v1/payment-channels` |
| `/stats` | `views/stats` | Static rows via `EntityListPage` |

`views/entity-pages/` contains generic `EntityListPage` / `EntityFormPage` / `EntityTable` components for simple list/form pages.

### State

Redux store (`src/store/`) holds only `snackbar` state. Pages fetch data on mount with local `useState` — there is no global API state. `redux-persist` is wired but only persists snackbar.

### Navigation menu

Defined in `src/menu-items/`. Add a `NavItemType` entry there to surface a route in the sidebar.

## Key conventions

- Form pages detect create / edit / view mode from `useLocation().pathname` — pass `state: { id }` via React Router `Link`/`navigate` for edit and view.
- MUI v7 `Grid` uses the `size` prop (`<Grid size={{ xs: 12, md: 6 }}>`) — not `xs`/`md` directly on `Grid`.
- `src/ui-component/` holds shared cards, charts, and MUI re-exports. Check there before adding a new wrapper component.
- Per-domain TypeScript types live alongside their view (`views/projects/types.ts`, etc.).

## Contributing

Contributions follow the rules in the top-level [`CONTRIBUTING.md`](../../CONTRIBUTING.md).

## License

Released under the [MIT License](../../LICENSE).
