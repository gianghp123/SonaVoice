-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS voice_ai_global_config (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    config      jsonb NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    deleted_at  timestamptz
);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS voice_ai_global_config;

-- +goose StatementEnd