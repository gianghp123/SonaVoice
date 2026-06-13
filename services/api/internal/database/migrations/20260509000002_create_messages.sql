-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS messages (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      uuid NOT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    role            varchar(255) NOT NULL,
    transcript      text,
    was_interrupted boolean NOT NULL DEFAULT false,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);

CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages (session_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS messages;
-- +goose StatementEnd
