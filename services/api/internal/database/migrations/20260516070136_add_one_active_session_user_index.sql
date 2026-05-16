-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS uq_one_active_session_per_user
ON sessions (user_id)
WHERE status IN ('active', 'pending');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_one_active_session_per_user;
-- +goose StatementEnd
