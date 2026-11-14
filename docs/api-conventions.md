# WriteWisely — API Conventions

> Status: Draft v1 · Last updated: 2026-11-14
> The actual endpoint spec is **generated** (OpenAPI via swaggo, served at `/swagger`). This doc defines the rules every endpoint follows + the MVP endpoint plan. Depends on: data-design.md, AGENTS.md.

## 1. Basics

- **Base path:** `/api/v1`
- **Format:** JSON only. `Content-Type: application/json` (except chart image upload, which is `multipart/form-data`).
- **Field naming:** `snake_case` in all JSON bodies (matches the DB and `chart_data` / `feedback` shapes in data-design). Go struct tags and generated TS types follow this.
- **Auth:** none in MVP. (Future: `Authorization: Bearer <token>`.) Do not add auth middleware yet, but do not block it architecturally.
- **IDs:** UUID strings.
- **Timestamps:** RFC 3339 / ISO 8601 UTC.

## 2. Responses

- **Success:** return the resource directly with the right status. No envelope.
- **Error:** always this shape —
  ```json
  { "error": { "code": "validation_error", "message": "target_band must be between 0 and 9" } }
  ```

### Status codes

| Code | When |
|------|------|
| `200 OK` | Successful GET or a completed action returning a result |
| `201 Created` | Resource created (submission, attempt) |
| `400 Bad Request` | Malformed body / failed validation (`validation_error`) |
| `404 Not Found` | Unknown id (`not_found`) |
| `409 Conflict` | Invalid state transition (`conflict`) |
| `422 Unprocessable Entity` | Well-formed but semantically invalid (e.g. Task 2 with chart data) |
| `502 Bad Gateway` | LLM provider call failed (`evaluation_failed`) |
| `500 Internal Server Error` | Unexpected (`internal_error`) |

### Error codes (stable strings)
`validation_error`, `not_found`, `conflict`, `unprocessable`, `evaluation_failed`, `internal_error`.

## 3. Pagination (list endpoints)

Query params: `?limit=20&offset=0` (defaults: limit 20, max 100; offset 0). List response shape:
```json
{ "items": [ /* … */ ], "total": 42, "limit": 20, "offset": 0 }
```

## 4. Swagger annotation rules (code-first)

Every handler MUST be annotated so the generated spec is complete. Minimum per endpoint:
- `@Summary` + `@Tags`
- `@Accept` / `@Produce json`
- `@Param` for path/query/body (body references a named DTO struct)
- `@Success` **and every** `@Failure` with status + the response DTO (success) or the error object (failure)
- `@Router /path [method]`

Example:
```go
// CreateAttempt submits an essay (original or revision) and returns its evaluation.
//
// @Summary  Submit an essay attempt
// @Tags     attempts
// @Accept   json
// @Produce  json
// @Param    id   path  string              true  "Submission ID"
// @Param    body body  CreateAttemptRequest true "Essay text"
// @Success  201  {object}  AttemptWithEvaluationResponse
// @Failure  400  {object}  ErrorResponse
// @Failure  404  {object}  ErrorResponse
// @Failure  502  {object}  ErrorResponse
// @Router   /submissions/{id}/attempts [post]
func (h *SubmissionHandler) CreateAttempt(c *fiber.Ctx) error { /* … */ }
```

Request/response DTOs are Go structs defined near the handler (not the domain entities). This keeps the wire contract decoupled from internal models.

## 5. Regeneration workflow

```
make docs   # runs: swag init  (→ server/docs/)  then  openapi-typescript (→ client/src/lib/api/types.gen.ts)
```
- Wired as `//go:generate swag init` so `go generate ./...` also works.
- **CI check:** regenerate in CI and fail if `git diff` is non-empty → forces a stale spec to be caught. This is how "automatic" is actually enforced.

## 6. Planned MVP endpoints (overview)

The generated spec is authoritative; this is the intended surface so the builder knows scope.

| Method & path | Purpose |
|---------------|---------|
| `POST /task1/practice` | Mode B: generate a Task 1 practice prompt + `chart_data` (stateless; not persisted until a submission is created) |
| `POST /uploads/chart` | Mode A: upload a chart image (`multipart`), returns `chart_image_path` |
| `POST /submissions` | Create the task context (task_type, prompt, chart, target_band). Returns the submission |
| `POST /submissions/{id}/attempts` | Submit an essay (original = #1, or a revision). Server runs repetition detection + evaluation, stores attempt + evaluation, returns both |
| `GET /submissions/{id}` | Full detail: submission + ordered attempts + evaluations |
| `GET /submissions` | History list (paginated) for the progress view |

### Flow
1. **Task 1 Mode B:** `POST /task1/practice` → show chart → user writes → `POST /submissions` → `POST /submissions/{id}/attempts`.
2. **Task 1 Mode A:** `POST /uploads/chart` → `POST /submissions` (with `chart_image_path`) → `POST .../attempts`.
3. **Task 2:** `POST /submissions` (prompt only) → `POST .../attempts`.
4. **Revision:** `POST /submissions/{id}/attempts` again → new `attempt_no`, re-scored. Client shows the band delta.
