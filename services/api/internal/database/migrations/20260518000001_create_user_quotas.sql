-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS user_quotas (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     varchar(255) NOT NULL,
    quota_key   varchar(255) NOT NULL,
    quota_date  date NOT NULL,
    remaining   bigint NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_quotas_user_key_date
ON user_quotas (user_id, quota_key, quota_date);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS user_quotas;
DROP INDEX IF EXISTS uq_user_quotas_user_key_date;
-- +goose StatementEnd