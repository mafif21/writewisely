# AGENTS.md — WriteWisely Engineering Guide

> This document is written for an **AI coding agent** building WriteWisely, and for the human reviewer checking its output.
> Keep this file at the repository root as `AGENTS.md` (and/or `CLAUDE.md`).
> Read `docs/project-brief.md` and `docs/data-design.md` before writing any code.

## 0. Working model

- Implementation is done by an **AI agent**. The human (Apip) acts as **reviewer**.
- Therefore: follow specs **literally**. If a spec is ambiguous or missing, **stop and ask** — do not invent product behavior, endpoints, schemas, or dependencies.
- Prefer small, reviewable changes with clear reasoning over large speculative ones.

## 1. What we are building (one paragraph)

An AI-powered IELTS Writing checker for **Task 1 (Academic)** and **Task 2**. A user submits an essay (optionally with a chart for Task 1); the app returns an estimated band score (per IELTS criterion + overall), inline grammar corrections with explanations, and lexical-variety suggestions as clickable badges. The user then **revises their own answer** and the AI re-scores the revision. MVP is single-user, no auth.

## 2. Repository structure

```
writewisely/
├── AGENTS.md                 # this file (root)
├── README.md
├── docker-compose.yml        # Postgres for local dev
├── Makefile                  # dev / test / docs targets
├── docs/                     # project docs
│   ├── project-brief.md
│   ├── data-design.md
│   ├── api-conventions.md
│   ├── roadmap.md
│   └── tasks/                # unit-of-work specs (01-scaffold.md, …)
│
├── server/                   # Go + Fiber backend
│   ├── cmd/
│   │   └── api/
│   │       └── main.go               # entrypoint: config, wiring, start Fiber
│   ├── internal/                     # not importable from outside the module
│   │   ├── feature/                  # business modules (package-by-feature)
│   │   │   ├── submission/
│   │   │   │   ├── handler.go         # Fiber handlers + swaggo annotations (HTTP only)
│   │   │   │   ├── service.go         # business logic
│   │   │   │   ├── repository.go      # Postgres access (+ persistence struct / DAO if DB shape differs)
│   │   │   │   ├── dto.go             # API request/response structs (json tags)
│   │   │   │   ├── model.go           # domain entity
│   │   │   │   └── service_test.go    # unit tests — colocated
│   │   │   └── evaluation/
│   │   │       ├── service.go         # orchestrates scoring + repetition
│   │   │       ├── repository.go
│   │   │       ├── evaluator.go       # Evaluator interface — defined where it is consumed
│   │   │       ├── repetition.go      # deterministic repetition detector
│   │   │       ├── dto.go
│   │   │       ├── model.go
│   │   │       └── repetition_test.go
│   │   ├── adapter/
│   │   │   └── anthropic/            # implements evaluation.Evaluator (Claude); imports evaluation
│   │   │       ├── client.go
│   │   │       ├── prompts.go        # band descriptors + few-shot calibration
│   │   │       └── schema.go         # structured-output parse + validate
│   │   ├── platform/                 # shared infrastructure (framework/IO)
│   │   │   ├── config/               # env loading
│   │   │   ├── database/             # pgxpool setup
│   │   │   └── server/               # Fiber app, router, middleware, health endpoint
│   │   └── constant/                 # shared constants, one file per topic
│   │       ├── health.go
│   │       ├── apierror.go           # error-code strings
│   │       └── enum.go               # task_type / chart_mode / prompt_source values
│   ├── docs/                         # swag-GENERATED OpenAPI (docs.go, swagger.json/yaml) — do not edit by hand
│   ├── migrations/                   # 0001_init.up.sql / .down.sql
│   ├── test/                         # integration / e2e tests ONLY (unit tests stay colocated)
│   ├── .env.example
│   ├── go.mod
│   └── go.sum
│
└── client/                   # Next.js frontend
    ├── src/
    │   ├── app/                      # App Router (routing = folders)
    │   │   ├── layout.tsx
    │   │   ├── page.tsx              # dashboard
    │   │   ├── submissions/
    │   │   │   ├── new/page.tsx
    │   │   │   └── [id]/
    │   │   │       ├── page.tsx          # result view
    │   │   │       └── revise/page.tsx   # revision view
    │   │   └── history/page.tsx
    │   ├── components/
    │   │   ├── chart/                # renders chart_data
    │   │   ├── feedback/             # inline corrections + repetition badges
    │   │   └── ui/
    │   ├── lib/
    │   │   ├── api/
    │   │   │   ├── client.ts         # typed fetch wrapper
    │   │   │   └── types.gen.ts      # GENERATED from OpenAPI — do not edit by hand
    │   │   └── utils/
    │   └── styles/
    ├── public/
    ├── .env.local.example
    ├── next.config.js
    ├── package.json
    └── tsconfig.json
```

