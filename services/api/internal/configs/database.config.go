package configs

import "github.com/gianghp123/SonaVoice/api/internal/utils"

type DatabaseConfig struct {
	DatabaseUrl string
}

func loadDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		DatabaseUrl: utils.GetEnv("DATABASE_URL", ""),
	}
}
