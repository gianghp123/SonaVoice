package dtos

type GlobalConfig struct {
	Config ConfigPayload `json:"config"`
}

type ConfigPayload struct {
	Enabled       bool         `json:"enabled"`
	ResetTimezone string       `json:"reset_timezone"`
	Limits        LimitsConfig `json:"limits"`
}

type LimitsConfig struct {
	User    UserLimitConfig    `json:"user"`
	Session SessionLimitConfig `json:"session"`
}

type UserLimitConfig struct {
	DailyVoiceSeconds int `json:"daily_voice_seconds"`
	DailyRequestCount int `json:"daily_request_count"`
}

type SessionLimitConfig struct {
	MaxSessionLockTTL int `json:"max_session_lockTTL"`
}