Two top-level app folders: **`server/`** and **`client/`**. Keep them independent (separate deps, separate build). The only contract between them is the HTTP API, expressed as an OpenAPI spec (see §9).

## 3. 🔒 Golden rules (hard constraints — never violate)

1. **Vendor abstraction.** All LLM calls go through the Go `Evaluator` interface (defined in the `evaluation` package, where it is consumed). No Anthropic SDK types leak outside the `adapter/anthropic` package. Business logic depends on the interface, never on the vendor.
2. **No AI-generated chart images.** For generated Task 1 charts, the AI returns **structured data** (see data design §5.1); the client renders it. Never ask the model for an image.
3. **Repetition detection is deterministic.** Detect repeated lemmas in server code (not via the LLM). The LLM is only asked for synonym *suggestions*.
4. **Never hand over a finished rewrite in MVP.** Output is suggestions + corrections only. The user rewrites in a revision field; the AI then scores the revision. Do not generate a full "model answer" essay for the user to copy.
5. **Design the evaluator for scoring consistency.** Use temperature ≈ 0, include the official IELTS band descriptors in the prompt, and use few-shot calibration anchors. Scoring must be reproducible for the same input. This is the product's #1 risk.
6. **Single-user, future-proofed.** No auth in MVP, but `user_id` exists (nullable) per the data design. Do not hard-delete the multi-user path.
7. **Validate all AI output.** The model returns JSON; parse it into typed structs and validate (bands in range & 0.5 steps, required fields present) before persisting or returning. Never trust raw model output.
8. **API contract is OpenAPI, code-first.** Every endpoint MUST carry swaggo annotations (summary, params, request DTO, every response + status). After adding or changing any endpoint, regenerate the spec (`make docs` / `go generate`). The client's request/response types are **generated** from that spec — never hand-write or hand-edit `client/src/lib/api/types.gen.ts`. A stale spec fails CI.
9. **One-way feature dependencies.** A feature may import another only in one direction (e.g. `submission` → `evaluation`, never the reverse). No import cycles between features. Interfaces are defined in the consuming package; adapters import the feature to implement them.

## 4. Server conventions (`server/`)

- **Language/framework:** Go + Fiber. REST, JSON.
- **Architecture — package-by-feature.** Each module under `internal/feature/<name>/` owns its full stack:
  - `handler.go` — Fiber handlers: parse request, call service, map result/error to HTTP. No business logic. Swagger annotations live here, above each handler.
  - `service.go` — business logic & orchestration.
  - `repository.go` — DB access (Postgres). SQL lives here only.
  - `dto.go` — API-facing request/response structs (json tags + used in swagger annotations).
  - `model.go` — the clean domain entity.
  - Shared infra (config, db pool, Fiber setup, health) lives in `internal/platform/`; the Claude client in `internal/adapter/anthropic/`.
- **DTO / DAO separation.**
  - **DTO (always):** never expose a domain entity or DB row directly to the client. API structs live in `dto.go` and are mapped to/from the domain model explicitly.
  - **DAO / persistence struct (only when needed):** when the DB representation differs from the domain model (e.g. `jsonb` fields like `feedback`/`chart_data`, or nullable columns), use a dedicated persistence struct in `repository.go`. For simple entities, scan straight into the domain model — do not add a DAO just for symmetry.
