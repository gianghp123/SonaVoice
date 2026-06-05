-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS user_profiles (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      varchar(255) NOT NULL,
    display_name varchar(255) NOT NULL,
    english_level varchar(50) NOT NULL,
    preferences  jsonb DEFAULT '{}',
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    deleted_at   timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_user_profiles_user_id
ON user_profiles (user_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS uq_user_profiles_user_id;
DROP TABLE IF EXISTS user_profiles;

-- +goose StatementEnd

