# Contributing to xpaywall

Thanks for contributing. This guide covers the repo layout, dev setup, and what every pull request should follow.

## Repository layout

xpaywall is a multi-service monorepo with one Git submodule:

| Path | Service | Tracked as |
|---|---|---|
| `xgateway/` | Payment-enforcing reverse proxy (Go) | **Submodule** (`cp0x-org/xgateway`) — see [`xgateway/CONTRIBUTING.md`](xgateway/CONTRIBUTING.md) |
| `control-api/` | REST control plane (Go) | In-tree |
| `frontend/adminpanel/` | React admin dashboard | In-tree |
| `examples/server/` | Sample upstream API for testing | In-tree |
| `docs/` | Project documentation | In-tree |

Clone and initialise the submodule:

```bash
git clone https://github.com/cp0x-org/xpaywall.git
cd xpaywall
git submodule update --init --recursive
```

Changes to `xgateway/` go to its own repo. Open a PR there for gateway-only changes, then bump the submodule pointer here in a follow-up PR.

## Prerequisites

- Go 1.26+
- Node.js 22+ and Yarn 4 (for the admin panel)
- PostgreSQL 16
- Docker / Docker Compose

## Getting the stack running

```bash
docker compose up -d
```

This brings up Postgres, control-api, xgateway, the example upstream, and the admin panel. See the [root README](README.md) for the host-port map and login credentials.

For local development against a single service, see that service's README.

## Pull request checklist

Every PR must pass:

| Area | Command |
|---|---|
| control-api | `go test ./...` |
| xgateway | `go test ./...` |
| adminpanel | `yarn lint && yarn build` |

Before opening a PR:

- Run the test/lint commands above.
- Keep the diff focused — don't bundle unrelated refactors.
- Match the surrounding code style. No `logrus`, no `pkg/errors`. Max Go line length 140.
- Update or add tests when you change behaviour.
- Update docs in `docs/` (and the service README) if your change is user-visible.

## Database changes (control-api)

- New schema goes through a goose migration in `control-api/migrations/`.
- Update SQL under `control-api/internal/storage/postgres/queries/` and regenerate sqlc:

  ```bash
  cd control-api
  sqlc generate
  ```

- **Never hand-edit** files under `control-api/internal/storage/postgres/generated/`.

## API changes

If you change a control-api handler signature or response shape:

1. Update the corresponding type in `frontend/adminpanel/src/types/` and the API wrapper in `frontend/adminpanel/src/api/`.
2. Update Swagger annotations and regenerate docs:

   ```bash
   cd control-api
   go tool swag init -g swagger_meta.go --dir ./cmd/control-api,./ --output docs --parseInternal
   ```

## Commit messages

Follow the Conventional Commits style used in this repo:

```
feat(scope): short summary
fix(scope): short summary
docs(scope): short summary
```

Examples from history: `feat(docker): update image tags`, `feat(migrate): add command for applying database migrations`.

## Branching

- `main` is the integration branch.
- Open PRs against `main` from a feature branch.
- Release branches use the pattern `xpaywall-<version>` (e.g. `xpaywall-0.1.0`).

## Reporting bugs and proposing features

Open a GitHub issue describing:

- What you expected vs what actually happened
- A minimal reproduction (config snippet, request, log output)
- Versions: `git rev-parse HEAD`, Go version, browser / OS where relevant

## Security

Do **not** open a public issue for security reports. See [`docs/09-security.md`](docs/09-security.md) for the disclosure process.

## License

By contributing, you agree your contributions will be licensed under the [MIT License](LICENSE).