- **Constants.** Shared / cross-cutting constants live in `internal/constant/`, one file per topic (`health.go`, `apierror.go`, `enum.go`). The enum values from data-design (`task_type`, `chart_mode`, `prompt_source`) and API error codes belong here — single source of truth, no magic strings. Constants used in exactly one place may stay local to avoid a god-package and import cycles.
- **Comments.** Exported identifiers get a concise one-line doc comment (Go/godoc convention). No narration of obvious code. Reserve longer comments for non-obvious logic or complex business rules.
- **API docs (Swagger / OpenAPI):** `swaggo/swag` generates the spec from handler annotations into `server/docs/`; `gofiber/swagger` serves the UI at `/swagger`. Regeneration via `//go:generate swag init` + `make docs`. Treat `server/docs/` as generated output — never edit by hand. See `docs/api-conventions.md`.
- **Config:** environment variables only (12-factor), via `internal/platform/config`. No secrets in code. Provide `.env.example`.
- **DB:** PostgreSQL via `jackc/pgx` v5 + `pgxpool`. Migrations in `server/migrations/` via `golang-migrate`. Never edit an applied migration.
- **Errors:** return typed errors from the service layer; handlers translate to HTTP status + JSON error body `{ "error": { "code": "...", "message": "..." } }` (codes in `constant/apierror.go`; see api-conventions).
- **Context:** pass `context.Context` through service → repository → adapter (timeouts, cancellation on LLM calls).
- **The `Evaluator` interface (shape to aim for, lives in `feature/evaluation`):**
  ```go
  type Evaluator interface {
      Evaluate(ctx context.Context, in EvaluationInput) (EvaluationResult, error)
      SuggestSynonyms(ctx context.Context, in SynonymInput) ([]SynonymSuggestion, error)
      GenerateTask1Chart(ctx context.Context, in ChartGenInput) (ChartData, error)
  }
  ```
  (Exact signatures finalized in the Technical Design doc.)

## 5. Client conventions (`client/`)

- **Framework:** Next.js (App Router) + TypeScript (strict).
- **API types are generated.** `client/src/lib/api/types.gen.ts` is generated from the server's OpenAPI spec via `openapi-typescript`. Never hand-edit it. If a type is wrong, fix the server annotation and regenerate.
- **API access:** a single typed client module (`lib/api/client.ts`) wraps `fetch` using the generated types. No scattered `fetch` calls across components.
- **Charts:** one chart library (Recharts or Chart.js — pick in TDD) for rendering `chart_data`.
- **Key screens (MVP):**
  1. New submission (choose task type; provide/generate prompt; upload or generate chart for Task 1; write essay).
  2. Result view (band vs target, inline corrections, repetition badges).
  3. Revision view (revision field → re-score → show delta).
  4. History / progress (list + band trend).
- **State:** keep it simple; server is source of truth. No heavy global state unless justified.
- **Components:** small, typed, reusable. Badges (clickable synonym suggestions) and inline-correction highlighting are shared components.

## 6. Testing expectations

- **Unit tests are colocated** — `foo_test.go` sits next to `foo.go` in the same package (Go convention; enables white-box testing and `go test ./...`). Do not move unit tests into a separate folder.
- **Integration / e2e tests** (need a real Postgres or a running server) live in `server/test/`.
- **Server:** unit-test the service layer; the repetition detector must have tests; AI calls are mocked via the `Evaluator` interface. Repository integration tests run against a test Postgres.
- **Client:** component tests for the critical interactive pieces (badges, correction rendering, revision flow).
- Do not mock away the thing under test. The `Evaluator` interface exists partly to make the service testable without hitting the real API.

## 7. Definition of Done (per task)

A task is done when:
1. It implements exactly what the referenced doc/ticket specifies — no more, no less.
2. Code follows the architecture (package-by-feature), golden rules, and conventions above.
3. AI output (if any) is parsed into typed structs and validated.
4. Endpoints added/changed: swagger annotations present and the spec + client types regenerated.
5. Tests for the new logic pass (unit colocated); existing tests still pass.
6. No secrets committed; new env vars added to `.env.example`.
7. A short note in the PR describes *what* and *why* (the reviewer reads this first).

## 8. When unsure

Stop and ask the reviewer. Specifically ask — don't guess — about: product behavior, scoring rubric details, API shapes not yet in the contract, new dependencies, and anything touching the scoring-consistency design.

## 9. API workflow (code-first OpenAPI)

The HTTP contract is not hand-maintained as prose — it is generated:

```
Write handler + swaggo annotations
        │  make docs  (swag init → go generate)
        ▼
server/docs/swagger.json         ──►  served at /swagger (Swagger UI, reviewer surface)
        │  openapi-typescript
        ▼
client/src/lib/api/types.gen.ts  ──►  consumed by lib/api/client.ts
```

Conventions (base path, status codes, error codes, pagination, annotation rules, planned MVP endpoints) live in `docs/api-conventions.md`.