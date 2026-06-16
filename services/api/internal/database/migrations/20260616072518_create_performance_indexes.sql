-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_sessions_speech_session_id ON sessions (speech_session_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions (status);
CREATE INDEX IF NOT EXISTS idx_grammar_analyses_session_created ON grammar_analyses (session_id, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_grammar_analyses_session_created;
DROP INDEX IF EXISTS idx_sessions_status;
DROP INDEX IF EXISTS idx_sessions_speech_session_id;
-- +goose StatementEnd
