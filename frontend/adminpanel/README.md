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
| `VITE_GOOGLE_CLIENT_ID` | no | — | Google OAuth client ID. Empty ⇒ the Google sign-in button is not rendered. Must match control-api's `GOOGLE_CLIENT_ID`. |
| `API_URL` | runtime | — | Browser-accessible control-api URL (injected via `public/config.js`) |
| `PROXY_URL` | runtime | — | Browser-accessible xgateway URL (injected via `public/config.js`) |

`API_URL` and `PROXY_URL` are runtime values read from `window.__CONFIG__`, so they can be changed without rebuilding the Docker image.

## Architecture

### API layer

All HTTP calls go through `src/utils/axios.ts`, which is an axios instance pointing at `VITE_APP_API_URL`. It injects the JWT from `localStorage.serviceToken` and redirects to `/login` on 401.

> Use `axiosServices` (the default export) for every control-api call — never raw `fetch` or bare `axios`.

### Auth

JWT-based, with optional **Google sign-in** (`views/pages/authentication/GoogleLoginButton.tsx`, gated by `VITE_GOOGLE_CLIENT_ID`). `utils/route-guard/AuthGuard` wraps all main routes; the auth context is `src/contexts/JWTContext.tsx`. The login/register pages also drive **Forgot password → Reset password** (the reset token arrives by email when SMTP is configured on control-api, otherwise it is returned by the API and the page redirects to the reset form). The template's other auth providers (Firebase, Auth0, AWS, Supabase) and their demo pages have been **removed** — only JWT + Google remain.

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
| `/projects` | `views/projects` | Full CRUD; archive (soft-delete) preserves request logs. Slug is unique **per owner**. |
| `/routes` | `views/routes-page` | Route CRUD with Formik + Yup. The **Proxy URL** column/preview is `/{username}/{slug}/{route}` (username from the logged-in user). |
| `/payment-methods` | `views/payment-methods` | CRUD; **Global** column. Global rows are view-only for non-superadmins (Edit/Delete superadmin-only). |
| `/payment-assets` | `views/payment-assets` | CRUD; same Global handling. |
| `/facilitators` | `views/facilitators` | CRUD; same Global handling. |
| `/project-payment-methods` | `views/project-payment-methods` | Links a project to a payment method + asset. |
| `/requests` | `views/requests` | Request logs + per-request events. |
| `/stats` | `views/stats` | Static rows via `EntityListPage`. |

The **Global (visible to all users)** toggle on the payment-method / asset / facilitator forms (`ui-component/GlobalScopeToggle.tsx`) is rendered only for superadmins. `views/entity-pages/` holds generic `EntityListPage` / `EntityFormPage` / `EntityTable` components for simple list/form pages.

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
