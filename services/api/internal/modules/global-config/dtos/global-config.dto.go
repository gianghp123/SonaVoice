package dtos

type GlobalConfig struct {
	Config ConfigPayload `json:"config"`
}

type ConfigPayload struct {
	Enabled       bool         `json:"enabled"`
	ResetTimezone string       `json:"resetTimezone"`
	Limits        LimitsConfig `json:"limits"`
}

type LimitsConfig struct {
	Guest UserLimitConfig `json:"guest"`
	User  UserLimitConfig `json:"user"`
}

type UserLimitConfig struct {
	DailyVoiceSeconds int `json:"dailyVoiceSeconds"`
	DailyRequestCount int `json:"dailyRequestCount"`
}
