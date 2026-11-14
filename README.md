# WriteWisely

AI-powered IELTS Writing checker for Task 1 (Academic) and Task 2. It gives an estimated band score per criterion, inline grammar corrections, and lexical-variety suggestions. The user then revises their own answer and the AI re-scores it.

- `server/`: Go + Fiber API
- `client/`: Next.js frontend (coming in PR-2)
- `docs/`: project brief, data design, API conventions, task specs
- Engineering rules: see [`AGENTS.md`](AGENTS.md)

## Prerequisites

- Go 1.24+ (developed on 1.25)
- Docker with Docker Compose v2
- GNU Make

You don't need to install `swag` or `migrate` globally. `swag` is pinned as a Go tool in `server/go.mod`, and the Makefile runs a pinned `golang-migrate` through `go run`.

## Run the server locally

All commands run from the repository root.

```sh
# 1. Start Postgres 16 (published on localhost:5433)
docker compose up -d
docker compose ps            # wait until postgres shows "healthy"

# 2. Configure env
cp server/.env.example server/.env

# 3. Apply migrations
make migrate-up

# 4. Start the API on :8080
make run
```

Then check:

- Health: `curl localhost:8080/api/v1/health` returns `{"status":"ok","db":"up"}`. If the database is unreachable it returns `503` with `{"status":"degraded","db":"down"}`.
- API docs (Swagger UI): <http://localhost:8080/swagger>

## Make targets

| Target | What it does |
|--------|--------------|
| `make run` | Start the API server (reads `server/.env`) |
| `make test` | Run server tests |
| `make docs` | Regenerate the OpenAPI spec into `server/docs/` (`swag init` via `go generate`) |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-down` | Revert the most recent migration |

The migrate targets use the `DB_*` variables from your environment, or from `server/.env` if `DB_HOST` isn't set. golang-migrate needs a URL, so for these targets `DB_PASSWORD` must be URL-safe (the server itself has no such restriction).

## Configuration

Configuration comes from environment variables only. In development, `server/.env` is loaded if it exists, and real environment variables take precedence over it.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DB_HOST` | yes | (none) | Postgres host |
| `DB_PORT` | no | `5432` | Postgres port (`5433` for the docker-compose service) |
| `DB_USER` | yes | (none) | Postgres user |
| `DB_PASSWORD` | yes | (none) | Postgres password |
| `DB_NAME` | yes | (none) | Database name |
| `DB_SSLMODE` | no | `disable` | libpq `sslmode` |
| `PORT` | no | `8080` | HTTP listen port |
| `APP_ENV` | no | (empty) | Environment name, e.g. `development` |

The server refuses to start if a required `DB_*` variable is missing. It connects with a libpq keyword/value string, so special characters in `DB_PASSWORD` need no escaping.

## API contract

The API is documented code-first. Each handler carries swaggo annotations, and `make docs` regenerates `server/docs/` (`swagger.json` / `swagger.yaml` / `docs.go`). Don't edit `server/docs/` by hand. After changing an endpoint, run `make docs` and commit the result. Conventions are in [`docs/api-conventions.md`](docs/api-conventions.md).
