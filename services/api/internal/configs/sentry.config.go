package configs

import "github.com/gianghp123/SonaVoice/api/internal/utils"

type SentryConfig struct {
	Dsn         string
	Environment string
}

func loadSentryConfig() SentryConfig {
	return SentryConfig{
		Dsn:         utils.GetEnv("SENTRY_DSN", ""),
		Environment: utils.GetEnv("ENVIRONMENT", "development"),
	}
}
