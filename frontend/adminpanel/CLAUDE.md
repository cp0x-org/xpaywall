# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
yarn start        # dev server on port 3000
yarn build        # tsc + vite build (production)
yarn lint         # eslint check
yarn lint:fix     # eslint auto-fix
yarn prettier     # format src/**
yarn tsc          # type-check only (no emit)
```

No test suite is configured — `yarn tsc` and `yarn lint` are the primary correctness checks.

## Architecture

Built on the **Berry MUI** template (MUI v7, React 19, Vite 7, Redux Toolkit). Only the domain-specific parts matter for this project; most of the template scaffolding (auth contexts, themes, ui-component overrides) is boilerplate.

### API layer

All HTTP calls go through `src/utils/axios.ts`, which is an axios instance pointing at `VITE_APP_API_URL` (defaults to `http://localhost:3010/`). It injects the JWT from `localStorage.serviceToken` and redirects to `/login` on 401. Use `axiosServices` (default export) for all control-api calls — never raw `fetch` or bare `axios`.

### Auth

JWT-based. `utils/route-guard/AuthGuard` wraps all main routes. The active auth context is `src/contexts/JWTContext.tsx` (others — Firebase, Auth0, AWS, Supabase — are template leftovers and unused).

### Routing

Three route groups in `src/routes/`:
- `LoginRoutes` — `/login`
- `MainRoutes` — authenticated app routes wrapped in `AuthGuard + MainLayout`
- `AuthenticationRoutes` — misc auth pages

All main pages are lazy-loaded via `ui-component/Loadable`.

### Pages (views/)

| Route | View | Notes |
|---|---|---|
| `/dashboard` | `views/dashboard` | Stats cards + chart; direct axios calls |
| `/projects` | `views/projects` | Full CRUD; `ProjectForm` handles create/edit/view by inspecting `pathname` and `location.state.id` |
| `/routes` | `views/routes-page` | Route CRUD; `RouteFormPage` uses Formik + Yup |
| `/payment-channels` | `views/payment-channels` | List from `/api/v1/payment-channels` |
| `/stats` | `views/stats` | Static rows via `EntityListPage` |

`views/entity-pages/` contains generic `EntityListPage` / `EntityFormPage` / `EntityTable` components for pages with static or simple data. Domain-specific pages (projects, routes, payment-channels) have their own table/form components.

### State

Redux store (`src/store/`) holds only `snackbar` state. No global API state — pages fetch on mount with local `useState`. `redux-persist` is wired but only persists snackbar.

### Navigation menu

Defined in `src/menu-items/`. Add a new `NavItemType` entry here to make a route appear in the sidebar.

## Key conventions

- Form pages detect create/edit/view mode from `useLocation().pathname` — pass `state: { id }` via React Router `Link`/`navigate` for edit and view.
- MUI v7 `Grid` uses the `size` prop (`<Grid size={{ xs: 12, md: 6 }}>`) — not `xs`/`md` directly on Grid.
- `src/ui-component/` has shared cards, charts, and MUI re-exports — check there before adding new MUI wrapper components.
- Types for each domain live alongside their view: `views/projects/types.ts`, `views/routes-page/types.ts`, etc.
