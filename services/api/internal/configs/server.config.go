package configs

import "github.com/gianghp123/SonaVoice/api/internal/utils"

type ServerConfig struct {
	Port string
	Mode string
}

func loadServerConfig() ServerConfig {
	return ServerConfig{
		Port: utils.GetEnv("PORT", "3001"),
		Mode: utils.GetEnv("GIN_MODE", "debug"),
	}
}
