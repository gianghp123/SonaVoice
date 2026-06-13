-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS grammar_analyses (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id          uuid NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    message_id          uuid NOT NULL UNIQUE REFERENCES messages (id) ON DELETE CASCADE,

    original_text       text NOT NULL,
    result              jsonb NOT NULL DEFAULT '{}',

    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now(),
    deleted_at          timestamptz
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS grammar_analyses;
-- +goose StatementEnd
