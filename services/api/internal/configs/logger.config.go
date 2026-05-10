package configs

import "github.com/gianghp123/SonaVoice/api/internal/utils"

type LoggerConfig struct {
	Mode string
}

func loadLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Mode: utils.GetEnv("ENVIRONMENT", "development"),
	}
}
