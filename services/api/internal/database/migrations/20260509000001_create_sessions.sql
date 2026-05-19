-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS sessions (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         varchar(255) NOT NULL,
    speech_session_id varchar(255),
    status          varchar(255) NOT NULL,
    max_duration    bigint NOT NULL DEFAULT 0,
    actual_usage    bigint NOT NULL DEFAULT 0,
    quota_date      date,
    started_at      timestamptz,
    ended_at        timestamptz,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_one_active_session_per_user
ON sessions (user_id)
WHERE status IN ('active', 'pending');
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS sessions;
DROP INDEX IF EXISTS uq_one_active_session_per_user;
-- +goose StatementEnd

