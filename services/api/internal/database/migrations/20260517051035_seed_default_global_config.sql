-- +goose Up
-- +goose StatementBegin

INSERT INTO global_config (config)
SELECT
    '{
      "enabled": true,
      "resetTimezone": "Asia/Ho_Chi_Minh",
      "limits": {
        "user": {
          "dailyVoiceSeconds": 300,
          "dailyRequestCount": 50
        },
        "session": {
          "maxSessionLockTTL": 5
        }
      }
    }'::jsonb
WHERE NOT EXISTS (
    SELECT 1
    FROM global_config
    WHERE deleted_at IS NULL
);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin

DELETE FROM global_config
WHERE config = '{
  "enabled": true,
  "resetTimezone": "Asia/Ho_Chi_Minh",
  "limits": {
    "user": {
      "dailyVoiceSeconds": 300,
      "dailyRequestCount": 50
    },
    "session": {
      "maxSessionLockTTL": 5
    }
  }
}'::jsonb;

-- +goose StatementEnd