# Task 01 — Project Scaffold

> Status: Todo · For: AI builder · Reviewer: Apip
> Reads first: `AGENTS.md`, `docs/data-design.md`, `docs/api-conventions.md`.
> Goal: a running skeleton for `server/` and `client/` with the OpenAPI + Postgres plumbing wired. **No business logic** (no evaluation, no submission endpoints) beyond a health check.

## Scope rules

- Build **only** what this task lists. Do not add submission/attempt/evaluation endpoints, the Evaluator, or any Claude integration yet — those are later tasks.
- Create the full folder structure from AGENTS.md §2, even where a folder only gets a placeholder file.
- **Deliver as two PRs** to keep review small: **PR-1 = server**, **PR-2 = client** (client depends on server's generated spec).

## Fixed technology choices (do not substitute)

| Concern | Choice |
|--------|--------|
| HTTP framework | Fiber v2 |
| Swagger generation | `swaggo/swag` + `github.com/gofiber/swagger` (serve UI at `/swagger`) |
| Postgres driver | `jackc/pgx` v5 with `pgxpool` |
| Migrations | `golang-migrate` (`.up.sql` / `.down.sql` in `server/migrations/`) |
| Local env loading | `joho/godotenv` for dev + `os.Getenv` (no viper) |
| Client | `create-next-app` (App Router, TypeScript strict) |
| Client type-gen | `openapi-typescript` → `client/src/lib/api/types.gen.ts` |
| Go module path | `writewisely/server` (local module; can point to a GitHub path later) |

---

## PR-1 — Server scaffold (`server/`)

1. **Module & structure:** `go mod init writewisely/server`. Create the full tree from AGENTS.md §2 (`cmd/`, `internal/{config,domain,handler,service,repository,adapter/anthropic}`, `migrations/`, `pkg/`). Empty-for-now packages get a short doc comment file (no stubs that pretend to do work).
2. **Config** (`internal/config`): load from env — `PORT` (default 8080), `DATABASE_URL`, `APP_ENV`. Use `godotenv` to load `.env` in dev. Fail fast with a clear error if `DATABASE_URL` is missing.
3. **DB** (`internal/repository` or a `db` helper): open a `pgxpool` from `DATABASE_URL` on startup and `Ping` it. Log success/failure. Expose the pool via dependency injection from `main.go` — no global.
4. **Health endpoint:** `GET /api/v1/health` → `200 {"status":"ok","db":"up"}` (report `"down"` + `503` if the DB ping fails). Fully annotated with swaggo (per api-conventions §4) so it appears in the spec.
5. **Swagger:** wire `swaggo/swag` with `//go:generate swag init` at the entrypoint; serve the UI at `/swagger` via `gofiber/swagger`. Generated output in `server/docs/` (git-committed, but never hand-edited).
6. **Migration `0001_init`:** create the schema from `docs/data-design.md` exactly — tables `writing_submission`, `attempt`, `evaluation`; the `text + CHECK` enums; FKs with `ON DELETE CASCADE`; the `UNIQUE` and index definitions; `gen_random_uuid()` defaults (enable `pgcrypto` if needed). Provide matching `.down.sql`.
7. **Makefile targets:** `run`, `test`, `migrate-up`, `migrate-down`, `docs` (runs `swag init`; the client type-gen step is added in PR-2).
8. **Files:** `.env.example` (all env vars), update root `docker-compose.yml` with a `postgres:16` service (volume, healthcheck, env), update `README.md` with run steps.

### PR-1 acceptance checklist (reviewer)
- [ ] `docker-compose up -d` starts Postgres; healthcheck passes.
- [ ] `make migrate-up` applies `0001_init`; all 3 tables + constraints + indexes exist (`\d+` check).
- [ ] `make migrate-down` cleanly reverts.
- [ ] Server starts; `GET /api/v1/health` returns `{"status":"ok","db":"up"}`.
- [ ] Stopping Postgres makes health return `503` + `"db":"down"` (no crash).
- [ ] `/swagger` loads and shows the health endpoint with its documented response.
- [ ] `make docs` regenerates `server/docs/` with no diff afterward (spec is in sync).
- [ ] No secrets committed; `.env.example` present. No endpoints beyond health.

---

## PR-2 — Client scaffold (`client/`)

1. **Bootstrap:** `create-next-app` (App Router, TS strict, ESLint). Create the tree from AGENTS.md §2 (`src/app`, `src/components/{chart,feedback,ui}`, `src/lib/{api,utils}`).
2. **Type-gen:** add `openapi-typescript` to generate `src/lib/api/types.gen.ts` from `server/docs/swagger.json`. Add the step to the root `make docs` target so one command regenerates both sides.
3. **API client** (`src/lib/api/client.ts`): a minimal typed `fetch` wrapper using the generated types + `NEXT_PUBLIC_API_BASE_URL`. Centralized error handling to the standard error shape.
4. **Smoke screen:** home `page.tsx` calls `GET /api/v1/health` and renders the status — proves client↔server connectivity end to end.
5. **Files:** `.env.local.example` (`NEXT_PUBLIC_API_BASE_URL=http://localhost:8080`), README run steps.

### PR-2 acceptance checklist (reviewer)
- [ ] `make docs` generates `types.gen.ts` from the server spec with no error.
- [ ] `npm run dev` starts the client.
- [ ] Home page shows the live API health status (green when server up, error state when down).
- [ ] No hand-written request/response types; only `client.ts` touches `fetch`.
- [ ] `tsc` passes in strict mode.

---

## Out of scope (explicitly later)
Submission/attempt/evaluation endpoints · repetition detector · Evaluator interface & Claude adapter · chart rendering · auth · CI pipeline. These come after scaffold is merged, via the roadmap.
