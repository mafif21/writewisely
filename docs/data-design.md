# WriteWisely — Data Design

> Status: Draft v1 · Last updated: 2026-11-14 · DB: PostgreSQL
> Depends on: project-brief.md · Feeds: api-conventions.md, Technical Design.

## 1. Design principles

- **Single-user now, multi-user later.** Every user-owned row carries a nullable `user_id`. Adding auth later = backfill + make it `NOT NULL`, no restructuring.
- **Separate the task from the writing.** One IELTS task (prompt + chart + target) can have many written versions (original + revisions). This is what powers the revision loop and progress tracking.
- **One evaluation per written version.** Each attempt is scored exactly once; comparing evaluations across attempts shows improvement.
- **Feedback is semi-structured.** Band numbers are columns (queryable for progress charts); rich feedback (corrections, badges, comments) is `jsonb`.

## 2. Entity-Relationship (overview)

```
writing_submission (the task context)
        │ 1
        │
        │ *
     attempt  (original essay = #1, each revision = #2, #3, …)
        │ 1
        │
        │ 1
   evaluation (AI scoring + feedback for that attempt)
```

- `writing_submission` **1 — \*** `attempt`
- `attempt` **1 — 1** `evaluation`

## 3. Enumerated values

Implemented as `TEXT` + `CHECK` constraints (simpler migrations than native enums; easy to extend).

| Field | Allowed values | Meaning |
|-------|----------------|---------|
| `task_type` | `task1_academic`, `task2` | Which IELTS writing task |
| `prompt_source` | `user_provided`, `ai_generated` | Did the user bring the prompt, or did AI generate it (Mode B) |
| `chart_mode` | `none`, `uploaded_image`, `generated_data` | How the Task 1 chart is supplied (`none` for Task 2) |

## 4. Tables

### 4.1 `writing_submission`
The task the user is working on — the prompt, optional chart, and target band. Does **not** hold the essay text.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `user_id` | `uuid` | NULL | Future multi-user; always NULL in MVP |
| `task_type` | `text` | NOT NULL, CHECK in enum | |
| `prompt_text` | `text` | NOT NULL | The question/prompt |
| `prompt_source` | `text` | NOT NULL, CHECK in enum | |
| `chart_mode` | `text` | NOT NULL, CHECK in enum | `none` for Task 2 |
| `chart_data` | `jsonb` | NULL | Structured chart (Mode B generated). Shape in §5.1 |
| `chart_image_path` | `text` | NULL | Stored reference to uploaded chart (Mode A) |
| `target_band` | `numeric(2,1)` | NOT NULL, default `7.0`, CHECK 0–9 | |
| `created_at` | `timestamptz` | NOT NULL, default `now()` | |

**Integrity rule (enforce in service layer, optionally a CHECK):**
- `task_type = task2` ⇒ `chart_mode = none`, `chart_data` and `chart_image_path` NULL.
- `chart_mode = uploaded_image` ⇒ `chart_image_path` NOT NULL.
- `chart_mode = generated_data` ⇒ `chart_data` NOT NULL.

### 4.2 `attempt`
One written version of an essay for a submission. `#1` is the original; revisions are `#2`, `#3`, …

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `submission_id` | `uuid` | NOT NULL, FK → `writing_submission(id)` ON DELETE CASCADE | |
| `attempt_no` | `int` | NOT NULL, CHECK ≥ 1 | 1 = original, ≥2 = revision |
| `essay_text` | `text` | NOT NULL | |
| `word_count` | `int` | NOT NULL | Computed server-side on insert |
| `created_at` | `timestamptz` | NOT NULL, default `now()` | |

**Constraint:** `UNIQUE (submission_id, attempt_no)`.

### 4.3 `evaluation`
AI scoring + feedback for one attempt.

| Column | Type | Constraints | Notes |
|--------|------|-------------|-------|
| `id` | `uuid` | PK, default `gen_random_uuid()` | |
| `attempt_id` | `uuid` | NOT NULL, UNIQUE, FK → `attempt(id)` ON DELETE CASCADE | One eval per attempt |
| `band_overall` | `numeric(2,1)` | NOT NULL, CHECK 0–9 | |
| `band_task` | `numeric(2,1)` | NOT NULL, CHECK 0–9 | Task Achievement / Response |
| `band_coherence` | `numeric(2,1)` | NOT NULL, CHECK 0–9 | Coherence & Cohesion |
| `band_lexical` | `numeric(2,1)` | NOT NULL, CHECK 0–9 | Lexical Resource |
| `band_grammar` | `numeric(2,1)` | NOT NULL, CHECK 0–9 | Grammatical Range & Accuracy |
| `feedback` | `jsonb` | NOT NULL | Corrections, badges, comments. Shape in §5.2 |
| `model` | `text` | NOT NULL | e.g. the Claude model id used, for traceability |
| `created_at` | `timestamptz` | NOT NULL, default `now()` | |

> Band values are multiples of 0.5 (IELTS scale). Enforce the 0.5 step in the service layer.

## 5. JSONB shapes

### 5.1 `writing_submission.chart_data` (Mode B — generated)
```json
{
  "chart_type": "line | bar | pie | table | process | map",
  "title": "Internet users by region, 2000–2020",
  "x_label": "Year",
  "y_label": "Users (millions)",
  "series": [
    { "name": "Asia", "points": [{ "x": "2000", "y": 110 }, { "x": "2010", "y": 820 }] }
  ],
  "notes": "Optional extra context for rendering"
}
```
The client renders this with a chart library. The same object is passed to the AI during evaluation so it checks the essay against the real data.

### 5.2 `evaluation.feedback`
```json
{
  "summary": "Short overall comment.",
  "criteria_comments": {
    "task": "…",
    "coherence": "…",
    "lexical": "…",
    "grammar": "…"
  },
  "inline_corrections": [
    {
      "criterion": "grammar",
      "original": "The number of users were increasing",
      "suggestion": "The number of users was increasing",
      "reason": "\"The number of\" takes a singular verb.",
      "examples": [
        "The number of students was high.",
        "The number of cars was rising.",
        "The number of errors was reduced."
      ]
    }
  ],
  "repetition_badges": [
    {
      "lemma": "increase",
      "paragraph_index": 1,
      "occurrences": 3,
      "suggestions": ["rise", "grow", "climb"]
    }
  ]
}
```

**Division of labor for `repetition_badges`:**
- Detection (`lemma`, `paragraph_index`, `occurrences`) = **deterministic, server-side** (lemmatization).
- `suggestions` = **AI-generated**, context-aware synonyms.
The service merges both before persisting.

## 6. Indexes

| Index | Purpose |
|-------|---------|
| `writing_submission (user_id, created_at DESC)` | List a user's submissions / progress timeline |
| `attempt (submission_id, attempt_no)` | Fetch attempts of a submission in order |
| `evaluation (attempt_id)` UNIQUE | One eval per attempt + fast lookup |

## 7. Progress tracking (how the data answers it)

- **Within one task:** compare `evaluation.band_overall` across a submission's attempts (`attempt_no` order) → did the revision improve?
- **Over time:** order a user's `evaluation` rows by `created_at` → band trend line toward target 7.

## 8. Migration note

- Enable `pgcrypto` (or `gen_random_uuid()` from Postgres 13+ core) for UUID defaults.
- Use a migration tool (`golang-migrate`) — migrations live in `server/migrations/`. One migration per table/change; never edit an applied migration.
