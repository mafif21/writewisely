# WriteWisely — Project Brief

> Status: Draft v1 · Owner: Apip · Last updated: 2026-11-14
> First foundational document. Supersedes nothing. Next docs: Data Design, Technical Design (TDD).

## 1. Overview

**WriteWisely** is an AI-powered IELTS Writing checker for **Task 1 (Academic)** and **Task 2**. A user submits an essay (optionally with a chart for Task 1), and the app returns an estimated band score, targeted grammar corrections, and lexical-variety suggestions — then invites the user to **revise their own answer** and re-scores the revision.

The product is built around a **learning loop**, not one-shot scoring: suggestions first, user revises, AI re-evaluates. This loop is the core differentiator.

**One-liner:** *Practice IELTS Writing, get band feedback, and learn by revising — not by being handed the answer.*

## 2. Problem

IELTS Writing is hard to self-study because high-quality, criterion-based feedback normally requires a human examiner or tutor. Learners get either (a) generic grammar checkers that don't understand IELTS band criteria, or (b) tools that hand over a "perfect" rewrite, which gives a score but teaches little. There is no cheap, fast feedback loop that both scores against the real IELTS rubric **and** forces active revision.

## 3. Goals & Objectives

**Primary objective (now):** a personal learning tool to support the owner's own IELTS prep (target band 7, test in 2027).

**Secondary objectives (future):** portfolio-grade product; potential paid public SaaS.

Because the primary objective is learning, the MVP prioritizes **useful, trustworthy feedback for one user** over scale, billing, or growth features. Business/monetization work is explicitly deferred until after the tool proves valuable to the owner.

## 4. Target Users

| Phase | User | Implication |
|-------|------|-------------|
| MVP (now) | The owner (single user) | No auth, no billing. Optimize for feedback quality. |
| Future (public) | IELTS test-takers self-studying Writing | Auth, multi-user, billing, scoring-reliability hardening, cost optimization. |

Data model is designed single-user **but future-proofed** (nullable `user_id` from day one) so multi-user is an additive change, not a rewrite.

## 5. MVP Scope

### Core features

1. **Band scoring — achieved vs target.** AI returns a band per IELTS criterion (Task Achievement/Response, Coherence & Cohesion, Lexical Resource, Grammatical Range & Accuracy) plus overall. Displayed against a target band (default 7; configurable later).
2. **Inline grammar correction.** Each detected error gets: the correction, a short reason *why* that grammar is used, and up to **3 examples**.
3. **Lexical repetition assistance.** Detect repeated verbs/words within a paragraph (e.g. reusing "increase"). Surface synonym suggestions as **clickable badges**. Detection is deterministic on the server (lemmatization); AI is used only to generate context-appropriate synonyms.
4. **Revision loop (core mechanic).** The app gives suggestions **but not a finished rewrite**. The user rewrites in a **revision field**, and the AI evaluates the revision so the user can see improvement.

### Task 1 (Academic) input modes

- **Mode A — Upload:** user uploads a chart/graph image; AI reads it (vision) and evaluates the essay against it.
- **Mode B — Generate:** AI generates a practice prompt with chart **data** (structured JSON), rendered client-side with a chart library. *AI does not generate chart images.*

### Persistence

- Store every submission + its evaluation + revisions, enabling a **progress-over-time view** (band trend). In MVP scope.

## 6. Non-Goals (explicitly out of MVP)

- Speaking module (and any non-Writing skill)
- Authentication, user accounts, multi-user
- Payment / billing / subscriptions
- Mobile app
- AI-generated chart **images**
- Advanced scoring-reliability techniques (self-consistency / ensembling)
- Public-facing analytics dashboards

## 7. High-Level Architecture

```
┌────────────────────┐        REST/JSON        ┌───────────────────────┐
│  Client (Next.js)  │  ───────────────────▶   │  Server (Go + Fiber)  │
│  - submission UI   │                         │  - REST API           │
│  - chart rendering │  ◀───────────────────   │  - repetition detect  │
│  - revision field  │                         │  - Evaluator interface│
│  - score / badges  │                         └──────────┬────────────┘
└────────────────────┘                                    │
                                                          ▼
                                            ┌──────────────────────────┐
                                            │  Anthropic Claude (API)   │
                                            │  vision + structured JSON │
                                            └──────────────────────────┘
                                                          │
                                                          ▼
                                            ┌──────────────────────────┐
                                            │  Datastore: PostgreSQL    │
                                            │  submissions / evals      │
                                            └──────────────────────────┘
```

**Stack**
- **Client:** Next.js (App Router), TypeScript, a chart library (Recharts or Chart.js) for Mode B.
- **Server:** Go + Fiber, REST API.
- **AI:** Anthropic Claude, behind a Go `Evaluator` interface (vendor-abstracted). Used for scoring, grammar feedback, and synonym suggestions. Vision used for Mode A chart reading.
- **Storage:** PostgreSQL (see data-design.md).

## 8. Key Technical Decisions & Risks

| Topic | Decision / Note |
|-------|-----------------|
| **LLM vendor** | Anthropic Claude for MVP (validated by owner's own usage quality). Revisit providers before going public. |
| **Vendor coupling** | Abstract behind an `Evaluator` interface so models can be swapped/benchmarked without touching business logic. |
| **Chart generation** | Never generate chart images via AI. AI outputs structured chart data; client renders it. |
| **Repetition detection** | Deterministic (lemmatization) on the server; AI only for synonym suggestions. Saves cost. |
| **⚠️ Scoring consistency (top risk)** | LLMs are probabilistic — the same essay can score differently across runs, and standards can drift between essays. For a credible (and future paid) IELTS tool this is risk #1. The evaluation module must be **designed for consistency from the start**, not patched later. Mitigations (for implementation phase): low temperature (~0), official band descriptors in the prompt, few-shot calibration with anchored example essays, and structured per-criterion output. |

## 9. Decisions Made / Next Steps

1. **Datastore:** PostgreSQL (decided). Chosen for owner proficiency, Docker-friendliness, `jsonb` support, and clean path to multi-user.
2. **Data Design** — done (`data-design.md`): `writing_submission`, `attempt`, `evaluation` with nullable `user_id`.
3. **API contract** — code-first OpenAPI via swaggo (`api-conventions.md`).
4. **Engineering guide** — `AGENTS.md` (golden rules, structure, conventions).
5. **Next:** scaffold (`tasks/01-scaffold.md`), then the Technical Design (Evaluator interface shape, Claude prompt/structured-output + consistency approach) and development roadmap.
