-- 0001_init: initial schema from docs/data-design.md (v1).
--
-- UUID defaults use gen_random_uuid(), which is built into PostgreSQL 13+,
-- so the pgcrypto extension is not required (target: postgres:16).

BEGIN;

-- §4.1 writing_submission: the task context (prompt, optional chart, target band).
-- Cross-column chart integrity rules (§4.1) are enforced in the service layer.
CREATE TABLE writing_submission (
    id               uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          uuid         NULL,
    task_type        text         NOT NULL
        CONSTRAINT writing_submission_task_type_check
        CHECK (task_type IN ('task1_academic', 'task2')),
    prompt_text      text         NOT NULL,
    prompt_source    text         NOT NULL
        CONSTRAINT writing_submission_prompt_source_check
        CHECK (prompt_source IN ('user_provided', 'ai_generated')),
    chart_mode       text         NOT NULL
        CONSTRAINT writing_submission_chart_mode_check
        CHECK (chart_mode IN ('none', 'uploaded_image', 'generated_data')),
    chart_data       jsonb        NULL,
    chart_image_path text         NULL,
    target_band      numeric(2,1) NOT NULL DEFAULT 7.0
        CONSTRAINT writing_submission_target_band_check
        CHECK (target_band BETWEEN 0 AND 9),
    created_at       timestamptz  NOT NULL DEFAULT now()
);

-- §6: list a user's submissions / progress timeline.
CREATE INDEX writing_submission_user_id_created_at_idx
    ON writing_submission (user_id, created_at DESC);

-- §4.2 attempt: one written version of an essay (1 = original, >= 2 = revision).
CREATE TABLE attempt (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id uuid        NOT NULL
        CONSTRAINT attempt_submission_id_fkey
        REFERENCES writing_submission (id) ON DELETE CASCADE,
    attempt_no    int         NOT NULL
        CONSTRAINT attempt_attempt_no_check
        CHECK (attempt_no >= 1),
    essay_text    text        NOT NULL,
    word_count    int         NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    -- §4.2 constraint; its unique index also serves the §6 index
    -- attempt (submission_id, attempt_no).
    CONSTRAINT attempt_submission_id_attempt_no_key
        UNIQUE (submission_id, attempt_no)
);

-- §4.3 evaluation: AI scoring + feedback for exactly one attempt.
CREATE TABLE evaluation (
    id             uuid         PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id     uuid         NOT NULL
        CONSTRAINT evaluation_attempt_id_fkey
        REFERENCES attempt (id) ON DELETE CASCADE,
    band_overall   numeric(2,1) NOT NULL
        CONSTRAINT evaluation_band_overall_check
        CHECK (band_overall BETWEEN 0 AND 9),
    band_task      numeric(2,1) NOT NULL
        CONSTRAINT evaluation_band_task_check
        CHECK (band_task BETWEEN 0 AND 9),
    band_coherence numeric(2,1) NOT NULL
        CONSTRAINT evaluation_band_coherence_check
        CHECK (band_coherence BETWEEN 0 AND 9),
    band_lexical   numeric(2,1) NOT NULL
        CONSTRAINT evaluation_band_lexical_check
        CHECK (band_lexical BETWEEN 0 AND 9),
    band_grammar   numeric(2,1) NOT NULL
        CONSTRAINT evaluation_band_grammar_check
        CHECK (band_grammar BETWEEN 0 AND 9),
    feedback       jsonb        NOT NULL,
    model          text         NOT NULL,
    created_at     timestamptz  NOT NULL DEFAULT now(),
    -- One evaluation per attempt; this unique index is the §6 index
    -- evaluation (attempt_id) UNIQUE.
    CONSTRAINT evaluation_attempt_id_key
        UNIQUE (attempt_id)
);

COMMIT;
