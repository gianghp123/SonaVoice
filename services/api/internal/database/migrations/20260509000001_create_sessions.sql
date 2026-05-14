-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS sessions (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    varchar(255) NOT NULL,
    speech_session_id varchar(255) NOT NULL,
    status     varchar(255) NOT NULL,
    started_at timestamptz,
    ended_at   timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd

