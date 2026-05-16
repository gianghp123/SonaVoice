package configs

import "github.com/gianghp123/SonaVoice/api/internal/utils"

type RedisConfig struct {
	RedisUrl string
}

func loadRedisConfig() RedisConfig {
	return RedisConfig{
		RedisUrl: utils.GetEnv("REDIS_URL", "redis://localhost:6379"),
	}
}
