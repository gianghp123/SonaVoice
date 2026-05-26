package configs

import (
	"log"

	dotenv "github.com/joho/godotenv"
)

type Config struct {
	Database  DatabaseConfig
	ClerkAuth ClerkConfig
	Logger    LoggerConfig
	Redis     RedisConfig
	Sentry    SentryConfig
}

func Load() *Config {

	err := dotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return &Config{
		Database:  loadDatabaseConfig(),
		ClerkAuth: loadClerkConfig(),
		Logger:    loadLoggerConfig(),
		Redis:     loadRedisConfig(),
		Sentry:    loadSentryConfig(),
	}
}
