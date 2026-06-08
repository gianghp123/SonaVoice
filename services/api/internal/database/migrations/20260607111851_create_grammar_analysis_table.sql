-- +goose Up
-- +goose StatementBegin


CREATE TABLE IF NOT EXISTS grammar_analyses (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id          uuid NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    message_id          uuid NOT NULL UNIQUE REFERENCES messages (id) ON DELETE CASCADE,

    original_text       text NOT NULL,
    corrected_text      text,
    explanation         text,

    has_correction      boolean NOT NULL DEFAULT true,
    severity            varchar(255) NOT NULL DEFAULT 'low',

    practice_sentence   text,
    practice_focus      text,
    practice_reason     text,

    metadata            jsonb NOT NULL DEFAULT '{}',

    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz,

    CONSTRAINT grammar_analyses_severity_check
        CHECK (severity IN ('low', 'medium', 'high'))
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS grammar_analyses;
-- +goose StatementEnd
